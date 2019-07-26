package localStorage

import (
	"context"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems"
	"google.golang.org/grpc"
	"net"
	"os"
	"sync"
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
	handler := &StorageService{}
	srv := grpc.NewServer(nil)
	elton_v2.RegisterStorageServiceServer(srv, handler)

	return grpcServeWithCtx(srv, ctx, s.listener)
}

// grpcServeWithCtx serve gRPC server.
// When context is cancelled, gRPC server shutdown gracefully.
func grpcServeWithCtx(srv *grpc.Server, ctx context.Context, listener net.Listener) error {
	srvCtx, cancel := context.WithCancel(ctx)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-srvCtx.Done()
		srv.GracefulStop()
	}()

	err := srv.Serve(listener)

	cancel()
	wg.Done()
	return err
}
