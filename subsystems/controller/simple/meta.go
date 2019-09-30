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

func newLocalMetaServer(ms controller_db.MetaStore) *localMetaServer {
	return &localMetaServer{
		ms: ms,
	}
}

type localMetaServer struct {
	lock sync.RWMutex
	ms   controller_db.MetaStore
}

func (m *localMetaServer) GetMeta(ctx context.Context, req *GetMetaRequest) (*GetMetaResponse, error) {
	prop, err := m.ms.Get(req.GetKey())
	if err != nil {
		if errors.Is(err, controller_db.ErrNotFoundProp) {
			return nil, status.Errorf(codes.NotFound, "property not found")
		}
		if errors.Is(err, &controller_db.InputError{}) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		log.Println("ERROR:", err)
		return nil, status.Error(codes.Internal, "database error")
	}
	return &GetMetaResponse{
		Key:  req.GetKey(),
		Body: prop,
	}, nil
}
func (m *localMetaServer) SetMeta(ctx context.Context, req *SetMetaRequest) (*SetMetaResponse, error) {
	old, err := m.ms.Set(req.GetKey(), req.GetBody(), req.GetMustCreate())
	if err != nil {
		if errors.Is(err, &controller_db.InputError{}) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		// TODO: improve error handling
		// return nil, status.Error(codes.AlreadyExists, "key is already exists")
		// return nil, status.Errorf(codes.Unauthenticated, "replacement not allowed")
		log.Println("ERROR:", err)
		return nil, status.Error(codes.Internal, "database error")
	}

	return &SetMetaResponse{
		Key:     req.GetKey(),
		OldBody: old,
		Created: old != nil,
	}, nil
}
