package simple

import (
	"context"
	"errors"
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	controller_db "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems/controller/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

func newLocalNodeServer(ns controller_db.NodeStore) *localNodeServer {
	return &localNodeServer{
		ns: ns,
	}
}

type localNodeServer struct {
	ns controller_db.NodeStore
}

func (n *localNodeServer) RegisterNode(ctx context.Context, req *RegisterNodeRequest) (*RegisterNodeResponse, error) {
	err := n.ns.Register(req.GetId(), req.GetNode())
	if errors.Is(err, controller_db.ErrNodeAlreadyExists) {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	if err != nil {
		log.Printf("[CRITICAL] Missing error handling: %+v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &RegisterNodeResponse{}, nil
}
func (n *localNodeServer) UnregisterNode(ctx context.Context, req *UnregisterNodeRequest) (*UnregisterNodeResponse, error) {
	err := n.ns.Unregister(req.GetId())
	if errors.Is(err, controller_db.ErrNotFoundNode) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		log.Printf("[CRITICAL] Missing error handling: %+v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &UnregisterNodeResponse{}, nil
}
func (n *localNodeServer) Ping(ctx context.Context, req *PingNodeRequest) (*PingNodeResponse, error) {
	err := n.ns.Update(req.GetId(), func(node *Node) error {
		// TODO: update uptime.
		return nil
	})
	if errors.Is(err, controller_db.ErrNotFoundNode) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		log.Printf("[CRITICAL] Missing error handling: %+v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &PingNodeResponse{}, nil
}
func (n *localNodeServer) ListNodes(req *ListNodesRequest, stream NodeService_ListNodesServer) error {
	var breakLoop = errors.New("break loop")
	err := n.ns.List(func(id *NodeID, node *Node) error {
		select {
		case <-stream.Context().Done():
			return breakLoop
		default:
			return stream.Send(&ListNodesResponse{
				Id:   id,
				Node: node,
			})
		}
	})
	if err != nil {
		if err == breakLoop {
			return status.Errorf(codes.Aborted, "interrupted")
		}
		log.Printf("[CRITICAL] Missing error handling: %+v", err)
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}
