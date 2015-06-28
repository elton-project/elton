package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

type EltonNode struct {
	nodefs.Node
	file     EltonFile
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

func (n *EltonNode) OpenDir(context *fuse.Context) (stream []fuse.DirEntry, code fuse.Status) {
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

func (n *EltonNode) Open(flags uint32, context *fuse.Context) (fuseFile nodefs.File, code fuse.Status) {
	if flags&fuse.O_ANYWRITE != 0 {
		return nil, fuse.EPERM
	}

	fullPath := n.getPath(n.file.Name())
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		//		client := pb.NewEltonServiceClient(fs.connection)
		//		stream, err := client.GetObject(context.Background(), pb.ObjectInfo{ObjectId: n.file.Name, Version: n.file.Version})
		res, err := http.Get("http://" + path.Join(n.fs.eltonURL, "api", "elton", n.file.Name()))
		if err != nil {
			return nil, fuse.ToStatus(err)
		}
		defer res.Body.Close()

		dir := filepath.Dir(fullPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err = os.MkdirAll(dir, 0700); err != nil {
				return nil, fuse.ToStatus(err)
			}
		}

		out, err := os.Create(fullPath)
		if err != nil {
			out.Close()
			return nil, fuse.ToStatus(err)
		}

		io.Copy(out, res.Body)
		out.Close()
	}

	f, err := os.OpenFile(fullPath, int(flags), 0)
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	return nodefs.NewLoopbackFile(f), fuse.OK
}

func (n *EltonNode) Deletable() bool {
	return false
}

func (n *EltonNode) GetAttr(out *fuse.Attr, file nodefs.File, context *fuse.Context) fuse.Status {
	if n.Inode().IsDir() {
		out.Mode = fuse.S_IFDIR | 0744
		return fuse.OK
	}
	n.file.Stat(out)
	return fuse.OK
}

func (n *EltonNode) getPath(relPath string) string {
	return filepath.Join(n.basePath, relPath)
}
