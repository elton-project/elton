// See https://gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/wikis/elton2/eltonfsrpc
package eltonfs_rpc

import (
	"fmt"
	"github.com/golang/protobuf/ptypes"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"golang.org/x/xerrors"
	"strconv"
	"strings"
	"time"
)

func timestamp2time(ts *tspb.Timestamp) time.Time {
	t, err := ptypes.Timestamp(ts)
	if err != nil {
		panic(xerrors.Errorf("convert timestamp: %w", err))
	}
	return t
}

func time2timestamp(t time.Time) *tspb.Timestamp {
	ts, err := ptypes.TimestampProto(t)
	if err != nil {
		panic(xerrors.Errorf("convert time: %w", err))
	}
	return ts
}

type EltonObjectID string

func (id EltonObjectID) ToGRPC() *elton_v2.ObjectKey {
	return &elton_v2.ObjectKey{
		Id: string(id),
	}
}

func (EltonObjectID) FromGRPC(key *elton_v2.ObjectKey) EltonObjectID {
	return EltonObjectID(key.GetId())
}

type CommitID string

func (CommitID) FromGRPC(id *elton_v2.CommitID) CommitID {
	return CommitID(fmt.Sprintf("%s:%d", id.GetId(), id.GetNumber()))
}

func (id CommitID) ToGRPC() *elton_v2.CommitID {
	components := strings.SplitN(string(id), ":", 2)
	n, err := strconv.ParseUint(components[1], 10, 64)
	if err != nil {
		panic(xerrors.Errorf("parse int: %w", err))
	}

	return &elton_v2.CommitID{
		Id: &elton_v2.VolumeID{
			Id: components[0],
		},
		Number: n,
	}
}

type VolumeID string

func (id VolumeID) ToGRC() *elton_v2.VolumeID {
	return &elton_v2.VolumeID{
		Id: string(id),
	}
}

const Setup1StructID = 1

type Setup1 struct {
	XXX_XDR_ID      struct{} `xdrid:"1"`
	ClientName      string   `xdr:"1"`
	VersionMajor    uint64   `xdr:"2"`
	VersionMinor    uint64   `xdr:"3"`
	VersionRevision uint64   `xdr:"4"`
}

const Setup2StructID = 2

type Setup2 struct {
	XXX_XDR_ID      struct{} `xdrid:"2"`
	Error           uint64   `xdr:"1"`
	Reason          string   `xdr:"2"`
	ServerName      string   `xdr:"3"`
	VersionMajor    uint64   `xdr:"4"`
	VersionMinor    uint64   `xdr:"5"`
	VersionRevision uint64   `xdr:"6"`
}

const PingStructID = 3

type Ping struct {
	XXX_XDR_ID struct{} `xdrid:"3"`
}

const EltonObjectInfoStructID = 5

type EltonObjectInfo struct {
	XXX_XDR_ID    struct{}  `xdrid:"5"`
	Hash          []byte    `xdr:"1"`
	HashAlgorithm string    `xdr:"2"`
	CreatedAt     time.Time `xdr:"3"`
	Size          uint64    `xdr:"4"`
}

const EltonObjectBodyStructID = 6

type EltonObjectBody struct {
	XXX_XDR_ID struct{} `xdrid:"6"`
	Contents   []byte   `xdr:"1"`
	Offset     uint64   `xdr:"2"`
}

func (EltonObjectBody) FromGRPC(body *elton_v2.ObjectBody) EltonObjectBody {
	return EltonObjectBody{
		Contents: body.GetContents(),
		Offset:   body.GetOffset(),
	}
}

const CommitInfoStructID = 7

type CommitInfo struct {
	XXX_XDR_ID    struct{}  `xdrid:"7"`
	CreatedAt     time.Time `xdr:"1"`
	LeftParentID  CommitID  `xdr:"2"`
	RightParentID CommitID  `xdr:"3"`
	Tree          *TreeInfo `xdr:"5"`
}

func (CommitInfo) FromGRPC(i *elton_v2.CommitInfo) *CommitInfo {
	info := &CommitInfo{}
	info.CreatedAt = timestamp2time(i.GetCreatedAt())
	info.LeftParentID = CommitID("").FromGRPC(i.GetLeftParentID())
	info.RightParentID = CommitID("").FromGRPC(i.GetRightParentID())
	info.Tree = TreeInfo{}.FromGRPC(i.GetTree())
	return info
}

func (info CommitInfo) ToGRPC() *elton_v2.CommitInfo {
	return &elton_v2.CommitInfo{
		CreatedAt:     time2timestamp(info.CreatedAt),
		LeftParentID:  info.LeftParentID.ToGRPC(),
		RightParentID: info.RightParentID.ToGRPC(),
		Tree:          info.Tree.ToGRPC(),
	}
}

const TreeInfoStructID = 8

type TreeInfo struct {
	XXX_XDR_ID struct{}              `xdrid:"8"`
	P2I        map[string]uint64     `xdr:"2"`
	I2F        map[uint64]*EltonFile `xdr:"3"`
}

func (TreeInfo) FromGRPC(t *elton_v2.Tree) *TreeInfo {
	i2f := map[uint64]*EltonFile{}
	for i, f := range t.I2F {
		i2f[i] = EltonFile{}.FromGRPC(f)
	}
	tree := &TreeInfo{}
	tree.P2I = t.GetP2I()
	tree.I2F = i2f
	return tree
}

