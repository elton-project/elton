// Code generated by protoc-gen-go. DO NOT EDIT.
// source: meta.proto

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

type GetMetaRequest struct {
	Key                  *PropertyID `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *GetMetaRequest) Reset()         { *m = GetMetaRequest{} }
func (m *GetMetaRequest) String() string { return proto.CompactTextString(m) }
func (*GetMetaRequest) ProtoMessage()    {}
func (*GetMetaRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_3b5ea8fe65782bcc, []int{0}
}

func (m *GetMetaRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetMetaRequest.Unmarshal(m, b)
}
func (m *GetMetaRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetMetaRequest.Marshal(b, m, deterministic)
}
func (m *GetMetaRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetMetaRequest.Merge(m, src)
}
func (m *GetMetaRequest) XXX_Size() int {
	return xxx_messageInfo_GetMetaRequest.Size(m)
}
func (m *GetMetaRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetMetaRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetMetaRequest proto.InternalMessageInfo

func (m *GetMetaRequest) GetKey() *PropertyID {
	if m != nil {
		return m.Key
	}
	return nil
}

type GetMetaResponse struct {
	// Requested key.
	Key *PropertyID `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	// Property value.  If property is not exists, it is null.
	Body                 *Property `protobuf:"bytes,2,opt,name=body,proto3" json:"body,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *GetMetaResponse) Reset()         { *m = GetMetaResponse{} }
func (m *GetMetaResponse) String() string { return proto.CompactTextString(m) }
func (*GetMetaResponse) ProtoMessage()    {}
func (*GetMetaResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_3b5ea8fe65782bcc, []int{1}
}

func (m *GetMetaResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetMetaResponse.Unmarshal(m, b)
}
func (m *GetMetaResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetMetaResponse.Marshal(b, m, deterministic)
}
func (m *GetMetaResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetMetaResponse.Merge(m, src)
}
func (m *GetMetaResponse) XXX_Size() int {
	return xxx_messageInfo_GetMetaResponse.Size(m)
}
func (m *GetMetaResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetMetaResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetMetaResponse proto.InternalMessageInfo

func (m *GetMetaResponse) GetKey() *PropertyID {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *GetMetaResponse) GetBody() *Property {
	if m != nil {
		return m.Body
	}
	return nil
}

type SetMetaRequest struct {
	Key                  *PropertyID `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Body                 *Property   `protobuf:"bytes,2,opt,name=body,proto3" json:"body,omitempty"`
	MustCreate           bool        `protobuf:"varint,4,opt,name=mustCreate,proto3" json:"mustCreate,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *SetMetaRequest) Reset()         { *m = SetMetaRequest{} }
func (m *SetMetaRequest) String() string { return proto.CompactTextString(m) }
func (*SetMetaRequest) ProtoMessage()    {}
func (*SetMetaRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_3b5ea8fe65782bcc, []int{2}
}

func (m *SetMetaRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SetMetaRequest.Unmarshal(m, b)
}
func (m *SetMetaRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SetMetaRequest.Marshal(b, m, deterministic)
}
func (m *SetMetaRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SetMetaRequest.Merge(m, src)
}
func (m *SetMetaRequest) XXX_Size() int {
	return xxx_messageInfo_SetMetaRequest.Size(m)
}
func (m *SetMetaRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SetMetaRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SetMetaRequest proto.InternalMessageInfo

func (m *SetMetaRequest) GetKey() *PropertyID {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *SetMetaRequest) GetBody() *Property {
	if m != nil {
		return m.Body
	}
	return nil
}

func (m *SetMetaRequest) GetMustCreate() bool {
	if m != nil {
		return m.MustCreate
	}
	return false
}

type SetMetaResponse struct {
	// Requested key.
	Key *PropertyID `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	// Old property value.  If property created, it is null.
	OldBody              *Property `protobuf:"bytes,2,opt,name=oldBody,proto3" json:"oldBody,omitempty"`
	Created              bool      `protobuf:"varint,4,opt,name=created,proto3" json:"created,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *SetMetaResponse) Reset()         { *m = SetMetaResponse{} }
func (m *SetMetaResponse) String() string { return proto.CompactTextString(m) }
func (*SetMetaResponse) ProtoMessage()    {}
func (*SetMetaResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_3b5ea8fe65782bcc, []int{3}
}

func (m *SetMetaResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SetMetaResponse.Unmarshal(m, b)
}
func (m *SetMetaResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SetMetaResponse.Marshal(b, m, deterministic)
}
func (m *SetMetaResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SetMetaResponse.Merge(m, src)
}
func (m *SetMetaResponse) XXX_Size() int {
	return xxx_messageInfo_SetMetaResponse.Size(m)
}
func (m *SetMetaResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SetMetaResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SetMetaResponse proto.InternalMessageInfo

func (m *SetMetaResponse) GetKey() *PropertyID {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *SetMetaResponse) GetOldBody() *Property {
	if m != nil {
		return m.OldBody
	}
	return nil
}

func (m *SetMetaResponse) GetCreated() bool {
	if m != nil {
		return m.Created
	}
	return false
}

func init() {
	proto.RegisterType((*GetMetaRequest)(nil), "elton.v2.GetMetaRequest")
	proto.RegisterType((*GetMetaResponse)(nil), "elton.v2.GetMetaResponse")
	proto.RegisterType((*SetMetaRequest)(nil), "elton.v2.SetMetaRequest")
	proto.RegisterType((*SetMetaResponse)(nil), "elton.v2.SetMetaResponse")
}

func init() { proto.RegisterFile("meta.proto", fileDescriptor_3b5ea8fe65782bcc) }

var fileDescriptor_3b5ea8fe65782bcc = []byte{
	// 255 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xca, 0x4d, 0x2d, 0x49,
	0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x48, 0xcd, 0x29, 0xc9, 0xcf, 0xd3, 0x2b, 0x33,
	0x92, 0xe2, 0x2e, 0xa9, 0x2c, 0x48, 0x2d, 0x86, 0x08, 0x2b, 0x59, 0x70, 0xf1, 0xb9, 0xa7, 0x96,
	0xf8, 0xa6, 0x96, 0x24, 0x06, 0xa5, 0x16, 0x96, 0xa6, 0x16, 0x97, 0x08, 0xa9, 0x71, 0x31, 0x67,
	0xa7, 0x56, 0x4a, 0x30, 0x2a, 0x30, 0x6a, 0x70, 0x1b, 0x89, 0xe8, 0xc1, 0xb4, 0xe9, 0x05, 0x14,
	0xe5, 0x17, 0xa4, 0x16, 0x95, 0x54, 0x7a, 0xba, 0x04, 0x81, 0x14, 0x28, 0x25, 0x72, 0xf1, 0xc3,
	0x75, 0x16, 0x17, 0xe4, 0xe7, 0x15, 0xa7, 0x12, 0xab, 0x55, 0x48, 0x8d, 0x8b, 0x25, 0x29, 0x3f,
	0xa5, 0x52, 0x82, 0x09, 0xac, 0x50, 0x08, 0x53, 0x61, 0x10, 0x58, 0x5e, 0xa9, 0x81, 0x91, 0x8b,
	0x2f, 0x98, 0x2c, 0xd7, 0x11, 0x6b, 0x85, 0x90, 0x1c, 0x17, 0x57, 0x6e, 0x69, 0x71, 0x89, 0x73,
	0x51, 0x6a, 0x62, 0x49, 0xaa, 0x04, 0x8b, 0x02, 0xa3, 0x06, 0x47, 0x10, 0x92, 0x88, 0x52, 0x23,
	0x23, 0x17, 0x7f, 0x30, 0x99, 0xde, 0xd4, 0xe1, 0x62, 0xcf, 0xcf, 0x49, 0x71, 0xc2, 0xef, 0x0c,
	0x98, 0x12, 0x21, 0x09, 0x2e, 0xf6, 0x64, 0xb0, 0x9d, 0x29, 0x50, 0x67, 0xc0, 0xb8, 0x46, 0xbd,
	0x8c, 0x5c, 0xdc, 0x20, 0x07, 0x04, 0xa7, 0x16, 0x95, 0x65, 0x26, 0xa7, 0x0a, 0xd9, 0x71, 0xb1,
	0x43, 0x43, 0x5e, 0x48, 0x02, 0x61, 0x22, 0x6a, 0x34, 0x4a, 0x49, 0x62, 0x91, 0x81, 0xba, 0xdf,
	0x8e, 0x8b, 0x3d, 0x18, 0x53, 0x7f, 0x30, 0x4e, 0xfd, 0x68, 0xfe, 0x4f, 0x62, 0x03, 0x27, 0x1d,
	0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0x7a, 0x56, 0xa9, 0x17, 0x5f, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MetaServiceClient is the client API for MetaService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MetaServiceClient interface {
	// Get a property value.
	GetMeta(ctx context.Context, in *GetMetaRequest, opts ...grpc.CallOption) (*GetMetaResponse, error)
	// Set a property key and value.
	//
	// Error:
	// - AlreadyExists: Failed to create the new property.
	// - Unauthenticated: Failed to replacement the exists property.
	SetMeta(ctx context.Context, in *SetMetaRequest, opts ...grpc.CallOption) (*SetMetaResponse, error)
}

type metaServiceClient struct {
	cc *grpc.ClientConn
}

func NewMetaServiceClient(cc *grpc.ClientConn) MetaServiceClient {
	return &metaServiceClient{cc}
}

func (c *metaServiceClient) GetMeta(ctx context.Context, in *GetMetaRequest, opts ...grpc.CallOption) (*GetMetaResponse, error) {
	out := new(GetMetaResponse)
	err := c.cc.Invoke(ctx, "/elton.v2.MetaService/GetMeta", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metaServiceClient) SetMeta(ctx context.Context, in *SetMetaRequest, opts ...grpc.CallOption) (*SetMetaResponse, error) {
	out := new(SetMetaResponse)
	err := c.cc.Invoke(ctx, "/elton.v2.MetaService/SetMeta", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MetaServiceServer is the server API for MetaService service.
type MetaServiceServer interface {
	// Get a property value.
	GetMeta(context.Context, *GetMetaRequest) (*GetMetaResponse, error)
	// Set a property key and value.
	//
	// Error:
	// - AlreadyExists: Failed to create the new property.
	// - Unauthenticated: Failed to replacement the exists property.
	SetMeta(context.Context, *SetMetaRequest) (*SetMetaResponse, error)
}

// UnimplementedMetaServiceServer can be embedded to have forward compatible implementations.
type UnimplementedMetaServiceServer struct {
}

func (*UnimplementedMetaServiceServer) GetMeta(ctx context.Context, req *GetMetaRequest) (*GetMetaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMeta not implemented")
}
func (*UnimplementedMetaServiceServer) SetMeta(ctx context.Context, req *SetMetaRequest) (*SetMetaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetMeta not implemented")
}

func RegisterMetaServiceServer(s *grpc.Server, srv MetaServiceServer) {
	s.RegisterService(&_MetaService_serviceDesc, srv)
}

func _MetaService_GetMeta_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMetaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetaServiceServer).GetMeta(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/elton.v2.MetaService/GetMeta",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetaServiceServer).GetMeta(ctx, req.(*GetMetaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetaService_SetMeta_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetMetaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetaServiceServer).SetMeta(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/elton.v2.MetaService/SetMeta",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetaServiceServer).SetMeta(ctx, req.(*SetMetaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _MetaService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "elton.v2.MetaService",
	HandlerType: (*MetaServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetMeta",
			Handler:    _MetaService_GetMeta_Handler,
		},
		{
			MethodName: "SetMeta",
			Handler:    _MetaService_SetMeta_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "meta.proto",
}
