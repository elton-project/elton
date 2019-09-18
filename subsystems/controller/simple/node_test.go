package simple

import (
	"context"
	"github.com/stretchr/testify/assert"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"google.golang.org/grpc"
	"io"
	"testing"
)

func TestLocalNodeServer_RegisterNode(t *testing.T) {
	t.Run("first_registration", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewNodeServiceClient(dial())
			res, err := client.RegisterNode(ctx, &elton_v2.RegisterNodeRequest{
				Id:   &elton_v2.NodeID{Id: "node-1"},
				Node: &elton_v2.Node{},
			})
			assert.NoError(t, err)
			assert.IsType(t, &elton_v2.RegisterNodeResponse{}, res)
		})
	})
	t.Run("second_time_registration", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewNodeServiceClient(dial())
			res, err := client.RegisterNode(ctx, &elton_v2.RegisterNodeRequest{})
			assert.NoError(t, err)
			assert.IsType(t, &elton_v2.RegisterNodeResponse{}, res)
		})
	})
}

func TestLocalNodeServer_UnregisterNode(t *testing.T) {
	t.Run("normal_case", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewNodeServiceClient(dial())
			res, err := client.UnregisterNode(ctx, &elton_v2.UnregisterNodeRequest{
				Id: &elton_v2.NodeID{Id: "node-1"},
			})
			assert.NoError(t, err)
			assert.IsType(t, &elton_v2.UnregisterNodeResponse{}, res)
		})
	})
}

func TestLocalNodeServer_Ping(t *testing.T) {
	t.Run("ping_from_registered_node", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewNodeServiceClient(dial())

			rres, err := client.RegisterNode(ctx, &elton_v2.RegisterNodeRequest{
				Id:   &elton_v2.NodeID{Id: "node-1"},
				Node: &elton_v2.Node{},
			})
			if !assert.NoError(t, err) || !assert.IsType(t, &elton_v2.RegisterNodeResponse{}, rres) {
				return
			}

			pres, err := client.Ping(ctx, &elton_v2.PingNodeRequest{
				Id: &elton_v2.NodeID{Id: "node-1"},
			})
			assert.NoError(t, err)
			assert.IsType(t, &elton_v2.PingNodeResponse{}, pres)
		})
	})
	t.Run("ping_from_not_registered_node", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewNodeServiceClient(dial())
			pres, err := client.Ping(ctx, &elton_v2.PingNodeRequest{
				Id: &elton_v2.NodeID{Id: "deleted-node"},
			})
			assert.Error(t, err, "not is not registered")
			assert.Nil(t, pres)
		})
	})
}

func TestLocalNodeServer_ListNodes(t *testing.T) {
	t.Run("emtpy_list", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewNodeServiceClient(dial())
			stream, err := client.ListNodes(ctx, &elton_v2.ListNodesRequest{})
			if !assert.NoError(t, err) {
				return
			}

			_, err = stream.Recv()
			assert.EqualError(t, err, io.EOF.Error())
		})
	})

	t.Run("should_returns_all_nodes", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewNodeServiceClient(dial())

			// Register two nodes.
			_, err := client.RegisterNode(ctx, &elton_v2.RegisterNodeRequest{
				Id:   &elton_v2.NodeID{Id: "node-1"},
				Node: &elton_v2.Node{Name: "node-1", Address: []string{"node1.local"}},
			})
			if !assert.NoError(t, err) {
				return
			}
			_, err = client.RegisterNode(ctx, &elton_v2.RegisterNodeRequest{
				Id:   &elton_v2.NodeID{Id: "node-2"},
				Node: &elton_v2.Node{Name: "node-2", Address: []string{"node2.local"}},
			})
			if !assert.NoError(t, err) {
				return
			}

			stream, err := client.ListNodes(ctx, &elton_v2.ListNodesRequest{})
			if !assert.NoError(t, err) {
				return
			}
			for {
				res, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if !assert.NoError(t, err) {
					return
				}
				// TODO: check res.Node field
				_ = res.Node
				_ = res
			}
		})
	})
}
