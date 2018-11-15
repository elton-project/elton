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
	ls unsafeListenerStore
	// delivererのリスト。
	ds unsafeDelivererStore
}

func (em *P2PEventManager) Listen(ctx context.Context, info *pb.EventListenerInfo) (*pb.ListenResult, error) {
	em.lock.Lock()
	defer em.lock.Unlock()

	em.ls.Add(info)
	return &pb.ListenResult{}, nil
}

func (em *P2PEventManager) Unlisten(ctx context.Context, info *pb.EventListenerInfo) (*pb.UnlistenResult, error) {
	em.lock.Lock()
	defer em.lock.Unlock()

	em.ls.Remove(info)
	return &pb.UnlistenResult{}, nil
}

func (em *P2PEventManager) ListenStatusChanges(ctx context.Context, info *pb.EventDelivererInfo) (*pb.ListenStatusChangesResult, error) {
	em.lock.Lock()
	defer em.lock.Unlock()

	em.ds.Add(info)
	return &pb.ListenStatusChangesResult{}, nil
}

func (em *P2PEventManager) UnlistenStatusChanges(ctx context.Context, info *pb.EventDelivererInfo) (*pb.UnlistenStatusChangesResult, error) {
	em.lock.Lock()
	defer em.lock.Unlock()

	em.ds.Remove(info)
	return &pb.UnlistenStatusChangesResult{}, nil
}
