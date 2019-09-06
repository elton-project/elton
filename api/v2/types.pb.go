// Code generated by protoc-gen-go. DO NOT EDIT.
// source: types.proto

package elton_v2

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

type FileType int32

const (
	FileType_Regular         FileType = 0
	FileType_Directory       FileType = 1
	FileType_SymbolicLink    FileType = 2
	FileType_FIFO            FileType = 3
	FileType_CharacterDevice FileType = 4
	FileType_BlockDevice     FileType = 5
	FileType_Socket          FileType = 6
)

var FileType_name = map[int32]string{
	0: "Regular",
	1: "Directory",
	2: "SymbolicLink",
	3: "FIFO",
	4: "CharacterDevice",
	5: "BlockDevice",
	6: "Socket",
}

var FileType_value = map[string]int32{
	"Regular":         0,
	"Directory":       1,
	"SymbolicLink":    2,
	"FIFO":            3,
	"CharacterDevice": 4,
	"BlockDevice":     5,
	"Socket":          6,
}

func (x FileType) String() string {
	return proto.EnumName(FileType_name, int32(x))
}

func (FileType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{0}
}

type ObjectKey struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ObjectKey) Reset()         { *m = ObjectKey{} }
func (m *ObjectKey) String() string { return proto.CompactTextString(m) }
func (*ObjectKey) ProtoMessage()    {}
func (*ObjectKey) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{0}
}

func (m *ObjectKey) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ObjectKey.Unmarshal(m, b)
}
func (m *ObjectKey) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ObjectKey.Marshal(b, m, deterministic)
}
func (m *ObjectKey) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ObjectKey.Merge(m, src)
}
func (m *ObjectKey) XXX_Size() int {
	return xxx_messageInfo_ObjectKey.Size(m)
}
func (m *ObjectKey) XXX_DiscardUnknown() {
	xxx_messageInfo_ObjectKey.DiscardUnknown(m)
}

var xxx_messageInfo_ObjectKey proto.InternalMessageInfo

func (m *ObjectKey) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type ObjectInfo struct {
	Hash []byte `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	// Supported algorithms:
	//  - "SHA1"
	HashAlgorithm        string               `protobuf:"bytes,4,opt,name=hashAlgorithm,proto3" json:"hashAlgorithm,omitempty"`
	CreateTime           *timestamp.Timestamp `protobuf:"bytes,2,opt,name=createTime,proto3" json:"createTime,omitempty"`
	Size                 uint64               `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *ObjectInfo) Reset()         { *m = ObjectInfo{} }
func (m *ObjectInfo) String() string { return proto.CompactTextString(m) }
func (*ObjectInfo) ProtoMessage()    {}
func (*ObjectInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{1}
}

func (m *ObjectInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ObjectInfo.Unmarshal(m, b)
}
func (m *ObjectInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ObjectInfo.Marshal(b, m, deterministic)
}
func (m *ObjectInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ObjectInfo.Merge(m, src)
}
func (m *ObjectInfo) XXX_Size() int {
	return xxx_messageInfo_ObjectInfo.Size(m)
}
func (m *ObjectInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_ObjectInfo.DiscardUnknown(m)
}

var xxx_messageInfo_ObjectInfo proto.InternalMessageInfo

func (m *ObjectInfo) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *ObjectInfo) GetHashAlgorithm() string {
	if m != nil {
		return m.HashAlgorithm
	}
	return ""
}

func (m *ObjectInfo) GetCreateTime() *timestamp.Timestamp {
	if m != nil {
		return m.CreateTime
	}
	return nil
}

func (m *ObjectInfo) GetSize() uint64 {
	if m != nil {
		return m.Size
	}
	return 0
}

