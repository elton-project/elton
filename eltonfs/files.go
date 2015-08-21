package main

import (
	"encoding/json"
	"io/ioutil"
	"sync"
	"syscall"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

type eltonFile struct {
	nodefs.File
	node *eltonNode

	key      string
	version  uint64
	delegate string
}

func (f *eltonFile) InnerFile() nodefs.File {
	return f.File
}

func (f *eltonFile) Flush() fuse.Status {
	code := f.File.Flush()
	if !code.Ok() {
		return code
	}

	st := syscall.Stat_t{}
	err := syscall.Stat(f.node.filename(), &st)
	f.node.info.Size = uint64(st.Size)
	f.node.info.Blocks = uint64(st.Blocks)
	return fuse.ToStatus(err)
}

func (f *eltonFile) Read(buf []byte, off int64) (res fuse.ReadResult, code fuse.Status) {
	return f.File.Read(buf, off)
}

func (f *eltonFile) Write(data []byte, off int64) (uint32, fuse.Status) {
	f.commit()
	return f.File.Write(data, off)
}

func (f *eltonFile) commit() {
	root := f.node.fs.Root()
	children := root.Inode().Children()

	var files []FileInfo
	for k, v := range children {
		if !v.IsDir() {
			n := v.Node().(*eltonNode)
			files = append(
				files,
				FileInfo{
					Name:     k,
					Key:      n.file.key,
					Version:  n.file.version,
					Delegate: n.file.delegate,
					Size:     n.info.Size,
					Time:     n.info.ChangeTime(),
				},
			)
		}
	}

	data, _ := json.Marshal(files)
	ioutil.WriteFile("/tmp/hogehogehoge", data, 0600)
}
