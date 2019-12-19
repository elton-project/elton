package elton_v2

import (
	"context"
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
func (ApiClient) CommitService() (CommitServiceClient, error) {
	cc, err := ApiClient{}.dial(subsystems.ControllerPort)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}
	return NewCommitServiceClient(cc), nil
}
func (ApiClient) StorageService() (StorageServiceClient, error) {
	cc, err := ApiClient{}.dial(subsystems.StoragePort)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}
	return NewStorageServiceClient(cc), nil
}

func (ApiClient) VolumeService() (VolumeServiceClient, error) {
	cc, err := ApiClient{}.dial(subsystems.ControllerPort)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}
	return NewVolumeServiceClient(cc), nil
}
