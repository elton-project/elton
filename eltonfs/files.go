package main

import (
	"fmt"
	"time"

	"github.com/hanwen/go-fuse/fuse"
)

type EltonFile interface {
	Stat(out *fuse.Attr)
	Name() string
}

type eltonFile struct {
	Key     string
	Version uint64
	Size    uint64
	Time    time.Time
}

func (f *eltonFile) Stat(out *fuse.Attr) {
	out.Mode = fuse.S_IFREG | 0644
	out.Size = f.Size
}

func (f *eltonFile) Name() string {
	return fmt.Sprintf("%s-%d", f.Key, f.Version)
}
