package main

import (
	"context"
	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/events/p2p"
	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"math/rand"
	"net"
	"os"
	"path"
	"sync"
	"time"
)

func generateTempFilePath() string {
	letters := "abcdefghijklmnopqrstuvwxyz"
	return path.Join("/tmp", "elton-"+string([]uint8{
		letters[rand.Int()%len(letters)],
		letters[rand.Int()%len(letters)],
		letters[rand.Int()%len(letters)],
		letters[rand.Int()%len(letters)],
		letters[rand.Int()%len(letters)],
		letters[rand.Int()%len(letters)],
		letters[rand.Int()%len(letters)],
		letters[rand.Int()%len(letters)],
	}))
}

func generateTempSock() (net.Listener, *grpc.ClientConn) {
	sockPath := generateTempFilePath()
	_ = os.Remove(sockPath)
	sock, err := net.Listen("unix", sockPath)
	if err != nil {
		panic(err)
	}

	conn, err := grpc.Dial("unix://"+sockPath, grpc.WithInsecure())
	checkError(err)
	return sock, conn
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func server(ctx context.Context, em, ed, el net.Listener) {
	ems := grpc.NewServer()
	proto2.RegisterEventManagerServer(ems, &p2p.P2PEventManager{
		L: zap.S().With("server", "P2PEventManager"),
	})

	eds := grpc.NewServer()
	d := &p2p.P2PEventDeliverer{
		L: zap.S().With("server", "P2PEventDeliverer"),
		Master: &proto2.Node{
			Id:      0,
			Group:   nil,
			Address: "unix://" + em.Addr().String(),
		},
		Self: &proto2.Node{
			Id:      1,
			Group:   nil,
			Address: "unix://" + ed.Addr().String(),
		},
	}
	proto2.RegisterEventDelivererServer(eds, d)

	wg := &sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		checkError(ems.Serve(em))
	}()
	go func() {
		defer wg.Done()
		checkError(eds.Serve(ed))
	}()
	go func() {
		defer wg.Done()
		defer ems.GracefulStop()
		defer eds.GracefulStop()
		defer func() {
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			checkError(d.Unregister(ctx))
		}()
		checkError(d.Register(ctx))
		<-ctx.Done()
	}()
	wg.Wait()
}

func client(ctx context.Context, em, ed, el *grpc.ClientConn) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	func() {
		emc := proto2.NewEventManagerClient(em)
		result, err := emc.Listen(ctx,
			&proto2.EventListenerInfo{
				Node: &proto2.Node{
					Id:      2,
					Group:   nil,
					Address: "unix://" + el.Target(),
				},
				Type: proto2.EventType_ET_OBJECT_CREATED,
			},
		)
		checkError(err)
		zap.S().Infow("Client/Listen", "result", result)
	}()

	func() {
		eld := proto2.NewEventDelivererClient(ed)
		result, err := eld.OnListenChanged(ctx, &proto2.AllEventListenerInfo{
			Nodes: []*proto2.EventListenerInfo{},
		})
		checkError(err)
		zap.S().Infow("Client/OnListenChanged", "result", result)
	}()

	//elc := proto2.NewEventListenerClient(el)
}

func main() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer zap.S().Sync()

	ctx, cancel := context.WithCancel(context.Background())

	ems, emc := generateTempSock()
	eds, edc := generateTempSock()
	els, elc := generateTempSock()

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		zap.S().Infow("Server", "status", "starting")
		defer wg.Done()
		server(ctx, ems, eds, els)
		zap.S().Infow("Server", "status", "stopped")
	}()
	go func() {
		zap.S().Infow("Client", "status", "starting")
		defer wg.Done()
		defer cancel()
		client(ctx, emc, edc, elc)
		zap.S().Infow("Client", "status", "stopped")
	}()
	zap.S().Infow("Main", "status", "waiting")
	wg.Wait()
	zap.S().Infow("Main", "status", "done")
}
