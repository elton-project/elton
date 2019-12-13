package simple

import (
	"github.com/deckarep/golang-set"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"reflect"
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
			if got := m.diff(tt.args.base, tt.args.other, tt.args.baseT, tt.args.otherT); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("diff() = %v, want %v", got, tt.want)
			}
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
		want   map[uint64][]uint64
	}{
		{
			name: "empty_tree",
			args: args{
				tree: &Tree{
					RootIno: 1,
					Inodes:  map[uint64]*File{},
				},
			},
			want: map[uint64][]uint64{},
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
								"bash": 4,
							},
						},
					},
				},
			},
			want: map[uint64][]uint64{
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
			if got := m.reverseIndex(tt.args.tree); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reverseIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}
