package main

import (
	"fmt"
	"time"

	"github.com/hanwen/go-fuse/fuse"
)

type EltonFile struct {
	Key      string
	Version  uint64
	Delegate string
	Size     uint64
	Time     time.Time
}

func (f *EltonFile) Stat(out *fuse.Attr) {
	out.Mode = fuse.S_IFREG | 0600
	out.Size = f.Size
}

func (f *EltonFile) Name() string {
	return fmt.Sprintf("%s-%d", f.Key, f.Version)
}
