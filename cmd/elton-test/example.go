package main

import (
	"context"
	"github.com/pkg/errors"
	. "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
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
	return SubsystemType_UnknownSubsystemType
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

	server ExampleDelivererServer
}

func (s *ExampleDelivererService) String() string {
	return "<Service: " + s.Name() + ">"
}
func (s *ExampleDelivererService) Name() string {
	return "Example/Deliverer"
}
func (s *ExampleDelivererService) SubsystemType() SubsystemType {
	return SubsystemType_UnknownSubsystemType
}
func (s *ExampleDelivererService) ServiceType() ServiceType {
	return ServiceType_DelivererServiceType
}
func (s *ExampleDelivererService) Serve(config *ServerConfig) error {
	return WithGrpcServer(config.Ctx, func(srv *grpc.Server) error {
		RegisterEventDelivererServer(srv, &s.server)
		return srv.Serve(config.Listener)
	})
}
func (s *ExampleDelivererService) Created(config *ServerConfig) error {
	return nil
}
func (s *ExampleDelivererService) Running(config *ServerConfig) error {
	return ConnectController(config.Ctx, config.Discoverer, func(c ControllerServiceClient) error {
		result, err := c.ListenStatusChanges(config.Ctx, &EventDelivererInfo{
			ServerInfo: &config.ServerInfo,
		})
		// TODO: resultの中身を見て、エラーチェックする
		_ = result
		return errors.Wrapf(err, "%s.Running()", s.Name())
	})
}
func (s *ExampleDelivererService) Prestop(config *ServerConfig) error {
	return ConnectController(config.Ctx, config.Discoverer, func(c ControllerServiceClient) error {
		result, err := c.UnlistenStatusChanges(config.Ctx, &EventDelivererInfo{
			ServerInfo: &config.ServerInfo,
		})
		// TODO: resultの中身を見て、エラーチェックする
		_ = result
		// TODO: システム停止時に、controllerが先に終了してしまうとここでエラーが発生する。仕方がないので、当面はこのエラーは握りつぶす。
		_ = err
		return nil
		//return errors.Wrapf(err, "%s.Prestop()", s.Name())
	})
}
func (s *ExampleDelivererService) Stopped(config *ServerConfig) error {
	return nil
}

type ExampleDelivererServer struct {
}

func (s *ExampleDelivererServer) OnListenChanged(ctx context.Context, info *AllEventListenerInfo) (*Empty, error) {
	// todo
	panic("not implemented")
}
