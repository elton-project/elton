package eltonfs_rpc

import (
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"strconv"
)

type apiClient struct{}

func (apiClient) dial(port int) (*grpc.ClientConn, error) {
	address := "localhost:" + strconv.Itoa(port)
	return grpc.Dial(address, grpc.WithInsecure())
}
func (apiClient) CommitService() (elton_v2.CommitServiceClient, error) {
	cc, err := apiClient{}.dial(subsystems.ControllerPort)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}
	return elton_v2.NewCommitServiceClient(cc), nil
}
func (apiClient) StorageService() (elton_v2.StorageServiceClient, error) {
	cc, err := apiClient{}.dial(subsystems.StoragePort)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}
	return elton_v2.NewStorageServiceClient(cc), nil
}
