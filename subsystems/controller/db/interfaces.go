package controller_db

import (
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
)

// Stores is accessors for various databases.
type Stores interface {
	MetaStore() MetaStore
	VolumeStore() VolumeStore
	CommitStore() CommitStore
	NodeStore() NodeStore
}

// MetaStore is an interface for properties database.
type MetaStore interface {
	// Get gets a property.
	//
	// Error:
	// - ErrNotFoundProp: If property is not found.
	// - InternalError
	Get(id *PropertyID) (*Property, error)
	// Set sets a property.  If property already exists, check mustCreate and prop.allowReplace field value. See "Error"
	// section for detail.  If property replaced, return the old property.
	//
	// Error:
	// - ErrAlreadyExists: If mustCreate=true and specified property is exists.
	// - ErrNotAllowedReplace: If specified property is exists and existing property is not allowed replace
	//                         (Property.allowReplace=false).
	// - InternalError
	Set(id *PropertyID, prop *Property, mustCreate bool) (old *Property, err error)
}

// VolumeStore is an interface for volumes database.
type VolumeStore interface {
	// Get gets a volume.
	//
	// Error:
	// - ErrNotFoundVolume: If volume is not found.
	// - InternalError
	Get(id *VolumeID) (*VolumeInfo, error)
	// GetByName gets volume ID and volume information by name.
	//
	// Error:
	// - ErrNotFoundVolume: If volume is not found.
	// - InternalError
	GetByName(name string) (*VolumeID, *VolumeInfo, error)
	// Exists checks whether the volume exists.
	//
	// Error:
	// - InternalError
	Exists(id *VolumeID) (bool, error)
	// Delete deletes a volume.
	//
	// Error:
	// - ErrNotFoundVolume: If volume is not found.
	// - InternalError
	Delete(id *VolumeID) error
	// Walk walks all volumes and calling fn for each volume.
	//
	// Error:
	// - InternalError
	Walk(fn func(id *VolumeID, info *VolumeInfo) error) error
	// Create creates a volume.
	//
	// Error:
	// - ErrDupVolumeID: If volume ID is duplicated.
	// - ErrDupVolumeName: If volume name is duplicated.
	// - InternalError
	Create(info *VolumeInfo) (*VolumeID, error)
}

// CommitStore is an interface for commits database.
type CommitStore interface {
	// Get gets a commit information.
	//
	// Error:
	// - NotFoundCommit: If commit is not found.
	// - InternalError
	Get(id *CommitID) (*CommitInfo, error)
	// Exists checks whether the commit exists.
	//
	// Error:
	// - InternalError
	Exists(id *CommitID) (bool, error)
	// Parents gets a parent commit IDs of specified commit.  In normal case, left is set the CommitID and right is nil.
	// If merge commit is specified, left and right is set the CommitID.
	//
	// Error:
	// - ErrNotFoundCommit: If commit is not found.
	// - InternalError
	Parents(id *CommitID) (left *CommitID, right *CommitID, err error)
	// Latest gets the latest CommitID.
	//
	// Error:
	// - ErrNotFoundCommit: If volume has no commits.
	// - InternalError
	Latest(vid *VolumeID) (*CommitID, error)
	// Create creates new commit.  If new commit is based on current latest commit, it updates latest commit to new commit id.
	//
	// Error:
	// - ErrCrossVolumeCommit: If mismatch vid and info.LeftParentID and info.RightParentID.
	// - ErrNotFoundVolume: If specified volume is not found.
	// - ErrInvalidParentCommit: If parent commit ID combination is invalid.
	// - ErrInvalidTree: If specified tree is invalid.
	// - InternalError
	Create(vid *VolumeID, info *CommitInfo, tree *Tree) (*CommitID, error)
	// Tree gets a tree information from the CommitID.
	//
	// Error:
	// - NotFoundCommit : If commit is not found.
	// - ErrNotFoundTree: If tree is not found.
	//                    TODO: コミットはあるのにtreeがない状況 !?
	// - InternalError
	Tree(id *CommitID) (*Tree, error)
}

type NodeStore interface {
	// Register a node.
	//
	// Error:
	// - NodeAlreadyExists
	// - InternalError
	Register(id *NodeID, node *Node) error
	// Unregister a node.
	//
	// Error:
	// - ErrNotFoundNode
	// - InternalError
	Unregister(id *NodeID) error
	// Update a node information inside callback().
	// The callback() should update Node fields during executing callback().  If callback() returns an error, changes
	// are discarded.
	//
	// Error:
	// - ErrNotFoundNode
	// - InternalError
	Update(id *NodeID, callback func(node *Node) error) error
	// List calls walker() for each node.  If walker() returns an error, return immediately it.
	//
	// Error:
	// - InternalError
	List(walker func(id *NodeID, node *Node) error) error
}
