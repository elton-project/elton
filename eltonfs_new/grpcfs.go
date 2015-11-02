package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	pb "git.t-lab.cs.teu.ac.jp/nashio/elton/grpc/proto"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

const _XATTRSEP = "@XATTR@"

type attrResponse struct {
	*fuse.Attr
	fuse.Status
}

type xattrResponse struct {
	data []byte
	fuse.Status
}

type dirResponse struct {
	entries []fuse.DirEntry
	fuse.Status
}

type linkResponse struct {
	linkContent string
	fuse.Status
}

type grpcFileSystem struct {
	pathfs.FileSystem

	connection *grpc.ClientConn

	attributes *TimedCache
	dirs       *TimedCache
	links      *TimedCache
	xattr      *TimedCache
}

func readDir(fs pathfs.FileSystem, name string) *dirResponse {
	origStream, code := fs.OpenDir(name, nil)

	r := &dirResponse{nil, code}
	if !code.Ok() {
		return r
	}
	r.entries = origStream
	return r
}

func getAttr(fs pathfs.FileSystem, name string) *attrResponse {
	a, code := fs.GetAttr(name, nil)
	return &attrResponse{
		Attr:   a,
		Status: code,
	}
}

func getXAttr(fs pathfs.FileSystem, nameAttr string) *xattrResponse {
	ns := strings.SplitN(nameAttr, _XATTRSEP, 2)
	a, code := fs.GetXAttr(ns[0], ns[1], nil)
	return &xattrResponse{
		data:   a,
		Status: code,
	}
}

func readLink(fs pathfs.FileSystem, name string) *linkResponse {
	a, code := fs.Readlink(name, nil)
	return &linkResponse{
		linkContent: a,
		Status:      code,
	}
}

func NewGrpcFileSystem(fs pathfs.FileSystem, ttl time.Duration) pathfs.FileSystem {
	c := new(grpcFileSystem)
	c.FileSystem = fs
	conn, err := grpc.Dial(fs.(*loopbackFileSystem).Opts.Target, []grpc.DialOption{grpc.WithInsecure()}...)
	if err != nil {
		return nil
	}

	c.connection = conn
	c.attributes = NewTimedCache(func(n string) (interface{}, bool) {
		a := getAttr(fs, n)
		return a, a.Ok()
	}, ttl)
	c.dirs = NewTimedCache(func(n string) (interface{}, bool) {
		d := readDir(fs, n)
		return d, d.Ok()
	}, ttl)
	c.links = NewTimedCache(func(n string) (interface{}, bool) {
		l := readLink(fs, n)
		return l, l.Ok()
	}, ttl)
	c.xattr = NewTimedCache(func(n string) (interface{}, bool) {
		l := getXAttr(fs, n)
		return l, l.Ok()
	}, ttl)
	return c
}

var ELTON_CONFIG_NAME = ".elton_config"
var ELTON_MOUNT_INFO = ".elton_mountinfo"

// func (fs *grpcFileSystem) getFilesInfo() (files []FileInfo, err error) {
// }

func (fs *grpcFileSystem) getMountInfo() (*pb.ObjectInfo, error) {
	p := fs.FileSystem.(*loopbackFileSystem).GetPath(ELTON_CONFIG_NAME)
	buf, err := ioutil.ReadFile(p)
	if err != nil {
		return fs.createMountInfo(p)
	}

	obj := new(pb.ObjectInfo)
	err = json.Unmarshal(buf, obj)
	return obj, err
}

func (fs *grpcFileSystem) createMountInfo(p string) (*pb.ObjectInfo, error) {
	obj, err := fs.generateObjectID(p)

	if err = ioutil.WriteFile(fs.FileSystem.(*loopbackFileSystem).GetPath(ELTON_MOUNT_INFO), []byte("[]"), 0644); err != nil {
		return obj, err
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return obj, err
	}

	if err = ioutil.WriteFile(p, data, 0644); err != nil {
		return obj, err
	}

	obj.RequestHostname = fmt.Sprintf("%s:%d", fs.FileSystem.(*loopbackFileSystem).Opts.HostName, fs.FileSystem.(*loopbackFileSystem).Opts.Port)
	client := pb.NewEltonServiceClient(fs.connection)
	_, err = client.CommitObjectInfo(context.Background(), obj)
	return obj, err
}

func (fs *grpcFileSystem) generateObjectID(p string) (*pb.ObjectInfo, error) {
	client := pb.NewEltonServiceClient(fs.connection)
	stream, err := client.GenerateObjectInfo(
		context.Background(),
		&pb.ObjectInfo{
			ObjectId: p,
		},
	)
	if err != nil {
		return nil, err
	}

	return stream.Recv()
}

func (fs *grpcFileSystem) DropCache() {
	for _, c := range []*TimedCache{fs.attributes, fs.dirs, fs.links, fs.xattr} {
		c.DropAll(nil)
	}
}

func (fs *grpcFileSystem) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	if name == _DROP_CACHE {
		return &fuse.Attr{
			Mode: fuse.S_IFREG | 0777,
		}, fuse.OK
	}

	r := fs.attributes.Get(name).(*attrResponse)
	return r.Attr, r.Status
}

func (fs *grpcFileSystem) GetXAttr(name string, attr string, context *fuse.Context) ([]byte, fuse.Status) {
	key := name + _XATTRSEP + attr
	r := fs.xattr.Get(key).(*xattrResponse)
	return r.data, r.Status
}

func (fs *grpcFileSystem) Readlink(name string, context *fuse.Context) (string, fuse.Status) {
	r := fs.links.Get(name).(*linkResponse)
	return r.linkContent, r.Status
}

func (fs *grpcFileSystem) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, status fuse.Status) {
	r := fs.dirs.Get(name).(*dirResponse)
	return r.entries, r.Status
}

func (fs *grpcFileSystem) String() string {
	return fmt.Sprintf("grpcFileSystem(%v)", fs.FileSystem)
}

func (fs *grpcFileSystem) Open(name string, flags uint32, context *fuse.Context) (f nodefs.File, status fuse.Status) {
	if flags&fuse.O_ANYWRITE != 0 && name == _DROP_CACHE {
		log.Println("Dropping cache for", fs)
		fs.DropCache()
	}
	return fs.FileSystem.Open(name, flags, context)
}
