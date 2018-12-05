package main

import (
	"context"
	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/events/p2p"
	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type ManagerService struct {
	L *zap.SugaredLogger

	m    p2p.P2PEventManager
	addr net.Addr
}

func (s *ManagerService) Name() string {
	return "<Service: P2PEventManager>"
}
func (s *ManagerService) SetAddr(addr net.Addr) {
	s.addr = addr
}
func (s *ManagerService) Register(ctx context.Context) error {
	return nil
}
func (s *ManagerService) Unregister(ctx context.Context) error {
	return nil
}
func (s *ManagerService) Serve(ctx context.Context, listener net.Listener) error {
	s.m.L = s.L
	return WithGrpcServer(ctx, func(srv *grpc.Server) error {
		proto2.RegisterEventManagerServer(srv, &s.m)
		return srv.Serve(listener)
	})
}
