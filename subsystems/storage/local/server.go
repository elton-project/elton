package localStorage

import (
	"context"
	"github.com/yuuki0xff/pathlib"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"google.golang.org/grpc"
	"net"
	"os"
)

const DefaultListenAddr = "0.0.0.0:0"
const DefaultCacheDir = "/var/tmp/elton-local-storage"
const DefaultMaxObjectSize = 10 << 20 // 10MiB

func NewLocalStorageServer() subsystems.Server {
	return &LocalStorage{
		ListenAddr: DefaultListenAddr,
		CacheDir:   DefaultCacheDir,
	}
}

type LocalStorage struct {
	ListenAddr string
	CacheDir   string

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
	if s.listener == nil {
		l, err := net.Listen("tcp", s.ListenAddr)
		if err != nil {
			return err
		}
		s.listener = l
	}
	return nil
}
func (s *LocalStorage) SetListener(l net.Listener) {
	s.listener = l
}
func (s *LocalStorage) Serve(ctx context.Context) error {
	handler := &StorageService{
		Repo: NewRepository(pathlib.New(s.CacheDir), nil, DefaultMaxObjectSize),
	}
	srv := grpc.NewServer(nil)
	elton_v2.RegisterStorageServiceServer(srv, handler)

	return utils.GrpcServeWithCtx(srv, ctx, s.listener)
}
