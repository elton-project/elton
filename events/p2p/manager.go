package p2p

import (
	"context"
	pb "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"go.uber.org/zap"
	"sync"
)

type NodeID uint64

// An implementation the EventManagerServer interface.
type P2PEventManager struct {
	L *zap.SugaredLogger

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

	em.L.Debugw("Listen", "args", info)
	em.ls.Add(info)
	em.notifyListenChanged()
	return &pb.ListenResult{}, nil
}
func (em *P2PEventManager) Unlisten(ctx context.Context, info *pb.EventListenerInfo) (*pb.UnlistenResult, error) {
	em.lock.Lock()
	defer em.lock.Unlock()

	em.L.Debugw("Unlisten", "args", info)
	em.ls.Remove(info)
	em.notifyListenChanged()
	return &pb.UnlistenResult{}, nil
}
func (em *P2PEventManager) ListenStatusChanges(ctx context.Context, info *pb.EventDelivererInfo) (*pb.ListenStatusChangesResult, error) {
	em.lock.Lock()
	defer em.lock.Unlock()

	em.L.Debugw("ListenStatusChanges", "args", info)
	em.ds.Add(info)
	return &pb.ListenStatusChangesResult{}, nil
}
func (em *P2PEventManager) UnlistenStatusChanges(ctx context.Context, info *pb.EventDelivererInfo) (*pb.UnlistenStatusChangesResult, error) {
	em.lock.Lock()
	defer em.lock.Unlock()

	em.L.Debugw("UnlistenStatusChanges", "args", info)
	em.ds.Remove(info)
	return &pb.UnlistenStatusChangesResult{}, nil
}
func (em *P2PEventManager) notifyListenChanged() {
	em.ds.Foreach(func(info *pb.EventDelivererInfo) error {
		// TODO: notify to other notes
		em.L.Debugw("notifyListenChanged", "to", info.Node)
		return nil
	})
}
