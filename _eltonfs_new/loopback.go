package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

type DirInfo struct {
	DirEntries []fuse.DirEntry
	DirAttr    *fuse.Attr
}

type loopbackFileSystem struct {
	pathfs.FileSystem
	Root string
	Opts Options

	FileNameMapMutex sync.RWMutex
	FileNameMap      map[string]string

	DirMapMutex sync.RWMutex
	DirMap      map[string]DirInfo
}

func NewLoopbackFileSystem(root string, opts Options) pathfs.FileSystem {
	return &loopbackFileSystem{
		FileSystem:  pathfs.NewDefaultFileSystem(),
		Root:        root,
		FileNameMap: make(map[string]string),
		DirMap:      make(map[string]DirInfo),
		Opts:        opts,
	}
}

func (fs *loopbackFileSystem) OnMount(nodeFs *pathfs.PathNodeFs) {
}

func (fs *loopbackFileSystem) OnUnmount() {
}

func (fs *loopbackFileSystem) GetPath(relPath string) string {
	fs.FileNameMapMutex.RLock()
	name, ok := fs.FileNameMap[relPath]
	fs.FileNameMapMutex.RUnlock()

	if !ok {
		name = fs.GeneratePath(relPath)
	}

	return filepath.Join(fs.Root, name[:2], name[2:])
}

func (fs *loopbackFileSystem) GeneratePath(relPath string) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s%d", relPath, time.Now().Nanosecond())))
	name := fmt.Sprintf("%x-%d", h.Sum(nil), 0)

	fs.FileNameMapMutex.Lock()
	defer fs.FileNameMapMutex.Unlock()
	fs.FileNameMap[relPath] = name

	return name
}

func (fs *loopbackFileSystem) GetAttr(name string, context *fuse.Context) (a *fuse.Attr, code fuse.Status) {
	fullPath := fs.GetPath(name)
	var err error = nil
	st := syscall.Stat_t{}
	if name == "" {
		err = syscall.Stat(fullPath, &st)
	} else {
		err = syscall.Lstat(fullPath, &st)
	}

	if err != nil {
		return nil, fuse.ToStatus(err)
	}
	a = &fuse.Attr{}
	a.FromStat(&st)
	return a, fuse.OK
}

func (fs *loopbackFileSystem) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, status fuse.Status) {
	fs.DirMapMutex.RLock()
	entries, ok := fs.DirMap[name].DirEntries
	fs.DirMapMutex.RUnlock()

	if !ok {
		return nil, fuse.ENOENT
	}

	return entries, fuse.OK
}

func (fs *loopbackFileSystem) Open(path string, flags uint32, context *fuse.Context) (fuseFile nodefs.File, status fuse.Status) {
	name := fs.GetPath(path)
	dir := filepath.Dir(name)
	if _, err := os.Stat(dir); err != nil {
		os.MkdirAll(dir, 0744)
	}

	f, err := os.OpenFile(fs.GetPath(name), int(flags), 0)
	if err != nil {
		return nil, fuse.ToStatus(err)
	}
	return nodefs.NewLoopbackFile(f), fuse.OK
}

func (fs *loopbackFileSystem) Chmod(path string, mode uint32, context *fuse.Context) (code fuse.Status) {
	err := os.Chmod(fs.GetPath(path), os.FileMode(mode))
	return fuse.ToStatus(err)
}

func (fs *loopbackFileSystem) Chown(path string, uid uint32, gid uint32, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(os.Chown(fs.GetPath(path), int(uid), int(gid)))
}

func (fs *loopbackFileSystem) Truncate(path string, offset uint64, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(os.Truncate(fs.GetPath(path), int64(offset)))
}

func (fs *loopbackFileSystem) Utimens(path string, Atime *time.Time, Mtime *time.Time, context *fuse.Context) (code fuse.Status) {
	var a time.Time
	if Atime != nil {
		a = *Atime
	}
	var m time.Time
	if Mtime != nil {
		m = *Mtime
	}
	return fuse.ToStatus(os.Chtimes(fs.GetPath(path), a, m))
}

func (fs *loopbackFileSystem) Readlink(name string, context *fuse.Context) (out string, code fuse.Status) {
	f, err := os.Readlink(fs.GetPath(name))
	return f, fuse.ToStatus(err)
}

func (fs *loopbackFileSystem) Mknod(name string, mode uint32, dev uint32, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(syscall.Mknod(fs.GetPath(name), mode, int(dev)))
}

func (fs *loopbackFileSystem) Mkdir(path string, mode uint32, context *fuse.Context) (code fuse.Status) {
	fs.DirMapMutex.Lock()
	defer fs.DirMapMutex.Unlock()
	fs.DirMap[path] = DirInfo{
		[]fuse.DirEntry{},
	}

	return fuse.OK
}

func (fs *loopbackFileSystem) Unlink(name string, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(syscall.Unlink(fs.GetPath(name)))
}

func (fs *loopbackFileSystem) Rmdir(name string, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(syscall.Rmdir(fs.GetPath(name)))
}

func (fs *loopbackFileSystem) Symlink(pointedTo string, linkName string, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(os.Symlink(pointedTo, fs.GetPath(linkName)))
}

func (fs *loopbackFileSystem) Rename(oldPath string, newPath string, context *fuse.Context) (codee fuse.Status) {
	err := os.Rename(fs.GetPath(oldPath), fs.GetPath(newPath))
	return fuse.ToStatus(err)
}

func (fs *loopbackFileSystem) Link(orig string, newName string, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(os.Link(fs.GetPath(orig), fs.GetPath(newName)))
}

func (fs *loopbackFileSystem) Access(name string, mode uint32, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(syscall.Access(fs.GetPath(name), mode))
}

func (fs *loopbackFileSystem) Create(path string, flags uint32, mode uint32, context *fuse.Context) (fuseFile nodefs.File, code fuse.Status) {
	name := fs.GetPath(path)
	dir := filepath.Dir(name)
	if _, err := os.Stat(dir); err != nil {
		os.MkdirAll(dir, 0744)
	}

	fs.DirMapMutex.Lock()
	fs.DirMap[filepath.Dir(path)] = append(
		fs.DirMap[filepath.Dir(path)],
		fuse.DirEntry{Name: path, Mode: mode})
	fs.DirMapMutex.Unlock()

	f, err := os.OpenFile(name, int(flags)|os.O_CREATE, os.FileMode(mode))
	return nodefs.NewLoopbackFile(f), fuse.ToStatus(err)
}