func (info TreeInfo) ToGRPC() *elton_v2.Tree {
	i2f := map[uint64]*elton_v2.File{}
	for i, f := range info.I2F {
		i2f[i] = f.ToGRPC()
	}
	return &elton_v2.Tree{
		P2I: info.P2I,
		I2F: i2f,
	}
}

const GetObjectRequestStructID = 9

type GetObjectRequest struct {
	XXX_XDR_ID struct{}      `xdrid:"9"`
	ID         EltonObjectID `xdr:"1"`
	Offset     uint64        `xdr:"2"`
	Size       uint64        `xdr:"3"`
}

func (req GetObjectRequest) ToGRPC() *elton_v2.GetObjectRequest {
	return &elton_v2.GetObjectRequest{
		Key:    req.ID.ToGRPC(),
		Offset: req.Offset,
		Size:   req.Size,
	}
}

const GetObjectResponseStructID = 10

type GetObjectResponse struct {
	XXX_XDR_ID struct{}        `xdrid:"8"`
	ID         EltonObjectID   `xdr:"1"`
	Body       EltonObjectBody `xdr:"3"`
}

func (GetObjectResponse) FromGRPC(res *elton_v2.GetObjectResponse) *GetObjectResponse {
	return &GetObjectResponse{
		ID:   EltonObjectID("").FromGRPC(res.GetKey()),
		Body: EltonObjectBody{}.FromGRPC(res.GetBody()),
	}
}

const CreateObjectRequestStructID = 11

type CreateObjectRequest struct {
	XXX_XDR_ID struct{}        `xdrid:"11"`
	Body       EltonObjectBody `xdr:"1"`
}

const CreateObjectResponseStructID = 12

type CreateObjectResponse struct {
	XXX_XDR_ID struct{}      `xdrid:"12"`
	ID         EltonObjectID `xdr:"1"`
}

const CreateCommitRequestStructID = 13

type CreateCommitRequest struct {
	XXX_XDR_ID struct{}   `xdrid:"13"`
	Info       CommitInfo `xdr:"1"`
	VID        VolumeID   `xdr:"2"`
}

func (req CreateCommitRequest) ToGRPC() *elton_v2.CommitRequest {
	return &elton_v2.CommitRequest{
		Info: req.Info.ToGRPC(),
		Id:   req.VID.ToGRC(),
	}
}

const CreateCommitResponseStructID = 14

type CreateCommitResponse struct {
	XXX_XDR_ID struct{} `xdrid:"14"`
	ID         CommitID `xdr:"1"`
}

func (CreateCommitResponse) FromGRPC(res *elton_v2.CommitResponse) *CreateCommitResponse {
	return &CreateCommitResponse{
		ID: CommitID("").FromGRPC(res.Id),
	}
}

const NotifyLatestCommitStructID = 15

type NotifyLatestCommit struct {
	XXX_XDR_ID struct{} `xdrid:"15"`
	ID         CommitID `xdr:"1"`
}

const GetCommitInfoRequestStructID = 16

type GetCommitInfoRequest struct {
	XXX_XDR_ID struct{} `xdrid:"16"`
	ID         CommitID `xdr:"1"`
}

func (req *GetCommitInfoRequest) ToGRPC() *elton_v2.GetCommitRequest {
	return &elton_v2.GetCommitRequest{
		Id: req.ID.ToGRPC(),
	}
}

const GetCommitInfoResponseStructID = 17

type GetCommitInfoResponse struct {
	XXX_XDR_ID struct{}    `xdrid:"17"`
	ID         CommitID    `xdr:"1"`
	Info       *CommitInfo `xdr:"2"`
}

func (GetCommitInfoResponse) FromGRPC(res *elton_v2.GetCommitResponse) *GetCommitInfoResponse {
	return &GetCommitInfoResponse{
		ID:   CommitID("").FromGRPC(res.GetId()),
		Info: CommitInfo{}.FromGRPC(res.GetInfo()),
	}
}

const EltonFileStructID = 18

type EltonFile struct {
	XXX_XDR_ID struct{}      `xdrid:"18"`
	ObjectID   EltonObjectID `xdr:"1"`
	FileType   uint8         `xdr:"2"`
	Mode       uint64        `xdr:"3"`
	Owner      uint64        `xdr:"4"`
	Group      uint64        `xdr:"5"`
	Atime      time.Time     `xdr:"6"`
	Mtime      time.Time     `xdr:"7"`
	Ctime      time.Time     `xdr:"8"`
	Major      uint64        `xdr:"9"`
	Minor      uint64        `xdr:"10"`
}

func (f EltonFile) ToGRPC() *elton_v2.File {
	return &elton_v2.File{
		ContentRef: &elton_v2.FileContentRef{
			Key: f.ObjectID.ToGRPC(),
		},
		FileType: elton_v2.FileType(f.FileType),
		Mode:     uint32(f.Mode),
		Owner:    uint32(f.Owner),
		Group:    uint32(f.Group),
		Atime:    time2timestamp(f.Atime),
		Mtime:    time2timestamp(f.Mtime),
		Ctime:    time2timestamp(f.Ctime),
		Major:    uint32(f.Major),
		Minor:    uint32(f.Minor),
	}
}

func (EltonFile) FromGRPC(f *elton_v2.File) *EltonFile {
	return &EltonFile{
		ObjectID: EltonObjectID("").FromGRPC(f.GetContentRef().GetKey()),
	}
}

const MaxStructID = 17
