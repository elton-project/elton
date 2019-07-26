package localStorage

import (
	"context"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems"
	"net"
	"os"
)

const DefaultListenAddr = "0.0.0.0:0"
const DefaultCacheDir = "/var/tmp/elton-local-storage"

func NewLocalStorageServer() subsystems.Server {
	return &LocalStorage{
		ListenAddr: DefaultListenAddr,
		CacheDir:   DefaultCacheDir,
	}
}

type LocalStorage struct {
	ListenAddr string
	CacheDir   string

	paht     string
	listener net.Listener
}

func (s *LocalStorage) Name() string {
	return "local-storage"
}
func (s *LocalStorage) Configure() error {
	_, err := os.Stat(s.CacheDir)
	if os.IsNotExist(err) {
		return os.Mkdir(s.CacheDir, 0700)
	}
	return nil
}
func (s *LocalStorage) Listen() error {
	l, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.listener = l
	return nil
}
func (s *LocalStorage) Serve(ctx context.Context) error {
	panic("not implemented")
}
