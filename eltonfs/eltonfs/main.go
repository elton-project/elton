package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/eltonfs"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/jessevdk/go-flags"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var opts eltonfs.Options
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "eltonfs"
	parser.Usage = "[OPTIONS] ELTON_HOST MOUNTPOINT"
	args, err := parser.Parse()
	if err != nil {
		os.Exit(2)
	}

	if len(args) != 2 {
		parser.WriteHelp(os.Stdout)
		os.Exit(2)
	}

	root, err := eltonfs.NewEltonFSRoot(args[0], &opts)
	if err != nil {
		log.Printf("NewEltonFSRoot failed: %v", err)
		os.Exit(2)
	}

	conn := nodefs.NewFileSystemConnector(root, nil)

	origAbs, _ := filepath.Abs(args[1])
	mOpts := &fuse.MountOptions{
		AllowOther: true,
		Name:       "eltonfs",
		Options:    opts.FuseOptions,
		FsName:     origAbs,
	}
	server, err := fuse.NewServer(conn.RawFS(), args[1], mOpts)
	if err != nil {
		log.Printf("Mount fail: %v", err)
		os.Exit(2)
	}

	server.SetDebug(opts.Debug)
	server.Serve()
}
