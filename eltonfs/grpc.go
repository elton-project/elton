package main

import (
	"encoding/base64"
	"fmt"
	"net"

	"golang.org/x/net/context"

	elton "git.t-lab.cs.teu.ac.jp/nashio/elton/api"
	pb "git.t-lab.cs.teu.ac.jp/nashio/elton/grpc/proto"
	"google.golang.org/grpc"
)

type EltonFSGrpcServer struct {
	Opts *Options
	FS   *elton.FileSystem

	Server *grpc.Server
}

func NewEltonFSGrpcServer(opts *Options) (*EltonFSGrpcServer, error) {
	return &EltonFSGrpcServer{Opts: opts, FS: elton.NewFileSystem(opts.LowerDir)}, nil
}

func (e *EltonFSGrpcServer) Serve() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", e.Opts.Port))
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	pb.RegisterEltonServiceServer(server, e)

	e.Server = server
	server.Serve(lis)
	return nil
}

func (e *EltonFSGrpcServer) Stop() {
	e.Server.Stop()
}

func (e *EltonFSGrpcServer) GetObject(o *pb.ObjectInfo, stream pb.EltonService_GetObjectServer) error {
	body, err := e.FS.Read(o.ObjectId, o.Version)
	if err != nil {
		return err
	}

	if err = stream.Send(
		&pb.Object{
			ObjectId: o.ObjectId,
			Version:  o.Version,
			Body:     base64.StdEncoding.EncodeToString(body),
		},
	); err != nil {
		return err
	}

	return nil
}

func (e *EltonFSGrpcServer) GenerateObjectInfo(o *pb.ObjectInfo, stream pb.EltonService_GenerateObjectInfoServer) error {
	return nil
}

func (e *EltonFSGrpcServer) CommitObjectInfo(c context.Context, o *pb.ObjectInfo) (r *pb.EmptyMessage, err error) {
	return
}

func (e *EltonFSGrpcServer) PutObject(c context.Context, o *pb.Object) (r *pb.EmptyMessage, err error) {
	return
}

func (e *EltonFSGrpcServer) DeleteObject(c context.Context, o *pb.ObjectInfo) (r *pb.EmptyMessage, err error) {
	return
}
