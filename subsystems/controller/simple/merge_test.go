package simple

import (
	"fmt"
	"github.com/deckarep/golang-set"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"golang.org/x/xerrors"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"
)

func mustProtoTime(t time.Time) *timestamp.Timestamp {
	ts, err := ptypes.TimestampProto(t)
	if err != nil {
		panic(err)
	}
	return ts
}

type treeBuilder struct {
	Tree
}

func newTreeBuilder() *treeBuilder {
	b := &treeBuilder{}
	b.Tree = Tree{
		RootIno: 1,
		Inodes:  map[uint64]*File{},
	}
	return b
}
func (b *treeBuilder) Dirs(inos ...uint64) *treeBuilder {
	for _, ino := range inos {
		if b.Tree.Inodes[ino] != nil {
			panic(xerrors.Errorf("dir inode(%d) is already added", ino))
		}
		b.Tree.Inodes[ino] = &File{
			FileType: FileType_Directory,
			Entries:  map[string]uint64{},
		}
	}
	return b
}
func (b *treeBuilder) Files(inos ...uint64) *treeBuilder {
	for _, ino := range inos {
		if b.Tree.Inodes[ino] != nil {
			panic(xerrors.Errorf("regular file inode(%d) is already added", ino))
		}
		b.Tree.Inodes[ino] = &File{
			FileType: FileType_Regular,
		}
	}
	return b
}

func (b *treeBuilder) File(ino uint64, mode os.FileMode, s string) *treeBuilder {
	if b.Tree.Inodes[ino] != nil {
		panic(xerrors.Errorf("regular file inode(%d) is already added", ino))
	}
	b.Tree.Inodes[ino] = &File{
		ContentRef: &FileContentRef{
			Key: &ObjectKey{
				Id: "id-" + s,
			},
		},
		FileType: FileType_Regular,
		Mode:     uint32(mode),
	}
	return b
}

func (b *treeBuilder) DirEntry(dirIno uint64, name string, fileIno uint64) *treeBuilder {
	if b.Tree.Inodes[dirIno] == nil {
		panic(xerrors.Errorf("dir inode(%d) is not found", dirIno))
	}
	b.Tree.Inodes[dirIno].Entries[name] = fileIno
	return b
}

type diffBuilder struct {
	Diff
}

func newDiffBuilder() *diffBuilder {
	return &diffBuilder{
		Diff{
			Added:    mapset.NewThreadUnsafeSet(),
			Deleted:  mapset.NewThreadUnsafeSet(),
			Modified: mapset.NewThreadUnsafeSet(),
		},
	}
}
func (b *diffBuilder) Add(inos ...uint64) *diffBuilder {
	for _, ino := range inos {
		b.Diff.Added.Add(ino)
	}
	return b
}
func (b *diffBuilder) Del(inos ...uint64) *diffBuilder {
	for _, ino := range inos {
		b.Diff.Deleted.Add(ino)
	}
	return b
}
func (b *diffBuilder) Modify(inos ...uint64) *diffBuilder {
	for _, ino := range inos {
		b.Diff.Modified.Add(ino)
	}
	return b
}

