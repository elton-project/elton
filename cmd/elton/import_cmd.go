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
	"path"
	"path/filepath"
	"strings"
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
	filesCh := make(chan string, 10)
	results, err := builder.PutFilesAsync(ctx, base, filesCh, 8)
	if err != nil {
		return xerrors.Errorf("PutFiles: %w", err)
	}
	eg := errgroup.Group{}
	eg.Go(func() error {
		for _, file := range files {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case filesCh <- file:
			}
		}
		close(filesCh)
		return nil
	})
	eg.Go(func() error {
		failed := false
		for result := range results {
			if result.error != nil {
				showError(result.error)
				failed = true
			}
		}
		if failed {
			return xerrors.Errorf("some tasks are failed")
		}
		return nil
	})
	eg.Wait()

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

// Context of putting files or directories to the tree.
type treePutter struct {
	*treeBuilder
	ctx context.Context
	// Queue for import requests.
	// If you want to send, MUST add reqWg counter before sending it.
	reqCh    chan putRequest
	entryCh  chan *fileEntry
	resultCh chan putResult
	// Wait for all goroutines of putter.
	wg sync.WaitGroup
	// Wait for entryCh senders.
	entryWg sync.WaitGroup
	// Wait for putRequest senders and requests.
	reqWg sync.WaitGroup
}
type putRequest struct {
	path string
	dir  *elton_v2.File
}
type fileEntry struct {
	path string
	dir  *elton_v2.File
	name string
	stat *unix.Stat_t
	// Content reader (regular file only)
	r io.ReadCloser
	// Contents of dir (directory only)
	entries []os.FileInfo
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
func (b *treeBuilder) PutFilesAsync(ctx context.Context, base string, in <-chan string, workers int) (<-chan putResult, error) {
	dirIno, err := searchFile(b.tree, base)
	if err != nil {
		return nil, xerrors.Errorf("base dir: %w", err)
	}
	dir, ok := b.tree.Inodes[dirIno]
	if !ok {
		return nil, xerrors.Errorf("base dir: not found inode: ino=%d", dirIno)
	}
	if dir.FileType != elton_v2.FileType_Directory {
		return nil, xerrors.Errorf("not a directory: ino=%d", dirIno)
	}

	p := &treePutter{
		treeBuilder: b,
		ctx:         ctx,
		reqCh:       make(chan putRequest),
		entryCh:     make(chan *fileEntry, 128),
		resultCh:    make(chan putResult, 10),
	}
	p.wg.Add(1)
	p.reqWg.Add(1)
	go func() {
		defer p.wg.Done()
		defer p.reqWg.Done()
		for filePath := range in {
			// Adding request to queue.
			p.reqWg.Add(1)
			p.reqCh <- putRequest{
				path: filePath,
				dir:  dir,
			}
		}
	}()
	p.wg.Add(workers)
	p.entryWg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer p.wg.Done()
			defer p.entryWg.Done()
			for req := range p.reqCh {
				// processRequest may be send putRequests and fileEntry.
				if err := p.processRequest(req); err != nil {
					err = xerrors.Errorf("request(%s): %w", req.path, err)
					p.resultCh <- putResult{
						error: err,
					}
				}
				// Done a request.
				p.reqWg.Done()
			}
		}()
	}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for entry := range p.entryCh {
			err := p.processEntry(entry)
			if err != nil {
				err = xerrors.Errorf("entry(%s): %w", entry.path, err)
			}
			p.resultCh <- putResult{
				error: err,
				Entry: entry,
			}
		}
		close(p.resultCh)
	}()
	go func() {
		p.reqWg.Wait()
		close(p.reqCh)
	}()
	go func() {
		p.entryWg.Wait()
		close(p.entryCh)
	}()
	return p.resultCh, nil
}

func (p *treePutter) processRequest(req putRequest) error {
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	default:
	}

	fname := filepath.Base(req.path)
	if !isValidFileName(fname) {
		return xerrors.Errorf("invalid path name: %s", req.path)
	}
	stat := &unix.Stat_t{}
	if err := unix.Stat(req.path, stat); err != nil {
		return xerrors.Errorf("stat(%s): %w", req.path, err)
	}
	switch stat.Mode & unix.S_IFMT {
	case unix.S_IFDIR:
		entries, err := ioutil.ReadDir(req.path)
		if err != nil {
			return xerrors.Errorf("ReadDir(%s): %w", req.path, err)
		}
		// Should add reqWg to prevent closing the reqCh unexpected timing.
		p.reqWg.Add(1)
		p.entryCh <- &fileEntry{
			path:    req.path,
			dir:     req.dir,
			name:    fname,
			stat:    stat,
			entries: entries,
		}
	case unix.S_IFREG:
		reader, err := os.Open(req.path)
		if err != nil {
			return err
		}
		p.entryCh <- &fileEntry{
			path: req.path,
			dir:  req.dir,
			name: fname,
			stat: stat,
			r:    reader,
		}
	case unix.S_IFLNK:
		dest, err := os.Readlink(req.path)
		if err != nil {
			return err
		}
		p.entryCh <- &fileEntry{
			path: req.path,
			dir:  req.dir,
			name: fname,
			stat: stat,
			r:    ioutil.NopCloser(strings.NewReader(dest)),
		}
	}
	return nil
}
func (p *treePutter) processEntry(entry *fileEntry) error {
	dir := entry.dir
	name := entry.name
	stat := entry.stat

	if entry.r != nil {
		defer entry.r.Close()
	}

	var ftype elton_v2.FileType
	switch stat.Mode & unix.S_IFMT {
	case unix.S_IFREG:
		ftype = elton_v2.FileType_Regular
	case unix.S_IFLNK:
		ftype = elton_v2.FileType_SymbolicLink
	case unix.S_IFDIR:
		ftype = elton_v2.FileType_Directory
	default:
		return xerrors.Errorf("unsupported file type")
	}

	var ref *elton_v2.FileContentRef
	if entry.r != nil {
		// todo: may be crash if file is large.
		body, err := ioutil.ReadAll(entry.r)
		if err != nil {
			return xerrors.Errorf("read file: %w", err)
		}

		res, err := p.sc.CreateObject(p.ctx, &elton_v2.CreateObjectRequest{
			Body: &elton_v2.ObjectBody{
				Contents: body,
			},
		})
		if err != nil {
			return xerrors.Errorf("create object: %w", err)
		}
		ref = &elton_v2.FileContentRef{
			Key: res.GetKey(),
		}
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	if dir.Entries == nil {
		dir.Entries = map[string]uint64{}
	}
	file := &elton_v2.File{
		ContentRef: ref,
		FileType:   ftype,
		Mode:       stat.Mode & ^uint32(unix.S_IFMT),
		Owner:      stat.Uid,
		Group:      stat.Gid,
		Atime:      mustConvertTimestmap(stat.Atim),
		Mtime:      mustConvertTimestmap(stat.Mtim),
		Ctime:      mustConvertTimestmap(stat.Ctim),
		Entries:    nil,
	}
	dir.Entries[name] = p.assignInode(file)

	if ftype == elton_v2.FileType_Directory {
		// Add directory contents.
		// p.reqWg counter already added by processRequest().  Should not add it here.
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			defer p.reqWg.Done()
			for _, ent := range entry.entries {
				// Adding request to queue.
				p.reqWg.Add(1)
				p.reqCh <- putRequest{
					path: path.Join(entry.path, ent.Name()),
					dir:  file,
				}
			}
		}()
	}
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
