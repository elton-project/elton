package proto2

import (
	"context"
	"github.com/pkg/errors"
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
		err = errors.Wrap(fn(srv), "WithGrpcServer(): fn()")
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
		return errors.Wrapf(err, "grpc.Dial(addr=%s)", addr)
	}
	defer conn.Close()
	return errors.Wrapf(fn(conn), "WithGrpcConn(addr=%s): fn()", addr)
}

func ConnectOtherSubsystem(ctx context.Context, subsystemType SubsystemType, discoverer ServiceDiscoverer, fn func(conn *grpc.ClientConn) error) error {
	addr, err := discoverer.Get(ctx, subsystemType)
	if err != nil {
		return errors.Wrapf(err, "ServiceDiscoverer.Get(subsystem=%s)")
	}
	return WithGrpcConn(addr, fn)
}
