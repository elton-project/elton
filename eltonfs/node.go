package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/net/context"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"

	pb "../grpc/proto"
)

type EltonNode struct {
	nodefs.Node
	file     *EltonFile
	fs       *EltonFS
	basePath string
}

func (n *EltonNode) OnMount(c *nodefs.FileSystemConnector) {
	n.fs.onMount()
}

func (n *EltonNode) OnUnmount() {
	n.fs.onUnmount()
}

func (n *EltonNode) Print(indent int) {
	s := ""
	for i := 0; i < indent; i++ {
		s = s + " "
	}

	children := n.Inode().Children()
	for k, v := range children {
		if v.IsDir() {
			fmt.Println(s + k + ":")
			mn, ok := v.Node().(*EltonNode)
			if ok {
				mn.Print(indent + 2)
			}
		} else {
			fmt.Println(s + k)
		}
	}
}

func (n *EltonNode) OpenDir(c *fuse.Context) (stream []fuse.DirEntry, code fuse.Status) {
	children := n.Inode().Children()
	stream = make([]fuse.DirEntry, 0, len(children))
	for k, v := range children {
		mode := fuse.S_IFREG | 0600
		if v.IsDir() {
			mode = fuse.S_IFDIR | 0700
		}
		stream = append(stream, fuse.DirEntry{
			Name: k,
			Mode: uint32(mode),
		})
	}
	return stream, fuse.OK
}

func (n *EltonNode) Open(flags uint32, c *fuse.Context) (fuseFile nodefs.File, code fuse.Status) {
	log.Println("open")
	if flags&fuse.O_ANYWRITE != 0 {
		if n.basePath == n.fs.lower {
			n.basePath = n.fs.upper
			n.file.Version++
		}

		fullPath := n.getPath(n.file.Name())
		dir := filepath.Dir(fullPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err = os.MkdirAll(dir, 0700); err != nil {
				return nil, fuse.ToStatus(err)
			}
		}

		f, err := os.Create(fullPath)
		if err != nil {
			return nil, fuse.ToStatus(err)
		}

		return nodefs.NewLoopbackFile(f), fuse.OK
	}

	fullPath := n.getPath(n.file.Name())
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		client := pb.NewEltonServiceClient(n.fs.connection)
		stream, err := client.GetObject(
			context.Background(),
			&pb.ObjectInfo{
				ObjectId: n.file.Key,
				Version:  n.file.Version,
				Delegate: n.file.Delegate,
			},
		)
		if err != nil {
			return nil, fuse.ToStatus(err)
		}

		obj, err := stream.Recv()
		if err != nil {
			return nil, fuse.ToStatus(err)
		}
		data, err := base64.StdEncoding.DecodeString(obj.Body)
		if err != nil {
			return nil, fuse.ToStatus(err)
		}

		if err = CreateFile(fullPath, data); err != nil {
			return nil, fuse.ToStatus(err)
		}
	}

	f, err := os.OpenFile(fullPath, int(flags), 0)
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	return nodefs.NewLoopbackFile(f), fuse.OK
}

func (n *EltonNode) Create(name string, flags uint32, mode uint32, c *fuse.Context) (file nodefs.File, child *nodefs.Inode, code fuse.Status) {
	fsnode := &EltonNode{
		Node:     nodefs.NewDefaultNode(),
		basePath: n.fs.upper,
		fs:       n.fs,
		file: &EltonFile{
			Key:      name,
			Version:  uint64(0),
			Delegate: "",
			Size:     uint64(0),
			Time:     time.Now(),
		},
	}

	f, err := os.Create(fsnode.getPath(fsnode.file.Name()))
	if err != nil {
		return nil, nil, fuse.ToStatus(err)
	}

	child = n.Inode().NewChild(name, false, fsnode)
	return nodefs.NewLoopbackFile(f), child, fuse.OK
}

// func (n *EltonNode) GetAttr(out *fuse.Attr, file nodefs.File, c *fuse.Context) fuse.Status {
// 	log.Println("hideo")
// 	if n.Inode().IsDir() {
// 		out.Mode = fuse.S_IFDIR | 0700
// 		return fuse.OK
// 	}
// 	n.file.Stat(out)
// 	return fuse.OK
// }

func (n *EltonNode) getPath(relPath string) string {
	return filepath.Join(n.basePath, fmt.Sprintf("%s/%s", relPath[:2], relPath[2:]))
}
