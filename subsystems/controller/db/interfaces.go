package controller_db

import (
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
)

type VolumeStore interface {
	Get(id *VolumeID) (*VolumeInfo, error)
	Exists(id *VolumeID) (bool, error)
	Delete(id *VolumeID) error
	Walk(func(id *VolumeID, info *VolumeInfo) error) error
}
type CommitStore interface {
	Get(id *CommitID) (*CommitInfo, error)
	Exists(id *CommitID) (bool, error)
	Parents(id *CommitID) (*CommitID, *CommitID, error)
	Latest() (*CommitID, error)
	Create(info *CommitInfo) (*CommitID, error)
}
