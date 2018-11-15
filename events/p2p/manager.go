package p2p

import (
	"context"
	pb "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"sync"
)

type NodeID uint64

// An implementation the EventManagerServer interface.
type P2PEventManager struct {
	lock sync.RWMutex
	// listenerのマップ。
	// イベントが発生したときの送信先になる。
	ls map[pb.EventType]map[NodeID]*pb.EventListenerInfo
	// delivererのリスト。
	ds map[NodeID]*pb.EventDelivererInfo
}

func (em *P2PEventManager) init() {
	if em.ls == nil {
		em.ls = map[pb.EventType]map[NodeID]*pb.EventListenerInfo{}
	}
	if em.ds == nil {
		em.ds = map[NodeID]*pb.EventDelivererInfo{}
	}
}

func (em *P2PEventManager) Listen(ctx context.Context, info *pb.EventListenerInfo) (*pb.ListenResult, error) {
	em.lock.Lock()
	defer em.lock.Unlock()
	em.init()

	if em.ls[info.Type] == nil {
		em.ls[info.Type] = map[NodeID]*pb.EventListenerInfo{}
	}
	em.ls[info.Type][NodeID(info.Node.Id)] = info
	return &pb.ListenResult{}, nil
}

func (em *P2PEventManager) Unlisten(ctx context.Context, info *pb.EventListenerInfo) (*pb.UnlistenResult, error) {
	em.lock.Lock()
	defer em.lock.Unlock()
	em.init()

	delete(em.ls[info.Type], NodeID(info.Node.Id))
	return &pb.UnlistenResult{}, nil
}

func (em *P2PEventManager) ListenStatusChanges(ctx context.Context, info *pb.EventDelivererInfo) (*pb.ListenStatusChangesResult, error) {
	em.lock.Lock()
	defer em.lock.Unlock()
	em.init()

	em.ds[NodeID(info.Node.Id)] = info
	return &pb.ListenStatusChangesResult{}, nil
}

func (em *P2PEventManager) UnlistenStatusChanges(ctx context.Context, info *pb.EventDelivererInfo) (*pb.UnlistenStatusChangesResult, error) {
	em.lock.Lock()
	defer em.lock.Unlock()
	em.init()

	delete(em.ds, NodeID(info.Node.Id))
	return &pb.UnlistenStatusChangesResult{}, nil
}
