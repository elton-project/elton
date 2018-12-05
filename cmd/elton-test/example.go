package main

import (
	"context"
	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type ExampleSubsystem struct {
	EMAddr net.Addr

	m      ServiceManager
	ds, ls *grpc.Server
}

func (s *ExampleSubsystem) Name() string {
	return "<Subsystem: Example>"
}
func (s *ExampleSubsystem) Setup(ctx context.Context) error {
	s.m.Services = []Service{
		&ExampleDelivererService{
			Subsystem: s,
		},
	}
	return s.m.Setup(ctx)
}
func (s *ExampleSubsystem) Serve(ctx context.Context) []error {
	zap.S().Debugw("subsystem.Serve", "status", "serving")
	return s.m.Serve(ctx)
}

type ExampleDelivererService struct {
	Subsystem *ExampleSubsystem

	addr   string
	server ExampleDelivererServer
}

func (s *ExampleDelivererService) Name() string {
	return "<Service: Example/Deliverer>"
}
func (s *ExampleDelivererService) SetAddr(addr string) {
	s.addr = addr
}
func (s *ExampleDelivererService) Register(ctx context.Context) error {
	return WithGrpcConn(s.addr, func(conn *grpc.ClientConn) error {
		c := proto2.NewEventManagerClient(conn)
		c.ListenStatusChanges(ctx, &proto2.EventDelivererInfo{
			ServerInfo: &proto2.ServerInfo{
				// TODO: 型名 (Node) というのが不適切。
				// Serverにして、そこにアドレス、ポート番号、提供するサービス(1個)やグループなど。
			},
		})
		return nil
	})
}
func (s *ExampleDelivererService) Unregister(ctx context.Context) error {
	// todo
	return WithGrpcConn(s.addr, func(conn *grpc.ClientConn) error {
		return nil
	})
}
func (s *ExampleDelivererService) Serve(ctx context.Context, listener net.Listener) error {
	return WithGrpcServer(ctx, func(srv *grpc.Server) error {
		proto2.RegisterEventDelivererServer(srv, &s.server)
		return srv.Serve(listener)
	})
}

type ExampleDelivererServer struct {
}

func (s *ExampleDelivererServer) OnListenChanged(ctx context.Context, info *proto2.AllEventListenerInfo) (*proto2.Empty, error) {
	// todo
	panic("not implemented")
}
