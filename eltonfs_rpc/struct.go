// See https://gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/wikis/elton2/eltonfsrpc
package eltonfs_rpc

import (
	"fmt"
	"github.com/golang/protobuf/ptypes"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"golang.org/x/xerrors"
	"strconv"
	"strings"
	"time"
)

type EltonObjectID string
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
		// todo
		Id: &elton_v2.VolumeID{
			Id: components[0],
		},
		Number: n,
	}
}

type TreeID string

func (TreeID) FromGRPC(id *elton_v2.TreeID) TreeID {
	return TreeID(id.Id)
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
	XXX_XDR_ID struct{} `xdrid:"5"`
	Id         string   `xdr:"1"`
}

const EltonObjectBodyStructID = 6

type EltonObjectBody struct {
	XXX_XDR_ID    struct{}  `xdrid:"6"`
	Hash          []byte    `xdr:"1"`
	HashAlgorithm string    `xdr:"2"`
	CreatedAt     time.Time `xdr:"3"`
	Size          uint64    `xdr:"4"`
	Contents      []byte    `xdr:"5"`
}

const CommitInfoStructID = 7

type CommitInfo struct {
	XXX_XDR_ID    struct{}  `xdrid:"7"`
	CreatedAt     time.Time `xdr:"1"`
	LeftParentID  CommitID  `xdr:"2"`
	RightParentID CommitID  `xdr:"3"`
	TreeID        TreeID    `xdr:"4"`
	Tree          *TreeInfo `xdr:"5"`
}

func (CommitInfo) FromGRPC(i *elton_v2.CommitInfo) *CommitInfo {
	var err error
	info := &CommitInfo{}
	info.CreatedAt, err = ptypes.Timestamp(i.GetCreatedAt())
	if err != nil {
		panic(xerrors.Errorf("convert timestamp: %w", err))
	}
	info.LeftParentID = CommitID("").FromGRPC(i.GetLeftParentID())
	info.RightParentID = CommitID("").FromGRPC(i.GetRightParentID())
	info.TreeID = TreeID("").FromGRPC(i.GetTreeID())
	info.Tree = TreeInfo{}.FromGRPC(i.GetTree())
	return info
}

const TreeInfoStructID = 8

type TreeInfo struct {
	XXX_XDR_ID struct{}              `xdrid:"8"`
	P2I        map[string]uint64     `xdr:"2"`
	I2F        map[uint64]*EltonFile `xdr:"3"`
}

func (TreeInfo) FromGRPC(t *elton_v2.Tree) *TreeInfo {
	tree := &TreeInfo{}
	tree.P2I = t.GetP2I()
	// todo: i2f
	return tree
}

const GetObjectRequestStructID = 9

type GetObjectRequest struct {
	XXX_XDR_ID struct{}      `xdrid:"9"`
	ID         EltonObjectID `xdr:"1"`
	Offset     uint64        `xdr:"2"`
}

const GetObjectResponseStructID = 10

type GetObjectResponse struct {
	XXX_XDR_ID struct{}        `xdrid:"8"`
	ID         EltonObjectID   `xdr:"1"`
	Offset     uint64          `xdr:"2"`
	Body       EltonObjectBody `xdr:"3"`
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
}

const CreateCommitResponseStructID = 14

type CreateCommitResponse struct {
	XXX_XDR_ID struct{} `xdrid:"14"`
	ID         CommitID `xdr:"1"`
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
	ObjectID   EltonObjectID `xdrid:"1"`
	FileType   uint8         `xdrid:"2"`
	Mode       uint64        `xdrid:"3"`
	Owner      uint64        `xdrid:"4"`
	Group      uint64        `xdrid:"5"`
	Atime      time.Time     `xdrid:"6"`
	Mtime      time.Time     `xdrid:"7"`
	Ctime      time.Time     `xdrid:"8"`
	Major      uint64        `xdrid:"9"`
	Minor      uint64        `xdrid:"10"`
}

const MaxStructID = 17
