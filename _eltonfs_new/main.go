package main

import (
	"log"
	"os"
	"time"

	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	Debug          bool    `long:"debug" default:"false" description:"print debbuging messages."`
	HostName       string  `long:"host" default:"localhost" description:"this host name"`
	Port           uint64  `short:"p" long:"port" default:"51823" description:"grpc listen port"`
	Target         string  `long:"target" default:"localhost:12345" description:"target eltom master host"`
	EntryTTL       float64 `long:"entry_ttl" default:"1.0" description:"fuse entry cache TTL."`
	NegativeTTL    float64 `long:"negative_ttl" default:"1.0" description:"fuse negative entry cache TTL."`
	DelcacheTTL    float64 `long:"delcache_ttl" default:"5.0" description:"Deletion cache TTL in seconds."`
	BranchcacheTTL float64 `long:"branchcacheTTL" default:"5.0" description:"Branch cache TTL in seconds."`
	Deldirname     string  `long:"deletion_dirname" default:"ELTONFS_DELETIONS" description:"Directory name to use for deletions."`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "eltonfs"
	parser.Usage = "[OPTIONS] MOUNTPOINT RW-DIRECTORY RO-DIRECTORY ..."

	args, err := parser.Parse()
	if err != nil {
		os.Exit(2)
	}

	if len(args) != 3 {
		parser.WriteHelp(os.Stdout)
		os.Exit(2)
	}

	efs, err := NewEltonFsFromRoots(args[1:], &opts)
	if err != nil {
		log.Fatal("Cannot create EltonFs ", err)
		os.Exit(1)
	}

	nodeFs := pathfs.NewPathNodeFs(efs, &pathfs.PathNodeFsOptions{ClientInodes: true})
	mOpts := nodefs.Options{
		EntryTimeout:    time.Duration(opts.EntryTTL * float64(time.Second)),
		AttrTimeout:     time.Duration(opts.EntryTTL * float64(time.Second)),
		NegativeTimeout: time.Duration(opts.NegativeTTL * float64(time.Second)),
		PortableInodes:  false,
	}

	mountState, _, err := nodefs.MountRoot(args[0], nodeFs.Root(), &mOpts)
	if err != nil {
		log.Fatal("Mount fail:", err)
	}

	mountState.SetDebug(opts.Debug)
	mountState.Serve()
}
