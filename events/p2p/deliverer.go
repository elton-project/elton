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
	ls   map[pb.EventType][]*pb.EventListenerInfo
}

func (ed *P2PEventDeliverer) init() {
	if ed.ls == nil {
		ed.ls = map[pb.EventType][]*pb.EventListenerInfo{}
	}
}

func (ed *P2PEventDeliverer) OnListenChanged(ctx context.Context, info *pb.AllEventListenerInfo) (*pb.Empty, error) {
	ed.lock.Lock()
	defer ed.lock.Unlock()
	ed.init()
	for _, l := range info.Nodes {
		ed.ls[l.Type] = append(ed.ls[l.Type], l)
	}
	return &pb.Empty{}, nil
}

func (ed *P2PEventDeliverer) Send(eventType pb.EventType) {
	ed.lock.Lock()
	defer ed.lock.Unlock()
	ed.init()

	for _, l := range ed.ls[eventType] {
		fmt.Printf("Send %s to %s\n", eventType.String(), l.Node.Address)
	}
}
