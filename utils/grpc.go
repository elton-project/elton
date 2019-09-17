package utils

import (
	"context"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"sync"
)

// GrpcServeWithCtx serve gRPC server.
// When context is cancelled, gRPC server shutdown gracefully.
func GrpcServeWithCtx(srv *grpc.Server, ctx context.Context, listener net.Listener) error {
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

func WithTestServer(srv subsystems.Server, callback func(ctx context.Context, dial func() *grpc.ClientConn)) {
	eg := errgroup.Group{}
	l := &bufconn.Listener{}

	ctx, cancel := context.WithCancel(context.Background())
	srv.SetListener(l)
	if err := srv.Configure(); err != nil {
		panic(err)
	}
	eg.Go(func() error {
		return srv.Serve(ctx)
	})
	defer func() {
		cancel()
		if err := eg.Wait(); err != nil {
			panic(err)
		}
	}()

	// Prepare
	opt := grpc.WithContextDialer(func(ctx context.Context, target string) (conn net.Conn, err error) {
		conn, err = l.Dial()
		if err != nil {
			panic(xerrors.Errorf("failed to dial: %w", err))
		}
		return
	})

	dial := func() *grpc.ClientConn {
		conn, err := grpc.Dial("", opt)
		if err != nil {
			panic(xerrors.Errorf("failed to grpc.Dial: %w", err))
		}
		return conn
	}
	callback(ctx, dial)
}
