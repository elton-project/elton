package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"git.t-lab.cs.teu.ac.jp/nashio/elton/eltonfs"
	pb "git.t-lab.cs.teu.ac.jp/nashio/elton/grpc/proto"

	"github.com/calavera/dkvolume"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

type eltonfsConfig struct {
	ServerURL string
	HostName  string
	Port      uint64
	Debug     bool
}

type eltonfsServer struct {
	*fuse.Server
	connections int
}

type eltonfsDriver struct {
	root        string
	config      eltonfsConfig
	servers     map[string]*eltonfsServer
	eltonServer *eltonfs.EltonFSGrpcServer
	m           *sync.Mutex
}

func newEltonfsDriver(root string, config eltonfsConfig) (d eltonfsDriver, err error) {
	lowerdir := filepath.Join(root, "eltonfs-lowerdir")
	fi, err := os.Lstat(lowerdir)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(lowerdir, 0755); err != nil {
			return
		}
	}
	if fi != nil && !fi.IsDir() {
		return
	}

	return eltonfsDriver{
		root:    root,
		config:  config,
		servers: map[string]*eltonfsServer{},
		eltonServer: eltonfs.NewEltonFSGrpcServer(
			&eltonfs.Options{
				Debug:      config.Debug,
				HostName:   config.HostName,
				Port:       config.Port,
				LowerDir:   lowerdir,
				StandAlone: false,
			},
		),
		m: new(sync.Mutex),
	}, nil
}

func (d eltonfsDriver) Create(r dkvolume.Request) dkvolume.Response {
	d.m.Lock()
	defer d.m.Unlock()

	m := d.mountpoint(r.Name)
	upper := d.upperDir(r.Name)

	if _, ok := d.servers[m]; ok {
		return dkvolume.Response{}
	}

	fi, err := os.Lstat(m)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(m, 0755); err != nil {
			return dkvolume.Response{Err: err.Error()}
		}
		if err := os.MkdirAll(upper, 0755); err != nil {
			return dkvolume.Response{Err: err.Error()}
		}

	} else if err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	if fi != nil && !fi.IsDir() {
		return dkvolume.Response{Err: fmt.Sprintf("%v already exist and it's not a directory", m)}
	}

	if _, ok := r.Options["object_id"]; ok {
		if _, ok = r.Options["version"]; ok {
			version, err := strconv.ParseUint(r.Options["version"], 10, 64)
			if err != nil {
				return dkvolume.Response{Err: err.Error()}
			}

			data, err := json.Marshal(pb.ObjectInfo{
				ObjectId: r.Options["object_id"],
				Version:  version,
				Delegate: r.Options["delegate"],
			})
			if err != nil {
				return dkvolume.Response{Err: err.Error()}
			}

			if err = ioutil.WriteFile(
				filepath.Join(upper, eltonfs.ELTONFS_CONFIG_NAME),
				data,
				0644,
			); err != nil {
				return dkvolume.Response{Err: err.Error()}
			}
		}
	}

	return dkvolume.Response{}
}

func (d eltonfsDriver) Remove(r dkvolume.Request) dkvolume.Response {
	d.m.Lock()
	defer d.m.Unlock()

	m := d.mountpoint(r.Name)

	if s, ok := d.servers[m]; ok {
		if s.connections <= 1 {
			delete(d.servers, m)
		}
	}

	return dkvolume.Response{}
}

func (d eltonfsDriver) Path(r dkvolume.Request) dkvolume.Response {
	return dkvolume.Response{Mountpoint: d.mountpoint(r.Name)}
}

func (d eltonfsDriver) Mount(r dkvolume.Request) dkvolume.Response {
	d.m.Lock()
	defer d.m.Unlock()
	m := d.mountpoint(r.Name)
	log.Printf("Mounting volume %s on %s\n", r.Name, m)

	s, ok := d.servers[m]
	if ok && s.connections > 0 {
		s.connections++
		return dkvolume.Response{Mountpoint: m}
	}

	fi, err := os.Lstat(m)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(m, 0755); err != nil {
			return dkvolume.Response{Err: err.Error()}
		}
	} else if err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	if fi != nil && !fi.IsDir() {
		return dkvolume.Response{Err: fmt.Sprintf("%v already exist and it's not a directory", m)}
	}

	server, err := d.mountServer(m)
	if err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	d.servers[m] = &eltonfsServer{Server: server, connections: 1}

	return dkvolume.Response{Mountpoint: m}
}

func (d eltonfsDriver) Unmount(r dkvolume.Request) dkvolume.Response {
	d.m.Lock()
	defer d.m.Unlock()

	m := d.mountpoint(r.Name)
	log.Printf("Unmounting volume %s from %s\n", r.Name, m)

	if s, ok := d.servers[m]; ok {
		if s.connections == 1 {
			s.Unmount()
		}
		s.connections--
	} else {
		return dkvolume.Response{Err: fmt.Sprintf("Unable to find volume mounted on %s", m)}
	}

	return dkvolume.Response{}
}

func (d eltonfsDriver) mountpoint(name string) string {
	return filepath.Join(d.root, name)
}

func (d eltonfsDriver) upperDir(name string) string {
	return filepath.Join(d.root, fmt.Sprintf("%s-%s", name, "upper"))
}

func (d eltonfsDriver) lowerDir() string {
	return filepath.Join(d.root, "eltonfs-lowerdir")
}

func (d eltonfsDriver) mountServer(mountpoint string) (*fuse.Server, error) {
	root, err := eltonfs.NewEltonFSRoot(
		d.config.ServerURL,
		&eltonfs.Options{
			Debug:      d.config.Debug,
			HostName:   d.config.HostName,
			Port:       d.config.Port,
			UpperDir:   fmt.Sprintf("%s-%s", mountpoint, "upper"),
			LowerDir:   d.lowerDir(),
			StandAlone: true,
		},
	)
	if err != nil {
		return nil, err
	}

	conn := nodefs.NewFileSystemConnector(root, nil)
	server, err := fuse.NewServer(conn.RawFS(), mountpoint, nil)
	if err != nil {
		return nil, err
	}

	go server.Serve()

	return server, nil
}
