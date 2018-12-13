package main

import (
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

func (s *ManagerService) String() string {
	return "<Service: " + s.Name() + ">"
}
func (s *ManagerService) Name() string {
	return "P2PEventManager"
}
func (s *ManagerService) SubsystemType() SubsystemType {
	return UnknownSubsystemType
}
func (s *ManagerService) ServiceType() ServiceType {
	return UnknownServiceType
}
func (s *ManagerService) Serve(info *ServerInfo) error {
	s.m.L = s.L
	return WithGrpcServer(info.Ctx, func(srv *grpc.Server) error {
		proto2.RegisterEventManagerServer(srv, &s.m)
		return srv.Serve(info.Listener)
	})
}
func (s *ManagerService) Created(info *ServerInfo) error { return nil }
func (s *ManagerService) Running(info *ServerInfo) error { return nil } // TODO: Register this service.
func (s *ManagerService) Prestop(info *ServerInfo) error { return nil } // TODO: Unregister this service.
func (s *ManagerService) Stopped(info *ServerInfo) error { return nil }
