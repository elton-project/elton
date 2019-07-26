package localStorage

import (
	"context"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StorageService struct{}

func (*StorageService) CreateObject(ctx context.Context, req *elton_v2.CreateObjectRequest) (*elton_v2.CreateObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateObject not implemented")
}
func (*StorageService) GetObject(ctx context.Context, req *elton_v2.GetObjectRequest) (*elton_v2.GetObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetObject not implemented")
}
func (*StorageService) DeleteObject(ctx context.Context, req *elton_v2.DeleteObjectRequest) (*elton_v2.DeleteObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteObject not implemented")
}
