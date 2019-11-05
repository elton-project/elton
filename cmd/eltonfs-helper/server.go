package main

import (
	"context"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/eltonfs_rpc"
	"golang.org/x/xerrors"
	"net"
)

func handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	s := eltonfs_rpc.NewClientSession(conn)
	defer s.Close()
	if err := s.Setup(); err != nil {
		return xerrors.Errorf("server: %w", err)
	}
	// TODO: wait for context finished or closed.
	return nil
}
