package simple

import (
	mapset "github.com/deckarep/golang-set"
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"golang.org/x/xerrors"
	"log"
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
	// Left = InodeDeleted
	{InodeDeleted, InodeDeleted}:     NoConflict,
	{InodeAdded, InodeDeleted}:       Conflict,
	{InodeModified, InodeDeleted}:    Conflict,
	{InodeNotModified, InodeDeleted}: NoConflict,
	// Left = InodeAdded
	{InodeDeleted, InodeAdded}:     Conflict,
	{InodeAdded, InodeAdded}:       NeedCheckContents,
	{InodeModified, InodeAdded}:    NeedCheckContents,
	{InodeNotModified, InodeAdded}: NoConflict,
	// Left = InodeModified
	{InodeDeleted, InodeModified}:     Conflict,
	{InodeAdded, InodeModified}:       NeedCheckContents,
	{InodeModified, InodeModified}:    NeedCheckContents,
	{InodeNotModified, InodeModified}: NoConflict,
	// Left = InodeNotModified
	{InodeDeleted, InodeNotModified}:     NoConflict,
	{InodeAdded, InodeNotModified}:       NoConflict,
	{InodeModified, InodeNotModified}:    NoConflict,
	{InodeNotModified, InodeNotModified}: NoConflict,
}

type InoSlice []uint64