type ObjectBody struct {
	Contents             []byte   `protobuf:"bytes,1,opt,name=contents,proto3" json:"contents,omitempty"`
	Offset               uint64   `protobuf:"varint,2,opt,name=offset,proto3" json:"offset,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ObjectBody) Reset()         { *m = ObjectBody{} }
func (m *ObjectBody) String() string { return proto.CompactTextString(m) }
func (*ObjectBody) ProtoMessage()    {}
func (*ObjectBody) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{2}
}

func (m *ObjectBody) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ObjectBody.Unmarshal(m, b)
}
func (m *ObjectBody) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ObjectBody.Marshal(b, m, deterministic)
}
func (m *ObjectBody) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ObjectBody.Merge(m, src)
}
func (m *ObjectBody) XXX_Size() int {
	return xxx_messageInfo_ObjectBody.Size(m)
}
func (m *ObjectBody) XXX_DiscardUnknown() {
	xxx_messageInfo_ObjectBody.DiscardUnknown(m)
}

var xxx_messageInfo_ObjectBody proto.InternalMessageInfo

func (m *ObjectBody) GetContents() []byte {
	if m != nil {
		return m.Contents
	}
	return nil
}

func (m *ObjectBody) GetOffset() uint64 {
	if m != nil {
		return m.Offset
	}
	return 0
}

type PropertyKey struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PropertyKey) Reset()         { *m = PropertyKey{} }
func (m *PropertyKey) String() string { return proto.CompactTextString(m) }
func (*PropertyKey) ProtoMessage()    {}
func (*PropertyKey) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{3}
}

func (m *PropertyKey) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PropertyKey.Unmarshal(m, b)
}
func (m *PropertyKey) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PropertyKey.Marshal(b, m, deterministic)
}
func (m *PropertyKey) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PropertyKey.Merge(m, src)
}
func (m *PropertyKey) XXX_Size() int {
	return xxx_messageInfo_PropertyKey.Size(m)
}
func (m *PropertyKey) XXX_DiscardUnknown() {
	xxx_messageInfo_PropertyKey.DiscardUnknown(m)
}

var xxx_messageInfo_PropertyKey proto.InternalMessageInfo

func (m *PropertyKey) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type PropertyBody struct {
	Body                 string   `protobuf:"bytes,1,opt,name=body,proto3" json:"body,omitempty"`
	AllowReplace         bool     `protobuf:"varint,2,opt,name=allowReplace,proto3" json:"allowReplace,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PropertyBody) Reset()         { *m = PropertyBody{} }
func (m *PropertyBody) String() string { return proto.CompactTextString(m) }
func (*PropertyBody) ProtoMessage()    {}
func (*PropertyBody) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{4}
}

func (m *PropertyBody) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PropertyBody.Unmarshal(m, b)
}
func (m *PropertyBody) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PropertyBody.Marshal(b, m, deterministic)
}
func (m *PropertyBody) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PropertyBody.Merge(m, src)
}
func (m *PropertyBody) XXX_Size() int {
	return xxx_messageInfo_PropertyBody.Size(m)
}
func (m *PropertyBody) XXX_DiscardUnknown() {
	xxx_messageInfo_PropertyBody.DiscardUnknown(m)
}

var xxx_messageInfo_PropertyBody proto.InternalMessageInfo

func (m *PropertyBody) GetBody() string {
	if m != nil {
		return m.Body
	}
	return ""
}

func (m *PropertyBody) GetAllowReplace() bool {
	if m != nil {
		return m.AllowReplace
	}
	return false
}

type NodeID struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NodeID) Reset()         { *m = NodeID{} }
func (m *NodeID) String() string { return proto.CompactTextString(m) }
func (*NodeID) ProtoMessage()    {}
func (*NodeID) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{5}
}

func (m *NodeID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeID.Unmarshal(m, b)
}
func (m *NodeID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeID.Marshal(b, m, deterministic)
}
func (m *NodeID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeID.Merge(m, src)
}
func (m *NodeID) XXX_Size() int {
	return xxx_messageInfo_NodeID.Size(m)
}
func (m *NodeID) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeID.DiscardUnknown(m)
}

var xxx_messageInfo_NodeID proto.InternalMessageInfo

