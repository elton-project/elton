package main

import (
	"log"
	"os"

	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	Debug    bool   `long:"debug" default:"false" description:"print debbuging messages."`
	HostName string `long:"host" default:"localhost" description:"this host name"`
	Port     uint64 `short:"p" long:"port" default:"51823" description:"grpc listen port"`
	UpperDir string `long:"upperdir" required:"true" description:"union mount to upper rw directory."`
	LowerDir string `long:"lowerdir" required:"true" description:"union mount to lower ro directory"`
}

func main() {
	var opts Options
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

	root, err := NewEltonFSRoot(args[0], &opts)
	if err != nil {
		log.Printf("NewEltonFSRoot failed: %v", err)
		os.Exit(2)
	}

	fs, _, err := nodefs.MountRoot(args[1], root, nil)
	if err != nil {
		log.Printf("Mount fail: %v", err)
		os.Exit(2)
	}

	fs.SetDebug(opts.Debug)
	fs.Serve()
}
