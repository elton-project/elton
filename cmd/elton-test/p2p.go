package main

import (
	"context"
	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/events/p2p"
	. "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type ControllerSubsystem struct {
}

func (s *ControllerSubsystem) String() string {
	return "<Subsystem: " + s.Name() + ">"
}
func (s *ControllerSubsystem) Name() string {
	return "P2PControllerSubsystem"
}
func (s *ControllerSubsystem) SubsystemType() SubsystemType {
	return SubsystemType_ControllerSubsystemType
}
func (s *ControllerSubsystem) Setup(ctx context.Context, manager *ServiceManager) []error {
	manager.Add(&ControllerService{
		L: zap.S(),
	})
	// TODO: error handling
	manager.Setup(ctx)
	return nil
}
func (s *ControllerSubsystem) Serve(ctx context.Context, manager *ServiceManager) []error {
	return manager.Serve(ctx)
}

type ControllerService struct {
	L *zap.SugaredLogger

	m    p2p.P2PEventManager
	addr net.Addr
}

func (s *ControllerService) String() string {
	return "<Service: " + s.Name() + ">"
}
func (s *ControllerService) Name() string {
	return "P2PController/EventManager"
}
func (s *ControllerService) SubsystemType() SubsystemType {
	return SubsystemType_ControllerSubsystemType
}
func (s *ControllerService) ServiceType() ServiceType {
	return ServiceType_EventManagerServiceType
}
func (s *ControllerService) Serve(config *ServerConfig) error {
	s.m.L = s.L
	return WithGrpcServer(config.Ctx, func(srv *grpc.Server) error {
		RegisterEventManagerServer(srv, &s.m)
		return srv.Serve(config.Listener)
	})
}
func (s *ControllerService) Created(config *ServerConfig) error { return nil }
func (s *ControllerService) Running(config *ServerConfig) error { return nil } // TODO: Register this service.
func (s *ControllerService) Prestop(config *ServerConfig) error { return nil } // TODO: Unregister this service.
func (s *ControllerService) Stopped(config *ServerConfig) error { return nil }
