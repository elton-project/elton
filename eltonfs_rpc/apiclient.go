package eltonfs_rpc

import (
	"context"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"strconv"
	"time"
)

const DefaultAPITimeout = 3 * time.Second

type ApiClient struct{}

func (ApiClient) dial(port int) (*grpc.ClientConn, error) {
	address := "localhost:" + strconv.Itoa(port)
	ctx, _ := context.WithTimeout(context.Background(), DefaultAPITimeout)
	return grpc.DialContext(ctx, address, grpc.WithInsecure())
}
func (ApiClient) CommitService() (elton_v2.CommitServiceClient, error) {
	cc, err := ApiClient{}.dial(subsystems.ControllerPort)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}
	return elton_v2.NewCommitServiceClient(cc), nil
}
func (ApiClient) StorageService() (elton_v2.StorageServiceClient, error) {
	cc, err := ApiClient{}.dial(subsystems.StoragePort)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}
	return elton_v2.NewStorageServiceClient(cc), nil
}
