package p2p

import (
	"context"
	"fmt"
	pb "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"sync"
)

// An implementation the EventManagerServer and EventSender interface.
type P2PEventDeliverer struct {
	lock sync.Mutex
	ls   unsafeListenerStore
}

func (ed *P2PEventDeliverer) OnListenChanged(ctx context.Context, info *pb.AllEventListenerInfo) (*pb.Empty, error) {
	ed.lock.Lock()
	defer ed.lock.Unlock()

	ed.ls.Clear()
	for _, l := range info.Nodes {
		ed.ls.Add(l)
	}
	return &pb.Empty{}, nil
}

func (ed *P2PEventDeliverer) Send(eventType pb.EventType) {
	ed.lock.Lock()
	defer ed.lock.Unlock()

	ed.ls.Foreach(eventType, func(info *pb.EventListenerInfo) error {
		fmt.Printf("Send %s to %s\n", eventType.String(), info.Node.Address)
		return nil
	})
}
