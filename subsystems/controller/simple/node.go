package simple

import (
	"context"
	"errors"
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	controller_db "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems/controller/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func newLocalNodeServer(ns controller_db.NodeStore) *localNodeServer {
	// TODO: nsを渡す
	return &localNodeServer{
		ns: ns,
	}
}

type localNodeServer struct {
	ns controller_db.NodeStore
	//lock  sync.RWMutex
	//nodes map[nodeKey]*nodeInfo
}
type nodeKey struct {
	NodeId string
}
type nodeInfo struct {
	Address []string
	Name    string

	lastPing time.Time
}

func (n *localNodeServer) RegisterNode(ctx context.Context, req *RegisterNodeRequest) (*RegisterNodeResponse, error) {
	n.ns.Register(req.GetId(), req.GetNode())
	// TODO: add error handling
	return &RegisterNodeResponse{}, nil
}
func (n *localNodeServer) UnregisterNode(ctx context.Context, req *UnregisterNodeRequest) (*UnregisterNodeResponse, error) {
	key := newNodeKey(req.GetId())
	n.ns.Unregister(req.GetId())
	// TODO: add error handling
	return &UnregisterNodeResponse{}, nil
}
func (n *localNodeServer) Ping(ctx context.Context, req *PingNodeRequest) (*PingNodeResponse, error) {
	n.ns.Update(req.GetId(), func(node *Node) error {
		// TODO: update uptime.
		return nil
	})
	//return nil, status.Errorf(codes.NotFound, "node is not registered")
	// TODO: add error handling
	return &PingNodeResponse{}, nil
}
func (n *localNodeServer) ListNodes(req *ListNodesRequest, stream NodeService_ListNodesServer) error {
	var breakLoop = errors.New("break loop")
	err := n.ns.List(func(id *NodeID, node *Node) error {
		select {
		case <-stream.Context().Done():
			return breakLoop
		default:
		}
		return nil
	})
	if err == breakLoop {
		return status.Errorf(codes.Aborted, "interrupted")
	}
	// TODO: add error handling
	return nil
}

func newNodeKey(id *NodeID) nodeKey {
	return nodeKey{
		NodeId: id.GetId(),
	}
}
func newNodeInfo(node *Node) *nodeInfo {
	return &nodeInfo{
		Address: node.GetAddress(),
		Name:    node.GetName(),
	}
}
