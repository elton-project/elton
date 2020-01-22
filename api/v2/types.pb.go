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

// Identify the object.
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

// Metadata for the object.
type ObjectInfo struct {
	// Hash value of the object.  Hash algorithm specified by other field.
	Hash []byte `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	// Supported algorithms:
	//  - "SHA1"
	HashAlgorithm string               `protobuf:"bytes,4,opt,name=hashAlgorithm,proto3" json:"hashAlgorithm,omitempty"`
	CreatedAt     *timestamp.Timestamp `protobuf:"bytes,2,opt,name=createdAt,proto3" json:"createdAt,omitempty"`
	// Size of the object.
	Size                 uint64   `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
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

func (m *ObjectInfo) GetCreatedAt() *timestamp.Timestamp {
	if m != nil {
		return m.CreatedAt
	}
	return nil
}

func (m *ObjectInfo) GetSize() uint64 {
	if m != nil {
		return m.Size
	}
	return 0
}

// Contents of the object.
// If (offset=0 && len(contents)=ObjectInfo.size) is satisfied, it means
// complete data. Otherwise, it means part of data.
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

// Identify the property.
type PropertyID struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PropertyID) Reset()         { *m = PropertyID{} }
func (m *PropertyID) String() string { return proto.CompactTextString(m) }
func (*PropertyID) ProtoMessage()    {}
func (*PropertyID) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{3}
}

func (m *PropertyID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PropertyID.Unmarshal(m, b)
}
func (m *PropertyID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PropertyID.Marshal(b, m, deterministic)
}
func (m *PropertyID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PropertyID.Merge(m, src)
}
func (m *PropertyID) XXX_Size() int {
	return xxx_messageInfo_PropertyID.Size(m)
}
func (m *PropertyID) XXX_DiscardUnknown() {
	xxx_messageInfo_PropertyID.DiscardUnknown(m)
}

var xxx_messageInfo_PropertyID proto.InternalMessageInfo

