package simple

import (
	"context"
	"github.com/stretchr/testify/assert"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"testing"
)

func TestLocalMetaServer_GetMeta(t *testing.T) {
	t.Run("should_success_if_property_is_not_exists", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewMetaServiceClient(dial())
			res, err := client.GetMeta(ctx, &elton_v2.GetMetaRequest{
				Key: &elton_v2.PropertyID{
					Id: "foo",
				},
			})
			assert.Equal(t, status.Convert(err).Message(), "not found property")
			assert.Equal(t, &elton_v2.GetMetaResponse{
				Key:  &elton_v2.PropertyID{Id: "foo"},
				Body: nil,
			}, res)
		})
	})
	t.Run("should_return_valid_body", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewMetaServiceClient(dial())
			sres, err := client.SetMeta(ctx, &elton_v2.SetMetaRequest{
				Key:  &elton_v2.PropertyID{Id: "foo"},
				Body: &elton_v2.Property{Body: "body"},
			})
			if !assert.NoError(t, err) || !assert.NotNil(t, sres) {
				return
			}

			gres, err := client.GetMeta(ctx, &elton_v2.GetMetaRequest{
				Key: &elton_v2.PropertyID{
					Id: "foo",
				},
			})
			assert.NoError(t, err)
			assert.Equal(t, &elton_v2.GetMetaResponse{
				Key:  &elton_v2.PropertyID{Id: "foo"},
				Body: &elton_v2.Property{Body: "body"},
			}, gres)
		})
	})
}

func TestLocalMetaServer_SetMeta(t *testing.T) {
	t.Run("should_fail_when_try_to_create_property", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewMetaServiceClient(dial())
			res, err := client.SetMeta(ctx, &elton_v2.SetMetaRequest{
				Key:        &elton_v2.PropertyID{Id: "foo"},
				Body:       &elton_v2.Property{Body: "version 1", AllowReplace: true},
				MustCreate: true,
			})
			if !assert.NoError(t, err) || !assert.NotNil(t, res) {
				return
			}

			res, err = client.SetMeta(ctx, &elton_v2.SetMetaRequest{
				Key:        &elton_v2.PropertyID{Id: "foo"},
				Body:       &elton_v2.Property{Body: "version 2", AllowReplace: true},
				MustCreate: true,
			})
			assert.Error(t, err, "key is already exists")
			assert.Nil(t, res)
		})
	})
	t.Run("should_fail_when_try_to_replace", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewMetaServiceClient(dial())
			res, err := client.SetMeta(ctx, &elton_v2.SetMetaRequest{
				Key:  &elton_v2.PropertyID{Id: "foo"},
				Body: &elton_v2.Property{Body: "version 1"},
			})
			if !assert.NoError(t, err) || !assert.NotNil(t, res) {
				return
			}

			res, err = client.SetMeta(ctx, &elton_v2.SetMetaRequest{
				Key:  &elton_v2.PropertyID{Id: "foo"},
				Body: &elton_v2.Property{Body: "version 2", AllowReplace: true},
			})
			assert.Error(t, err, "replacement not allowed")
			assert.Nil(t, res)
		})
	})
	t.Run("should_return_old_body", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewMetaServiceClient(dial())
			res, err := client.SetMeta(ctx, &elton_v2.SetMetaRequest{
				Key:  &elton_v2.PropertyID{Id: "foo"},
				Body: &elton_v2.Property{Body: "version 1", AllowReplace: true},
			})
			if !assert.NoError(t, err) || !assert.NotNil(t, res) {
				return
			}

			res, err = client.SetMeta(ctx, &elton_v2.SetMetaRequest{
				Key:  &elton_v2.PropertyID{Id: "foo"},
				Body: &elton_v2.Property{Body: "version 2", AllowReplace: true},
			})
			assert.NoError(t, err)
			assert.Equal(t, "version 1", res.GetOldBody().GetBody())
		})
	})
}
