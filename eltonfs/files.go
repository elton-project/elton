package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	if f.key == ELTONFS_COMMIT_NAME {
		f.commit()
	}

	return f.File.Write(data, off)
}

func (f *eltonFile) commit() {
	f.node.fs.mux.Lock()
	defer f.node.fs.mux.Unlock()
	files := f.getFileTree("", f.node.fs.Root())

	data, _ := json.Marshal(files)
	ioutil.WriteFile("/tmp/hogehogehoge", data, 0600)
}

func (f *eltonFile) getFileTree(prefix string, root nodefs.Node) (files []FileInfo) {
	if prefix != "" {
		prefix += "/"
	}

	for k, v := range root.Inode().Children() {
		if k == ELTONFS_CONFIG_DIR {
			continue
		}

		p := fmt.Sprintf("%s%s", prefix, k)
		if v.IsDir() {
			files = append(files, f.getFileTree(p, v.Node())...)
			continue
		}

		n := v.Node().(*eltonNode)
		if n.basePath == n.fs.upper {
		}
		files = append(
			files,
			FileInfo{
				Name:     p,
				Key:      n.file.key,
				Version:  n.file.version,
				Delegate: n.file.delegate,
				Size:     n.info.Size,
				Time:     n.info.ChangeTime(),
			},
		)
	}

	return
}
