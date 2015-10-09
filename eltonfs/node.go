package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/net/context"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"

	pb "../grpc/proto"
)

type eltonNode struct {
	nodefs.Node
	file     *eltonFile
	fs       *eltonFS
	basePath string
	link     string
	info     fuse.Attr
}

func (n *eltonNode) OnMount(c *nodefs.FileSystemConnector) {
	go n.fs.Server.Serve()
	n.fs.newEltonTree()
}

func (n *eltonNode) OnUnmount() {
	n.fs.Server.Stop()
	n.fs.connection.Close()
}

func (n *eltonNode) newNode(name string, isDir bool) (newNode *eltonNode, err error) {
	if isDir {
		newNode = &eltonNode{
			Node: nodefs.NewDefaultNode(),
			fs:   n.fs,
		}

		t := time.Now()
		newNode.info.SetTimes(&t, &t, &t)
	} else {
		newNode, err = n.fs.newNode(name, time.Now())
		if err != nil {
			return nil, err
		}
	}

	n.Inode().NewChild(name, isDir, newNode)
	return newNode, nil
}

func (n *eltonNode) filename() string {
	if n.file.key == ELTONFS_COMMIT_NAME || n.file.key == ELTONFS_CONFIG_NAME {
		return filepath.Join(n.basePath, n.file.key)
	}

	return filepath.Join(n.basePath, n.file.key[:2], fmt.Sprintf("%s-%d", n.file.key[2:], n.file.version))
}

func (n *eltonNode) Deletable() bool {
	return true
}

func (n *eltonNode) Readlink(c *fuse.Context) ([]byte, fuse.Status) {
	return []byte(n.link), fuse.OK
}

func (n *eltonNode) StatFs() *fuse.StatfsOut {
	return new(fuse.StatfsOut)
}

func (n *eltonNode) Mkdir(name string, mode uint32, c *fuse.Context) (newNode *nodefs.Inode, code fuse.Status) {
	ch, err := n.newNode(name, true)
	if err != nil {
		return nil, fuse.ENOENT
	}

	ch.info.Mode = fuse.S_IFDIR
	return ch.Inode(), fuse.OK
}

func (n *eltonNode) Unlink(name string, c *fuse.Context) (code fuse.Status) {
	ch := n.Inode().RmChild(name)
	if ch == nil {
		return fuse.ENOENT
	}

	return fuse.OK
}

func (n *eltonNode) Rmdir(name string, c *fuse.Context) (code fuse.Status) {
	return n.Unlink(name, c)
}

func (n *eltonNode) Symlink(name string, content string, c *fuse.Context) (newNode *nodefs.Inode, code fuse.Status) {
	ch, err := n.newNode(name, false)
	if err != nil {
		return nil, fuse.ENOENT
	}

	ch.info.Mode = fuse.S_IFLNK | 0744
	ch.link = content

	return ch.Inode(), fuse.OK
}

func (n *eltonNode) Rename(oldName string, newParent nodefs.Node, newName string, c *fuse.Context) (code fuse.Status) {
	ch := n.Inode().RmChild(oldName)
	newParent.Inode().RmChild(newName)
	newParent.Inode().AddChild(newName, ch)

	return fuse.OK
}

func (n *eltonNode) Link(name string, existing nodefs.Node, c *fuse.Context) (*nodefs.Inode, fuse.Status) {
	n.Inode().AddChild(name, existing.Inode())
	return existing.Inode(), fuse.OK
}

func (n *eltonNode) Create(name string, flags uint32, mode uint32, c *fuse.Context) (file nodefs.File, child *nodefs.Inode, code fuse.Status) {
	ch, err := n.newNode(name, false)
	if err != nil {
		return nil, nil, fuse.ENOENT
	}

	ch.info.Mode = mode | fuse.S_IFREG

	fullPath := ch.filename()
	dir := filepath.Dir(fullPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0744); err != nil {
			return nil, nil, fuse.ToStatus(err)
		}
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return nil, nil, fuse.ToStatus(err)
	}

	return ch.newFile(f), ch.Inode(), fuse.OK
}

func (n *eltonNode) newFile(f *os.File) nodefs.File {
	n.file.File = nodefs.NewLoopbackFile(f)
	n.file.node = n
	return n.file
}

func (n *eltonNode) Open(flags uint32, c *fuse.Context) (fuseFile nodefs.File, code fuse.Status) {
	if flags&fuse.O_ANYWRITE != 0 {
		return n.write(flags, c)
	}

	fullPath := n.filename()
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		if err = n.getFile(flags, c); err != nil {
			return nil, fuse.ToStatus(err)
		}
	}

	f, err := os.OpenFile(fullPath, int(flags), 0)
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	return n.newFile(f), fuse.OK
}

func (n *eltonNode) write(flags uint32, c *fuse.Context) (fuseFile nodefs.File, code fuse.Status) {
	if n.basePath == n.fs.lower {
		n.basePath = n.fs.upper
	}

	fullPath := n.filename()
	dir := filepath.Dir(fullPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0744); err != nil {
			return nil, fuse.ToStatus(err)
		}
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	return n.newFile(f), fuse.OK
}

func (n *eltonNode) getFile(flags uint32, c *fuse.Context) (err error) {
	client := pb.NewEltonServiceClient(n.fs.connection)
	stream, err := client.GetObject(
		context.Background(),
		&pb.ObjectInfo{
			ObjectId:        n.file.key,
			Version:         n.file.version,
			Delegate:        n.file.delegate,
			RequestHostname: fmt.Sprintf("%s:%d", n.fs.options.HostName, n.fs.options.Port),
		},
	)
	if err != nil {
		return
	}

	obj, err := stream.Recv()
	if err != nil {
		return
	}

	data, err := base64.StdEncoding.DecodeString(obj.Body)
	if err != nil {
		return
	}

	return CreateFile(n.filename(), data)
}

func (n *eltonNode) GetAttr(out *fuse.Attr, file nodefs.File, c *fuse.Context) fuse.Status {
	if n.Inode().IsDir() {
		out.Mode = fuse.S_IFDIR | 0744
		return fuse.OK
	}

	*out = n.info
	return fuse.OK
}

func (n *eltonNode) Truncate(file nodefs.File, size uint64, c *fuse.Context) (code fuse.Status) {
	if file != nil {
		code = file.Truncate(size)
	} else {
		err := os.Truncate(n.filename(), int64(size))
		code = fuse.ToStatus(err)
	}

	if code.Ok() {
		now := time.Now()
		// TODO - should update mtime too?
		n.info.SetTimes(nil, nil, &now)
		n.info.Size = size
	}

	return code
}

func (n *eltonNode) Utimens(file nodefs.File, atime *time.Time, mtime *time.Time, c *fuse.Context) (code fuse.Status) {
	now := time.Now()
	n.info.SetTimes(atime, mtime, &now)
	return fuse.OK
}

func (n *eltonNode) Chmod(file nodefs.File, perms uint32, c *fuse.Context) (code fuse.Status) {
	n.info.Mode = (n.info.Mode ^ 07777) | perms
	now := time.Now()
	n.info.SetTimes(nil, nil, &now)
	return fuse.OK
}

func (n *eltonNode) Chown(file nodefs.File, uid uint32, gid uint32, c *fuse.Context) (code fuse.Status) {
	n.info.Uid = uid
	n.info.Gid = gid
	now := time.Now()
	n.info.SetTimes(nil, nil, &now)
	return fuse.OK
}