func (m *NodeID) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type Node struct {
	// IP addresses or DNS name.
	Address []string `protobuf:"bytes,1,rep,name=address,proto3" json:"address,omitempty"`
	// Human readable name.
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// Uptime
	Uptime               uint64   `protobuf:"varint,3,opt,name=uptime,proto3" json:"uptime,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Node) Reset()         { *m = Node{} }
func (m *Node) String() string { return proto.CompactTextString(m) }
func (*Node) ProtoMessage()    {}
func (*Node) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{6}
}

func (m *Node) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Node.Unmarshal(m, b)
}
func (m *Node) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Node.Marshal(b, m, deterministic)
}
func (m *Node) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Node.Merge(m, src)
}
func (m *Node) XXX_Size() int {
	return xxx_messageInfo_Node.Size(m)
}
func (m *Node) XXX_DiscardUnknown() {
	xxx_messageInfo_Node.DiscardUnknown(m)
}

var xxx_messageInfo_Node proto.InternalMessageInfo

func (m *Node) GetAddress() []string {
	if m != nil {
		return m.Address
	}
	return nil
}

func (m *Node) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Node) GetUptime() uint64 {
	if m != nil {
		return m.Uptime
	}
	return 0
}

type VolumeID struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *VolumeID) Reset()         { *m = VolumeID{} }
func (m *VolumeID) String() string { return proto.CompactTextString(m) }
func (*VolumeID) ProtoMessage()    {}
func (*VolumeID) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{7}
}

func (m *VolumeID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VolumeID.Unmarshal(m, b)
}
func (m *VolumeID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VolumeID.Marshal(b, m, deterministic)
}
func (m *VolumeID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VolumeID.Merge(m, src)
}
func (m *VolumeID) XXX_Size() int {
	return xxx_messageInfo_VolumeID.Size(m)
}
func (m *VolumeID) XXX_DiscardUnknown() {
	xxx_messageInfo_VolumeID.DiscardUnknown(m)
}

var xxx_messageInfo_VolumeID proto.InternalMessageInfo

func (m *VolumeID) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type VolumeInfo struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *VolumeInfo) Reset()         { *m = VolumeInfo{} }
func (m *VolumeInfo) String() string { return proto.CompactTextString(m) }
func (*VolumeInfo) ProtoMessage()    {}
func (*VolumeInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{8}
}

func (m *VolumeInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VolumeInfo.Unmarshal(m, b)
}
func (m *VolumeInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VolumeInfo.Marshal(b, m, deterministic)
}
func (m *VolumeInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VolumeInfo.Merge(m, src)
}
func (m *VolumeInfo) XXX_Size() int {
	return xxx_messageInfo_VolumeInfo.Size(m)
}
func (m *VolumeInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_VolumeInfo.DiscardUnknown(m)
}

var xxx_messageInfo_VolumeInfo proto.InternalMessageInfo

func (m *VolumeInfo) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type CommitID struct {
	Id                   *VolumeID `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Number               uint64    `protobuf:"varint,2,opt,name=number,proto3" json:"number,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *CommitID) Reset()         { *m = CommitID{} }
func (m *CommitID) String() string { return proto.CompactTextString(m) }
func (*CommitID) ProtoMessage()    {}
func (*CommitID) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{9}
}

func (m *CommitID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CommitID.Unmarshal(m, b)
}
func (m *CommitID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CommitID.Marshal(b, m, deterministic)
}
func (m *CommitID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CommitID.Merge(m, src)
}
func (m *CommitID) XXX_Size() int {
	return xxx_messageInfo_CommitID.Size(m)
}
func (m *CommitID) XXX_DiscardUnknown() {
	xxx_messageInfo_CommitID.DiscardUnknown(m)
}

var xxx_messageInfo_CommitID proto.InternalMessageInfo

func (m *CommitID) GetId() *VolumeID {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *CommitID) GetNumber() uint64 {
	if m != nil {
		return m.Number
	}
	return 0
}

type CommitInfo struct {
	CreatedAt            *timestamp.Timestamp `protobuf:"bytes,1,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	ParentID             *CommitID            `protobuf:"bytes,2,opt,name=parentID,proto3" json:"parentID,omitempty"`
	TreeID               *TreeID              `protobuf:"bytes,3,opt,name=treeID,proto3" json:"treeID,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *CommitInfo) Reset()         { *m = CommitInfo{} }
func (m *CommitInfo) String() string { return proto.CompactTextString(m) }
func (*CommitInfo) ProtoMessage()    {}
func (*CommitInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{10}
}

func (m *CommitInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CommitInfo.Unmarshal(m, b)
}
func (m *CommitInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CommitInfo.Marshal(b, m, deterministic)
}
func (m *CommitInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CommitInfo.Merge(m, src)
}
func (m *CommitInfo) XXX_Size() int {
	return xxx_messageInfo_CommitInfo.Size(m)
}
func (m *CommitInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_CommitInfo.DiscardUnknown(m)
}

var xxx_messageInfo_CommitInfo proto.InternalMessageInfo

func (m *CommitInfo) GetCreatedAt() *timestamp.Timestamp {
	if m != nil {
		return m.CreatedAt
	}
	return nil
}

func (m *CommitInfo) GetParentID() *CommitID {
	if m != nil {
		return m.ParentID
	}
	return nil
}

func (m *CommitInfo) GetTreeID() *TreeID {
	if m != nil {
		return m.TreeID
	}
	return nil
}

type TreeID struct {
	Key                  *ObjectKey `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *TreeID) Reset()         { *m = TreeID{} }
