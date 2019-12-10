package elton_v2

import (
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

func (t *Tree) FastValidate() error {
	if t == nil {
		return xerrors.New("tree is nil")
	}
	if len(t.GetInodes()) == 0 {
		return xerrors.New("t.Inodes is empty")
	}
	inodes := t.GetInodes()
	if inodes[t.GetRootIno()] == nil {
		return xerrors.New("root inode is not found")
	}

	// TODO: ちゃんとチェックする
	return nil
}