func (s InoSlice) Len() int           { return len(s) }
func (s InoSlice) Less(i, j int) bool { return s[i] < s[j] }
func (s InoSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

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

func (d *Diff) Filter(fn func(ino uint64) bool) *Diff {
	return &Diff{
		Added:    d.filterByFunc(d.Added, fn),
		Deleted:  d.filterByFunc(d.Deleted, fn),
		Modified: d.filterByFunc(d.Modified, fn),
	}
}

func (d *Diff) filterByFunc(set mapset.Set, fn func(ino uint64) bool) mapset.Set {
	out := mapset.NewThreadUnsafeSet()
	for _ino := range set.Iter() {
		ino := _ino.(uint64)
		if fn(ino) {
			out.Add(ino)
		}
	}
	return out
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
	newCurrent := m.shiftIno(latestDiff, currentDiff)

	// Check conflicts.
	if err := m.checkFileConflict(latestDiff, currentDiff, newCurrent); err != nil {
		return nil, err
	}
	if err := m.checkDirConflict(latestDiff, currentDiff, newCurrent); err != nil {
		return nil, err
	}

	rb := m.reverseIndex(m.Base)
	rl := m.reverseIndex(m.Latest)
	rc := m.reverseIndex(m.Current)

	_ = rb
	_ = rl
	_ = rc

	// TODO: Apply merge policies.
	log.Println("[WARN] merge policies check is not implemented")

	return m.mergeTree(latestDiff, currentDiff, newCurrent)
}

// mergeTree creates merged tree by apply currentDiff to latest tree.
func (m *Merger) mergeTree(latestDiff *Diff, currentDiff *Diff, newCurrent *Tree) (*Tree, error) {
	tree := m.Latest.DeepCopy()
	for _ino := range currentDiff.Added.Iter() {
		ino := _ino.(uint64)
		tree.Inodes[ino] = newCurrent.Inodes[ino]
	}
	for _ino := range currentDiff.Modified.Iter() {
		ino := _ino.(uint64)
		if newCurrent.Inodes[ino].FileType == FileType_Directory {
			// Copy directory attributes from current tree.
			i := newCurrent.Inodes[ino].DeepCopy()
			tree.Inodes[ino] = i

			// Apply changes of directory entries.
			nameDiff := newEntryDiff(m.Base.Inodes[ino], newCurrent.Inodes[ino])
			for _name := range nameDiff.Added.Iter() {
				name := _name.(string)
				i.Entries[name] = newCurrent.Inodes[ino].Entries[name]
			}
			for _name := range nameDiff.Deleted.Iter() {
				name := _name.(string)
				delete(i.Entries, name)
			}
			for _name := range nameDiff.Modified.Iter() {
				name := _name.(string)
				i.Entries[name] = newCurrent.Inodes[ino].Entries[name]
			}
		} else { // file
			tree.Inodes[ino] = newCurrent.Inodes[ino]
		}
	}
	for _ino := range currentDiff.Deleted.Iter() {
		ino := _ino.(uint64)
		delete(tree.Inodes, ino)
	}
	return tree, nil
}

// shiftIno shifts inode number (ino) of added inodes to prevent conflict.  m.Current tree is kept original status.
func (m *Merger) shiftIno(latestDiff, currentDiff *Diff) *Tree {
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
	return newCurrent
}
func (m *Merger) inodeSet(tree *Tree) mapset.Set {
	inodes := mapset.NewThreadUnsafeSet()
	for ino := range tree.GetInodes() {
		inodes.Add(ino)
	}
	return inodes
}
func (m *Merger) reverseIndex(tree *Tree) map[uint64]InoSlice {
	rev := map[uint64]InoSlice{}
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
	out := mapset.NewThreadUnsafeSet()
	for _ino := range inodes.Iter() {
		ino := _ino.(uint64)
		bino := base.Inodes[ino]
		oino := other.Inodes[ino]

		if bino == nil {
			panic(xerrors.New("bino is nil"))
		}
		if oino == nil {
			panic(xerrors.New("oino is nil"))
		}
		if bino.FileType != oino.FileType {
			panic(xerrors.Errorf("mismatch file type: bino=%s, oino=%s", bino, oino))
		}

		changed := false
		switch bino.FileType {
		case FileType_Directory:
			// Compare two directory inodes.
			changed = !bino.EqualsDirWithoutContents(oino)
			if !changed {
				changed = !bino.EqualsDirContents(oino)
			}
		default: // files
			// Compare two non-directory files.
			changed = !bino.EqualsFile(oino)
		}
		if changed {
			out.Add(ino)
		}
	}
	return out
}

func (m *Merger) checkFileConflict(latestDiffAll, currentDiffAll *Diff, newCurrent *Tree) error {
	// Filter by file type.
	latestDiff := latestDiffAll.Filter(func(ino uint64) bool {
		return m.Latest.Inodes[ino].FileType != FileType_Directory
	})
	currentDiff := currentDiffAll.Filter(func(ino uint64) bool {
		return newCurrent.Inodes[ino].FileType != FileType_Directory
	})

	return conflictRule{}.CheckConflictRulesFile(latestDiff, currentDiff, m.Latest, newCurrent)
}

func (m *Merger) checkDirConflict(latestDiffAll, currentDiffAll *Diff, newCurrent *Tree) error {
	// Filter by file type.
	latestDiff := latestDiffAll.Filter(func(ino uint64) bool {
		return m.Latest.Inodes[ino].FileType == FileType_Directory
	})
	currentDiff := currentDiffAll.Filter(func(ino uint64) bool {
		return newCurrent.Inodes[ino].FileType == FileType_Directory
	})

	if err := (conflictRule{}).CheckConflictRulesFile(latestDiff, currentDiff, m.Latest, newCurrent); err != nil {
		return err
	}
	return conflictRule{}.CheckConflictRulesDir(latestDiff, currentDiff, m.Base, m.Latest, newCurrent)
}

type conflictRule struct{}

// CheckConflictRulesFile checks conflict of files and directories (attributes only).
func (conflictRule) CheckConflictRulesFile(a, b *Diff, aTree, bTree *Tree) error {
	if inoset := a.Deleted.Intersect(b.Added); inoset.Cardinality() > 0 {
		err := xerrors.Errorf("conflict(del-add): %s", inoset)
		log.Printf("[WARN] %s", err)
		return err
	}
	if inoset := a.Deleted.Intersect(b.Modified); inoset.Cardinality() > 0 {
		return xerrors.Errorf("conflict(del-mod): %s", inoset)
	}
	if inoset := a.Added.Intersect(b.Deleted); inoset.Cardinality() > 0 {
		err := xerrors.Errorf("conflict(add-del): %s", inoset)
		log.Printf("[WARN] %s", err)
		return err
	}
	if inoset := a.Added.Intersect(b.Added); inoset.Cardinality() > 0 {
		err := xerrors.Errorf("conflict(add-add): %s", inoset)
		log.Printf("[WARN] %s", err)
		return err
	}
	if inoset := a.Added.Intersect(b.Modified); inoset.Cardinality() > 0 {
		err := xerrors.Errorf("conflict(add-mod): %s", inoset)
		log.Printf("[WARN] %s", err)
		return err
	}
	if inoset := a.Modified.Intersect(b.Deleted); inoset.Cardinality() > 0 {
		return xerrors.Errorf("conflict(mod-del): %s", inoset)
	}
	if inoset := a.Modified.Intersect(b.Added); inoset.Cardinality() > 0 {
		err := xerrors.Errorf("conflict(mod-add): %s", inoset)
		log.Printf("[WARN] %s", err)
		return err
	}
	if inoset := a.Modified.Intersect(b.Modified); inoset.Cardinality() > 0 {
		for _ino := range inoset.Iter() {
			// TODO: goroutine leak.  ループ中にbreakまたはreturnすると、Iter()内部で作成した1つのgoroutineがリークする。

			ino := _ino.(uint64)
			aino := aTree.Inodes[ino]
			bino := bTree.Inodes[ino]

			if aino.FileType == FileType_Directory {
				if !aino.EqualsDirWithoutContents(bino) {
					// THe result is not same.
					return xerrors.Errorf("conflict(mod-mod): a=%s, b=%s", aino, bino)
				}
			} else {
				if !aino.EqualsFile(bino) {
					// The result is not same.
					return xerrors.Errorf("conflict(mod-mod): a=%s, b=%s", aino, bino)
				}
			}
			// Changed same file by two ways (base->latest and base->current), but the result is same.
			// This changes are should allow.
		}
	}
	return nil
}

// CheckConflictRulesDir checks conflict of directory entries.
func (conflictRule) CheckConflictRulesDir(a, b *Diff, baseTree, aTree, bTree *Tree) error {
	for _ino := range a.Modified.Intersect(b.Modified).Iter() {
		ino := _ino.(uint64)
		baseFile := baseTree.Inodes[ino]
		aFile := aTree.Inodes[ino]
		bFile := bTree.Inodes[ino]

		aDiff := newEntryDiff(baseFile, aFile)
		bDiff := newEntryDiff(baseFile, bFile)

		if nameSet := aDiff.Deleted.Intersect(bDiff.Added); nameSet.Cardinality() > 0 {
			err := xerrors.Errorf("conflict(del-add): base=%s a=%s b=%s conflicted_entries=%s", baseFile, aFile, bFile, nameSet)
			log.Printf("[WARN] %s", err)
			return err
		}
		if nameSet := aDiff.Deleted.Intersect(bDiff.Modified); nameSet.Cardinality() > 0 {
			return xerrors.Errorf("conflict(del-mod): base=%s a=%s b=%s conflicted_entries=%s", baseFile, aFile, bFile, nameSet)
		}
		if nameSet := aDiff.Added.Intersect(bDiff.Deleted); nameSet.Cardinality() > 0 {
			err := xerrors.Errorf("conflict(add-del): base=%s a=%s b=%s conflicted_entries=%s", baseFile, aFile, bFile, nameSet)
			log.Printf("[WARN] %s", err)
			return err
		}
		if nameSet := aDiff.Added.Intersect(bDiff.Added); nameSet.Cardinality() > 0 {
			// todo: 挙動未定
			err := xerrors.Errorf("conflict(add-add): base=%s a=%s b=%s conflicted_entries=%s", baseFile, aFile, bFile, nameSet)
			log.Printf("[WARN] %s", err)
			return err
		}
		if nameSet := aDiff.Added.Intersect(bDiff.Modified); nameSet.Cardinality() > 0 {
			err := xerrors.Errorf("conflict(add-mod): base=%s a=%s b=%s conflicted_entries=%s", baseFile, aFile, bFile, nameSet)
			log.Printf("[WARN] %s", err)
			return err
		}
		if nameSet := aDiff.Modified.Intersect(bDiff.Deleted); nameSet.Cardinality() > 0 {
			return xerrors.Errorf("conflict(mod-del): base=%s a=%s b=%s conflicted_entries=%s", baseFile, aFile, bFile, nameSet)
		}
		if nameSet := aDiff.Modified.Intersect(bDiff.Added); nameSet.Cardinality() > 0 {
			err := xerrors.Errorf("conflict(mod-add): base=%s a=%s b=%s conflicted_entries=%s", baseFile, aFile, bFile, nameSet)
			log.Printf("[WARN] %s", err)
			return err
		}
		if nameSet := aDiff.Modified.Intersect(bDiff.Modified); nameSet.Cardinality() > 0 {
			for _name := range nameSet.Iter() {
				name := _name.(string)
				if aFile.Entries[name] != bFile.Entries[name] {
					// The referenced inode number of directory entries associated "name" is changed.  And it is not match.
					return xerrors.Errorf("conflict(mod-mod): base=%s a=%s b=%s conflicted_entries=%s", baseFile, aFile, bFile, nameSet)
				}
				// The referenced inode number of directory entries associated "name" is changed.  But it is referenced
				// same inode number.
			}
		}
	}
	return nil
}

type entryDiff struct {
	Added    mapset.Set
	Deleted  mapset.Set
	Modified mapset.Set
}

func newEntryDiff(base, changed *File) *entryDiff {
	baseNames := entryDiff{}.nameSet(base)
	changedNames := entryDiff{}.nameSet(changed)

	modified := mapset.NewThreadUnsafeSet()
	for _name := range baseNames.Intersect(changedNames).Iter() {
		name := _name.(string)
		if base.Entries[name] != changed.Entries[name] {
			modified.Add(name)
		}
	}

	return &entryDiff{
		Added:    changedNames.Difference(baseNames),
		Deleted:  baseNames.Difference(changedNames),
		Modified: modified,
	}
}
func (entryDiff) nameSet(f *File) mapset.Set {
	if f.FileType != FileType_Directory {
		err := xerrors.Errorf("f is not directory: %s", f)
		panic(err)
	}

	ent := mapset.NewThreadUnsafeSet()
	for name := range f.Entries {
		ent.Add(name)
	}
	return ent
}
