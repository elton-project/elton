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

	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/eltonfs"
	pb "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto"

	"github.com/docker/go-plugins-helpers/volume"
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

func (d eltonfsDriver) Create(r *volume.CreateRequest) error {
	d.m.Lock()
	defer d.m.Unlock()

	m := d.mountpoint(r.Name)
	upper := d.upperDir(r.Name)

	if _, ok := d.servers[m]; ok {
		return nil
	}

	fi, err := os.Lstat(m)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(m, 0755); err != nil {
			return err
		}
		if err := os.MkdirAll(upper, 0755); err != nil {
			return err
		}

	} else if err != nil {
		return err
	}

	if fi != nil && !fi.IsDir() {
		return fmt.Errorf("%v already exist and it's not a directory", m)
	}

	if _, ok := r.Options["object_id"]; ok {
		if _, ok = r.Options["version"]; ok {
			version, err := strconv.ParseUint(r.Options["version"], 10, 64)
			if err != nil {
				return err
			}

			data, err := json.Marshal(pb.ObjectInfo{
				ObjectId: r.Options["object_id"],
				Version:  version,
				Delegate: r.Options["delegate"],
			})
			if err != nil {
				return err
			}

			if err = ioutil.WriteFile(
				filepath.Join(upper, eltonfs.ELTONFS_CONFIG_NAME),
				data,
				0644,
			); err != nil {
				return err
			}
		}
	}

	return nil
}
func (d eltonfsDriver) List() (*volume.ListResponse, error) {
	d.m.Lock()
	defer d.m.Unlock()

	var volumes []*volume.Volume
	for name := range d.servers {
		volumes = append(volumes, d.volume(name))
	}

	return &volume.ListResponse{
		Volumes: volumes,
	}, nil
}
func (d eltonfsDriver) Get(r *volume.GetRequest) (*volume.GetResponse, error) {
	d.m.Lock()
	defer d.m.Unlock()

	if !d.isExists(r.Name) {
		return nil, fmt.Errorf("volume not found: name=%s", r.Name)
	}

	return &volume.GetResponse{
		Volume: d.volume(r.Name),
	}, nil
}

func (d eltonfsDriver) Remove(r *volume.RemoveRequest) error {
	d.m.Lock()
	defer d.m.Unlock()

	m := d.mountpoint(r.Name)

	if s, ok := d.servers[m]; ok {
		if s.connections <= 1 {
			delete(d.servers, m)
		}
	}

	return nil
}

func (d eltonfsDriver) Path(r *volume.PathRequest) (*volume.PathResponse, error) {
	return &volume.PathResponse{Mountpoint: d.mountpoint(r.Name)}, nil
}

func (d eltonfsDriver) Mount(r *volume.MountRequest) (*volume.MountResponse, error) {
	d.m.Lock()
	defer d.m.Unlock()
	m := d.mountpoint(r.Name)
	log.Printf("Mounting volume %s on %s\n", r.Name, m)

	s, ok := d.servers[m]
	if ok && s.connections > 0 {
		s.connections++
		return &volume.MountResponse{Mountpoint: m}, nil
	}

	fi, err := os.Lstat(m)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(m, 0755); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	if fi != nil && !fi.IsDir() {
		return nil, fmt.Errorf("%v already exist and it's not a directory", m)
	}

	server, err := d.mountServer(m)
	if err != nil {
		return nil, err
	}

	d.servers[m] = &eltonfsServer{Server: server, connections: 1}

	return &volume.MountResponse{Mountpoint: m}, nil
}

func (d eltonfsDriver) Unmount(r *volume.UnmountRequest) error {
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
		return fmt.Errorf("Unable to find volume mounted on %s", m)
	}

	return nil
}

func (d eltonfsDriver) Capabilities() *volume.CapabilitiesResponse {
	return &volume.CapabilitiesResponse{
		Capabilities: volume.Capability{
			Scope: "global",
		},
	}
}

func (d eltonfsDriver) isExists(name string) bool {
	_, ok := d.servers[name]
	return ok
}
func (d eltonfsDriver) volume(name string) *volume.Volume {
	return &volume.Volume{
		Name:       name,
		Mountpoint: d.mountpoint(name),
	}
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

	origAbs, _ := filepath.Abs(mountpoint)
	mOpts := &fuse.MountOptions{
		AllowOther: true,
		Name:       "eltonfs",
		FsName:     origAbs,
	}
	server, err := fuse.NewServer(conn.RawFS(), mountpoint, mOpts)
	if err != nil {
		return nil, err
	}

	go server.Serve()

	return server, nil
}
