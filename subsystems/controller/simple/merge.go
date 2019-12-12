package simple

import (
	mapset "github.com/deckarep/golang-set"
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"golang.org/x/xerrors"
)

type ModificationType uint8

const (
	InodeNotModified ModificationType = iota
	InodeAdded
	InodeDeleted
	InodeModified
)

type ConflictType uint8

const (
	_ ConflictType = iota
	NoConflict
	Conflict
	NeedCheckContents
)

type fileConflictKey struct{ Right, Left ModificationType }

var fileConflictTable = map[fileConflictKey]ConflictType{
	{InodeDeleted, InodeDeleted}:     NoConflict,
	{InodeAdded, InodeDeleted}:       Conflict,
	{InodeModified, InodeDeleted}:    Conflict,
	{InodeNotModified, InodeDeleted}: NoConflict,

	{InodeDeleted, InodeAdded}:     Conflict,
	{InodeAdded, InodeAdded}:       NeedCheckContents,
	{InodeModified, InodeAdded}:    NeedCheckContents,
	{InodeNotModified, InodeAdded}: NoConflict,

	{InodeDeleted, InodeModified}:     Conflict,
	{InodeAdded, InodeModified}:       NeedCheckContents,
	{InodeModified, InodeModified}:    NeedCheckContents,
	{InodeNotModified, InodeModified}: NoConflict,

	{InodeDeleted, InodeNotModified}:     NoConflict,
	{InodeAdded, InodeNotModified}:       NoConflict,
	{InodeModified, InodeNotModified}:    NoConflict,
	{InodeNotModified, InodeNotModified}: NoConflict,
}

type Diff struct {
	Added    mapset.Set
	Deleted  mapset.Set
	Modified mapset.Set
}

func (d *Diff) Changed() mapset.Set {
	return d.Added.Union(d.Deleted).Union(d.Modified)
}
func (d *Diff) HowChanges(ino uint64) ModificationType {
	if d.Added.Contains(ino) {
		return InodeAdded
	}
	if d.Deleted.Contains(ino) {
		return InodeDeleted
	}
	if d.Modified.Contains(ino) {
		return InodeModified
	}
	return InodeNotModified
}
func (d *Diff) HowChangesDir(ino uint64) ModificationType {
	panic("todo")
}

// Merger merges Latest tree and Current tree by the 3 way merge algorithm.
type Merger struct {
	Info    *CommitInfo
	Base    *Tree
	Latest  *Tree
	Current *Tree
}

func (m *Merger) Merge() (*Tree, error) {
	b := m.inodeSet(m.Base)
	l := m.inodeSet(m.Latest)
	c := m.inodeSet(m.Current)

	latestDiff := m.diff(b, l, m.Base, m.Latest)
	currentDiff := m.diff(b, c, m.Base, m.Current)

	// Fix inode number to prevent conflict.  Result is stored to newCurrent.  m.Current tree is kept original status.
	newCurrent := m.Current.DeepCopy()
	for _oldIno := range latestDiff.Added.Intersect(currentDiff.Added).Iter() {
		oldIno := _oldIno.(uint64)
		newIno := newCurrent.NextIno(m.Base, m.Latest)

		if !(m.Latest.Inodes[newIno] == nil && newCurrent.Inodes[newIno] == nil) {
			panic("bug")
		}

		// Fix inodes table.
		newCurrent.Inodes[newIno] = newCurrent.Inodes[oldIno]
		delete(newCurrent.Inodes, oldIno)

		// Fix directory entries.
		for _, inode := range newCurrent.Inodes {
			if inode.FileType == FileType_Directory {
				for name, to := range inode.Entries {
					if to == oldIno {
						inode.Entries[name] = newIno
					}
				}
			}
		}
	}

	// Check conflict.
	for _ino := range latestDiff.Changed().Union(currentDiff.Changed()).Iter() {
		ino := _ino.(uint64)

		lino := m.Latest.Inodes[ino]
		cino := newCurrent.Inodes[ino]
		if lino.FileType != cino.FileType {
			err := xerrors.Errorf("changed file type: ino=%d, latest=%s, current=%s", ino, m.Latest.Inodes[ino], newCurrent.Inodes[ino])
			panic(err)
		}

		switch lino.FileType {
		case FileType_Directory:
			// Check changes to directory inode are acceptable.
			c1 := latestDiff.HowChangesDir(ino)
			c2 := currentDiff.HowChangesDir(ino)
			switch fileConflictTable[fileConflictKey{c1, c2}] {
			case NoConflict:
				// ok
			case Conflict:
				// todo: エラーメッセージを詳細にする
				return nil, xerrors.Errorf("conflict")
			case NeedCheckContents:
				// todo
				panic("todo")
			default:
				panic("todo")
			}
			// Check directory entries.
			// todo

		default: // files
			c1 := latestDiff.HowChanges(ino)
			c2 := currentDiff.HowChanges(ino)
			switch fileConflictTable[fileConflictKey{c1, c2}] {
			case NoConflict:
				// OK
			case Conflict:
				// todo: エラーメッセージを詳細にする
				return nil, xerrors.Errorf("conflict")
			case NeedCheckContents:
				// todo: 何を確認すれば良いのか？
				panic("todo")
			default:
				panic("todo")
			}
		}

	}

	rb := m.reverseIndex(m.Base)
	rl := m.reverseIndex(m.Latest)
	rc := m.reverseIndex(m.Current)

	_ = rb
	_ = rl
	_ = rc

	// Apply merge policies.
	// todo

	// Create merged tree by apply currentDiff.
	// todo
	panic("todo")
}
func (m *Merger) inodeSet(tree *Tree) mapset.Set {
	inodes := mapset.NewThreadUnsafeSet()
	for ino := range tree.GetInodes() {
		inodes.Add(ino)
	}
	return inodes
}
func (m *Merger) reverseIndex(tree *Tree) map[uint64][]uint64 {
	rev := map[uint64][]uint64{}
	for ino, f := range tree.GetInodes() {
		if f.GetFileType() == FileType_Directory {
			for _, ent := range f.GetEntries() {
				rev[ent] = append(rev[ent], ino)
			}
		}
	}
	return rev
}
func (m *Merger) diff(base, other mapset.Set, baseT, otherT *Tree) *Diff {
	return &Diff{
		Added:    other.Difference(base),                                          // other - base = added inodes
		Deleted:  base.Difference(other),                                          // base - other = deleted inodes
		Modified: m.filterNotModifiedInodes(base.Intersect(other), baseT, otherT), // filter(base & other)
	}
}
func (m *Merger) filterNotModifiedInodes(inodes mapset.Set, base, other *Tree) mapset.Set {
	b := base.GetInodes()
	o := other.GetInodes()

	out := mapset.NewThreadUnsafeSet()
	for ino := range inodes.Iter() {
		_ = b
		_ = o
		_ = ino
		// TODO: fileの比較処理
	}
	return out
}
