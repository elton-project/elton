package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/calavera/dkvolume"
)

const (
	eltonfsId     = "_eltonfs"
	socketAddress = "/run/docker/plugins/eltonfs.sock"
)

var (
	defaultPath = filepath.Join(dkvolume.DefaultDockerRootDirectory, eltonfsId)

	root     = flag.String("root", defaultPath, "Docker volumes root directory")
	hostname = flag.String("hostname", "localhost", "Local hostname")
	port     = flag.Uint64("port", 42339, "listen elton port")
	debug    = flag.Bool("debug", false, "Debug")
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] elton_server\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		Usage()
		os.Exit(1)
	}

	config := eltonfsConfig{
		ServerURL: flag.Args()[0],
		HostName:  *hostname,
		Port:      *port,
		Debug:     *debug,
	}

	d, err := newEltonfsDriver(*root, config)
	if err != nil {
		fmt.Errorf(err.Error())
		os.Exit(2)
	}

	h := dkvolume.NewHandler(d)
	go d.eltonServer.Serve()
	defer d.eltonServer.Stop()

	fmt.Printf("Listening on %s\n", socketAddress)
	fmt.Println(h.ServeUnix("root", socketAddress))
}
