package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/docker/go-plugins-helpers/volume"
)

const (
	eltonfsId     = "_eltonfs"
	socketAddress = "/run/docker/plugins/eltonfs.sock"
	socketGid     = 0 // GID 0 is root group
)

var (
	defaultPath = filepath.Join(volume.DefaultDockerRootDirectory, eltonfsId)

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
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}

	h := volume.NewHandler(d)
	go d.eltonServer.Serve()
	defer d.eltonServer.Stop()

	fmt.Printf("Listening on %s\n", socketAddress)
	fmt.Println(h.ServeUnix(socketAddress, socketGid))
}
