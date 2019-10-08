// Code generated by protoc-gen-go. DO NOT EDIT.
// source: storage.proto

package elton_v2

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type CreateObjectRequest struct {
	Body                 *ObjectBody `protobuf:"bytes,2,opt,name=body,proto3" json:"body,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *CreateObjectRequest) Reset()         { *m = CreateObjectRequest{} }
func (m *CreateObjectRequest) String() string { return proto.CompactTextString(m) }
func (*CreateObjectRequest) ProtoMessage()    {}
func (*CreateObjectRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_0d2c4ccf1453ffdb, []int{0}
}

func (m *CreateObjectRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateObjectRequest.Unmarshal(m, b)
}
func (m *CreateObjectRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateObjectRequest.Marshal(b, m, deterministic)
}
func (m *CreateObjectRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateObjectRequest.Merge(m, src)
}
func (m *CreateObjectRequest) XXX_Size() int {
	return xxx_messageInfo_CreateObjectRequest.Size(m)
}
func (m *CreateObjectRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateObjectRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CreateObjectRequest proto.InternalMessageInfo

func (m *CreateObjectRequest) GetBody() *ObjectBody {
	if m != nil {
		return m.Body
	}
	return nil
}

type CreateObjectResponse struct {
	Key                  *ObjectKey `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *CreateObjectResponse) Reset()         { *m = CreateObjectResponse{} }
func (m *CreateObjectResponse) String() string { return proto.CompactTextString(m) }
func (*CreateObjectResponse) ProtoMessage()    {}
func (*CreateObjectResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_0d2c4ccf1453ffdb, []int{1}
}

func (m *CreateObjectResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateObjectResponse.Unmarshal(m, b)
}
func (m *CreateObjectResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateObjectResponse.Marshal(b, m, deterministic)
}
func (m *CreateObjectResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateObjectResponse.Merge(m, src)
}
func (m *CreateObjectResponse) XXX_Size() int {
	return xxx_messageInfo_CreateObjectResponse.Size(m)
}
func (m *CreateObjectResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateObjectResponse.DiscardUnknown(m)
}

var xxx_messageInfo_CreateObjectResponse proto.InternalMessageInfo

func (m *CreateObjectResponse) GetKey() *ObjectKey {
	if m != nil {
		return m.Key
	}
	return nil
}

type GetObjectRequest struct {
	Key                  *ObjectKey `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Offset               uint64     `protobuf:"varint,2,opt,name=offset,proto3" json:"offset,omitempty"`
	Size                 uint64     `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *GetObjectRequest) Reset()         { *m = GetObjectRequest{} }
func (m *GetObjectRequest) String() string { return proto.CompactTextString(m) }
func (*GetObjectRequest) ProtoMessage()    {}
func (*GetObjectRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_0d2c4ccf1453ffdb, []int{2}
}

