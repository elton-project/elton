package controller_db

import (
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
)

type MetaStore interface {
	Get(id *PropertyID) (prop *Property, err error)
	Set(id *PropertyID) (old *Property, err error)
}
type VolumeStore interface {
	Get(id *VolumeID) (*VolumeInfo, error)
	GetByName(name string) (*VolumeID, *VolumeInfo, error)
	Exists(id *VolumeID) (bool, error)
	Delete(id *VolumeID) error
	Walk(func(id *VolumeID, info *VolumeInfo) error) error
	Create(info *VolumeInfo) (*VolumeID, error)
}
type CommitStore interface {
	Get(id *CommitID) (*CommitInfo, error)
	Exists(id *CommitID) (bool, error)
	Parents(id *CommitID) (*CommitID, *CommitID, error)
	Latest() (*CommitID, error)
	Create(vid *VolumeID, info *CommitInfo, tree *Tree) (*CommitID, error)
	Tree(id *CommitID) (*Tree, error)
	TreeByTreeID(id *TreeID) (*Tree, error)
}
