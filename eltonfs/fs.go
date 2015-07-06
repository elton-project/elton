package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"google.golang.org/grpc"

	pb "../grpc/proto"
)

type FileInfo struct {
	Name     string
	Key      string
	Version  uint64
	Delegate string
	Size     uint64
	Time     time.Time
}

var FILEMODE os.FileMode = 0600
var ELTONFS_CONFIG_DIR string = ".eltonfs"
var ELTONFS_CONFIG_NAME string = "CONFIG"
var ELTONFS_COMMIT_NAME string = "COMMIT"

func NewEltonFSRoot(target string, opts *Options) (root nodefs.Node, err error) {
	dir := filepath.Join(opts.UpperDir, ELTONFS_CONFIG_DIR)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}

	efs, err := NewEltonFS(opts.LowerDir, opts.UpperDir, target)
	if err != nil {
		return nil, err
	}

	if err := fs.newEltonTree(); err != nil {
		return nil, err
	}

	return efs.Root(), nil
}

func NewEltonFS(lower, upper, eltonURL string) (*eltonFS, error) {
	conn, err := grpc.Dial(eltonURL)
	if err != nil {
		return nil, err
	}

	fs := &eltonFS{
		lower:      lower,
		upper:      upper,
		eltonURL:   eltonURL,
		connection: conn,
	}

	fs.root = fs.newNode(upper, time.Now())
	return fs, nil
}

type eltonFS struct {
	root       *EltonNode
	files      map[string]*eltonFile
	lower      string
	upper      string
	eltonURL   string
	connection *grpc.ClientConn
	mux        sync.Mutex
}

func (fs *eltonFS) String() string {
	return fmt.Sprintf("EltonFS(%s)", "elton")
}

func (fs *eltonFS) SetDebug(bool) {
}

func (fs *eltonFS) Root() nodefs.Node {
	return fs.root
}

func (n *eltonFS) OnMount(c *nodefs.FileSystemConnector) {
	for k, v := range fs.files {
		fs.addFile(k, v)
	}
	fs.files = nil
}

func (fs *eltonFS) OnUnmount() {
	fs.connection.Close()
}

func (fs *eltonFS) newNode(base string, t time.Time) *eltonNode {
	fs.mux.Lock()
	n := &eltonNode{
		Node:     nodefs.NewDefaultNode(),
		basePath: base,
		fs:       fs,
	}

	fs.info.SetTimes(&t, &t, &t)
	n.info.Mode = fuse.S_IFDIR | 0700
	fs.mux.Unlock()
	return n
}

func (fs *eltonFS) Filename(n *Inode) string {
	mn := n.Node().(*eltonNode)
	return mn.filename()
}

func (fs *eltonFS) addFile(name string, f *eltonFile) {
	comps := strings.Split(name, "/")

	node := fs.root.Inode()
	for i, c := range comps {
		child := node.GetChild(c)
		if child == nil {
			fsnode := fs.newNode(fs.lower, f.Time)
			if i == len(comps)-1 {
				fsnode.file = f
			}

			child = node.NewChild(c, fsnode.file == new(eltonFile), fsnode)
		}
		node = child
	}
}

func (fs *eltonFS) newEltonTree() error {
	files, err := getFilesInfo(host, opts)
	if err != nil {
		return nil, err
	}

	out := make(map[string]*eltonFile)
	for _, f := range files {
		out[f.Name] = &eltonFile{
			key:      f.Key,
			version:  f.Version,
			delegate: f.Delegate,
		}
	}

	fs.files = out
	return nil
}

func getFilesInfo(host string, opts *Options) (files []FileInfo, err error) {
	conn, err := grpc.Dial(host)
	if err != nil {
		return
	}
	defer conn.Close()

	mi, err := getMountInfo(conn, opts)
	p := filepath.Join(opts.LowerDir, fmt.Sprintf("%s/%s-%d", mi.ObjectId[:2], mi.ObjectId[2:], mi.Version))
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

		if err = CreateFile(p, data); err != nil {
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

func CreateFile(p string, data []byte) error {
	dir := filepath.Dir(p)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(p, data, FILEMODE)
}

func getMountInfo(conn *grpc.ClientConn, opts *Options) (obj *pb.ObjectInfo, err error) {
	p := filepath.Join(opts.UpperDir, ELTONFS_CONFIG_DIR, ELTONFS_CONFIG_NAME)
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

	if err = CreateFile(
		filepath.Join(
			opts.LowerDir,
			fmt.Sprintf(
				"%s/%s-%d",
				obj.ObjectId[:2],
				obj.ObjectId[2:],
				obj.Version,
			),
		),
		[]byte("[]"),
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
