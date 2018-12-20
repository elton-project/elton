package p2p

import (
	"context"
	. "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"go.uber.org/zap"
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
	return manager.Setup(ctx)
}
func (s *ControllerSubsystem) Serve(ctx context.Context, manager *ServiceManager) []error {
	return manager.Serve(ctx)
}
