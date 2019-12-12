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
func (id *VolumeID) Equals(other *VolumeID) bool {
	return id.GetId() == other.GetId()
}
func (id *CommitID) Empty() bool {
	return id.GetId().Empty() && id.GetNumber() == 0
}
func (id *CommitID) Equals(other *CommitID) bool {
	return id.GetId().Equals(other.GetId()) && id.GetNumber() == other.GetNumber()
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
func (t *Tree) DeepCopy() *Tree {
	inodes := make(map[uint64]*File, len(t.GetInodes()))
	for ino, f := range t.GetInodes() {
		inodes[ino] = f.DeepCopy()
	}
	return &Tree{
		RootIno: t.GetRootIno(),
		Inodes:  inodes,
	}
}
func (f *File) DeepCopy() *File {
	x := &File{}
	*x = *f
	x.Entries = make(map[string]uint64, len(f.Entries))
	for name, ino := range f.Entries {
		x.Entries[name] = ino
	}
	return x
}
