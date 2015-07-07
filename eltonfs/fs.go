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

	if err := efs.newEltonTree(); err != nil {
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

	fs.root = fs.newNode("", time.Now())
	return fs, nil
}

type eltonFS struct {
	root       *eltonNode
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

func (fs *eltonFS) OnMount(c *nodefs.FileSystemConnector) {
}

func (fs *eltonFS) OnUnmount() {
	fs.connection.Close()
}

func (fs *eltonFS) newNode(name string, t time.Time) *eltonNode {
	fs.mux.Lock()
	n := &eltonNode{
		Node:     nodefs.NewDefaultNode(),
		basePath: fs.upper,
		fs:       fs,
		file:     new(eltonFile),
	}

	n.file.key = name

	n.info.SetTimes(&t, &t, &t)
	n.info.Mode = fuse.S_IFDIR | 0700
	fs.mux.Unlock()
	return n
}

func (fs *eltonFS) Filename(n *nodefs.Inode) string {
	mn := n.Node().(*eltonNode)
	return mn.filename()
}

func (fs *eltonFS) newEltonTree() error {
	files, err := fs.getFilesInfo()
	if err != nil {
		return err
	}

	for _, f := range files {
		comps := strings.Split(f.Name, "/")

		node := fs.root.Inode()
		for i, c := range comps {
			child := node.GetChild(c)
			if child == nil {
				fsnode := fs.newNode(c, f.Time)
				if i == len(comps)-1 {
					fsnode.file.key = f.Key
					fsnode.file.version = f.Version
					fsnode.file.delegate = f.Delegate
					fsnode.basePath = fs.lower
				}
			}
			node = child
		}
	}

	return nil
}

func (fs *eltonFS) getFilesInfo() (files []FileInfo, err error) {
	mi, err := fs.getMountInfo()
	p := filepath.Join(fs.lower, fmt.Sprintf("%s/%s-%d", mi.ObjectId[:2], mi.ObjectId[2:], mi.Version))
	if _, err = os.Stat(p); os.IsNotExist(err) {
		client := pb.NewEltonServiceClient(fs.connection)
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

func (fs *eltonFS) getMountInfo() (obj *pb.ObjectInfo, err error) {
	p := filepath.Join(fs.upper, ELTONFS_CONFIG_DIR, ELTONFS_CONFIG_NAME)
	buf, err := ioutil.ReadFile(p)
	if err != nil {
		return fs.createMountInfo(p)
	}

	err = json.Unmarshal(buf, obj)
	return
}

func (fs *eltonFS) createMountInfo(p string) (obj *pb.ObjectInfo, err error) {
	obj, err = fs.generateObjectID(p)
	client := pb.NewEltonServiceClient(fs.connection)
	stream, err := client.CreateObjectInfo(context.Background(), obj)
	if err != nil {
		return obj, err
	}

	if obj, err = stream.Recv(); err != nil {
		return obj, err
	}

	if err = CreateFile(
		filepath.Join(
			fs.lower,
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

func (fs *eltonFS) generateObjectID(p string) (obj *pb.ObjectInfo, err error) {
	client := pb.NewEltonServiceClient(fs.connection)
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
