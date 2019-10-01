package simple

import (
	"context"
	"errors"
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	controller_db "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems/controller/db"
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
	// TODO: impl
	UnimplementedCommitServiceServer
	lock sync.RWMutex
	vs   controller_db.VolumeStore
	cs   controller_db.CommitStore
}

func (v *localVolumeServer) CreateVolume(ctx context.Context, req *CreateVolumeRequest) (*CreateVolumeResponse, error) {
	if req.GetInfo() == nil {
		return nil, status.Error(codes.InvalidArgument, "info is null")
	}

	vid, err := v.vs.Create(req.GetInfo())
	if err != nil {
		if errors.Is(err, controller_db.ErrDupVolumeID) || errors.Is(err, controller_db.ErrDupVolumeName) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
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
		if errors.Is(err, controller_db.ErrNotFoundVolume) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
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
	bothEmpty := req.GetId() == nil && req.GetName() == ""
	bothNonEmpty := req.GetId() != nil && req.GetName() != ""
	if bothEmpty || bothNonEmpty {
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
		vid, vi, err := v.vs.GetByName(req.GetName())
		if err != nil {
			if errors.Is(err, &controller_db.InputError{}) {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			log.Println("ERROR:", err)
			return nil, status.Error(codes.Internal, "database error")
		}
		return &InspectVolumeResponse{
			Id:   vid,
			Info: vi,
		}, err
	} else {
		panic("unreachable")
	}
}