func (m *TreeID) String() string { return proto.CompactTextString(m) }
func (*TreeID) ProtoMessage()    {}
func (*TreeID) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{11}
}

func (m *TreeID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TreeID.Unmarshal(m, b)
}
func (m *TreeID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TreeID.Marshal(b, m, deterministic)
}
func (m *TreeID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TreeID.Merge(m, src)
}
func (m *TreeID) XXX_Size() int {
	return xxx_messageInfo_TreeID.Size(m)
}
func (m *TreeID) XXX_DiscardUnknown() {
	xxx_messageInfo_TreeID.DiscardUnknown(m)
}

var xxx_messageInfo_TreeID proto.InternalMessageInfo

func (m *TreeID) GetKey() *ObjectKey {
	if m != nil {
		return m.Key
	}
	return nil
}

type TreeEntry struct {
	Path                 string   `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	FileID               *FileID  `protobuf:"bytes,2,opt,name=fileID,proto3" json:"fileID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TreeEntry) Reset()         { *m = TreeEntry{} }
func (m *TreeEntry) String() string { return proto.CompactTextString(m) }
func (*TreeEntry) ProtoMessage()    {}
func (*TreeEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{12}
}

func (m *TreeEntry) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TreeEntry.Unmarshal(m, b)
}
func (m *TreeEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TreeEntry.Marshal(b, m, deterministic)
}
func (m *TreeEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TreeEntry.Merge(m, src)
}
func (m *TreeEntry) XXX_Size() int {
	return xxx_messageInfo_TreeEntry.Size(m)
}
func (m *TreeEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_TreeEntry.DiscardUnknown(m)
}

var xxx_messageInfo_TreeEntry proto.InternalMessageInfo

func (m *TreeEntry) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *TreeEntry) GetFileID() *FileID {
	if m != nil {
		return m.FileID
	}
	return nil
}

type FileID struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FileID) Reset()         { *m = FileID{} }
func (m *FileID) String() string { return proto.CompactTextString(m) }
func (*FileID) ProtoMessage()    {}
func (*FileID) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{13}
}

func (m *FileID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FileID.Unmarshal(m, b)
}
func (m *FileID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FileID.Marshal(b, m, deterministic)
}
func (m *FileID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileID.Merge(m, src)
}
func (m *FileID) XXX_Size() int {
	return xxx_messageInfo_FileID.Size(m)
}
func (m *FileID) XXX_DiscardUnknown() {
	xxx_messageInfo_FileID.DiscardUnknown(m)
}

var xxx_messageInfo_FileID proto.InternalMessageInfo

