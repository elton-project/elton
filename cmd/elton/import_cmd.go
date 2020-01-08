package main

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/spf13/cobra"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/unix"
	"golang.org/x/xerrors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
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
	sc, err := elton_v2.StorageService()
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
	builder := newTreeBuilder(sc, tree)
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

	fentries := make(chan *fileEntry, 10)
	eg := errgroup.Group{}
	eg.Go(func() error {
		for _, file := range files {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

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

			fentries <- &fileEntry{
				dir:  dir,
				name: fname,
				stat: stat,
				r:    reader,
			}
		}
		return nil
	})
	eg.Go(func() error {
		failed := false
		results := builder.PutFilesAsync(ctx, fentries, 4)
		for result := range results {
			if result.error == nil {
				continue
			}
			err := xerrors.Errorf("PutFiles(%s): %w", result.Entry.name, result.error)
			showError(err)
			failed = true
		}
		if failed {
			return xerrors.Errorf("failed to PutFiles()")
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return err
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
		return xerrors.Errorf("commit: %w", err)
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

type treeBuilder struct {
	lock sync.Mutex
	sc   elton_v2.StorageServiceClient
	tree *elton_v2.Tree
	ino  uint64
}
type fileEntry struct {
	dir  *elton_v2.File
	name string
	stat *unix.Stat_t
	r    io.ReadCloser
}
type putResult struct {
	error
	Entry *fileEntry
}

func newTreeBuilder(sc elton_v2.StorageServiceClient, tree *elton_v2.Tree) *treeBuilder {
	return &treeBuilder{
		lock: sync.Mutex{},
		sc:   sc,
		tree: tree,
		ino:  1,
	}
}

func (b *treeBuilder) PutFilesAsync(ctx context.Context, in <-chan *fileEntry, workers int) <-chan putResult {
	out := make(chan putResult, 128)
	// Start workers.
	wg := sync.WaitGroup{}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for entry := range in {
				err := b.putFile(ctx, entry.dir, entry.name, entry.stat, entry.r)
				out <- putResult{
					error: err,
					Entry: entry,
				}
			}
		}()
	}
	// Close out when all workers are finished.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func (b *treeBuilder) putFile(ctx context.Context, dir *elton_v2.File, name string, stat *unix.Stat_t, r io.ReadCloser) error {
	defer r.Close()

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

	res, err := b.sc.CreateObject(ctx, &elton_v2.CreateObjectRequest{
		Body: &elton_v2.ObjectBody{
			Contents: body,
		},
	})
	if err != nil {
		return xerrors.Errorf("create object: %w", err)
	}

	b.lock.Lock()
	defer b.lock.Unlock()
	if dir.Entries == nil {
		dir.Entries = map[string]uint64{}
	}
	dir.Entries[name] = b.assignInode(&elton_v2.File{
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
func (b *treeBuilder) assignInode(file *elton_v2.File) uint64 {
	for {
		_, ok := b.tree.Inodes[b.ino]
		if !ok {
			break
		}
		// This ino is already used.
		// todo: check ino range
		b.ino++
	}
	b.tree.Inodes[b.ino] = file
	return b.ino
}
func mustConvertTimestmap(timespec unix.Timespec) *tspb.Timestamp {
	ts, err := ptypes.TimestampProto(time.Unix(timespec.Unix()))
	if err != nil {
		panic(err)
	}
	return ts
}
