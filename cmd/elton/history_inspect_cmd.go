package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"golang.org/x/xerrors"
	"path/filepath"
	"strings"
)

func historyInspectFn(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return xerrors.Errorf("invalid args")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	strCid := args[0]
	files := args[1:]
	cid, err := elton_v2.ParseCommitID(strCid)
	if err != nil {
		showError(err)
		return nil
	}

	if err := _historyInspectFn(ctx, cid, files); err != nil {
		showError(err)
	}
	return nil
}

func _historyInspectFn(ctx context.Context, cid *elton_v2.CommitID, files []string) error {
	c, err := elton_v2.ApiClient{}.CommitService()
	if err != nil {
		return xerrors.Errorf("api client: %w", err)
	}

	res, err := c.GetCommit(ctx, &elton_v2.GetCommitRequest{
		Id: cid,
	})
	if err != nil {
		return xerrors.Errorf("get commit: %w", err)
	}
	info := res.GetInfo()

	if len(files) == 0 {
		fmt.Print(dumpCommitInfo(info))
	} else {
		for _, f := range files {
			s, err := dumpFileInfo(info, f)
			if err != nil {
				showError(err)
				return nil
			}
			fmt.Print(s)
		}
	}
	return nil
}

func dumpCommitInfo(info *elton_v2.CommitInfo) string {
	var buff strings.Builder
	buff.WriteString(fmt.Sprintf("CreatedAt: %s\n", info.GetCreatedAt().String()))
	buff.WriteString(fmt.Sprintf("Left: %s\n", info.GetLeftParentID().ConvertString()))
	buff.WriteString(fmt.Sprintf("Right: %s\n", info.GetRightParentID().ConvertString()))
	buff.WriteString(fmt.Sprintf("RootIno: %d\n", info.GetTree().GetRootIno()))
	buff.WriteString(fmt.Sprintf("Inodes: %d\n", len(info.GetTree().GetInodes())))
	return buff.String()
}
func dumpFileInfo(info *elton_v2.CommitInfo, fpath string) (string, error) {
	ino, err := searchFile(info.GetTree(), fpath)
	if err != nil {
		return "", xerrors.Errorf("dump file info: %w", err)
	}
	inode, ok := info.Tree.Inodes[ino]
	if !ok {
		return "", xerrors.Errorf("not found inode: ino=%d", ino)
	}

	var buff strings.Builder
	buff.WriteString(fmt.Sprintf("Ino: %d\n", ino))
	buff.WriteString(fmt.Sprintf("ContentRef: %s\n", inode.GetContentRef().GetKey().GetId()))
	buff.WriteString(fmt.Sprintf("FileType: %s\n", inode.GetFileType().String()))
	buff.WriteString(fmt.Sprintf("Mode: 0%o\n", inode.GetMode()))
	buff.WriteString(fmt.Sprintf("Owner: %d\n", inode.GetOwner()))
	buff.WriteString(fmt.Sprintf("Groups: %d\n", inode.GetGroup()))
	buff.WriteString(fmt.Sprintf("Atime: %s\n", inode.GetAtime().String()))
	buff.WriteString(fmt.Sprintf("Mtime: %s\n", inode.GetMtime().String()))
	buff.WriteString(fmt.Sprintf("Ctime: %s\n", inode.GetCtime().String()))
	buff.WriteString(fmt.Sprintf("Major: %d\n", inode.GetMajor()))
	buff.WriteString(fmt.Sprintf("Minor: %d\n", inode.GetMinor()))
	buff.WriteString(fmt.Sprintf("Entries:\n"))
	for name, ino := range inode.GetEntries() {
		buff.WriteString(fmt.Sprintf("  %q => %d\n", name, ino))
	}
	return buff.String(), nil
}
func searchFile(tree *elton_v2.Tree, fpath string) (uint64, error) {
	fpath = filepath.Clean(fpath)
	fpath = filepath.ToSlash(fpath)
	if fpath == "/" {
		return tree.RootIno, nil
	}
	components := strings.Split(fpath, "/")
	if components[0] != "" {
		return 0, xerrors.Errorf("file path must start with slash (\"/\")")
	}

	ino := tree.RootIno
	for _, name := range components[1:] {
		inode, ok := tree.Inodes[ino]
		if !ok {
			return 0, xerrors.Errorf("not found inode: ino=%d", ino)
		}

		switch inode.FileType {
		case elton_v2.FileType_Directory:
			newIno, ok := inode.Entries[name]
			if !ok {
				return 0, xerrors.Errorf("not found entry: ino=%d, name=%s", ino, name)
			}
			ino = newIno
		default:
			return 0, xerrors.Errorf("not a directory: ino=%d", ino)
		}
	}
	return ino, nil
}