func (m *FileID) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

// Linuxのinodeに相当する。
type File struct {
	// If file is regular or symlink, a valid reference is set to the contentRef.
	// Otherwise, it is set to the null reference.
	ContentRef *FileContentRef `protobuf:"bytes,1,opt,name=contentRef,proto3" json:"contentRef,omitempty"`
	FileType   FileType        `protobuf:"varint,2,opt,name=fileType,proto3,enum=elton.v2.FileType" json:"fileType,omitempty"`
	// 実際には16bitで十分だが、protocol bufferは16bit integerをサポートしていないため、32bitで表現している。
	Mode  uint32               `protobuf:"varint,3,opt,name=mode,proto3" json:"mode,omitempty"`
	Owner uint32               `protobuf:"varint,4,opt,name=owner,proto3" json:"owner,omitempty"`
	Group uint32               `protobuf:"varint,5,opt,name=group,proto3" json:"group,omitempty"`
	Atime *timestamp.Timestamp `protobuf:"bytes,6,opt,name=atime,proto3" json:"atime,omitempty"`
	Mtime *timestamp.Timestamp `protobuf:"bytes,7,opt,name=mtime,proto3" json:"mtime,omitempty"`
	Ctime *timestamp.Timestamp `protobuf:"bytes,8,opt,name=ctime,proto3" json:"ctime,omitempty"`
	// For device file.
	Major                uint32   `protobuf:"varint,9,opt,name=major,proto3" json:"major,omitempty"`
	Minor                uint32   `protobuf:"varint,10,opt,name=minor,proto3" json:"minor,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *File) Reset()         { *m = File{} }
func (m *File) String() string { return proto.CompactTextString(m) }
func (*File) ProtoMessage()    {}
func (*File) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{14}
}

func (m *File) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_File.Unmarshal(m, b)
}
func (m *File) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_File.Marshal(b, m, deterministic)
}
func (m *File) XXX_Merge(src proto.Message) {
	xxx_messageInfo_File.Merge(m, src)
}
func (m *File) XXX_Size() int {
	return xxx_messageInfo_File.Size(m)
}
func (m *File) XXX_DiscardUnknown() {
	xxx_messageInfo_File.DiscardUnknown(m)
}

var xxx_messageInfo_File proto.InternalMessageInfo

func (m *File) GetContentRef() *FileContentRef {
	if m != nil {
		return m.ContentRef
	}
	return nil
}

func (m *File) GetFileType() FileType {
	if m != nil {
		return m.FileType
	}
	return FileType_Regular
}

func (m *File) GetMode() uint32 {
	if m != nil {
		return m.Mode
	}
	return 0
}

func (m *File) GetOwner() uint32 {
	if m != nil {
		return m.Owner
	}
	return 0
}

func (m *File) GetGroup() uint32 {
	if m != nil {
		return m.Group
	}
	return 0
}

func (m *File) GetAtime() *timestamp.Timestamp {
	if m != nil {
		return m.Atime
	}
	return nil
}

func (m *File) GetMtime() *timestamp.Timestamp {
	if m != nil {
		return m.Mtime
	}
	return nil
}

func (m *File) GetCtime() *timestamp.Timestamp {
	if m != nil {
		return m.Ctime
	}
	return nil
}

func (m *File) GetMajor() uint32 {
	if m != nil {
		return m.Major
	}
	return 0
}

func (m *File) GetMinor() uint32 {
	if m != nil {
		return m.Minor
	}
	return 0
}

// Fileの中身への参照
type FileContentRef struct {
	Key                  *ObjectKey `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *FileContentRef) Reset()         { *m = FileContentRef{} }
func (m *FileContentRef) String() string { return proto.CompactTextString(m) }
func (*FileContentRef) ProtoMessage()    {}
func (*FileContentRef) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{15}
}