func (m *PropertyID) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type Property struct {
	Body                 string   `protobuf:"bytes,1,opt,name=body,proto3" json:"body,omitempty"`
	AllowReplace         bool     `protobuf:"varint,2,opt,name=allowReplace,proto3" json:"allowReplace,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Property) Reset()         { *m = Property{} }
func (m *Property) String() string { return proto.CompactTextString(m) }
func (*Property) ProtoMessage()    {}
func (*Property) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{4}
}

func (m *Property) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Property.Unmarshal(m, b)
}
func (m *Property) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Property.Marshal(b, m, deterministic)
}
func (m *Property) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Property.Merge(m, src)
}
func (m *Property) XXX_Size() int {
	return xxx_messageInfo_Property.Size(m)
}
func (m *Property) XXX_DiscardUnknown() {
	xxx_messageInfo_Property.DiscardUnknown(m)
}

var xxx_messageInfo_Property proto.InternalMessageInfo

func (m *Property) GetBody() string {
	if m != nil {
		return m.Body
	}
	return ""
}

func (m *Property) GetAllowReplace() bool {
	if m != nil {
		return m.AllowReplace
	}
	return false
}

// Identify the node.
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

// Identify the volume.
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

// Metadata for the volume.
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

// Identify the commit.
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

// TODO: rename
type CommitInfo struct {
	CreatedAt *timestamp.Timestamp `protobuf:"bytes,1,opt,name=createdAt,proto3" json:"createdAt,omitempty"`
	// 親コミットのID
	// 通常のコミットはleftのみを指定する。
	LeftParentID *CommitID `protobuf:"bytes,2,opt,name=leftParentID,proto3" json:"leftParentID,omitempty"`
	// nil以外の場合は、このコミットはマージコミット。
	// もう一つの親コミットIDを指定する。
	RightParentID        *CommitID `protobuf:"bytes,4,opt,name=rightParentID,proto3" json:"rightParentID,omitempty"`
	Tree                 *Tree     `protobuf:"bytes,5,opt,name=tree,proto3" json:"tree,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
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

func (m *CommitInfo) GetLeftParentID() *CommitID {
	if m != nil {
		return m.LeftParentID
	}
	return nil
}

func (m *CommitInfo) GetRightParentID() *CommitID {
	if m != nil {
		return m.RightParentID
	}
	return nil
}

func (m *CommitInfo) GetTree() *Tree {
	if m != nil {
		return m.Tree
	}
	return nil
}

// Tree keeps encoded data of directory tree structure in the commit.
type Tree struct {
	RootIno              uint64           `protobuf:"varint,3,opt,name=root_ino,json=rootIno,proto3" json:"root_ino,omitempty"`
	Inodes               map[uint64]*File `protobuf:"bytes,4,rep,name=inodes,proto3" json:"inodes,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *Tree) Reset()         { *m = Tree{} }
func (m *Tree) String() string { return proto.CompactTextString(m) }
func (*Tree) ProtoMessage()    {}
func (*Tree) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{11}
}

func (m *Tree) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Tree.Unmarshal(m, b)
}
func (m *Tree) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Tree.Marshal(b, m, deterministic)
}
func (m *Tree) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Tree.Merge(m, src)
}
func (m *Tree) XXX_Size() int {
	return xxx_messageInfo_Tree.Size(m)
}
func (m *Tree) XXX_DiscardUnknown() {
	xxx_messageInfo_Tree.DiscardUnknown(m)
}

var xxx_messageInfo_Tree proto.InternalMessageInfo

func (m *Tree) GetRootIno() uint64 {
	if m != nil {
		return m.RootIno
	}
	return 0
}

func (m *Tree) GetInodes() map[uint64]*File {
	if m != nil {
		return m.Inodes
	}
	return nil
}

// File presents the Linux inode.
type File struct {
	// If file is regular or symlink, a valid reference is set to the contentRef.
	// Otherwise, it is set to the null reference.
	ContentRef *FileContentRef `protobuf:"bytes,1,opt,name=contentRef,proto3" json:"contentRef,omitempty"`
	FileType   FileType        `protobuf:"varint,2,opt,name=fileType,proto3,enum=elton.v2.FileType" json:"fileType,omitempty"`
	// 実際には16bitで十分だが、protocol bufferは16bit
	// integerをサポートしていないため、32bitで表現している。
	Mode  uint32               `protobuf:"varint,3,opt,name=mode,proto3" json:"mode,omitempty"`
	Owner uint32               `protobuf:"varint,4,opt,name=owner,proto3" json:"owner,omitempty"`
	Group uint32               `protobuf:"varint,5,opt,name=group,proto3" json:"group,omitempty"`
	Atime *timestamp.Timestamp `protobuf:"bytes,6,opt,name=atime,proto3" json:"atime,omitempty"`
	Mtime *timestamp.Timestamp `protobuf:"bytes,7,opt,name=mtime,proto3" json:"mtime,omitempty"`
	Ctime *timestamp.Timestamp `protobuf:"bytes,8,opt,name=ctime,proto3" json:"ctime,omitempty"`
	// For device file.
	Major uint32 `protobuf:"varint,9,opt,name=major,proto3" json:"major,omitempty"`
	Minor uint32 `protobuf:"varint,10,opt,name=minor,proto3" json:"minor,omitempty"`
	// For directory.
	Entries              map[string]uint64 `protobuf:"bytes,11,rep,name=entries,proto3" json:"entries,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *File) Reset()         { *m = File{} }
func (m *File) String() string { return proto.CompactTextString(m) }
func (*File) ProtoMessage()    {}
func (*File) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{12}
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

func (m *File) GetEntries() map[string]uint64 {
	if m != nil {
		return m.Entries
	}
	return nil
}

// Reference to file content.
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
	return fileDescriptor_d938547f84707355, []int{13}
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
	proto.RegisterType((*PropertyID)(nil), "elton.v2.PropertyID")
	proto.RegisterType((*Property)(nil), "elton.v2.Property")
	proto.RegisterType((*NodeID)(nil), "elton.v2.NodeID")
	proto.RegisterType((*Node)(nil), "elton.v2.Node")
	proto.RegisterType((*VolumeID)(nil), "elton.v2.VolumeID")
	proto.RegisterType((*VolumeInfo)(nil), "elton.v2.VolumeInfo")
	proto.RegisterType((*CommitID)(nil), "elton.v2.CommitID")
	proto.RegisterType((*CommitInfo)(nil), "elton.v2.CommitInfo")
	proto.RegisterType((*Tree)(nil), "elton.v2.Tree")
	proto.RegisterMapType((map[uint64]*File)(nil), "elton.v2.Tree.InodesEntry")
	proto.RegisterType((*File)(nil), "elton.v2.File")
	proto.RegisterMapType((map[string]uint64)(nil), "elton.v2.File.EntriesEntry")
	proto.RegisterType((*FileContentRef)(nil), "elton.v2.FileContentRef")
}

func init() { proto.RegisterFile("types.proto", fileDescriptor_d938547f84707355) }

var fileDescriptor_d938547f84707355 = []byte{
	// 813 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x54, 0x6d, 0x8b, 0xdb, 0x46,
	0x10, 0xae, 0xec, 0xb5, 0x2d, 0x8f, 0xed, 0x8b, 0xd8, 0x84, 0xa2, 0x3a, 0x85, 0x1a, 0x91, 0xc2,
	0xd1, 0x0f, 0x4a, 0x71, 0x69, 0x7b, 0xe4, 0x53, 0x73, 0xe7, 0x18, 0x7c, 0x0d, 0x4d, 0xd8, 0x1c,
	0xfd, 0x5a, 0x64, 0x69, 0x6c, 0x2b, 0x96, 0xb4, 0x66, 0xb5, 0xbe, 0xa0, 0xfe, 0x8e, 0xfe, 0x8c,
	0xfe, 0xaa, 0x42, 0xff, 0x47, 0xd9, 0x17, 0xd9, 0x12, 0xbd, 0x72, 0xf4, 0xd3, 0xed, 0xcc, 0x33,
	0xcf, 0xcc, 0xa3, 0x99, 0xe7, 0x0c, 0x23, 0x59, 0x1d, 0xb0, 0x0c, 0x0f, 0x82, 0x4b, 0x4e, 0x5d,
	0xcc, 0x24, 0x2f, 0xc2, 0xfb, 0xf9, 0xf4, 0xab, 0x2d, 0xe7, 0xdb, 0x0c, 0x5f, 0xea, 0xfc, 0xfa,
	0xb8, 0x79, 0x29, 0xd3, 0x1c, 0x4b, 0x19, 0xe5, 0x07, 0x53, 0x1a, 0x3c, 0x87, 0xe1, 0xbb, 0xf5,
	0x47, 0x8c, 0xe5, 0xcf, 0x58, 0xd1, 0x0b, 0xe8, 0xa4, 0x89, 0xef, 0xcc, 0x9c, 0xcb, 0x21, 0xeb,
	0xa4, 0x49, 0xf0, 0x87, 0x03, 0x60, 0xd0, 0x55, 0xb1, 0xe1, 0x94, 0x02, 0xd9, 0x45, 0xe5, 0x4e,
	0x17, 0x8c, 0x99, 0x7e, 0xd3, 0x17, 0x30, 0x51, 0x7f, 0x5f, 0x67, 0x5b, 0x2e, 0x52, 0xb9, 0xcb,
	0x7d, 0xa2, 0xd9, 0xed, 0x24, 0xbd, 0x82, 0x61, 0x2c, 0x30, 0x92, 0x98, 0xbc, 0x96, 0x7e, 0x67,
	0xe6, 0x5c, 0x8e, 0xe6, 0xd3, 0xd0, 0x48, 0x0b, 0x6b, 0x69, 0xe1, 0x5d, 0x2d, 0x8d, 0x9d, 0x8b,
	0xd5, 0xcc, 0x32, 0xfd, 0x1d, 0xfd, 0xee, 0xcc, 0xb9, 0x24, 0x4c, 0xbf, 0x83, 0x9f, 0x6a, 0x55,
	0xd7, 0x3c, 0xa9, 0xe8, 0x14, 0xdc, 0x98, 0x17, 0x12, 0x0b, 0x59, 0x5a, 0x65, 0xa7, 0x98, 0x7e,
	0x0e, 0x7d, 0xbe, 0xd9, 0x94, 0x68, 0x86, 0x12, 0x66, 0xa3, 0xe0, 0x4b, 0x80, 0xf7, 0x82, 0x1f,
	0x50, 0xc8, 0x6a, 0xb5, 0xf8, 0xd7, 0x67, 0x5f, 0x83, 0x5b, 0xa3, 0x6a, 0xfe, 0x9a, 0x27, 0x95,
	0x45, 0xf5, 0x9b, 0x06, 0x30, 0x8e, 0xb2, 0x8c, 0x7f, 0x62, 0x78, 0xc8, 0xa2, 0x18, 0x75, 0x6f,
	0x97, 0xb5, 0x72, 0x81, 0x0f, 0xfd, 0x5f, 0x78, 0x82, 0x0f, 0x74, 0x7f, 0x0b, 0x44, 0x21, 0xd4,
	0x87, 0x41, 0x94, 0x24, 0x02, 0x4b, 0x25, 0xbb, 0x7b, 0x39, 0x64, 0x75, 0xa8, 0x66, 0x16, 0x51,
	0x6e, 0xfa, 0x0e, 0x99, 0x7e, 0xab, 0x2f, 0x39, 0x1e, 0xd4, 0xf1, 0xec, 0x26, 0x6c, 0x14, 0x4c,
	0xc1, 0xfd, 0x95, 0x67, 0xc7, 0xfc, 0xa1, 0x49, 0x33, 0x00, 0x8b, 0xd9, 0xeb, 0xe9, 0xae, 0xce,
	0xb9, 0x6b, 0xb0, 0x04, 0xf7, 0x86, 0xe7, 0x79, 0x2a, 0x57, 0x0b, 0x1a, 0x9c, 0xd8, 0xa3, 0x39,
	0x0d, 0x6b, 0x07, 0x85, 0x75, 0x77, 0xd5, 0x51, 0xa9, 0x28, 0x8e, 0xf9, 0x1a, 0x45, 0xbd, 0x4f,
	0x13, 0x05, 0x7f, 0x39, 0x00, 0xb6, 0x91, 0x1a, 0xd5, 0x3a, 0xb7, 0xf3, 0x7f, 0xce, 0xfd, 0x03,
	0x8c, 0x33, 0xdc, 0xc8, 0xf7, 0x91, 0xc0, 0x42, 0xae, 0x16, 0xd6, 0x2b, 0x0d, 0x39, 0xb5, 0x5c,
	0xd6, 0xaa, 0xa3, 0x57, 0x30, 0x11, 0xe9, 0x76, 0x77, 0x26, 0x92, 0xff, 0x24, 0xb6, 0x0b, 0x69,
	0x00, 0x44, 0x0a, 0x44, 0xbf, 0xa7, 0x09, 0x17, 0x67, 0xc2, 0x9d, 0x40, 0x64, 0x1a, 0xbb, 0x25,
	0x6e, 0xd7, 0x23, 0xc1, 0x9f, 0x0e, 0x10, 0x95, 0xa4, 0x5f, 0x80, 0x2b, 0x38, 0x97, 0xbf, 0xa5,
	0x05, 0xb7, 0xd7, 0x18, 0xa8, 0x78, 0x55, 0x70, 0x3a, 0x87, 0x7e, 0x5a, 0xf0, 0x04, 0x4b, 0x9f,
	0xcc, 0xba, 0xfa, 0xb3, 0x5b, 0xfd, 0xc2, 0x95, 0x06, 0xdf, 0x14, 0x52, 0x54, 0xcc, 0x56, 0x4e,
	0x57, 0x30, 0x6a, 0xa4, 0xa9, 0x07, 0xdd, 0x3d, 0x1a, 0xc3, 0x11, 0xa6, 0x9e, 0xf4, 0x05, 0xf4,
	0xee, 0xa3, 0xec, 0x88, 0x76, 0x1b, 0x0d, 0x8d, 0xcb, 0x34, 0x43, 0x66, 0xc0, 0x57, 0x9d, 0x2b,
	0xe7, 0x96, 0xb8, 0x8e, 0xd7, 0xb9, 0x25, 0x6e, 0xc7, 0xeb, 0x06, 0x7f, 0x77, 0x81, 0x28, 0x9c,
	0x5e, 0x01, 0xd8, 0x7f, 0x08, 0x86, 0x1b, 0x7b, 0x0e, 0xbf, 0xdd, 0xe3, 0xe6, 0x84, 0xb3, 0x46,
	0x2d, 0x0d, 0xc1, 0xdd, 0xa4, 0x19, 0xde, 0x55, 0x07, 0x33, 0xfb, 0xa2, 0xb9, 0xd0, 0xa5, 0x45,
	0xd8, 0xa9, 0x46, 0x59, 0x2c, 0xe7, 0x89, 0xb1, 0xe8, 0x84, 0xe9, 0x37, 0x7d, 0x06, 0x3d, 0xfe,
	0xa9, 0x40, 0xa1, 0x2f, 0x32, 0x61, 0x26, 0x50, 0xd9, 0xad, 0xe0, 0xc7, 0x83, 0x5e, 0xfb, 0x84,
	0x99, 0x80, 0x7e, 0x0b, 0xbd, 0x48, 0x7b, 0xbc, 0xff, 0xa8, 0x67, 0x4c, 0xa1, 0x62, 0xe4, 0x9a,
	0x31, 0x78, 0x9c, 0x91, 0xd7, 0x8c, 0x58, 0x33, 0xdc, 0xc7, 0x19, 0xba, 0x50, 0x69, 0xcd, 0xa3,
	0x8f, 0x5c, 0xf8, 0x43, 0xa3, 0x55, 0x07, 0x3a, 0x9b, 0x16, 0x5c, 0xf8, 0x60, 0xb3, 0x2a, 0xa0,
	0xdf, 0xc3, 0x00, 0x0b, 0x29, 0x52, 0x2c, 0xfd, 0x91, 0x36, 0xc0, 0xf3, 0xf6, 0xc2, 0xc2, 0x37,
	0x06, 0x35, 0x0e, 0xa8, 0x6b, 0xa7, 0xaf, 0x60, 0xdc, 0x04, 0x9a, 0x1e, 0x18, 0x1a, 0x0f, 0x3c,
	0x6b, 0x7a, 0x80, 0x34, 0x6e, 0x1e, 0xfc, 0x08, 0x17, 0xed, 0x13, 0xd2, 0xaf, 0xcf, 0xec, 0xd1,
	0xfc, 0xe9, 0x59, 0xc0, 0xe9, 0x87, 0x5e, 0xb7, 0xfc, 0x46, 0x82, 0x5b, 0xdf, 0x90, 0x8e, 0x60,
	0xc0, 0x70, 0x7b, 0xcc, 0x22, 0xe1, 0x7d, 0x46, 0x27, 0x30, 0x5c, 0xa4, 0x02, 0x63, 0xc9, 0x45,
	0xe5, 0x39, 0xd4, 0x83, 0xf1, 0x87, 0x2a, 0x5f, 0xf3, 0x2c, 0x8d, 0xdf, 0xa6, 0xc5, 0xde, 0xeb,
	0x50, 0x17, 0xc8, 0x72, 0xb5, 0x7c, 0xe7, 0x75, 0xe9, 0x53, 0x78, 0x72, 0xb3, 0x8b, 0x44, 0x14,
	0x4b, 0x14, 0x0b, 0xbc, 0x4f, 0x63, 0xf4, 0x08, 0x7d, 0x02, 0xa3, 0xeb, 0x8c, 0xc7, 0x7b, 0x9b,
	0xe8, 0x51, 0x80, 0xfe, 0x07, 0x1e, 0xef, 0x51, 0x7a, 0xfd, 0x75, 0x5f, 0x2f, 0xfa, 0xbb, 0x7f,
	0x02, 0x00, 0x00, 0xff, 0xff, 0x3b, 0xe3, 0x27, 0x79, 0xb1, 0x06, 0x00, 0x00,
}
