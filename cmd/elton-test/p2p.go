package main

import (
	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/events/p2p"
	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type ControllerService struct {
	L *zap.SugaredLogger

	m    p2p.P2PEventManager
	addr net.Addr
}

func (s *ControllerService) String() string {
	return "<Service: " + s.Name() + ">"
}
func (s *ControllerService) Name() string {
	return "P2PEventManager"
}
func (s *ControllerService) SubsystemType() SubsystemType {
	return UnknownSubsystemType
}
func (s *ControllerService) ServiceType() ServiceType {
	return UnknownServiceType
}
func (s *ControllerService) Serve(info *ServerInfo) error {
	s.m.L = s.L
	return WithGrpcServer(info.Ctx, func(srv *grpc.Server) error {
		proto2.RegisterEventManagerServer(srv, &s.m)
		return srv.Serve(info.Listener)
	})
}
func (s *ControllerService) Created(info *ServerInfo) error { return nil }
func (s *ControllerService) Running(info *ServerInfo) error { return nil } // TODO: Register this service.
func (s *ControllerService) Prestop(info *ServerInfo) error { return nil } // TODO: Unregister this service.
func (s *ControllerService) Stopped(info *ServerInfo) error { return nil }
