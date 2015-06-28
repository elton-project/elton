package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/hanwen/go-fuse/fuse/nodefs"
	"google.golang.org/grpc"

	pb "../grpc/proto"
)

type FileInfo struct {
	Name    string
	Key     string
	Version uint64
	Size    uint64
	Time    time.Time
}

type EltonFS struct {
	root       *EltonNode
	files      map[string]EltonFile
	lower      string
	upper      string
	eltonURL   string //本当はURLがいい
	connection *grpc.ClientConn
}

var FILEMODE os.FileMode = 0600
var ELTONFS_FILE_NAME string = ".eltonfs"

func NewEltonFS(files map[string]EltonFile, lower, upper, eltonURL string) (*EltonFS, error) {
	conn, err := grpc.Dial(eltonURL)
	if err != nil {
		return nil, err
	}

	fs := &EltonFS{
		root:       &EltonNode{Node: nodefs.NewDefaultNode()},
		files:      files,
		lower:      lower,
		upper:      upper,
		eltonURL:   eltonURL,
		connection: conn,
	}

	fs.root.fs = fs
	return fs, nil
}

func NewEltonTree(host string, opts *Options) (map[string]EltonFile, error) {
	files, err := getFilesInfo(host, opts)
	if err != nil {
		return nil, err
	}

	out := make(map[string]EltonFile)
	for _, f := range files {
		out[f.Name] = &eltonFile{
			Key:     f.Key,
			Version: f.Version,
			Size:    f.Size,
			Time:    f.Time,
		}
	}

	return out, nil
}

func getFilesInfo(host string, opts *Options) (files []FileInfo, err error) {
	conn, err := grpc.Dial(host)
	if err != nil {
		return
	}
	defer conn.Close()

	mi, err := getMountInfo(conn, opts)
	p := path.Join(opts.LowerDir, fmt.Sprintf("%s-%d", mi.ObjectId, mi.Version))
	if _, err = os.Stat(p); os.IsNotExist(err) {
		client := pb.NewEltonServiceClient(conn)
		stream, err := client.GetObject(context.Background(), mi)
		if err != nil {
			return nil, err
		}

		obj, err := stream.Recv()
		if err != nil {
			return nil, err
		}

		data, err := base64.StdEncoding.DecodeString(obj.Body)
		if err != nil {
			return nil, err
		}

		if err = ioutil.WriteFile(p, data, FILEMODE); err != nil {
			return nil, err
		}
	}

	body, err := ioutil.ReadFile(p)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &files)
	return
}

func getMountInfo(conn *grpc.ClientConn, opts *Options) (obj *pb.ObjectInfo, err error) {
	p := filepath.Join(opts.LowerDir, ELTONFS_FILE_NAME)
	buf, err := ioutil.ReadFile(p)
	if err != nil {
		return createMountInfo(conn, p, opts)
	}

	err = json.Unmarshal(buf, obj)
	return
}

func createMountInfo(conn *grpc.ClientConn, p string, opts *Options) (obj *pb.ObjectInfo, err error) {
	obj, err = generateObjectID(conn, p)
	client := pb.NewEltonServiceClient(conn)
	stream, err := client.CreateObjectInfo(context.Background(), obj)
	if err != nil {
		return obj, err
	}

	if obj, err = stream.Recv(); err != nil {
		return obj, err
	}

	if err = ioutil.WriteFile(
		path.Join(
			opts.LowerDir,
			fmt.Sprintf("%s-%d", obj.ObjectId, obj.Version)),
		[]byte("[]"),
		FILEMODE,
	); err != nil {
		return
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return
	}
	if err = ioutil.WriteFile(p, data, FILEMODE); err != nil {
		return
	}

	return
}

func generateObjectID(conn *grpc.ClientConn, p string) (obj *pb.ObjectInfo, err error) {
	client := pb.NewEltonServiceClient(conn)
	stream, err := client.GenerateObjectID(
		context.Background(),
		&pb.ObjectName{
			Names: []string{p},
		},
	)
	if err != nil {
		return
	}

	obj, err = stream.Recv()
	return
}

func NewEltonFileSystem(target string, opts *Options) (root nodefs.Node, err error) {
	files, err := NewEltonTree(target, opts)
	if err != nil {
		return nil, err
	}

	efs, err := NewEltonFS(files, opts.LowerDir, opts.UpperDir, target)
	if err != nil {
		return nil, err
	}

	return efs.Root(), nil
}

func (fs *EltonFS) String() string {
	return fs.upper
}

func (fs *EltonFS) SetDebug(bool) {
}

func (fs *EltonFS) Root() nodefs.Node {
	return fs.root
}

func (fs *EltonFS) onMount() {
	for k, v := range fs.files {
		fs.addFile(k, v)
	}
	fs.files = nil
}

func (fs *EltonFS) onUnmount() {
	fs.connection.Close()
}

func (n *EltonFS) addFile(name string, f EltonFile) {
	comps := strings.Split(name, "/")

	node := n.root.Inode()
	for i, c := range comps {
		child := node.GetChild(c)
		if child == nil {
			fsnode := &EltonNode{
				Node:     nodefs.NewDefaultNode(),
				basePath: n.lower,
				fs:       n,
			}
			if i == len(comps)-1 {
				fsnode.file = f
			}

			child = node.NewChild(c, fsnode.file == nil, fsnode)
		}
		node = child
	}
}
