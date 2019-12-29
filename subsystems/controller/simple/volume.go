package simple

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	controller_db "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems/controller/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

func newLocalVolumeServer(vs controller_db.VolumeStore, cs controller_db.CommitStore) *localVolumeServer {
	return &localVolumeServer{
		vs: vs,
		cs: cs,
	}
}

type localVolumeServer struct {
	vs controller_db.VolumeStore
	cs controller_db.CommitStore
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
			if errors.Is(err, controller_db.ErrNotFoundVolume) {
				return nil, status.Error(codes.NotFound, err.Error())
			}
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
			if errors.Is(err, controller_db.ErrNotFoundVolume) {
				return nil, status.Error(codes.NotFound, err.Error())
			}
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

func (v *localVolumeServer) GetLastCommit(ctx context.Context, req *GetLastCommitRequest) (*GetLastCommitResponse, error) {
	vid := req.GetVolumeId()
	if vid.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "vid should not empty")
	}

	cid, err := v.cs.Latest(vid)
	if err != nil {
		if errors.Is(err, controller_db.ErrNotFoundCommit) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, &controller_db.InputError{}) {
			log.Printf("[CRITICAL] Missing error handling: %+v", err)
			return nil, status.Error(codes.Internal, err.Error())
		}
		log.Printf("[ERROR] %+v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	info, err := v.cs.Get(cid)
	if err != nil {
		if errors.Is(err, controller_db.ErrNotFoundCommit) {
			// The commit deleted before get detail info.
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, &controller_db.InputError{}) {
			log.Printf("[CRITICAL] Missing error handling: %+v", err)
			return nil, status.Error(codes.Internal, err.Error())
		}
		log.Printf("[ERROR] %+v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &GetLastCommitResponse{
		Id:   cid,
		Info: info,
	}, nil
}
func (v *localVolumeServer) ListCommits(req *ListCommitsRequest, srv CommitService_ListCommitsServer) error {
	if req.GetNext() != "" {
		return status.Error(codes.InvalidArgument, "next parameter is not supported")
	}
	limit := req.GetLimit()

	vid := req.GetId()
	cid, err := v.cs.Latest(vid)
	if err != nil {
		if errors.Is(err, controller_db.ErrNotFoundCommit) {
			// The volume has no commit.
			return nil
		}
		if errors.Is(err, &controller_db.InputError{}) {
			log.Printf("[CRITICAL] Missing error handling: %+v", err)
			return status.Error(codes.Internal, err.Error())
		}
		log.Printf("[ERROR] %+v", err)
		return status.Error(codes.Internal, err.Error())
	}

	count := uint64(0)
	for cid.GetId().GetId() != "" {
		select {
		case <-srv.Context().Done():
			return status.Error(codes.Canceled, "canceled")
		default:
			err = srv.Send(&ListCommitsResponse{
				Next: "",
				Id:   cid,
			})
			if err != nil {
				return fmt.Errorf("failed to send response: %w", err)
			}
			count++
			if limit > 0 && count >= limit {
				// Limit reached.
				break
			}
		}

		info, err := v.cs.Get(cid)
		if err != nil {
			if errors.Is(err, controller_db.ErrNotFoundCommit) {
				// The commit deleted during processing.
				return nil
			}
			if errors.Is(err, &controller_db.InputError{}) {
				log.Printf("[CRITICAL] Missing error handling: %+v", err)
				return status.Error(codes.Internal, err.Error())
			}
			log.Printf("[ERROR] %+v", err)
			return status.Error(codes.Internal, err.Error())
		}
		cid = info.GetLeftParentID()
	}
	return nil
}
func (v *localVolumeServer) GetCommit(ctx context.Context, req *GetCommitRequest) (*GetCommitResponse, error) {
	if req.GetId() == nil {
		return nil, status.Error(codes.InvalidArgument, "id should not nil")
	}

	info, err := v.cs.Get(req.GetId())
	if err != nil {
		if errors.Is(err, controller_db.ErrNotFoundCommit) {
			return nil, status.Errorf(codes.NotFound, "not found commit: %s", req.GetId())
		}
		if errors.Is(err, &controller_db.InputError{}) {
			log.Printf("[CRITICAL] Missing error handling: %+v", err)
			return nil, status.Error(codes.Internal, err.Error())
		}
		log.Printf("[ERROR] %+v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &GetCommitResponse{
		Id:   req.GetId(),
		Info: info,
	}, nil
}
func (v *localVolumeServer) Commit(ctx context.Context, req *CommitRequest) (*CommitResponse, error) {
	if req.GetId() == nil {
		return nil, status.Error(codes.InvalidArgument, "id should not nil")
	}
	if req.GetInfo() == nil {
		return nil, status.Error(codes.InvalidArgument, "info should not nil")
	}
	if req.GetInfo().GetTree() == nil {
		return nil, status.Error(codes.InvalidArgument, "tree should not nil")
	}

	// Base info
	baseID := req.GetInfo().GetLeftParentID()
	resCmt, err := v.GetCommit(ctx, &GetCommitRequest{
		Id: baseID,
	})
	if err != nil {
		return nil, wrapStatus(err, codes.InvalidArgument, "left parent")
	}
	baseTree := resCmt.GetInfo().GetTree()

	// Last info
	resLast, err := v.GetLastCommit(ctx, &GetLastCommitRequest{
		VolumeId: req.Id,
	})
	if err != nil {
		return nil, wrapStatus(err, codes.InvalidArgument, "last commit")
	}
	lastID := resLast.GetId()
	lastTree := resLast.GetInfo().GetTree()

	if baseID.Equals(lastID) {
		cid, err := v.commit(req.GetId(), req.GetInfo())
		if err != nil {
			return nil, wrapStatus(err, 0, "saving new commit")
		}
		return &CommitResponse{Id: cid}, nil
	} else {
		// Some transactions are committed during this transaction processing.  Should try to merge two commits.
		m := &Merger{
			Info:    req.GetInfo(),
			Base:    baseTree,
			Latest:  lastTree,
			Current: req.GetInfo().GetTree(),
		}
		mergedTree, err := m.Merge()
		if err != nil {
			return nil, wrapStatus(err, 0, "merge two commits")
		}

		// We succeed merge latest tree and current tree.  Commit latest current tree and merged tree.
		// todo: latestを進めずにコミットする
		currentCid, err := v.commit(req.GetId(), req.GetInfo())
		if err != nil {
			return nil, wrapStatus(err, 0, "saving current commit")
		}
		mergedCid, err := v.commit(req.GetId(), &CommitInfo{
			CreatedAt:     ptypes.TimestampNow(),
			LeftParentID:  lastID,
			RightParentID: currentCid,
			Tree:          mergedTree,
		})
		if err != nil {
			return nil, wrapStatus(err, 0, "saving merged commit")
		}
		return &CommitResponse{Id: mergedCid}, nil
	}
}
func (v *localVolumeServer) commit(vid *VolumeID, info *CommitInfo) (*CommitID, error) {
	cid, err := v.cs.Create(vid, info, info.GetTree())
	if err != nil {
		if errors.Is(err, controller_db.ErrCrossVolumeCommit) ||
			errors.Is(err, controller_db.ErrNotFoundVolume) ||
			errors.Is(err, controller_db.ErrInvalidParentCommit) ||
			errors.Is(err, controller_db.ErrInvalidTree) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, &controller_db.InputError{}) {
			log.Printf("[CRITICAL] Missing error handling: %+v", err)
			return nil, status.Error(codes.Internal, err.Error())
		}
		log.Printf("[ERROR] %+v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return cid, nil
}

// wrapStatus returns new gRPC error object with specified code and prefix.
// If base error code is codes.Internal or code==0, the code argument is ignored and keeps original gRPC error code.
func wrapStatus(err error, code codes.Code, prefix string) error {
	newCode := status.Code(err)
	if newCode != codes.Internal && code != 0 {
		newCode = code
	}
	return status.Errorf(newCode, "%s: %s", prefix, status.Convert(err).Message())
}
