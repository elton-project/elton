package elton_v2

import (
	"context"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

const ControllerConfigFile = "/etc/elton/controller-node"
const StorageConfigFile = "/etc/elton/storage-node"
const DefaultAPITimeout = 3 * time.Second

var controllerURI string
var storageURI string

func init() {
	controllerURI = readFileOrDefault(ControllerConfigFile, "localhost:"+strconv.Itoa(subsystems.ControllerPort))
	storageURI = readFileOrDefault(StorageConfigFile, "localhost:"+strconv.Itoa(subsystems.StoragePort))
}

func readFileOrDefault(file string, defaultData string) string {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			log.Printf("%s: not found.  Using default settings", file)
			return defaultData
		}
		log.Fatal(xerrors.Errorf("%s: %w", file, err))
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(xerrors.Errorf("%s: %w", file, err))
	}
	return string(data)
}

type ApiClient struct{}

func (ApiClient) dial(address string) (*grpc.ClientConn, error) {
	ctx, _ := context.WithTimeout(context.Background(), DefaultAPITimeout)
	return grpc.DialContext(ctx, address, grpc.WithInsecure())
}
func (ApiClient) CommitService() (CommitServiceClient, error) {
	cc, err := ApiClient{}.dial(controllerURI)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}
	return NewCommitServiceClient(cc), nil
}
func (ApiClient) StorageService() (StorageServiceClient, error) {
	cc, err := ApiClient{}.dial(storageURI)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}
	return NewStorageServiceClient(cc), nil
}

func (ApiClient) VolumeService() (VolumeServiceClient, error) {
	cc, err := ApiClient{}.dial(controllerURI)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}
	return NewVolumeServiceClient(cc), nil
}
