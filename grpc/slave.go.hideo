package grpc

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"../elton"
	pb "./proto"
	"google.golang.org/grpc"
)

type EltonSlave struct {
	FS         elton.FileSystem
	Connection *grpc.ClientConn
	Conf       elton.Config
}

func NewEltonSlave(conf elton.Config) (*EltonSlave, error) {
	fs := elton.NewFileSystem(conf.Serve.Dir)
	conn, err := grpc.Dial(conf.Serve.MasterHostName)
	if err != nil {
		return nil, err
	}

	return &EltonSlave{FS: fs, Connection: conn, Conf: conf}
}

func (e *EltonSlave) Serve() {
	defer e.Connection.Close()

	wg := new(sync.WaitGroup)
	go func() {
		var srv http.Server

	}()

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", e.Conf.Slave.IP, e.Conf.Slave.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterEltonServiceServer(server, e)
	log.Fatal(server.Serve(lis))
}
