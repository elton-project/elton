package elton_v2

import (
	"github.com/deckarep/golang-set"
	"golang.org/x/xerrors"
)

func (id *ObjectKey) Empty() bool {
	return id.GetId() == ""
}
func (id *PropertyID) Empty() bool {
	return id.GetId() == ""
}
func (id *NodeID) Empty() bool {
	return id.GetId() == ""
}
func (id *VolumeID) Empty() bool {
	return id.GetId() == ""
}
func (id *CommitID) Empty() bool {
	return id.GetId().Empty() && id.GetNumber() == 0
}
func (id *TreeID) Empty() bool {
	return id.GetId() == ""
}
func (id *FileID) Empty() bool {
	return id.GetId() == ""
}

func (t *Tree) FastValidate() error {
	if t == nil {
		return xerrors.New("tree is nil")
	}
	if t.GetP2I() == nil {
		return xerrors.New("P2I is nil")
	}
	if t.GetI2F() == nil {
		return xerrors.New("I2F is nil")
	}

	// Check whether the
	p2i := t.GetP2I()
	inodes1 := mapset.NewThreadUnsafeSet()
	for _, i := range p2i {
		inodes1.Add(i)
	}

	i2f := t.GetI2F()
	inodes2 := mapset.NewThreadUnsafeSet()
	for i := range i2f {
		inodes2.Add(i)
	}

	if !inodes1.Equal(inodes2) {
		for _, i := range p2i {
			if _, ok := i2f[i]; !ok {
				return xerrors.Errorf("no I2F entry: inode=%d", i)
			}
		}
		for i := range i2f {
			if !inodes1.Contains(i) {
				return xerrors.Errorf("unused I2F entry found: inode=%d", i)
			}
		}
	}
	return nil
}
