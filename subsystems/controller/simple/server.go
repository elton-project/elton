package simple

import (
	"context"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"google.golang.org/grpc"
	"net"
)

const DefaultListenAddr = "0.0.0.0:0"

type Server struct {
	ListenAddr string

	listener net.Listener
}

func (s *Server) Name() string {
	return "simple-controller"
}
func (s *Server) Configure() error {
	// Do nothing.
	return nil
}
func (s *Server) Listen() error {
	l, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.listener = l
	return nil
}
func (s *Server) Server(ctx context.Context) error {
	handler := NewController()
	srv := grpc.NewServer(nil)
	elton_v2.RegisterMetaServiceServer(srv, handler)
	elton_v2.RegisterNodeServiceServer(srv, handler)
	elton_v2.RegisterVolumeServiceServer(srv, handler)
	elton_v2.RegisterCommitServiceServer(srv, handler)

	return utils.GrpcServeWithCtx(srv, ctx, s.listener)
}

func NewServer() *Server {
	return &Server{
		ListenAddr: DefaultListenAddr,
	}
}