func (m *GetObjectRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetObjectRequest.Unmarshal(m, b)
}
func (m *GetObjectRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetObjectRequest.Marshal(b, m, deterministic)
}
func (m *GetObjectRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetObjectRequest.Merge(m, src)
}
func (m *GetObjectRequest) XXX_Size() int {
	return xxx_messageInfo_GetObjectRequest.Size(m)
}
func (m *GetObjectRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetObjectRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetObjectRequest proto.InternalMessageInfo

func (m *GetObjectRequest) GetKey() *ObjectKey {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *GetObjectRequest) GetOffset() uint64 {
	if m != nil {
		return m.Offset
	}
	return 0
}

func (m *GetObjectRequest) GetSize() uint64 {
	if m != nil {
		return m.Size
	}
	return 0
}

type GetObjectResponse struct {
	Key                  *ObjectKey  `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Body                 *ObjectBody `protobuf:"bytes,2,opt,name=body,proto3" json:"body,omitempty"`
	Info                 *ObjectInfo `protobuf:"bytes,3,opt,name=info,proto3" json:"info,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *GetObjectResponse) Reset()         { *m = GetObjectResponse{} }
func (m *GetObjectResponse) String() string { return proto.CompactTextString(m) }
func (*GetObjectResponse) ProtoMessage()    {}
func (*GetObjectResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_0d2c4ccf1453ffdb, []int{3}
}

func (m *GetObjectResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetObjectResponse.Unmarshal(m, b)
}
func (m *GetObjectResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetObjectResponse.Marshal(b, m, deterministic)
}
func (m *GetObjectResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetObjectResponse.Merge(m, src)
}
func (m *GetObjectResponse) XXX_Size() int {
	return xxx_messageInfo_GetObjectResponse.Size(m)
}
func (m *GetObjectResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetObjectResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetObjectResponse proto.InternalMessageInfo

func (m *GetObjectResponse) GetKey() *ObjectKey {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *GetObjectResponse) GetBody() *ObjectBody {
	if m != nil {
		return m.Body
	}
	return nil
}

func (m *GetObjectResponse) GetInfo() *ObjectInfo {
	if m != nil {
		return m.Info
	}
	return nil
}

type DeleteObjectRequest struct {
	Key                  *ObjectKey `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *DeleteObjectRequest) Reset()         { *m = DeleteObjectRequest{} }
func (m *DeleteObjectRequest) String() string { return proto.CompactTextString(m) }
func (*DeleteObjectRequest) ProtoMessage()    {}
func (*DeleteObjectRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_0d2c4ccf1453ffdb, []int{4}
}

func (m *DeleteObjectRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteObjectRequest.Unmarshal(m, b)
}
func (m *DeleteObjectRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteObjectRequest.Marshal(b, m, deterministic)
}
func (m *DeleteObjectRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteObjectRequest.Merge(m, src)
}
func (m *DeleteObjectRequest) XXX_Size() int {
	return xxx_messageInfo_DeleteObjectRequest.Size(m)
}
func (m *DeleteObjectRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteObjectRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteObjectRequest proto.InternalMessageInfo

func (m *DeleteObjectRequest) GetKey() *ObjectKey {
	if m != nil {
		return m.Key
	}
	return nil
}

type DeleteObjectResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteObjectResponse) Reset()         { *m = DeleteObjectResponse{} }
func (m *DeleteObjectResponse) String() string { return proto.CompactTextString(m) }
func (*DeleteObjectResponse) ProtoMessage()    {}
func (*DeleteObjectResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_0d2c4ccf1453ffdb, []int{5}
}

func (m *DeleteObjectResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteObjectResponse.Unmarshal(m, b)
}
func (m *DeleteObjectResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteObjectResponse.Marshal(b, m, deterministic)
}
func (m *DeleteObjectResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteObjectResponse.Merge(m, src)
}
func (m *DeleteObjectResponse) XXX_Size() int {
	return xxx_messageInfo_DeleteObjectResponse.Size(m)
}
func (m *DeleteObjectResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteObjectResponse.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteObjectResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*CreateObjectRequest)(nil), "elton.v2.CreateObjectRequest")
	proto.RegisterType((*CreateObjectResponse)(nil), "elton.v2.CreateObjectResponse")
	proto.RegisterType((*GetObjectRequest)(nil), "elton.v2.GetObjectRequest")
	proto.RegisterType((*GetObjectResponse)(nil), "elton.v2.GetObjectResponse")
	proto.RegisterType((*DeleteObjectRequest)(nil), "elton.v2.DeleteObjectRequest")
	proto.RegisterType((*DeleteObjectResponse)(nil), "elton.v2.DeleteObjectResponse")
}

func init() { proto.RegisterFile("storage.proto", fileDescriptor_0d2c4ccf1453ffdb) }

var fileDescriptor_0d2c4ccf1453ffdb = []byte{
	// 302 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x52, 0xd1, 0x4a, 0xc3, 0x40,
	0x10, 0x24, 0x36, 0x14, 0xdd, 0x56, 0xd1, 0x4b, 0x28, 0x21, 0xa2, 0x48, 0x40, 0xe8, 0x53, 0x1e,
	0xe2, 0xab, 0x22, 0x68, 0x41, 0x44, 0x8a, 0x90, 0x7e, 0x41, 0xd3, 0x6e, 0x24, 0x5a, 0xb2, 0x31,
	0x77, 0x16, 0xce, 0x8f, 0xf0, 0x6b, 0xfd, 0x00, 0xe9, 0x25, 0xda, 0x4b, 0x38, 0x21, 0x7d, 0x4b,
	0x76, 0x67, 0x67, 0x67, 0x67, 0x0e, 0x0e, 0xb9, 0xa0, 0x72, 0xfe, 0x82, 0x61, 0x51, 0x92, 0x20,
	0xb6, 0x8f, 0x2b, 0x41, 0x79, 0xb8, 0x8e, 0xfc, 0x81, 0x90, 0x05, 0xf2, 0xaa, 0x1c, 0xdc, 0x82,
	0x73, 0x5f, 0xe2, 0x5c, 0xe0, 0x73, 0xf2, 0x8a, 0x0b, 0x11, 0xe3, 0xfb, 0x07, 0x72, 0xc1, 0xc6,
	0x60, 0x27, 0xb4, 0x94, 0xde, 0xde, 0x85, 0x35, 0x1e, 0x44, 0x6e, 0xf8, 0x3b, 0x1c, 0x56, 0xb0,
	0x3b, 0x5a, 0xca, 0x58, 0x21, 0x82, 0x1b, 0x70, 0x9b, 0x04, 0xbc, 0xa0, 0x9c, 0x23, 0xbb, 0x84,
	0xde, 0x1b, 0x4a, 0xcf, 0x52, 0x04, 0x4e, 0x9b, 0xe0, 0x09, 0x65, 0xbc, 0xe9, 0x07, 0x08, 0xc7,
	0x0f, 0x28, 0x9a, 0xcb, 0xbb, 0x8d, 0xb2, 0x11, 0xf4, 0x29, 0x4d, 0x39, 0x0a, 0xa5, 0xd2, 0x8e,
	0xeb, 0x3f, 0xc6, 0xc0, 0xe6, 0xd9, 0x27, 0x7a, 0x3d, 0x55, 0x55, 0xdf, 0xc1, 0x97, 0x05, 0x27,
	0xda, 0x9e, 0x9d, 0x34, 0x76, 0x37, 0x63, 0x83, 0xcc, 0xf2, 0x94, 0xd4, 0x6a, 0x03, 0xf2, 0x31,
	0x4f, 0x29, 0x56, 0x88, 0xe0, 0x1a, 0x9c, 0x09, 0xae, 0xb0, 0xed, 0x7b, 0x47, 0xd7, 0x46, 0xe0,
	0x36, 0xa7, 0xab, 0x83, 0xa2, 0x6f, 0x0b, 0x8e, 0x66, 0x55, 0xec, 0x33, 0x2c, 0xd7, 0xd9, 0x02,
	0xd9, 0x14, 0x86, 0x7a, 0x3e, 0xec, 0x6c, 0x4b, 0x6a, 0x08, 0xde, 0x3f, 0xff, 0xaf, 0x5d, 0x5b,
	0x36, 0x81, 0x83, 0x3f, 0x1f, 0x99, 0xbf, 0x05, 0xb7, 0x43, 0xf4, 0x4f, 0x8d, 0xbd, 0x9a, 0x65,
	0x0a, 0x43, 0x5d, 0xbf, 0x2e, 0xca, 0xe0, 0x8a, 0x2e, 0xca, 0x74, 0x76, 0xd2, 0x57, 0x6f, 0xf9,
	0xea, 0x27, 0x00, 0x00, 0xff, 0xff, 0xb3, 0x83, 0x28, 0x89, 0xf3, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// StorageServiceClient is the client API for StorageService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type StorageServiceClient interface {
	// Create and save an object.
	//
	// Error:
	// - InvalidArgument: If specified object is invalid.
	// - AlreadyExists: ???  TODO
	CreateObject(ctx context.Context, in *CreateObjectRequest, opts ...grpc.CallOption) (*CreateObjectResponse, error)
	// Get an object.
	//
	// Error:
	// - InvalidArgument
	// - Internal
	GetObject(ctx context.Context, in *GetObjectRequest, opts ...grpc.CallOption) (*GetObjectResponse, error)
	// Delete an object.
	//
	// Error:
	// - InvalidArgument
	// - Internal
	DeleteObject(ctx context.Context, in *DeleteObjectRequest, opts ...grpc.CallOption) (*DeleteObjectResponse, error)
}

type storageServiceClient struct {
	cc *grpc.ClientConn
}

func NewStorageServiceClient(cc *grpc.ClientConn) StorageServiceClient {
	return &storageServiceClient{cc}
}

func (c *storageServiceClient) CreateObject(ctx context.Context, in *CreateObjectRequest, opts ...grpc.CallOption) (*CreateObjectResponse, error) {
	out := new(CreateObjectResponse)
	err := c.cc.Invoke(ctx, "/elton.v2.StorageService/CreateObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageServiceClient) GetObject(ctx context.Context, in *GetObjectRequest, opts ...grpc.CallOption) (*GetObjectResponse, error) {
	out := new(GetObjectResponse)
	err := c.cc.Invoke(ctx, "/elton.v2.StorageService/GetObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageServiceClient) DeleteObject(ctx context.Context, in *DeleteObjectRequest, opts ...grpc.CallOption) (*DeleteObjectResponse, error) {
	out := new(DeleteObjectResponse)
	err := c.cc.Invoke(ctx, "/elton.v2.StorageService/DeleteObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StorageServiceServer is the server API for StorageService service.
type StorageServiceServer interface {
	// Create and save an object.
	//
	// Error:
	// - InvalidArgument: If specified object is invalid.
	// - AlreadyExists: ???  TODO
	CreateObject(context.Context, *CreateObjectRequest) (*CreateObjectResponse, error)
	// Get an object.
	//
	// Error:
	// - InvalidArgument
	// - Internal
	GetObject(context.Context, *GetObjectRequest) (*GetObjectResponse, error)
	// Delete an object.
	//
	// Error:
	// - InvalidArgument
	// - Internal
	DeleteObject(context.Context, *DeleteObjectRequest) (*DeleteObjectResponse, error)
}

// UnimplementedStorageServiceServer can be embedded to have forward compatible implementations.
type UnimplementedStorageServiceServer struct {
}

func (*UnimplementedStorageServiceServer) CreateObject(ctx context.Context, req *CreateObjectRequest) (*CreateObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateObject not implemented")
}
func (*UnimplementedStorageServiceServer) GetObject(ctx context.Context, req *GetObjectRequest) (*GetObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetObject not implemented")
}
func (*UnimplementedStorageServiceServer) DeleteObject(ctx context.Context, req *DeleteObjectRequest) (*DeleteObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteObject not implemented")
}

func RegisterStorageServiceServer(s *grpc.Server, srv StorageServiceServer) {
	s.RegisterService(&_StorageService_serviceDesc, srv)
}

func _StorageService_CreateObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageServiceServer).CreateObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/elton.v2.StorageService/CreateObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageServiceServer).CreateObject(ctx, req.(*CreateObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StorageService_GetObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageServiceServer).GetObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/elton.v2.StorageService/GetObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageServiceServer).GetObject(ctx, req.(*GetObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StorageService_DeleteObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageServiceServer).DeleteObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/elton.v2.StorageService/DeleteObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageServiceServer).DeleteObject(ctx, req.(*DeleteObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _StorageService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "elton.v2.StorageService",
	HandlerType: (*StorageServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateObject",
			Handler:    _StorageService_CreateObject_Handler,
		},
		{
			MethodName: "GetObject",
			Handler:    _StorageService_GetObject_Handler,
		},
		{
			MethodName: "DeleteObject",
			Handler:    _StorageService_DeleteObject_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "storage.proto",
}
