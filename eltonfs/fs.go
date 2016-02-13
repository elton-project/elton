package eltonfs

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"google.golang.org/grpc"

	pb "git.t-lab.cs.teu.ac.jp/nashio/elton/grpc/proto"
)

type Options struct {
	Debug       bool     `long:"debug" default:"false" description:"print debbuging messages."`
	FuseOptions []string `short:"o" description:"fuse options"`
	HostName    string   `long:"host" default:"localhost" description:"this host name"`
	Port        uint64   `short:"p" long:"port" default:"51823" description:"grpc listen port"`
	UpperDir    string   `long:"upperdir" required:"true" description:"union mount to upper rw directory."`
	LowerDir    string   `long:"lowerdir" required:"true" description:"union mount to lower ro directory"`
	StandAlone  bool     `long:"standalone" default:"false" description:"stand-alone mode"`
}

type FileInfo struct {
	Name     string
	Key      string
	Version  uint64
	Delegate string
	Size     uint64
	Time     time.Time
}

const (
	FILEMODE            os.FileMode = 0644
	ELTONFS_CONFIG_DIR  string      = ".eltonfs"
	ELTONFS_CONFIG_NAME string      = "CONFIG"
	ELTONFS_COMMIT_NAME string      = "COMMIT"
)

func NewEltonFSRoot(target string, opts *Options) (root nodefs.Node, err error) {
	if err = ioutil.WriteFile(filepath.Join(opts.UpperDir, ELTONFS_COMMIT_NAME), []byte(""), FILEMODE); err != nil {
		return
	}

	server := NewEltonFSGrpcServer(opts)
	efs, err := NewEltonFS(opts.LowerDir, opts.UpperDir, target, opts, server)
	if err != nil {
		return nil, err
	}

	return efs.Root(), nil
}

func NewEltonFS(lower, upper, eltonURL string, opts *Options, server *EltonFSGrpcServer) (*eltonFS, error) {
	conn, err := grpc.Dial(eltonURL, []grpc.DialOption{grpc.WithInsecure()}...)
	if err != nil {
		return nil, err
	}

	fs := &eltonFS{
		lower:      lower,
		upper:      upper,
		eltonURL:   eltonURL,
		connection: conn,
		options:    opts,
		Server:     server,
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
	options    *Options
	Server     *EltonFSGrpcServer
}

func (fs *eltonFS) String() string {
	return fmt.Sprintf("EltonFS(%s)", "elton")
}

func (fs *eltonFS) Root() nodefs.Node {
	return fs.root
}

func (fs *eltonFS) newNode(name string, t time.Time) *eltonNode {
	n := &eltonNode{
		Node:     nodefs.NewDefaultNode(),
		basePath: fs.upper,
		fs:       fs,
	}

	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s%d", name, t.Nanosecond())))
	n.key = string(hex.EncodeToString(hasher.Sum(nil)))

	n.info.SetTimes(&t, &t, &t)
	n.info.Mode = fuse.S_IFDIR | 0644

	return n
}

func (fs *eltonFS) Filename(n *nodefs.Inode) string {
	mn := n.Node().(*eltonNode)
	return mn.filename()
}

func (fs *eltonFS) newEltonTree() {
	files, err := fs.getFilesInfo()
	if err != nil {
		log.Fatalln(err)
	}

	for _, f := range files {
		comps := strings.Split(f.Name, "/")

		node := fs.root.Inode()
		for i, c := range comps {
			child := node.GetChild(c)
			if child == nil {
				fsnode := fs.newNode(c, f.Time)
				if i == len(comps)-1 {
					fsnode.key = f.Key
					fsnode.version = f.Version
					fsnode.delegate = f.Delegate
					fsnode.basePath = fs.lower
					fsnode.info.Mode = fuse.S_IFREG | 0644
					fsnode.info.Size = f.Size
				}

				child = node.NewChild(c, fsnode.key == c, fsnode)
			}
			node = child
		}
	}

	child := fs.root.Inode().NewChild(ELTONFS_CONFIG_DIR, true, fs.newNode(ELTONFS_CONFIG_DIR, time.Now()))

	config := fs.newNode(ELTONFS_CONFIG_NAME, time.Now())
	config.key = ELTONFS_CONFIG_NAME
	config.info.Mode = fuse.S_IFREG | 0644
	child.NewChild(ELTONFS_CONFIG_NAME, false, config)

	commit := fs.newNode(ELTONFS_COMMIT_NAME, time.Now())
	commit.key = ELTONFS_COMMIT_NAME
	commit.info.Mode = fuse.S_IFREG | 0644
	child.NewChild(ELTONFS_COMMIT_NAME, false, commit)
}

func (fs *eltonFS) getFilesInfo() (files []FileInfo, err error) {
	mi, err := fs.getMountInfo()
	if err != nil {
		return nil, err
	}

	p := filepath.Join(fs.lower, mi.ObjectId[:2], fmt.Sprintf("%s-%d", mi.ObjectId[2:], mi.Version))
	if _, err = os.Stat(p); os.IsNotExist(err) {
		client := pb.NewEltonServiceClient(fs.connection)
		stream, err := client.GetObject(context.Background(), mi)
		if err != nil {
			return nil, err
		}

		fp, err := Create(p)
		if err != nil {
			return nil, err
		}

		writer := bufio.NewWriter(fp)
		for {
			obj, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			_, err = writer.Write(obj.Body)
			if err != nil {
				return nil, err
			}
			writer.Flush()
		}

		fp.Close()
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
		if err = os.MkdirAll(dir, 0744); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(p, data, FILEMODE)
}

func Create(p string) (fp *os.File, err error) {
	dir := filepath.Dir(p)
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0744); err != nil {
			return
		}
	}

	return os.OpenFile(p, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}

func (fs *eltonFS) getMountInfo() (obj *pb.ObjectInfo, err error) {
	p := filepath.Join(fs.upper, ELTONFS_CONFIG_NAME)
	buf, err := ioutil.ReadFile(p)
	if err != nil {
		return fs.createMountInfo(p)
	}

	obj = new(pb.ObjectInfo)
	err = json.Unmarshal(buf, obj)
	return
}

func (fs *eltonFS) createMountInfo(p string) (obj *pb.ObjectInfo, err error) {
	obj, err = fs.generateObjectID(p)
	if err = CreateFile(
		filepath.Join(
			fs.lower,
			obj.ObjectId[:2],
			fmt.Sprintf(
				"%s-%d",
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

	obj.RequestHostname = fmt.Sprintf("%s:%d", fs.options.HostName, fs.options.Port)
	client := pb.NewEltonServiceClient(fs.connection)
	_, err = client.CommitObjectInfo(context.Background(), obj)
	return
}

func (fs *eltonFS) generateObjectID(p string) (obj *pb.ObjectInfo, err error) {
	client := pb.NewEltonServiceClient(fs.connection)
	stream, err := client.GenerateObjectInfo(
		context.Background(),
		&pb.ObjectInfo{
			ObjectId: p,
		},
	)
	if err != nil {
		return
	}

	return stream.Recv()
}
