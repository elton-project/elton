package simple

import (
	"context"
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	controller_db "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems/controller/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

func newLocalNodeServer(ns controller_db.NodeStore) *localNodeServer {
	// TODO: nsを渡す
	return &localNodeServer{
		nodes: map[nodeKey]*nodeInfo{},
	}
}

type localNodeServer struct {
	lock  sync.RWMutex
	nodes map[nodeKey]*nodeInfo
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
	key := newNodeKey(req.GetId())
	info := newNodeInfo(req.GetNode())
	info.lastPing = time.Now()

	n.lock.Lock()
	defer n.lock.Unlock()

	n.nodes[key] = info
	return &RegisterNodeResponse{}, nil
}
func (n *localNodeServer) UnregisterNode(ctx context.Context, req *UnregisterNodeRequest) (*UnregisterNodeResponse, error) {
	key := newNodeKey(req.GetId())

	n.lock.Lock()
	defer n.lock.Unlock()

	delete(n.nodes, key)
	return &UnregisterNodeResponse{}, nil
}
func (n *localNodeServer) Ping(ctx context.Context, req *PingNodeRequest) (*PingNodeResponse, error) {
	key := newNodeKey(req.GetId())
	n.lock.Lock()
	defer n.lock.Unlock()

	info := n.nodes[key]
	if info == nil {
		return nil, status.Errorf(codes.NotFound, "node is not registered")
	}
	info.lastPing = time.Now()
	return &PingNodeResponse{}, nil
}
func (n *localNodeServer) ListNodes(req *ListNodesRequest, stream NodeService_ListNodesServer) error {
	n.lock.RLock()
	defer n.lock.RUnlock()

	for key, info := range n.nodes {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			res := &ListNodesResponse{
				Id: &NodeID{
					Id: key.NodeId,
				},
				Node: &Node{
					Address: info.Address,
					Name:    info.Name,
					Uptime:  0, // todo
				},
			}
			if err := stream.Send(res); err != nil {
				return err
			}
		}
	}
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
