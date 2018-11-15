package p2p

import (
	pb "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"io"
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

func (em *P2PEventManager) Listen(stream pb.EventManager_ListenServer) error {
	em.lock.Lock()
	defer em.lock.Unlock()
	em.init()

	for {
		l, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if em.ls[l.Type] == nil {
			em.ls[l.Type] = map[NodeID]*pb.EventListenerInfo{}
		}
		em.ls[l.Type][NodeID(l.Node.Id)] = l
	}
	return nil
}

func (em *P2PEventManager) Unlisten(stream pb.EventManager_UnlistenServer) error {
	em.lock.Lock()
	defer em.lock.Unlock()
	em.init()

	for {
		u, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		delete(em.ls[u.Type], NodeID(u.Node.Id))
	}
	return nil
}

func (em *P2PEventManager) ListenStatusChanges(stream pb.EventManager_ListenStatusChangesServer) error {
	em.lock.Lock()
	defer em.lock.Unlock()
	em.init()

	for {
		l, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		em.ds[NodeID(l.Node.Id)] = l
	}
	return nil
}

func (em *P2PEventManager) UnlistenStatusChanges(stream pb.EventManager_UnlistenStatusChangesServer) error {
	em.lock.Lock()
	defer em.lock.Unlock()
	em.init()

	for {
		u, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		delete(em.ds, NodeID(u.Node.Id))
	}
	return nil
}
