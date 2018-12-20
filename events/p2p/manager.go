package p2p

import (
	"context"
	"errors"
	"fmt"
	pb "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"sync"
)

var stopIteration = errors.New("stop iteration")

// An implementation the EventManagerServer interface.
type Controller struct {
	L *zap.SugaredLogger

	lock sync.RWMutex
	ss   unsafeServerStore
	// listenerのマップ。
	// イベントが発生したときの送信先になる。
	ls unsafeListenerStore
	// delivererのリスト。
	ds unsafeDelivererStore
}

func (em *Controller) StopSystem(ctx context.Context, empty *pb.Empty) (*pb.Empty, error) {
	panic("Controller.StopSystem: not implemented") // TODO
}
func (em *Controller) AddServerInfo(ctx context.Context, info *pb.ServerInfo) (*pb.Empty, error) {
	em.lock.Lock()
	defer em.lock.Unlock()
	em.ss.Add(info)
	return &pb.Empty{}, nil
}
func (em *Controller) StopServerInfo(ctx context.Context, info *pb.ServerInfo) (*pb.Empty, error) {
	panic("Controller.StopServerInfo: not implemented") // TODO
}
func (em *Controller) DetachServerInfo(ctx context.Context, info *pb.ServerInfo) (*pb.Empty, error) {
	panic("Controller.DetachServerInfo: not implemented") // TODO
}
func (em *Controller) GetServerInfo(ctx context.Context, query *pb.ServerQuery) (*pb.ServerInfo, error) {
	em.lock.Lock()
	defer em.lock.Unlock()

	var info *pb.ServerInfo
	err := em.ss.Search(query, func(i *pb.ServerInfo) error {
		info = i
		return stopIteration
	})
	if err == nil {
		return nil, fmt.Errorf("not found server: query=%s", query.String())
	} else if err == stopIteration {
		return info, nil
	} else {
		return nil, err
	}
}
func (em *Controller) Listen(ctx context.Context, info *pb.EventListenerInfo) (*pb.ListenResult, error) {
	em.lock.Lock()
	defer em.lock.Unlock()

	em.L.Debugw("Listen", "args", info)
	em.ls.Add(info)
	em.notifyListenChanged(ctx, info.Type)
	return &pb.ListenResult{}, nil
}
func (em *Controller) Unlisten(ctx context.Context, info *pb.EventListenerInfo) (*pb.UnlistenResult, error) {
	em.lock.Lock()
	defer em.lock.Unlock()

	em.L.Debugw("Unlisten", "args", info)
	em.ls.Remove(info)
	em.notifyListenChanged(ctx, info.Type)
	return &pb.UnlistenResult{}, nil
}
func (em *Controller) ListenStatusChanges(ctx context.Context, info *pb.EventDelivererInfo) (*pb.ListenStatusChangesResult, error) {
	em.lock.Lock()
	defer em.lock.Unlock()

	em.L.Debugw("ListenStatusChanges", "args", info)
	em.ds.Add(info)
	return &pb.ListenStatusChangesResult{}, nil
}
func (em *Controller) UnlistenStatusChanges(ctx context.Context, info *pb.EventDelivererInfo) (*pb.UnlistenStatusChangesResult, error) {
	em.lock.Lock()
	defer em.lock.Unlock()

	em.L.Debugw("UnlistenStatusChanges", "args", info)
	em.ds.Remove(info)
	return &pb.UnlistenStatusChangesResult{}, nil
}
func (em *Controller) notifyListenChanged(ctx context.Context, eventType pb.EventType) {
	var wg sync.WaitGroup
	defer wg.Wait()

	allNodes := &pb.AllEventListenerInfo{
		Servers: em.ls.ListListeners(eventType),
		Type:    eventType,
	}

	em.ds.Foreach(func(info *pb.EventDelivererInfo) error {
		l := em.L.With("to", info.ServerInfo)

		l.Debugw("notifyListenChanged")
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := grpc.Dial(info.ServerInfo.Address, grpc.WithInsecure())
			if err != nil {
				l.Errorw("notifyListenChanged",
					"status", "failed",
					"phase", "connecting",
					"error", err.Error())
				return
			}
			defer conn.Close()

			client := pb.NewEventDelivererClient(conn)
			if _, err := client.OnListenChanged(ctx, allNodes); err != nil {
				l.Errorw("notifyListenChanged",
					"status", "failed",
					"phase", "calling",
					"error", err.Error())
			}
		}()
		return nil
	})

	wg.Wait()
}
