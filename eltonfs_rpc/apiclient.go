package eltonfs_rpc

import (
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"strconv"
)

type apiClient struct{}

func (apiClient) CommitService() (elton_v2.CommitServiceClient, error) {
	server := "localhost:" + strconv.Itoa(subsystems.ControllerPort)
	cc, err := grpc.Dial(server, nil)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}
	return elton_v2.NewCommitServiceClient(cc), nil
}
func (apiClient) StorageService() (elton_v2.StorageServiceClient, error) {
	server := "localhost:" + strconv.Itoa(subsystems.StoragePort)
	cc, err := grpc.Dial(server, nil)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}
	return elton_v2.NewStorageServiceClient(cc), nil
}
