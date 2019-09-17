package simple

import (
	"context"
	"github.com/stretchr/testify/assert"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"google.golang.org/grpc"
	"testing"
)

func TestLocalMetaServer_GetMeta(t *testing.T) {
	utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
		client := elton_v2.NewMetaServiceClient(dial())
		res, err := client.GetMeta(ctx, &elton_v2.GetMetaRequest{
			Key: &elton_v2.PropertyKey{
				Id: "foo",
			},
		})
		assert.Error(t, err)
		assert.Nil(t, res)
	})

}