func (m *FileContentRef) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FileContentRef.Unmarshal(m, b)
}
func (m *FileContentRef) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FileContentRef.Marshal(b, m, deterministic)
}
func (m *FileContentRef) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileContentRef.Merge(m, src)
}
func (m *FileContentRef) XXX_Size() int {
	return xxx_messageInfo_FileContentRef.Size(m)
}
func (m *FileContentRef) XXX_DiscardUnknown() {
	xxx_messageInfo_FileContentRef.DiscardUnknown(m)
}

var xxx_messageInfo_FileContentRef proto.InternalMessageInfo

func (m *FileContentRef) GetKey() *ObjectKey {
	if m != nil {
		return m.Key
	}
	return nil
}

func init() {
	proto.RegisterEnum("elton.v2.FileType", FileType_name, FileType_value)
	proto.RegisterType((*ObjectKey)(nil), "elton.v2.ObjectKey")
	proto.RegisterType((*ObjectInfo)(nil), "elton.v2.ObjectInfo")
	proto.RegisterType((*ObjectBody)(nil), "elton.v2.ObjectBody")
	proto.RegisterType((*PropertyKey)(nil), "elton.v2.PropertyKey")
	proto.RegisterType((*PropertyBody)(nil), "elton.v2.PropertyBody")
	proto.RegisterType((*NodeID)(nil), "elton.v2.NodeID")
	proto.RegisterType((*Node)(nil), "elton.v2.Node")
	proto.RegisterType((*VolumeID)(nil), "elton.v2.VolumeID")
	proto.RegisterType((*VolumeInfo)(nil), "elton.v2.VolumeInfo")
	proto.RegisterType((*CommitID)(nil), "elton.v2.CommitID")
	proto.RegisterType((*CommitInfo)(nil), "elton.v2.CommitInfo")
	proto.RegisterType((*TreeID)(nil), "elton.v2.TreeID")
	proto.RegisterType((*TreeEntry)(nil), "elton.v2.TreeEntry")
	proto.RegisterType((*FileID)(nil), "elton.v2.FileID")
	proto.RegisterType((*File)(nil), "elton.v2.File")
	proto.RegisterType((*FileContentRef)(nil), "elton.v2.FileContentRef")
}

func init() { proto.RegisterFile("types.proto", fileDescriptor_d938547f84707355) }

