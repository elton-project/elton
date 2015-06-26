package grpc

import (
	"io"
	"testing"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	pb "./proto"
)

func TestGenerateObjectID(t *testing.T) {
	conn, _ := grpc.Dial("localhost:12345")
	defer conn.Close()

	client := pb.NewEltonServiceClient(conn)
	stream, _ := client.GenerateObjectID(context.Background(), &pb.ObjectName{Names: []string{"a.txt", "b.txt"}})

	array := []*pb.ObjectInfo{}
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}

		array = append(array, obj)
	}

	if len(array) != 2 {
		t.Fatalf("Expected length: 2, Got: %d", len(array))
	}
}

func TestCreateObjectInfo(t *testing.T) {
	conn, _ := grpc.Dial("localhost:12345")
	defer conn.Close()

	client := pb.NewEltonServiceClient(conn)
	stream, _ := client.GenerateObjectID(context.Background(), &pb.ObjectName{Names: []string{"a.txt", "b.txt"}})

	array := []*pb.ObjectInfo{}
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}

		array = append(array, obj)
	}

	if len(array) != 2 {
		t.Fatalf("Expected length: 2, Got: %d", len(array))
	}
}