func TestMerger_filterNotModifiedInodes(t *testing.T) {
	type fields struct {
		Info    *CommitInfo
		Base    *Tree
		Latest  *Tree
		Current *Tree
	}
	type args struct {
		inodes mapset.Set
		base   *Tree
		other  *Tree
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   mapset.Set
	}{
		{
			name: "inogre_inodes_not_included_in_inodes_args",
			args: args{
				inodes: mapset.NewThreadUnsafeSetFromSlice([]interface{}{
					uint64(1),
				}),
				base: &Tree{
					Inodes: map[uint64]*File{
						1: {FileType: FileType_Regular, Mode: 0644},
						2: {FileType: FileType_Regular, Mode: 0644},
					},
				},
				other: &Tree{
					Inodes: map[uint64]*File{
						1: {FileType: FileType_Regular, Mode: 0644},
						2: {FileType: FileType_Regular, Mode: 0777},
					},
				},
			},
			want: mapset.NewThreadUnsafeSetFromSlice([]interface{}{}),
		}, {
			name: "ignore_mtime_changes_of_directories",
			args: args{
				inodes: mapset.NewThreadUnsafeSetFromSlice([]interface{}{
					uint64(1),
					uint64(2),
					uint64(3),
				}),
				base: &Tree{
					Inodes: map[uint64]*File{
						1: {
							FileType: FileType_Directory,
							Atime:    mustProtoTime(time.Unix(10, 20)),
							Mtime:    mustProtoTime(time.Unix(30, 40)),
							Ctime:    mustProtoTime(time.Unix(50, 60)),
							Entries:  map[string]uint64{},
						},
						2: {
							FileType: FileType_Directory,
							Atime:    mustProtoTime(time.Unix(10, 20)),
							Mtime:    mustProtoTime(time.Unix(30, 40)),
							Ctime:    mustProtoTime(time.Unix(50, 60)),
							Entries:  map[string]uint64{},
						},
						3: {
							FileType: FileType_Directory,
							Atime:    mustProtoTime(time.Unix(10, 20)),
							Mtime:    mustProtoTime(time.Unix(30, 40)),
							Ctime:    mustProtoTime(time.Unix(50, 60)),
							Entries:  map[string]uint64{},
						},
					},
				},
				other: &Tree{
					Inodes: map[uint64]*File{
						1: {
							FileType: FileType_Directory,
							Atime:    mustProtoTime(time.Unix(99, 99)), // Changed
							Mtime:    mustProtoTime(time.Unix(30, 40)),
							Ctime:    mustProtoTime(time.Unix(50, 60)),
							Entries:  map[string]uint64{},
						},
						2: {
							FileType: FileType_Directory,
							Atime:    mustProtoTime(time.Unix(10, 20)),
							Mtime:    mustProtoTime(time.Unix(99, 99)), // Changed
							Ctime:    mustProtoTime(time.Unix(50, 60)),
							Entries:  map[string]uint64{},
						},
						3: {
							FileType: FileType_Directory,
							Atime:    mustProtoTime(time.Unix(10, 20)),
							Mtime:    mustProtoTime(time.Unix(30, 40)),
							Ctime:    mustProtoTime(time.Unix(99, 99)), // Changed
							Entries:  map[string]uint64{},
						},
					},
				},
			},
			want: mapset.NewThreadUnsafeSetFromSlice([]interface{}{
				uint64(1),
				uint64(3),
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Merger{
				Info:    tt.fields.Info,
				Base:    tt.fields.Base,
				Latest:  tt.fields.Latest,
				Current: tt.fields.Current,
			}
			if got := m.filterNotModifiedInodes(tt.args.inodes, tt.args.base, tt.args.other); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterNotModifiedInodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMerger_diff(t *testing.T) {
	type fields struct {
		Info    *CommitInfo
		Base    *Tree
		Latest  *Tree
		Current *Tree
	}
	type args struct {
		base   mapset.Set
		other  mapset.Set
		baseT  *Tree
		otherT *Tree
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Diff
	}{
		{
			name: "test_added",
			args: args{
				base:  mapset.NewThreadUnsafeSetFromSlice([]interface{}{}),
				other: mapset.NewThreadUnsafeSetFromSlice([]interface{}{uint64(1)}),
			},
			want: &Diff{
				Added:    mapset.NewThreadUnsafeSetFromSlice([]interface{}{uint64(1)}),
				Deleted:  mapset.NewThreadUnsafeSetFromSlice([]interface{}{}),
				Modified: mapset.NewThreadUnsafeSetFromSlice([]interface{}{}),
			},
		}, {
			name: "test_deleted",
			args: args{
				base:  mapset.NewThreadUnsafeSetFromSlice([]interface{}{uint64(1)}),
				other: mapset.NewThreadUnsafeSetFromSlice([]interface{}{}),
			},
			want: &Diff{
				Added:    mapset.NewThreadUnsafeSetFromSlice([]interface{}{}),
				Deleted:  mapset.NewThreadUnsafeSetFromSlice([]interface{}{uint64(1)}),
				Modified: mapset.NewThreadUnsafeSetFromSlice([]interface{}{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Merger{
				Info:    tt.fields.Info,
				Base:    tt.fields.Base,
				Latest:  tt.fields.Latest,
				Current: tt.fields.Current,
			}
			got := m.diff(tt.args.base, tt.args.other, tt.args.baseT, tt.args.otherT)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMerger_reverseIndex(t *testing.T) {
	type fields struct {
		Info    *CommitInfo
		Base    *Tree
		Latest  *Tree
		Current *Tree
	}
	type args struct {
		tree *Tree
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[uint64]InoSlice
	}{
		{
			name: "empty_tree",
			args: args{
				tree: &Tree{
					RootIno: 1,
					Inodes:  map[uint64]*File{},
				},
			},
			want: map[uint64]InoSlice{},
		}, {
			name: "normal_tree",
			args: args{
				tree: &Tree{
					RootIno: 1,
					Inodes: map[uint64]*File{
						1: { // /
							FileType: FileType_Directory,
							Entries: map[string]uint64{
								"bin": 2,
								"tmp": 5,
							},
						},
						2: { // /bin/
							FileType: FileType_Directory,
							Entries: map[string]uint64{
								"sh":   3,
								"bash": 4,
							},
						},
						3: { // /bin/sh
							FileType: FileType_SymbolicLink,
						},
						4: { // /bin/bash
							FileType: FileType_Regular,
						},
						5: { // /tmp/
							FileType: FileType_Directory,
							Entries: map[string]uint64{
								"bash": 4, // (hard link)
							},
						},
					},
				},
			},
			want: map[uint64]InoSlice{
				2: {1},
				3: {2},
				4: {2, 5},
				5: {1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Merger{
				Info:    tt.fields.Info,
				Base:    tt.fields.Base,
				Latest:  tt.fields.Latest,
				Current: tt.fields.Current,
			}
			got := m.reverseIndex(tt.args.tree)
			for _, parents := range got {
				// The order of inode numbers is not defined.  We should sort it to ensure stability.
				sort.Sort(parents)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reverseIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMerger_shiftIno(t *testing.T) {
	type fields struct {
		Info    *CommitInfo
		Base    *Tree
		Latest  *Tree
		Current *Tree
	}
	type args struct {
		latestDiff  *Diff
		currentDiff *Diff
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Tree
	}{
		{
			name: "no_file_added",
			fields: fields{
				Current: &newTreeBuilder().Dirs(1).Files(2, 4, 5, 6).Tree,
			},
			args: args{
				latestDiff: &Diff{
					Added:    mapset.NewThreadUnsafeSetFromSlice([]interface{}{}),
					Deleted:  mapset.NewThreadUnsafeSetFromSlice([]interface{}{uint64(2), uint64(3)}),
					Modified: mapset.NewThreadUnsafeSetFromSlice([]interface{}{uint64(4), uint64(5)}),
				},
				currentDiff: &Diff{
					Added:    mapset.NewThreadUnsafeSetFromSlice([]interface{}{}),
					Deleted:  mapset.NewThreadUnsafeSetFromSlice([]interface{}{uint64(3)}),
					Modified: mapset.NewThreadUnsafeSetFromSlice([]interface{}{uint64(4), uint64(6)}),
				},
			},
			want: &newTreeBuilder().Dirs(1).Files(2, 4, 5, 6).Tree,
		},
		{
			name: "add_inodes_with_same_ino",
			fields: fields{
				Base:    &newTreeBuilder().Dirs(1).Files(2).Tree,
				Latest:  &newTreeBuilder().Dirs(1).Files(2, 3, 4).Tree,
				Current: &newTreeBuilder().Dirs(1).Files(2, 3, 4).Tree,
			},
			args: args{
				latestDiff: &Diff{
					Added:    mapset.NewThreadUnsafeSetFromSlice([]interface{}{uint64(3), uint64(4)}),
					Deleted:  mapset.NewThreadUnsafeSetFromSlice([]interface{}{}),
					Modified: mapset.NewThreadUnsafeSetFromSlice([]interface{}{}),
				},
				currentDiff: &Diff{
					Added:    mapset.NewThreadUnsafeSetFromSlice([]interface{}{uint64(3), uint64(4)}),
					Deleted:  mapset.NewThreadUnsafeSetFromSlice([]interface{}{}),
					Modified: mapset.NewThreadUnsafeSetFromSlice([]interface{}{}),
				},
			},
			// 3 and 4 are shifted to 5 and 6.
			want: &newTreeBuilder().Dirs(1).Files(2, 5, 6).Tree,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Merger{
				Info:    tt.fields.Info,
				Base:    tt.fields.Base,
				Latest:  tt.fields.Latest,
				Current: tt.fields.Current,
			}
			got := m.shiftIno(tt.args.latestDiff, tt.args.currentDiff)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_newEntryDiff(t *testing.T) {
	type args struct {
		base    *File
		changed *File
	}
	tests := []struct {
		name string
		args args
		want *entryDiff
	}{
		{
			name: "no_change",
			args: args{
				base: &File{
					FileType: FileType_Directory,
					Entries: map[string]uint64{
						"foo": 1,
						"bar": 2,
					},
				},
				changed: &File{
					FileType: FileType_Directory,
					Entries: map[string]uint64{
						"foo": 1,
						"bar": 2,
					},
				},
			},
			want: &entryDiff{
				Added:    mapset.NewThreadUnsafeSetFromSlice(nil),
				Deleted:  mapset.NewThreadUnsafeSetFromSlice(nil),
				Modified: mapset.NewThreadUnsafeSetFromSlice(nil),
			},
		}, {
			name: "add_and_delete",
			args: args{
				base: &File{
					FileType: FileType_Directory,
					Entries: map[string]uint64{
						"foo": 1,
					},
				},
				changed: &File{
					FileType: FileType_Directory,
					Entries: map[string]uint64{
						"bar": 2,
					},
				},
			},
			want: &entryDiff{
				Added:    mapset.NewThreadUnsafeSetFromSlice([]interface{}{"bar"}),
				Deleted:  mapset.NewThreadUnsafeSetFromSlice([]interface{}{"foo"}),
				Modified: mapset.NewThreadUnsafeSetFromSlice(nil),
			},
		}, {
			name: "change_referenced_inode_number",
			args: args{
				base: &File{
					FileType: FileType_Directory,
					Entries: map[string]uint64{
						"foo": 1,
					},
				},
				changed: &File{
					FileType: FileType_Directory,
					Entries: map[string]uint64{
						"foo": 2,
					},
				},
			},
			want: &entryDiff{
				Added:    mapset.NewThreadUnsafeSetFromSlice(nil),
				Deleted:  mapset.NewThreadUnsafeSetFromSlice(nil),
				Modified: mapset.NewThreadUnsafeSetFromSlice([]interface{}{"foo"}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newEntryDiff(tt.args.base, tt.args.changed)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_conflictRule_CheckConflictRulesFile(t *testing.T) {
	type OP string
	const (
		addOP OP = "add"
		delOP OP = "del"
		modOP OP = "mod"
		noOP  OP = "nop"
	)
	type args struct {
		a     *Diff
		b     *Diff
		aTree *Tree
		bTree *Tree
	}
	tests := []struct {
		aOP OP
		bOP OP
		OK  bool
	}{
		// del-*
		{delOP, delOP, true},
		{delOP, addOP, false},
		{delOP, modOP, false},
		{delOP, noOP, true},
		// add-*
		{addOP, delOP, false},
		{addOP, addOP, false},
		{addOP, modOP, false},
		{addOP, noOP, true},
		// mod-*
		{modOP, delOP, false},
		{modOP, addOP, false},
		{modOP, modOP, true}, // NOTE: このテストでは、変更後のファイルが同一なので常にtrueになる。
		{modOP, noOP, true},
		// noop-*
		{noOP, delOP, true},
		{noOP, addOP, true},
		{noOP, modOP, true},
		{noOP, noOP, true},
	}

	op2diffTree := func(op OP) (*Diff, *Tree) {
		d := newDiffBuilder()
		t := newTreeBuilder().Dirs(1)
		switch op {
		case addOP:
			d.Add(2)
			t.Files(2)
		case delOP:
			d.Del(2)
		case modOP:
			d.Modify(2)
			t.Files(2)
		case noOP:
			t.Files(2)
		default:
			panic(xerrors.Errorf("unexpected op: %s", op))
		}
		return &d.Diff, &t.Tree
	}
	op2args := func(aOP, bOP OP) args {
		a, at := op2diffTree(aOP)
		b, bt := op2diffTree(bOP)
		return args{
			a:     a,
			b:     b,
			aTree: at,
			bTree: bt,
		}
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s-%s", tt.aOP, tt.bOP)
		t.Run(name, func(t *testing.T) {
			arg := op2args(tt.aOP, tt.bOP)
			co := conflictRule{}
			t.Logf("arg: %+v", arg)
			err := co.CheckConflictRulesFile(arg.a, arg.b, arg.aTree, arg.bTree)
			if tt.OK {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}

	t.Run("mod-mod/mismatch", func(t *testing.T) {
		a := &newDiffBuilder().Modify(2).Diff
		b := &newDiffBuilder().Modify(2).Diff
		at := &newTreeBuilder().Dirs(1).File(2, 0644, "foo").Tree
		bt := &newTreeBuilder().Dirs(1).File(2, 0644, "bar").Tree
		err := conflictRule{}.CheckConflictRulesFile(a, b, at, bt)
		assert.Error(t, err)
	})
}

func Test_conflictRule_CheckConflictRulesDir(t *testing.T) {
	type OP string
	const (
		addOP OP = "add"
		delOP OP = "del"
		modOP OP = "mod"
		noOP  OP = "nop"
	)
	type args struct {
		a        *Diff
		b        *Diff
		baseTree *Tree
		aTree    *Tree
		bTree    *Tree
	}
	tests := []struct {
		aOP OP
		bOP OP
		OK  bool
	}{
		// del-*
		{delOP, delOP, true},
		// This situation is not occur.
		// {delOP, addOP, false},
		{delOP, modOP, false},
		{delOP, noOP, true},
		// add-*
		// This situation is not occur.
		// {addOP, delOP, false},
		{addOP, addOP, false},
		{addOP, modOP, false},
		{addOP, noOP, true},
		// mod-*
		{modOP, delOP, false},
		{modOP, addOP, false},
		{modOP, modOP, true}, // NOTE: このテストでは、変更後のファイルが同一なので常にtrueになる。
		{modOP, noOP, true},
		// noop-*
		{noOP, delOP, true},
		{noOP, addOP, true},
		{noOP, modOP, true},
		{noOP, noOP, true},
	}

	op2diffTree := func(op OP) (*Diff, *Tree) {
		d := newDiffBuilder()
		t := newTreeBuilder().Dirs(1)
		switch op {
		case addOP:
			d.Modify(1)
			t.DirEntry(1, "foo", 2)
			t.DirEntry(1, "bar", 3).Files(2) // Add "bar" file.
		case delOP:
			d.Modify(1)
			// Delete "foo" file.
		case modOP:
			d.Modify(1)
			t.DirEntry(1, "foo", 3).Files(2) // Change "foo" file.
		case noOP:
			t.DirEntry(1, "foo", 2).Files(2)
		default:
			panic(xerrors.Errorf("unexpected op: %s", op))
		}
		return &d.Diff, &t.Tree
	}
	op2args := func(aOP, bOP OP) args {
		a, at := op2diffTree(aOP)
		b, bt := op2diffTree(bOP)
		return args{
			a:        a,
			b:        b,
			baseTree: &newTreeBuilder().Dirs(1).Tree,
			aTree:    at,
			bTree:    bt,
		}
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s-%s", tt.aOP, tt.bOP)
		t.Run(name, func(t *testing.T) {
			arg := op2args(tt.aOP, tt.bOP)
			co := conflictRule{}
			err := co.CheckConflictRulesDir(arg.a, arg.b, arg.baseTree, arg.aTree, arg.bTree)
			if tt.OK {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}

	t.Run("mod-mod/mismatch", func(t *testing.T) {
		a := &newDiffBuilder().Modify(1).Diff
		b := &newDiffBuilder().Modify(1).Diff
		base := &newTreeBuilder().Dirs(1).Tree
		at := &newTreeBuilder().Dirs(1).DirEntry(1, "foo", 3).Files(3).Tree
		bt := &newTreeBuilder().Dirs(1).DirEntry(1, "bar", 4).Files(4).Tree
		err := conflictRule{}.CheckConflictRulesDir(a, b, base, at, bt)
		assert.Error(t, err)
	})
}
