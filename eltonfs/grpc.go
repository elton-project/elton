package eltonfs

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"log"

	"golang.org/x/net/context"

	pb "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto"
	elton "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/server"
	"google.golang.org/grpc"
)

const bufferSize int = 64 * 1024

type EltonFSGrpcServer struct {
	Opts *Options
	FS   *elton.FileSystem

	Server *grpc.Server
}

func NewEltonFSGrpcServer(opts *Options) *EltonFSGrpcServer {
	return &EltonFSGrpcServer{Opts: opts, FS: elton.NewFileSystem(opts.LowerDir, false)}
}

func (e *EltonFSGrpcServer) Serve() error {
	if e.Opts.StandAlone {
		return nil
	}

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
	e.FS.PurgeTimer.Stop()

	if e.Opts.StandAlone {
		return
	}

	e.Server.Stop()
}

func (e *EltonFSGrpcServer) GetObject(o *pb.ObjectInfo, stream pb.EltonService_GetObjectServer) error {
	// EltonFSGrpcServerが利用されていないことを確認するためのコード。
	// もし利用されていなたら、このコードは除去してください。
	log.Fatal("GetObjectが呼び出された！！！！")

	fp, err := e.FS.Open(o.ObjectId, o.Version)
	if err != nil {
		return err
	}
	defer fp.Close()

	reader := bufio.NewReaderSize(fp, bufferSize)
	buf := make([]byte, bufferSize)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if err = stream.Send(
			&pb.Object{
				ObjectId: o.ObjectId,
				Version:  o.Version,
				Body:     buf[:n],
			},
		); err != nil {
			return err
		}
	}

	return nil
}

func (e *EltonFSGrpcServer) GenerateObjectInfo(o *pb.ObjectInfo, stream pb.EltonService_GenerateObjectInfoServer) error {
	return nil
}

func (e *EltonFSGrpcServer) CommitObjectInfo(o *pb.ObjectInfo, stream pb.EltonService_CommitObjectInfoServer) error {
	return nil
}

func (e *EltonFSGrpcServer) PutObject(c context.Context, o *pb.Object) (r *pb.EmptyMessage, err error) {
	return
}

func (e *EltonFSGrpcServer) DeleteObject(o *pb.ObjectInfo, stream pb.EltonService_DeleteObjectServer) error {
	return nil
}
