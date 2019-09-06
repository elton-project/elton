package utils

import (
	"context"
	"google.golang.org/grpc"
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
