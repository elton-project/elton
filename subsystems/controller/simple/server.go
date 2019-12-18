package simple

import (
	"context"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"io/ioutil"
	"net"
	"os"
	"strconv"
)

type Server struct {
	ListenAddr string
	Listener   net.Listener
	// Path to database directory.
	DatabaseAddr string
}

func (s *Server) Name() string {
	return "simple-controller"
}
func (s *Server) Configure() error {
	// Do nothing.
	return nil
}
func (s *Server) Listen() error {
	if s.Listener == nil {
		l, err := net.Listen("tcp", s.ListenAddr)
		if err != nil {
			return err
		}
		s.Listener = l
	}
	return nil
}
func (s *Server) SetListener(l net.Listener) {
	s.Listener = l
}
func (s *Server) Serve(ctx context.Context) error {
	if s.DatabaseAddr == "" {
		dir, err := ioutil.TempDir("", "")
		if err != nil {
			return xerrors.Errorf("failed to create tempdir: %w", err)
		}
		defer os.RemoveAll(dir)
		s.DatabaseAddr = dir
	}
	handler, dbClose := NewController(s.DatabaseAddr)
	defer dbClose()

	srv := grpc.NewServer()
	elton_v2.RegisterMetaServiceServer(srv, handler)
	elton_v2.RegisterNodeServiceServer(srv, handler)
	elton_v2.RegisterVolumeServiceServer(srv, handler)
	elton_v2.RegisterCommitServiceServer(srv, handler)

	return utils.GrpcServeWithCtx(srv, ctx, s.Listener)
}

func NewServer() *Server {
	return &Server{
		ListenAddr: "0.0.0.0:" + strconv.Itoa(subsystems.ControllerPort),
	}
}
