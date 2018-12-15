package proto2

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"sync"
)

func WithGrpcServer(parent_ctx context.Context, fn func(srv *grpc.Server) error) (err error) {
	var wg sync.WaitGroup
	srv := grpc.NewServer()

	ctx, cancel := context.WithCancel(parent_ctx)
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		srv.GracefulStop()
	}()
	go func() {
		defer wg.Done()
		err = fn(srv)
		cancel()
	}()

	zap.S().Debugw("withGS", "status", "waiting")
	wg.Wait()
	return
}

func WithGrpcConn(addr net.Addr, fn func(conn *grpc.ClientConn) error) error {
	target := addr.Network() + "://" + addr.String()
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()
	return fn(conn)
}

func ConnectOtherSubsystem(ctx context.Context, subsystemType SubsystemType, discoverer ServiceDiscoverer, fn func(conn *grpc.ClientConn) error) error {
	addr, err := discoverer.Get(ctx, subsystemType)
	if err != nil {
		return err
	}

	return WithGrpcConn(addr, fn)
}
