package localStorage

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StorageService struct {
	Repo *Repository
}

func (s *StorageService) CreateObject(ctx context.Context, req *elton_v2.CreateObjectRequest) (*elton_v2.CreateObjectResponse, error) {
	if req.GetBody().GetOffset() != 0 {
		return nil, status.Errorf(codes.InvalidArgument, "offset must zero when creating the object")
	}

	body := req.GetBody().GetContents()
	key, err := s.Repo.Create(body, Info{})

	if err != nil {
		return nil, status.Errorf(codes.AlreadyExists, "%s already exists", key.ID)
	}

	res := &elton_v2.CreateObjectResponse{
		Key: &elton_v2.ObjectKey{
			Id: key.ID,
		},
	}
	return res, nil
}
func (s *StorageService) GetObject(ctx context.Context, req *elton_v2.GetObjectRequest) (*elton_v2.GetObjectResponse, error) {
	key := Key{
		ID: req.GetKey().GetId(),
	}
	if key.ID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "key must not empty string")
	}

	body, info, err := s.Repo.Get(key, req.Offset, req.Size)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "local storage: failed to read the object: %s", err.Error())
	}

	createTime, err := ptypes.TimestampProto(info.CreateTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "local storage: failed to convert timestamp: %s", err.Error())
	}
	return &elton_v2.GetObjectResponse{
		Key: &elton_v2.ObjectKey{
			Id: key.ID,
		},
		Body: &elton_v2.ObjectBody{
			Contents: body,
		},
		Info: &elton_v2.ObjectInfo{
			Hash:          info.Hash,
			HashAlgorithm: info.HashAlgorithm,
			CreatedAt:     createTime,
			Size:          info.Size,
		},
	}, nil
}
func (s *StorageService) DeleteObject(ctx context.Context, req *elton_v2.DeleteObjectRequest) (*elton_v2.DeleteObjectResponse, error) {
	key := Key{
		ID: req.GetKey().GetId(),
	}
	if key.ID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "key must not empty string")
	}

	_, err := s.Repo.Delete(key)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "local storage: failed to delete the object: %s", err.Error())
	}

	return &elton_v2.DeleteObjectResponse{}, nil
}
