package p2p

import (
	"context"
	pb "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"go.uber.org/zap"
	"sync"
)

// An implementation the EventManagerServer and EventSender interface.
type P2PEventDeliverer struct {
	L *zap.SugaredLogger

	lock sync.Mutex
	ls   unsafeListenerStore
}

func (ed *P2PEventDeliverer) OnListenChanged(ctx context.Context, info *pb.AllEventListenerInfo) (*pb.Empty, error) {
	ed.lock.Lock()
	defer ed.lock.Unlock()

	ed.L.Debugw("OnListenChanged", "args", info)
	ed.ls.Clear()
	for _, l := range info.Nodes {
		ed.ls.Add(l)
	}
	return &pb.Empty{}, nil
}

func (ed *P2PEventDeliverer) send(eventType pb.EventType) {
	ed.lock.Lock()
	defer ed.lock.Unlock()

	ed.ls.Foreach(eventType, func(info *pb.EventListenerInfo) error {
		ed.L.Debugw("Send", "eventType", eventType, "to", info.Node)
		return nil
	})
}
