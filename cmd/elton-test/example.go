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

	ds, ls *grpc.Server
}

func (s *ExampleSubsystem) String() string {
	return "<Subsystem: " + s.Name() + ">"
}
func (s *ExampleSubsystem) Name() string {
	return "Example"
}
func (s *ExampleSubsystem) SubsystemType() SubsystemType {
	return UnknownSubsystemType
}
func (s *ExampleSubsystem) Setup(ctx context.Context, manager *ServiceManager) []error {
	manager.Add(&ExampleDelivererService{
		Subsystem: s,
	})
	return manager.Setup(ctx)
}
func (s *ExampleSubsystem) Serve(ctx context.Context, manager *ServiceManager) []error {
	zap.S().Debugw("subsystem.Serve", "status", "serving")
	return manager.Serve(ctx)
}

type ExampleDelivererService struct {
	Subsystem *ExampleSubsystem

	addr   net.Addr
	server ExampleDelivererServer
}

func (s *ExampleDelivererService) String() string {
	return "<Service: " + s.Name() + ">"
}
func (s *ExampleDelivererService) Name() string {
	return "Example/Deliverer"
}
func (s *ExampleDelivererService) SubsystemType() SubsystemType {
	return UnknownSubsystemType
}
func (s *ExampleDelivererService) ServiceType() ServiceType {
	return DelivererServiceType
}
func (s *ExampleDelivererService) Serve(info *ServerInfo) error {
	return WithGrpcServer(info.Ctx, func(srv *grpc.Server) error {
		proto2.RegisterEventDelivererServer(srv, &s.server)
		return srv.Serve(info.Listener)
	})
}
func (s *ExampleDelivererService) Created(info *ServerInfo) error {
	return nil
}
func (s *ExampleDelivererService) Running(info *ServerInfo) error {
	return WithGrpcConn(s.addr, func(conn *grpc.ClientConn) error {
		c := proto2.NewEventManagerClient(conn)
		c.ListenStatusChanges(info.Ctx, &proto2.EventDelivererInfo{
			ServerInfo: proto2.NewServerInfo(s.addr),
		})
		return nil
	})
}
func (s *ExampleDelivererService) Prestop(info *ServerInfo) error {
	// todo
	return WithGrpcConn(s.addr, func(conn *grpc.ClientConn) error {
		return nil
	})
}
func (s *ExampleDelivererService) Stopped(info *ServerInfo) error {
	return nil
}

type ExampleDelivererServer struct {
}

func (s *ExampleDelivererServer) OnListenChanged(ctx context.Context, info *proto2.AllEventListenerInfo) (*proto2.Empty, error) {
	// todo
	panic("not implemented")
}
