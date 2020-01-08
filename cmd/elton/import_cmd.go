package main

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/spf13/cobra"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"golang.org/x/sys/unix"
	"golang.org/x/xerrors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func importFn(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return xerrors.Errorf("invalid args")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	strCid := args[0]
	base := args[1]
	files := args[2:]

	cid, err := elton_v2.ParseCommitID(strCid)
	if err != nil {
		showError(err)
		return nil
	}

	if err := _importFn(ctx, cid, base, files); err != nil {
		showError(err)
	}
	return nil
}

func _importFn(ctx context.Context, cid *elton_v2.CommitID, base string, files []string) error {
	c, err := elton_v2.CommitService()
	if err != nil {
		return xerrors.Errorf("api client: %w", err)
	}
	defer elton_v2.Close(c)

	res, err := c.GetCommit(ctx, &elton_v2.GetCommitRequest{
		Id: cid,
	})
	if err != nil {
		return xerrors.Errorf("get commit: %w", err)
	}

	tree := res.GetInfo().GetTree()
	dirIno, err := searchFile(tree, base)
	if err != nil {
		return xerrors.Errorf("base dir: %w", err)
	}
	dir, ok := tree.Inodes[dirIno]
	if !ok {
		return xerrors.Errorf("base dir: not found inode: ino=%d", dirIno)
	}
	if dir.FileType != elton_v2.FileType_Directory {
		return xerrors.Errorf("not a directory: ino=%d", dirIno)
	}

	for _, file := range files {
		fname := filepath.Base(file)
		if !isValidFileName(fname) {
			return xerrors.Errorf("invalid file name: %s", file)
		}
		stat := &unix.Stat_t{}
		if err := unix.Stat(file, stat); err != nil {
			return xerrors.Errorf("stat(%s): %w", file, err)
		}
		reader, err := os.Open(file)
		if err != nil {
			return err
		}

		if err := putFile(ctx, tree, dir, fname, stat, reader); err != nil {
			return xerrors.Errorf("putFile(%s): %w", file, err)
		}
		if err := reader.Close(); err != nil {
			return err
		}
	}

	// todo: remove unreachable inodes.

	_, err = c.Commit(ctx, &elton_v2.CommitRequest{
		Info: &elton_v2.CommitInfo{
			CreatedAt:    ptypes.TimestampNow(),
			LeftParentID: cid,
			Tree:         tree,
		},
		Id: cid.GetId(),
	})
	if err != nil {
		xerrors.Errorf("commit: %w", err)
	}
	return nil
}

func isValidFileName(file string) bool {
	switch file {
	case ".":
		fallthrough
	case "..":
		fallthrough
	case "/":
		return false
	default:
		return true
	}
}
func putFile(ctx context.Context, tree *elton_v2.Tree, dir *elton_v2.File, name string, stat *unix.Stat_t, r io.Reader) error {
	var ftype elton_v2.FileType
	switch stat.Mode & unix.S_IFMT {
	case unix.S_IFREG:
		ftype = elton_v2.FileType_Regular
	case unix.S_IFLNK:
		ftype = elton_v2.FileType_SymbolicLink
	default:
		return xerrors.Errorf("unsupported file type")
	}

	// todo: may be crash if file is large.
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return xerrors.Errorf("read file: %w", err)
	}

	// todo: use conn cache
	c, err := elton_v2.StorageService()
	if err != nil {
		return xerrors.Errorf("api client: %w", err)
	}
	defer elton_v2.Close(c)

	res, err := c.CreateObject(ctx, &elton_v2.CreateObjectRequest{
		Body: &elton_v2.ObjectBody{
			Contents: body,
		},
	})
	if err != nil {
		return xerrors.Errorf("create object: %w", err)
	}

	if dir.Entries == nil {
		dir.Entries = map[string]uint64{}
	}
	dir.Entries[name] = assignInode(tree, &elton_v2.File{
		ContentRef: &elton_v2.FileContentRef{
			Key: res.GetKey(),
		},
		FileType: ftype,
		Mode:     stat.Mode & ^uint32(unix.S_IFMT),
		Owner:    stat.Uid,
		Group:    stat.Gid,
		Atime:    mustConvertTimestmap(stat.Atim),
		Mtime:    mustConvertTimestmap(stat.Mtim),
		Ctime:    mustConvertTimestmap(stat.Ctim),
		Entries:  nil,
	})
	return nil
}
func assignInode(tree *elton_v2.Tree, file *elton_v2.File) uint64 {
	ino := uint64(1)
	for ; ; ino++ {
		_, ok := tree.Inodes[ino]
		if !ok {
			break
		}
		// This ino is already used.
		// todo: check ino range
	}
	tree.Inodes[ino] = file
	return ino
}
func mustConvertTimestmap(timespec unix.Timespec) *tspb.Timestamp {
	ts, err := ptypes.TimestampProto(time.Unix(timespec.Unix()))
	if err != nil {
		panic(err)
	}
	return ts
}
