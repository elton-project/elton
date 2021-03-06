package simple

import (
	"context"
	"github.com/stretchr/testify/assert"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"testing"
)

func TestLocalNodeServer_RegisterNode(t *testing.T) {
	t.Run("first_registration", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewNodeServiceClient(dial())
			res, err := client.RegisterNode(ctx, &elton_v2.RegisterNodeRequest{
				Id:   &elton_v2.NodeID{Id: "node-1"},
				Node: &elton_v2.Node{Name: "foo"},
			})
			assert.NoError(t, err)
			assert.IsType(t, &elton_v2.RegisterNodeResponse{}, res)
		})
	})
	t.Run("second_time_registration", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewNodeServiceClient(dial())
			res, err := client.RegisterNode(ctx, &elton_v2.RegisterNodeRequest{
				Id:   &elton_v2.NodeID{Id: "node-1"},
				Node: &elton_v2.Node{Name: "foo"},
			})
			assert.NoError(t, err)
			assert.IsType(t, &elton_v2.RegisterNodeResponse{}, res)

			res, err = client.RegisterNode(ctx, &elton_v2.RegisterNodeRequest{
				Id:   &elton_v2.NodeID{Id: "node-1"},
				Node: &elton_v2.Node{Name: "bar"},
			})
			assert.Equal(t, codes.AlreadyExists, status.Code(err))
			assert.Contains(t, status.Convert(err).Message(), "node already exists: ")
			assert.Nil(t, res)
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
			assert.Equal(t, codes.NotFound, status.Code(err))
			assert.Contains(t, status.Convert(err).Message(), "not found node: ")
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

			allNodes := map[string]bool{
				"node-1": false,
				"node-2": false,
				"node-3": false,
			}

			// Register 3 nodes.
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
			_, err = client.RegisterNode(ctx, &elton_v2.RegisterNodeRequest{
				Id:   &elton_v2.NodeID{Id: "node-3"},
				Node: &elton_v2.Node{Name: "node-3", Address: []string{"node3.local"}},
			})
			if !assert.NoError(t, err) {
				return
			}

			// Receive nodes list.
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

				// Check res.Node field
				name := res.GetNode().GetName()
				t.Logf("Node.Name: %s", name)
				_, ok := allNodes[name]
				if !assert.True(t, ok, "unexpected node name") {
					return
				}
				if !assert.False(t, allNodes[name], "same node appeared") {
					return
				}
				allNodes[name] = true
			}

			// Check if all nodes are appeared.
			assert.Equal(t, map[string]bool{
				"node-1": true,
				"node-2": true,
				"node-3": true,
			}, allNodes)
		})
	})
}
