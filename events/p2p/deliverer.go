package p2p

import (
	"fmt"
	pb "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"io"
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

func (ed *P2PEventDeliverer) OnListenChanged(stream pb.EventDeliverer_OnListenChangedServer) error {
	ed.lock.Lock()
	defer ed.lock.Unlock()
	ed.init()

	for {
		l, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		ed.ls[l.Type] = append(ed.ls[l.Type], l)
	}
	return nil
}

func (ed *P2PEventDeliverer) Send(eventType pb.EventType) {
	ed.lock.Lock()
	defer ed.lock.Unlock()
	ed.init()

	for _, l := range ed.ls[eventType] {
		fmt.Printf("Send %s to %s\n", eventType.String(), l.Node.Address)
	}
}
