package p2p

import (
	. "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type ControllerService struct {
	L *zap.SugaredLogger

	m    Controller // TODO: 足りないメソッドを実装。
	addr net.Addr
}

func (s *ControllerService) String() string {
	return "<Service: " + s.Name() + ">"
}
func (s *ControllerService) Name() string {
	return subsystemName + "/Controller"
}
func (s *ControllerService) SubsystemType() SubsystemType {
	return SubsystemType_ControllerSubsystemType
}
func (s *ControllerService) ServiceType() ServiceType {
	// TODO: どのような値を返せば良いのか
	return ServiceType_UnknownServiceType
}
func (s *ControllerService) Serve(config *ServerConfig) error {
	s.m.L = s.L
	return WithGrpcServer(config.Ctx, func(srv *grpc.Server) error {
		RegisterControllerServiceServer(srv, &s.m)
		return srv.Serve(config.Listener)
	})
}
func (s *ControllerService) Created(config *ServerConfig) error { return nil }
func (s *ControllerService) Running(config *ServerConfig) error { return nil } // TODO: Register this service.
func (s *ControllerService) Prestop(config *ServerConfig) error { return nil } // TODO: Unregister this service.
func (s *ControllerService) Stopped(config *ServerConfig) error { return nil }
