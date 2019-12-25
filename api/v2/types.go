package elton_v2

import (
	"fmt"
	"golang.org/x/xerrors"
	"log"
	"strconv"
	"strings"
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
func (id *CommitID) ConvertString() string {
	if id == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s/%d", id.GetId().GetId(), id.GetNumber())
}
func ParseCommitID(s string) (*CommitID, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return nil, xerrors.Errorf("invalid commit id: not found separator")
	}

	id := parts[0]
	num, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, xerrors.Errorf("invalid commit id: %w", err)
	}
	return &CommitID{
		Id: &VolumeID{
			Id: id,
		},
		Number: num,
	}, nil
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
	if f.FileType == FileType_Directory {
		x.Entries = make(map[string]uint64, len(f.Entries))
		for name, ino := range f.Entries {
			x.Entries[name] = ino
		}
	} else if f.Entries != nil {
		log.Println("[WARN] allocated File.Entries to a non-directory inode")
	}
	return x
}
func (t *Tree) maxIno() uint64 {
	max := uint64(0)
	for ino := range t.GetInodes() {
		if max < ino {
			max = ino
		}
	}
	return max
}

// NextIno returns next available inode number.
// If additional arguments are specified, NexIno choose an inode number that does not exist in all specified trees.
func (t *Tree) NextIno(trees ...*Tree) uint64 {
	trees = append(trees, t)
	ino := t.maxIno() + 1
	for {
		found := false
		for _, tree := range trees {
			if tree.Inodes[ino] != nil {
				found = true
				break
			}
		}
		if !found {
			return ino
		}
		ino++
	}
}

type comparableFile struct {
	ContentRef     string
	FileType       FileType
	Mode           uint32
	Owner          uint32
	Group          uint32
	AtimeUnixEpoch int64
	AtimeNsec      int32
	MtimeUnixEpoch int64
	MtimeNsec      int32
	CtimeUnixEpoch int64
	CtimeNsec      int32
	Major          uint32
	Minor          uint32
}

func (f *File) toComparableFile() comparableFile {
	return comparableFile{
		ContentRef:     f.GetContentRef().GetKey().GetId(),
		FileType:       f.GetFileType(),
		Mode:           f.GetMode(),
		Owner:          f.GetOwner(),
		Group:          f.GetGroup(),
		AtimeUnixEpoch: f.GetAtime().GetSeconds(),
		AtimeNsec:      f.GetAtime().GetNanos(),
		MtimeUnixEpoch: f.GetMtime().GetSeconds(),
		MtimeNsec:      f.GetMtime().GetNanos(),
		CtimeUnixEpoch: f.GetCtime().GetSeconds(),
		CtimeNsec:      f.GetCtime().GetNanos(),
		Major:          f.GetMajor(),
		Minor:          f.GetMinor(),
	}
}
func (a *File) EqualsFile(b *File) bool {
	if a.GetFileType() == FileType_Directory || b.GetFileType() == FileType_Directory {
		log.Println("[WARN] Trying to compare directory as non-directory file")
	}
	if len(a.GetEntries()) > 0 || len(b.GetEntries()) > 0 {
		log.Println("[WARN] Contains directory entries in non-directory file")
	}
	return a.toComparableFile() == b.toComparableFile()
}

// EqualsDirWithoutContents compares two directory files and returns result.
// This comparision function ignores changes of mtime and entries fields.  The caller should compare those fields.
func (a *File) EqualsDirWithoutContents(b *File) bool {
	if a.GetFileType() != FileType_Directory || b.GetFileType() != FileType_Directory {
		log.Println("[WARN] Trying to compare non-directory file as directory")
	}

	a2 := a.toComparableFile()
	b2 := b.toComparableFile()

	// MUST ignore changes of mtime.
	a2.MtimeUnixEpoch = 0
	a2.MtimeNsec = 0
	b2.MtimeUnixEpoch = 0
	b2.MtimeNsec = 0

	return a2 == b2
}

// EqualsDirContents compares contained files of two directories.
func (a *File) EqualsDirContents(b *File) bool {
	if a.GetFileType() != FileType_Directory || b.GetFileType() != FileType_Directory {
		log.Println("[WARN] Trying to compare non-directory file as directory")
	}

	if len(a.Entries) != len(b.Entries) {
		// Fast path
		return false
	}
	// Slow path
	for k, v := range a.Entries {
		if v != b.Entries[k] {
			return false
		}
	}
	return true
}
