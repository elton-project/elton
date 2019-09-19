package simple

import (
	"context"
	"errors"
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	controller_db "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems/controller/db"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems/idgen"
	"golang.org/x/xerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"sync"
)

func newLocalVolumeServer(vs controller_db.VolumeStore, cs controller_db.CommitStore) *localVolumeServer {
	return &localVolumeServer{
		vs: vs,
		cs: cs,
	}
}

type localVolumeServer struct {
	UnimplementedCommitServiceServer
	lock sync.RWMutex
	vs   controller_db.VolumeStore
	cs   controller_db.CommitStore
}
type volumeKey struct {
	Id string
}
type volumeInfo struct {
	Id   string
	Name string
}

func (v *localVolumeServer) CreateVolume(ctx context.Context, req *CreateVolumeRequest) (*CreateVolumeResponse, error) {
	if req.GetInfo() == nil {
		return nil, status.Error(codes.InvalidArgument, "info is null")
	}

	vid, err := v.vs.Create(req.GetInfo())
	if err != nil {
		if errors.Is(err, &controller_db.InputError{}) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		log.Println("ERROR:", err)
		return nil, status.Error(codes.Internal, "database error")
	}
	return &CreateVolumeResponse{
		Id: vid,
	}, nil
}
func (v *localVolumeServer) DeleteVolume(ctx context.Context, req *DeleteVolumeRequest) (*DeleteVolumeResponse, error) {
	if req.GetId() == nil {
		return nil, status.Error(codes.InvalidArgument, "id is null")
	}

	err := v.vs.Delete(req.GetId())
	if err != nil {
		if errors.Is(err, &controller_db.InputError{}) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		log.Println("ERROR:", err)
		return nil, status.Error(codes.Internal, "database error")
	}
	return &DeleteVolumeResponse{}, nil
}
func (v *localVolumeServer) ListVolumes(req *ListVolumesRequest, stream VolumeService_ListVolumesServer) error {
	if req.GetNext() != "" {
		return status.Error(codes.FailedPrecondition, "next parameter is not supported") // TODO
	}
	limit := req.GetLimit()

	count := uint64(0)
	breakLoop := errors.New("break loop")
	err := v.vs.Walk(func(id *VolumeID, info *VolumeInfo) error {
		select {
		case <-stream.Context().Done():
			// Context canceled.
			return breakLoop
		default:
			res := &ListVolumesResponse{
				Id:   id,
				Info: info,
			}
			if err := stream.Send(res); err != nil {
				return err
			}

			count++
			if limit > 0 && count >= limit {
				// Limit reached.
				return breakLoop
			}
			return nil
		}
	})
	if err == breakLoop {
		err = nil
	}
	if err != nil {
		if errors.Is(err, &controller_db.InputError{}) {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		log.Println("ERROR:", err)
		return status.Error(codes.Internal, "database error")
	}
	return nil
}
func (v *localVolumeServer) InspectVolume(ctx context.Context, req *InspectVolumeRequest) (*InspectVolumeResponse, error) {
	if req.GetId() != nil && req.GetName() != "" {
		return nil, status.Error(codes.FailedPrecondition, "id and info is exclusive")
	}

	if req.GetId() != nil {
		// Search by id
		vi, err := v.vs.Get(req.GetId())
		if err != nil {
			if errors.Is(err, &controller_db.InputError{}) {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			log.Println("ERROR:", err)
			return nil, status.Error(codes.Internal, "database error")
		}
		return &InspectVolumeResponse{
			Id:   req.GetId(),
			Info: vi,
		}, nil
	} else if req.GetName() != "" {
		// Search by name
		// todo
		panic("not implemented")
	} else {
		panic("unreachable")
	}
}

func generateVolumeKey() volumeKey {
	s, err := idgen.Gen.NextStringID()
	if err != nil {
		err = xerrors.Errorf("failed to generate id: %w", err)
		log.Println(err)
		panic(err)
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
