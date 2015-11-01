package grpc

import (
	"io"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "git.t-lab.cs.teu.ac.jp/nashio/elton/grpc/proto"
)

var array []*pb.ObjectInfo
var conn1, conn2 *grpc.ClientConn

func init() {
	conn1, _ = grpc.Dial("localhost:12345")
	conn2, _ = grpc.Dial("localhost:23456")
}

func TestGenerateObjectID(t *testing.T) {
	requestGenerateObjectID(conn1)
	requestGenerateObjectID(conn2)

	t.Logf("%v", array)
	if len(array) != 16 {
		t.Fatalf("Expected length: 16, Got: %d", len(array))
	}
}

func requestGenerateObjectID(conn *grpc.ClientConn) {
	client := pb.NewEltonServiceClient(conn)
	stream, _ := client.GenerateObjectID(context.Background(), &pb.ObjectName{Names: []string{"a.txt", "b.txt", "a.txt", "b.txt", "a.txt", "b.txt", "a.txt", "b.txt"}})

	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}

		array = append(array, obj)
	}
}

func TestCreateObjectInfo(t *testing.T) {
	defer conn1.Close()
	defer conn2.Close()

	client := pb.NewEltonServiceClient(conn1)
	for _, o := range array {
		stream, _ := client.CreateObjectInfo(context.Background(), o)
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}

		t.Logf("%v", obj)
		if obj.Version == 0 {
			t.Fatalf("Expected new version, Got: %d", obj.Version)
		}
	}
}
