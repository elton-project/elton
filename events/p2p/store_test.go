package p2p

import (
	"github.com/stretchr/testify/assert"
	pb "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"testing"
)

func TestUnsafeListenerStore_Add(t *testing.T) {
	t.Run("multiple_servers", func(t *testing.T) {
		s := unsafeListenerStore{}
		s.Add(&pb.EventListenerInfo{
			ServerInfo: &pb.ServerInfo{
				Guid: "srv1",
			},
			Type: pb.EventType_ET_ALL,
		})
		s.Add(&pb.EventListenerInfo{
			ServerInfo: &pb.ServerInfo{
				Guid: "srv2",
			},
			Type: pb.EventType_ET_OBJECT_CREATED,
		})

		a := assert.New(t)
		a.Equal(2, len(s.ListListeners(pb.EventType_ET_OBJECT_CREATED)))
	})
}
