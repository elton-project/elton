package elton_v2

import (
	"context"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"log"
	"math"
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

type _conn_commitServiceClient struct {
	io.Closer
	CommitServiceClient
}
type _conn_StorageServiceClient struct {
	io.Closer
	StorageServiceClient
}
type _conn_VolumeServiceClient struct {
	io.Closer
	VolumeServiceClient
}

func dial(address string) (*grpc.ClientConn, error) {
	ctx, _ := context.WithTimeout(context.Background(), DefaultAPITimeout)
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	))
	if err != nil {
		return nil, err
	}
	return conn, nil
}
func Close(closer interface{}) error {
	return closer.(io.Closer).Close()
}
func CommitService() (CommitServiceClient, error) {
	cc, err := dial(controllerURI)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}
	return &_conn_commitServiceClient{
		Closer:              cc,
		CommitServiceClient: NewCommitServiceClient(cc),
	}, nil
}
func StorageService() (StorageServiceClient, error) {
	cc, err := dial(storageURI)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}
	return &_conn_StorageServiceClient{
		Closer:               cc,
		StorageServiceClient: NewStorageServiceClient(cc),
	}, nil
}
func VolumeService() (VolumeServiceClient, error) {
	cc, err := dial(controllerURI)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}
	return &_conn_VolumeServiceClient{
		Closer:              cc,
		VolumeServiceClient: NewVolumeServiceClient(cc),
	}, nil
}
