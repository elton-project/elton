package main

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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

func WithGrpcConn(addr string, fn func(conn *grpc.ClientConn) error) error {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()
	return fn(conn)
}
