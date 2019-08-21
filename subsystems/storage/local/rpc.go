package localStorage

import (
	"context"
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
	key, err := s.Repo.Create(body)

	if err != nil {
		return nil, status.Errorf(codes.AlreadyExists, "%s (version %s) already exists")
	}

	res := &elton_v2.CreateObjectResponse{
		Key: &elton_v2.ObjectKey{
			Id: key.ID,
		},
	}
	return res, nil
}
func (*StorageService) GetObject(ctx context.Context, req *elton_v2.GetObjectRequest) (*elton_v2.GetObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetObject not implemented")
}
func (*StorageService) DeleteObject(ctx context.Context, req *elton_v2.DeleteObjectRequest) (*elton_v2.DeleteObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteObject not implemented")
}
