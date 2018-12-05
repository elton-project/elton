package p2p

import (
	"context"
	"fmt"
	pb "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"sync"
)

// An implementation the EventManagerServer and EventSender interface.
type P2PEventDeliverer struct {
	L      *zap.SugaredLogger
	Master *pb.ServerInfo
	Self   *pb.ServerInfo

	lock sync.Mutex
	ls   unsafeListenerStore
}

func (ed *P2PEventDeliverer) OnListenChanged(ctx context.Context, info *pb.AllEventListenerInfo) (*pb.Empty, error) {
	ed.lock.Lock()
	defer ed.lock.Unlock()

	ed.L.Debugw("OnListenChanged", "args", info)
	ed.ls.Clear()
	for _, l := range info.Servers {
		ed.ls.Add(l)
	}
	return &pb.Empty{}, nil
}

func (ed *P2PEventDeliverer) send(eventType pb.EventType) {
	ed.lock.Lock()
	defer ed.lock.Unlock()

	ed.ls.Foreach(eventType, func(info *pb.EventListenerInfo) error {
		ed.L.Debugw("Send", "eventType", eventType, "to", info.ServerInfo)
		return nil
	})
}

func (ed *P2PEventDeliverer) Register(ctx context.Context) error {
	return ed.withMasterConn(func(master pb.EventManagerClient) error {
		result, err := master.ListenStatusChanges(ctx, ed.selfInfo())
		if err != nil {
			return err
		}
		if result.Error != "" {
			return fmt.Errorf("error response: ListenStatusChanges() returns \"%s\"", result.Error)
		}
		return nil
	})
}

func (ed *P2PEventDeliverer) Unregister(ctx context.Context) error {
	return ed.withMasterConn(func(master pb.EventManagerClient) error {
		result, err := master.UnlistenStatusChanges(ctx, ed.selfInfo())
		if err != nil {
			return err
		}
		if result.Error != "" {
			return fmt.Errorf("error response: UnlistenStatusChanges() returns \"%s\"", result.Error)
		}
		return nil
	})
}

func (ed *P2PEventDeliverer) withMasterConn(fn func(master pb.EventManagerClient) error) error {
	conn, err := grpc.Dial(ed.Master.Address, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()
	return fn(pb.NewEventManagerClient(conn))
}

func (ed *P2PEventDeliverer) selfInfo() *pb.EventDelivererInfo {
	return &pb.EventDelivererInfo{
		ServerInfo: ed.Self,
	}
}
