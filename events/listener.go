package events

import (
	"context"
	pb "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
)

type NullEventListener struct{}

func (el *NullEventListener) OnNodeAdded(ctx context.Context, node *pb.Node) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}
func (el *NullEventListener) OnNodeStopping(context.Context, *pb.Node) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}
func (el *NullEventListener) OnNodeStopped(context.Context, *pb.Node) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}
func (el *NullEventListener) OnNodeDetaching(context.Context, *pb.Node) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}
func (el *NullEventListener) OnNodeDetached(context.Context, *pb.Node) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}
func (el *NullEventListener) OnObjectCreated(context.Context, *pb.ObjectInfo) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}
func (el *NullEventListener) OnObjectDeleted(context.Context, *pb.ObjectInfo) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}
func (el *NullEventListener) OnObjectCacheCreated(context.Context, *pb.ObjectInfo) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}
func (el *NullEventListener) OnObjectCacheDeleted(context.Context, *pb.ObjectInfo) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}