var fileDescriptor_d938547f84707355 = []byte{
	// 715 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x53, 0xcf, 0x6e, 0xd3, 0x4e,
	0x10, 0xfe, 0x39, 0x71, 0x5c, 0x7b, 0x92, 0xb4, 0xd6, 0xf6, 0x27, 0x64, 0x05, 0x21, 0x22, 0x0b,
	0xa4, 0x88, 0x83, 0x8b, 0xc2, 0x81, 0x3f, 0x27, 0xda, 0x86, 0x48, 0x11, 0x15, 0x45, 0xdb, 0x88,
	0x2b, 0x72, 0xec, 0x49, 0xe2, 0xc6, 0xf6, 0x5a, 0x9b, 0x4d, 0x2b, 0xf3, 0x20, 0xdc, 0x79, 0x45,
	0x9e, 0x00, 0xed, 0x7a, 0x9d, 0x26, 0x6a, 0x45, 0x39, 0x79, 0xe6, 0x9b, 0x6f, 0x66, 0xbe, 0x9d,
	0x19, 0x43, 0x5b, 0x94, 0x05, 0xae, 0x83, 0x82, 0x33, 0xc1, 0x88, 0x8d, 0xa9, 0x60, 0x79, 0x70,
	0x33, 0xec, 0x3d, 0x5f, 0x30, 0xb6, 0x48, 0xf1, 0x44, 0xe1, 0xb3, 0xcd, 0xfc, 0x44, 0x24, 0x19,
	0xae, 0x45, 0x98, 0x15, 0x15, 0xd5, 0x7f, 0x0a, 0xce, 0xe5, 0xec, 0x1a, 0x23, 0xf1, 0x19, 0x4b,
	0x72, 0x08, 0x8d, 0x24, 0xf6, 0x8c, 0xbe, 0x31, 0x70, 0x68, 0x23, 0x89, 0xfd, 0x9f, 0x06, 0x40,
	0x15, 0x9d, 0xe4, 0x73, 0x46, 0x08, 0x98, 0xcb, 0x70, 0xbd, 0x54, 0x84, 0x0e, 0x55, 0x36, 0x79,
	0x01, 0x5d, 0xf9, 0x3d, 0x4d, 0x17, 0x8c, 0x27, 0x62, 0x99, 0x79, 0xa6, 0xca, 0xde, 0x07, 0xc9,
	0x07, 0x80, 0x88, 0x63, 0x28, 0x70, 0x9a, 0x64, 0xe8, 0x35, 0xfa, 0xc6, 0xa0, 0x3d, 0xec, 0x05,
	0x95, 0xb6, 0xa0, 0xd6, 0x16, 0x4c, 0x6b, 0x6d, 0x74, 0x87, 0x2d, 0xbb, 0xae, 0x93, 0x1f, 0xe8,
	0x35, 0xfb, 0xc6, 0xc0, 0xa4, 0xca, 0xf6, 0x3f, 0xd6, 0xba, 0xce, 0x58, 0x5c, 0x92, 0x1e, 0xd8,
	0x11, 0xcb, 0x05, 0xe6, 0x62, 0xad, 0xb5, 0x6d, 0x7d, 0xf2, 0x04, 0x2c, 0x36, 0x9f, 0xaf, 0x51,
	0xa8, 0xae, 0x26, 0xd5, 0x9e, 0xff, 0x0c, 0xda, 0x5f, 0x39, 0x2b, 0x90, 0x8b, 0xf2, 0xa1, 0x97,
	0x8f, 0xa1, 0x53, 0x87, 0x55, 0x0b, 0x02, 0xe6, 0x8c, 0xc5, 0xa5, 0x66, 0x28, 0x9b, 0xf8, 0xd0,
	0x09, 0xd3, 0x94, 0xdd, 0x52, 0x2c, 0xd2, 0x30, 0xaa, 0x9e, 0x65, 0xd3, 0x3d, 0xcc, 0xf7, 0xc0,
	0xfa, 0xc2, 0x62, 0x9c, 0x8c, 0xee, 0x75, 0xb8, 0x00, 0x53, 0x46, 0x88, 0x07, 0x07, 0x61, 0x1c,
	0x73, 0x5c, 0x4b, 0xed, 0xcd, 0x81, 0x43, 0x6b, 0x57, 0xf6, 0xcc, 0x43, 0x3d, 0x2e, 0x87, 0x2a,
	0x5b, 0x3e, 0x67, 0x53, 0xc8, 0x1d, 0xea, 0x71, 0x68, 0xcf, 0xef, 0x81, 0xfd, 0x8d, 0xa5, 0x9b,
	0xec, 0xa1, 0x4e, 0x7d, 0x00, 0x1d, 0xd3, 0x4b, 0x54, 0x55, 0x8d, 0xbb, 0xaa, 0xfe, 0x18, 0xec,
	0x73, 0x96, 0x65, 0x89, 0x98, 0x8c, 0x88, 0xbf, 0xcd, 0x6e, 0x0f, 0x49, 0x50, 0x1f, 0x52, 0x50,
	0x57, 0x97, 0x15, 0xa5, 0x8a, 0x7c, 0x93, 0xcd, 0x90, 0xd7, 0x43, 0xad, 0x3c, 0xff, 0x97, 0x01,
	0xa0, 0x0b, 0xc9, 0x56, 0xef, 0xeb, 0xad, 0xc7, 0xdf, 0x43, 0xa1, 0x4b, 0xfe, 0x6d, 0xeb, 0x8e,
	0x66, 0x9f, 0x0a, 0x12, 0x80, 0x5d, 0x84, 0x1c, 0x73, 0x31, 0x19, 0xe9, 0x73, 0xd9, 0xd1, 0x52,
	0x6b, 0xa5, 0x5b, 0x0e, 0x19, 0x80, 0x25, 0x38, 0xe2, 0x64, 0xa4, 0xe6, 0xd2, 0x1e, 0xba, 0x77,
	0xec, 0xa9, 0xc2, 0xa9, 0x8e, 0xfb, 0x27, 0x60, 0x55, 0x08, 0x79, 0x09, 0xcd, 0x15, 0x96, 0x5a,
	0xd7, 0xf1, 0x5d, 0xc2, 0xf6, 0x7f, 0xa0, 0x32, 0xee, 0x4f, 0xc0, 0x91, 0x09, 0x9f, 0x72, 0xc1,
	0xd5, 0x1d, 0x14, 0xa1, 0x58, 0xd6, 0xd3, 0x93, 0xb6, 0xec, 0x3d, 0x4f, 0x52, 0xdc, 0x2a, 0xdd,
	0xe9, 0x3d, 0x56, 0x38, 0xd5, 0x71, 0x79, 0x0d, 0x15, 0x72, 0x6f, 0x47, 0xbf, 0x1b, 0x60, 0xca,
	0x10, 0x79, 0x07, 0xa0, 0x6f, 0x97, 0xe2, 0x5c, 0x6b, 0xf3, 0xf6, 0x0b, 0x9e, 0x6f, 0xe3, 0x74,
	0x87, 0x2b, 0x47, 0x26, 0xdb, 0x4c, 0xcb, 0xa2, 0x3a, 0x99, 0xc3, 0xdd, 0x91, 0x8d, 0x75, 0x84,
	0x6e, 0x39, 0xf2, 0x29, 0x19, 0x8b, 0xab, 0x43, 0xea, 0x52, 0x65, 0x93, 0xff, 0xa1, 0xc5, 0x6e,
	0x73, 0xe4, 0xea, 0x2f, 0xee, 0xd2, 0xca, 0x91, 0xe8, 0x82, 0xb3, 0x4d, 0xe1, 0xb5, 0x2a, 0x54,
	0x39, 0xe4, 0x35, 0xb4, 0x42, 0x75, 0x89, 0xd6, 0xa3, 0x8b, 0xad, 0x88, 0x32, 0x23, 0x53, 0x19,
	0x07, 0x8f, 0x67, 0x64, 0x75, 0x46, 0xa4, 0x32, 0xec, 0xc7, 0x33, 0x14, 0x51, 0x6a, 0xcd, 0xc2,
	0x6b, 0xc6, 0x3d, 0xa7, 0xd2, 0xaa, 0x1c, 0x85, 0x26, 0x39, 0xe3, 0x1e, 0x68, 0x54, 0x3a, 0xfe,
	0x5b, 0x38, 0xdc, 0x9f, 0xe7, 0x3f, 0x9e, 0xc4, 0x2b, 0x01, 0x76, 0x3d, 0x50, 0xd2, 0x86, 0x03,
	0x8a, 0x8b, 0x4d, 0x1a, 0x72, 0xf7, 0x3f, 0xd2, 0x05, 0x67, 0x94, 0x70, 0x8c, 0x04, 0xe3, 0xa5,
	0x6b, 0x10, 0x17, 0x3a, 0x57, 0x65, 0x36, 0x63, 0x69, 0x12, 0x5d, 0x24, 0xf9, 0xca, 0x6d, 0x10,
	0x1b, 0xcc, 0xf1, 0x64, 0x7c, 0xe9, 0x36, 0xc9, 0x31, 0x1c, 0x9d, 0x2f, 0x43, 0x1e, 0x46, 0x02,
	0xf9, 0x08, 0x6f, 0x92, 0x08, 0x5d, 0x93, 0x1c, 0x41, 0xfb, 0x2c, 0x65, 0xd1, 0x4a, 0x03, 0x2d,
	0x02, 0x60, 0x5d, 0xb1, 0x68, 0x85, 0xc2, 0xb5, 0x66, 0x96, 0x7a, 0xf5, 0x9b, 0x3f, 0x01, 0x00,
	0x00, 0xff, 0xff, 0x98, 0x0d, 0xdd, 0x76, 0xeb, 0x05, 0x00, 0x00,
}
