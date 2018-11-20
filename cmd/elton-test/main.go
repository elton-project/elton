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

func server(em, ed, el net.Listener) {
	ems := grpc.NewServer()
	proto2.RegisterEventManagerServer(ems, &p2p.P2PEventManager{
		L: zap.S().With("server", "P2PEventManager"),
	})

	eds := grpc.NewServer()
	proto2.RegisterEventDelivererServer(eds, &p2p.P2PEventDeliverer{
		L: zap.S().With("server", "P2PEventDeliverer"),
	})

	go func() {
		checkError(ems.Serve(em))
	}()
	go func() {
		checkError(eds.Serve(ed))
	}()
	time.Sleep(1 * time.Second)
}

func client(em, ed, el *grpc.ClientConn) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	func() {
		emc := proto2.NewEventManagerClient(em)
		result, err := emc.Listen(ctx,
			&proto2.EventListenerInfo{
				Node: &proto2.Node{
					Id:      1,
					Address: el.Target(),
					Group:   nil,
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

	ems, emc := generateTempSock()
	eds, edc := generateTempSock()
	els, elc := generateTempSock()
	server(ems, eds, els)
	client(emc, edc, elc)
}
