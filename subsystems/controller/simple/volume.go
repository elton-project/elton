package simple

import (
	"context"
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems/idgen"
	"golang.org/x/xerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
)

func newLocalVolumeServer() *localVolumeServer {
	return &localVolumeServer{
		volumes: map[volumeKey]*volumeInfo{},
	}
}

type localVolumeServer struct {
	UnimplementedCommitServiceServer
	lock    sync.RWMutex
	volumes map[volumeKey]*volumeInfo
}
type volumeKey struct {
	Id string
}
type volumeInfo struct {
	Id   string
	Name string
}

func (v *localVolumeServer) CreateVolume(ctx context.Context, req *CreateVolumeRequest) (*CreateVolumeResponse, error) {
	key := generateVolumeKey()
	info := newVolumeInfo(req.GetInfo())

	v.lock.Lock()
	defer v.lock.Unlock()

	if v.volumes[key] != nil {
		return nil, status.Error(codes.AlreadyExists, "volume already exists")
	}
	v.volumes[key] = info
	return &CreateVolumeResponse{
		Id: &VolumeID{
			Id: key.Id,
		},
	}, nil
}
func (v *localVolumeServer) DeleteVolume(ctx context.Context, req *DeleteVolumeRequest) (*DeleteVolumeResponse, error) {
	key := newVolumeKey(req.GetId())

	v.lock.Lock()
	defer v.lock.Unlock()

	delete(v.volumes, key)
	return &DeleteVolumeResponse{}, nil
}
func (v *localVolumeServer) ListVolumes(req *ListVolumesRequest, stream VolumeService_ListVolumesServer) error {
	if req.GetNext() != "" {
		return status.Error(codes.FailedPrecondition, "next parameter is not supported") // TODO
	}
	limit := req.GetLimit()

	v.lock.RLock()
	defer v.lock.RUnlock()

	count := uint64(0)
	for _, info := range v.volumes {
		select {
		case <-stream.Context().Done():
			// Context canceled.
			return nil
		default:
			res := &ListVolumesResponse{
				Id: &VolumeID{
					Id: info.Id,
				},
				Info: &VolumeInfo{
					Name: info.Name,
				},
			}
			if err := stream.Send(res); err != nil {
				return err
			}

			count++
			if limit > 0 && count > limit {
				// Limit reached.
				return nil
			}
		}
	}
	return nil
}
func (v *localVolumeServer) InspectVolume(ctx context.Context, req *InspectVolumeRequest) (*InspectVolumeResponse, error) {
	if req.GetId() != nil && req.GetName() != "" {
		return nil, status.Error(codes.FailedPrecondition, "id and info is exclusive")
	}

	if req.GetId() != nil {
		// Search by id
		v.lock.RLock()
		defer v.lock.RUnlock()

		key := newVolumeKey(req.GetId())
		info := v.volumes[key]
		return &InspectVolumeResponse{
			Id:   key.ToID(),
			Info: info.ToInfo(),
		}, nil
	} else if req.GetName() != "" {
		// Search by name
		name := req.GetName()

		v.lock.RLock()
		defer v.lock.RUnlock()

		for _, info := range v.volumes {
			if info.Name == name {
				// Found the volume.
				return &InspectVolumeResponse{
					Id: &VolumeID{
						Id: info.Id,
					},
					Info: &VolumeInfo{
						Name: info.Name,
					},
				}, nil
			}
		}
		// Not found.
		return nil, status.Error(codes.NotFound, "volume not found")
	} else {
		panic("unreachable")
	}
}

func generateVolumeKey() volumeKey {
	s, err := idgen.Gen.NextStringID()
	if err != nil {
		panic(xerrors.Errorf("failed to generate id: %w", err))
	}
	return volumeKey{
		Id: s,
	}
}
func newVolumeKey(id *VolumeID) volumeKey {
	return volumeKey{
		Id: id.GetId(),
	}
}
func (k *volumeKey) ToID() *VolumeID {
	if k == nil {
		return nil
	}
	return &VolumeID{
		Id: k.Id,
	}
}
func newVolumeInfo(info *VolumeInfo) *volumeInfo {
	return &volumeInfo{
		Name: info.GetName(),
	}
}
func (v *volumeInfo) ToInfo() *VolumeInfo {
	if v == nil {
		return nil
	}
	return &VolumeInfo{
		Name: v.Name,
	}
}
