package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

func NewEltonFsFromRoots(roots []string, opts *Options) (pathfs.FileSystem, error) {
	fses := make([]pathfs.FileSystem, 0)
	for i, r := range roots {
		var fs pathfs.FileSystem
		fi, err := os.Stat(r)
		if err != nil {
			return nil, err
		}

		if fi.IsDir() {
			fs = NewLoopbackFileSystem(r, *opts)
		}

		if fs == nil {
			return nil, err
		}

		if i > 0 {
			fs = NewGrpcFileSystem(fs, 0)
		}

		if fs == nil {
			return nil, err
		}

		fses = append(fses, fs)
	}

	server, err := NewEltonFSGrpcServer(roots[1], opts)
	if err != nil {
		return nil, err
	}

	return NewEltonFs(fses, *opts, server)
}

func filePathHash(path string) string {
	dir, base := filepath.Split(path)

	h := md5.New()
	h.Write([]byte(dir))
	return fmt.Sprintf("%x-%s", h.Sum(nil)[:8], base)
}

type eltonFS struct {
	pathfs.FileSystem
	fileSystems   []pathfs.FileSystem
	deletionCache *dirCache
	branchCache   *TimedCache
	opts          *Options
	nodeFs        *pathfs.PathNodeFs
	server        *EltonFSGrpcServer
}

const (
	_DROP_CACHE = ".drop_cache"
)

func NewEltonFs(fileSystems []pathfs.FileSystem, opts Options, server *EltonFSGrpcServer) (pathfs.FileSystem, error) {
	g := &eltonFS{
		FileSystem:  pathfs.NewDefaultFileSystem(),
		fileSystems: fileSystems,
		opts:        &opts,
		server:      server,
	}

	writable := g.fileSystems[0]
	code := g.createDeletionStore()
	if !code.Ok() {
		return nil, fmt.Errorf("could not create deletion path %v: %v", opts.Deldirname, code)
	}

	g.deletionCache = newDirCache(writable, opts.Deldirname, time.Duration(opts.DelcacheTTL*float64(time.Second)))
	g.branchCache = NewTimedCache(
		func(n string) (interface{}, bool) {
			return g.getBranchAttrNoCache(n), true
		},
		time.Duration(opts.BranchcacheTTL*float64(time.Second)),
	)

	return g, nil
}

func (fs *eltonFS) OnMount(nodeFs *pathfs.PathNodeFs) {
	fs.nodeFs = nodeFs
}

func (fs *eltonFS) OnUnmount() {
	fs.server.Stop()
	for _, f := range fs.fileSystems {
		f.(*grpcFileSystem).connection.Close()
	}
}

func (fs *eltonFS) isDeleted(name string) (deleted bool, code fuse.Status) {
	marker := fs.deletionPath(name)
	haveCache, found := fs.deletionCache.HasEntry(filepath.Base(marker))
	if haveCache {
		return found, fuse.OK
	}

	_, code = fs.fileSystems[0].GetAttr(marker, nil)

	if code == fuse.OK {
		return true, code
	}
	if code == fuse.ENOENT {
		return false, fuse.OK
	}

	log.Println("error accessing deletion marker: ", marker)
	return false, fuse.Status(syscall.EROFS)
}

// TODO: Config Dir の実装
// func (fs *eltonFS) createEltonConfigStore() (code fuse.Status) {
// 	writable := fs.fileSystems[0]
// 	fi, code := writable.GetAttr(_ELTONFS_CONFIG_DIR, nil)
// 	if code == fuse.ENOENT {
// 		code = writable.Mkdir(_ELTONFS_CONFIG_DIR, 0755, nil)
// 		if code.Ok() {
// 			fi, code = writable.GetAttr(_ELTONFS_CONFIG_DIR, nil)
// 		}
// 	}

// 	if !code.Ok() || !fi.IsDir() {
// 		code = fuse.Status(syscall.EROFS)
// 	}

// 	return code
// }

func (fs *eltonFS) createDeletionStore() (code fuse.Status) {
	writable := fs.fileSystems[0]
	fi, code := writable.GetAttr(fs.opts.Deldirname, nil)
	if code == fuse.ENOENT {
		code = writable.Mkdir(fs.opts.Deldirname, 0755, nil)
		if code.Ok() {
			fi, code = writable.GetAttr(fs.opts.Deldirname, nil)
		}
	}

	if !code.Ok() || !fi.IsDir() {
		code = fuse.Status(syscall.EROFS)
	}

	return code
}

func (fs *eltonFS) getBranch(name string) branchResult {
	name = stripSlash(name)
	r := fs.branchCache.Get(name)
	return r.(branchResult)
}

type branchResult struct {
	attr   *fuse.Attr
	code   fuse.Status
	branch int
}

func (fs branchResult) String() string {
	return fmt.Sprintf("{%v %v branch %d}", fs.attr, fs.code, fs.branch)
}

func (fs *eltonFS) getBranchAttrNoCache(name string) branchResult {
	name = stripSlash(name)

	parent, base := path.Split(name)
	parent = stripSlash(parent)

	parentBranch := 0
	if base != "" {
		parentBranch = fs.getBranch(parent).branch
	}

	for i, fs := range fs.fileSystems {
		if i < parentBranch {
			continue
		}

		a, s := fs.GetAttr(name, nil)
		if s.Ok() {
			if i > 0 {
				a.Ino = 0
			}
			return branchResult{
				attr:   a,
				code:   s,
				branch: i,
			}
		} else {
			if s != fuse.ENOENT {
				log.Printf("getattr: %v:  Got error %v from branch %v", name, s, i)
			}
		}
	}
	return branchResult{nil, fuse.ENOENT, -1}
}

func (fs *eltonFS) deletionPath(name string) string {
	return filepath.Join(fs.opts.Deldirname, filePathHash(name))
}

func (fs *eltonFS) removeDeletion(name string) {
	marker := fs.deletionPath(name)
	fs.deletionCache.RemoveEntry(path.Base(marker))

	code := fs.fileSystems[0].Unlink(marker, nil)
	if !code.Ok() && code != fuse.ENOENT {
		log.Printf("error unlinking %s: %v", marker, code)
	}
}

func (fs *eltonFS) putDeletion(name string) (code fuse.Status) {
	code = fs.createDeletionStore()
	if !code.Ok() {
		return code
	}

	marker := fs.deletionPath(name)
	fs.deletionCache.AddEntry(path.Base(marker))

	writable := fs.fileSystems[0]
	fi, code := writable.GetAttr(marker, nil)
	if code.Ok() && fi.Size == uint64(len(name)) {
		return fuse.OK
	}

	var f nodefs.File
	if code == fuse.ENOENT {
		f, code = writable.Create(marker, uint32(os.O_TRUNC|os.O_WRONLY), 0644, nil)
	} else {
		writable.Chmod(marker, 0644, nil)
		f, code = writable.Open(marker, uint32(os.O_TRUNC|os.O_WRONLY), nil)
	}
	if !code.Ok() {
		log.Printf("could not create deletion file %v: %v", marker, code)
		return fuse.EPERM
	}
	defer f.Release()
	defer f.Flush()
	n, code := f.Write([]byte(name), 0)
	if int(n) != len(name) || !code.Ok() {
		panic(fmt.Sprintf("Error for writing %v: %v, %v (exp %v) %v", name, marker, n, len(name), code))
	}
	return fuse.OK
}

func (fs *eltonFS) Promote(name string, srcResult branchResult, context *fuse.Context) (code fuse.Status) {
	writable := fs.fileSystems[0]
	sourceFs := fs.fileSystems[srcResult.branch]

	fs.promoteDirsTo(name)

	if srcResult.attr.IsRegular() {
		code = pathfs.CopyFile(sourceFs, writable, name, name, context)

		if code.Ok() {
			code = writable.Chmod(name, srcResult.attr.Mode&07777|0200, context)
		}
		if code.Ok() {
			aTime := srcResult.attr.AccessTime()
			mTime := srcResult.attr.ModTime()
			code = writable.Utimens(name, &aTime, &mTime, context)
		}

		files := fs.nodeFs.AllFiles(name, 0)
		for _, fileWrapper := range files {
			if !code.Ok() {
				break
			}
			var ef *eltonFsFile
			f := fileWrapper.File
			for f != nil {
				ok := false
				ef, ok = f.(*eltonFsFile)
				if ok {
					break
				}
				f = f.InnerFile()
			}
			if ef == nil {
				panic("no eltonFsFile found inside")
			}

			if ef.layer > 0 {
				ef.layer = 0
				f := ef.File
				ef.File, code = fs.fileSystems[0].Open(name, fileWrapper.OpenFlags, context)
				f.Flush()
				f.Release()
			}
		}
	} else if srcResult.attr.IsSymlink() {
		link := ""
		link, code = sourceFs.Readlink(name, context)
		if !code.Ok() {
			log.Println("can't read link in source fs", name)
		} else {
			code = writable.Symlink(link, name, context)
		}
	} else if srcResult.attr.IsDir() {
		code = writable.Mkdir(name, srcResult.attr.Mode&07777|0200, context)
	} else {
		log.Println("Unknown file type: ", srcResult.attr)
		return fuse.ENOSYS
	}

	if !code.Ok() {
		fs.branchCache.GetFresh(name)
		return code
	} else {
		r := fs.getBranch(name)
		r.branch = 0
		fs.branchCache.Set(name, r)
	}

	return fuse.OK
}

func (fs *eltonFS) Link(orig string, newName string, context *fuse.Context) (code fuse.Status) {
	origResult := fs.getBranch(orig)
	code = origResult.code
	if code.Ok() && origResult.branch > 0 {
		code = fs.Promote(orig, origResult, context)
	}
	if code.Ok() && origResult.branch > 0 {
		fs.branchCache.GetFresh(orig)
		inode := fs.nodeFs.Node(orig)
		var a fuse.Attr
		inode.Node().GetAttr(&a, nil, nil)
	}
	if code.Ok() {
		code = fs.promoteDirsTo(newName)
	}
	if code.Ok() {
		code = fs.fileSystems[0].Link(orig, newName, context)
	}
	if code.Ok() {
		fs.removeDeletion(newName)
		fs.branchCache.GetFresh(newName)
	}

	return code
}

func (fs *eltonFS) Rmdir(path string, context *fuse.Context) (code fuse.Status) {
	r := fs.getBranch(path)
	if r.code != fuse.OK {
		return r.code
	}
	if !r.attr.IsDir() {
		return fuse.Status(syscall.ENOTDIR)
	}

	stream, code := fs.OpenDir(path, context)
	found := false

	for _ = range stream {
		found = true
	}
	if found {
		return fuse.Status(syscall.ENOTEMPTY)
	}

	if r.branch > 0 {
		code = fs.putDeletion(path)
		return code
	}
	code = fs.fileSystems[0].Rmdir(path, context)
	if code != fuse.OK {
		return code
	}

	r = fs.branchCache.GetFresh(path).(branchResult)
	if r.branch > 0 {
		code = fs.putDeletion(path)
	}
	return code
}

func (fs *eltonFS) Mkdir(path string, mode uint32, context *fuse.Context) (code fuse.Status) {
	deleted, code := fs.isDeleted(path)
	if !code.Ok() {
		return code
	}

	if !deleted {
		r := fs.getBranch(path)
		if r.code != fuse.ENOENT {
			return fuse.Status(syscall.EEXIST)
		}

	}

	code = fs.promoteDirsTo(path)
	if code.Ok() {
		code = fs.fileSystems[0].Mkdir(path, mode, context)
	}
	if code.Ok() {
		fs.removeDeletion(path)
		attr := &fuse.Attr{
			Mode: fuse.S_IFDIR | mode,
		}
		fs.branchCache.Set(path, branchResult{attr, fuse.OK, 0})
	}

	var stream []fuse.DirEntry
	stream, code = fs.OpenDir(path, context)
	if code.Ok() {
		for _, entry := range stream {
			fs.putDeletion(filepath.Join(path, entry.Name))
		}
	}

	return code
}

func (fs *eltonFS) Symlink(pointedTo string, linkName string, context *fuse.Context) (code fuse.Status) {
	code = fs.promoteDirsTo(linkName)
	if code.Ok() {
		code = fs.fileSystems[0].Symlink(pointedTo, linkName, context)
	}
	if code.Ok() {
		fs.removeDeletion(linkName)
		fs.branchCache.GetFresh(linkName)
	}
	return code
}

func (fs *eltonFS) Truncate(path string, size uint64, context *fuse.Context) (code fuse.Status) {
	if path == _DROP_CACHE {
		return fuse.OK
	}

	r := fs.getBranch(path)
	if r.branch > 0 {
		code = fs.Promote(path, r, context)
		r.branch = 0
	}

	if code.Ok() {
		code = fs.fileSystems[0].Truncate(path, size, context)
	}
	if code.Ok() {
		r.attr.Size = size
		now := time.Now()
		r.attr.SetTimes(nil, &now, &now)
		fs.branchCache.Set(path, r)
	}
	return code
}

func (fs *eltonFS) Utimens(name string, atime *time.Time, mtime *time.Time, context *fuse.Context) (code fuse.Status) {
	name = stripSlash(name)
	r := fs.getBranch(name)

	code = r.code
	if code.Ok() && r.branch > 0 {
		code = fs.Promote(name, r, context)
		r.branch = 0
	}
	if code.Ok() {
		code = fs.fileSystems[0].Utimens(name, atime, mtime, context)
	}
	if code.Ok() {
		now := time.Now()
		r.attr.SetTimes(atime, mtime, &now)
		fs.branchCache.Set(name, r)
	}

	return code
}

func (fs *eltonFS) Chown(name string, uid uint32, gid uint32, context *fuse.Context) (code fuse.Status) {
	name = stripSlash(name)
	r := fs.getBranch(name)
	if r.attr == nil || r.code != fuse.OK {
		return r.code
	}

	if os.Geteuid() != 0 {
		return fuse.EPERM
	}

	if r.attr.Uid != uid || r.attr.Gid != gid {
		if r.branch > 0 {
			code := fs.Promote(name, r, context)
			if code != fuse.OK {
				return code
			}
			r.branch = 0
		}
		fs.fileSystems[0].Chown(name, uid, gid, context)
	}
	r.attr.Uid = uid
	r.attr.Gid = gid
	now := time.Now()
	r.attr.SetTimes(nil, nil, &now)
	fs.branchCache.Set(name, r)
	return fuse.OK
}

func (fs *eltonFS) Chmod(name string, mode uint32, context *fuse.Context) (code fuse.Status) {
	name = stripSlash(name)
	r := fs.getBranch(name)
	if r.attr == nil {
		return r.code
	}

	if r.code != fuse.OK {
		return r.code
	}

	permMask := uint32(07777)

	oldMode := r.attr.Mode & permMask

	if oldMode != mode {
		if r.branch > 0 {
			code := fs.Promote(name, r, context)
			if code != fuse.OK {
				return code
			}
			r.branch = 0
		}
		fs.fileSystems[0].Chmod(name, mode, context)
	}
	r.attr.Mode = (r.attr.Mode &^ permMask) | mode
	now := time.Now()
	r.attr.SetTimes(nil, nil, &now)
	fs.branchCache.Set(name, r)
	return fuse.OK
}

func (fs *eltonFS) Access(name string, mode uint32, context *fuse.Context) (code fuse.Status) {
	mode = mode &^ fuse.W_OK
	if name == "" {
		return fuse.OK
	}
	r := fs.getBranch(name)
	if r.branch >= 0 {
		return fs.fileSystems[r.branch].Access(name, mode, context)
	}
	return fuse.ENOENT
}

func (fs *eltonFS) Unlink(name string, context *fuse.Context) (code fuse.Status) {
	r := fs.getBranch(name)
	if r.branch == 0 {
		code = fs.fileSystems[0].Unlink(name, context)
		if code != fuse.OK {
			return code
		}
		r = fs.branchCache.GetFresh(name).(branchResult)
	}

	if r.branch > 0 {
		code = fs.putDeletion(name)
	}
	return code
}

func (fs *eltonFS) Readlink(name string, context *fuse.Context) (out string, code fuse.Status) {
	r := fs.getBranch(name)
	if r.branch >= 0 {
		return fs.fileSystems[r.branch].Readlink(name, context)
	}
	return "", fuse.ENOENT
}

func stripSlash(fn string) string {
	return strings.TrimRight(fn, string(filepath.Separator))
}

func (fs *eltonFS) promoteDirsTo(filename string) fuse.Status {
	dirName, _ := filepath.Split(filename)
	dirName = stripSlash(dirName)

	var todo []string
	var results []branchResult
	for dirName != "" {
		r := fs.getBranch(dirName)

		if !r.code.Ok() {
			log.Println("path component does not exist", filename, dirName)
		}
		if !r.attr.IsDir() {
			log.Println("path component is not a directory.", dirName, r)
			return fuse.EPERM
		}
		if r.branch == 0 {
			break
		}
		todo = append(todo, dirName)
		results = append(results, r)
		dirName, _ = filepath.Split(dirName)
		dirName = stripSlash(dirName)
	}

	for i := range todo {
		j := len(todo) - i - 1
		d := todo[j]
		r := results[j]
		code := fs.fileSystems[0].Mkdir(d, r.attr.Mode&07777|0200, nil)
		if code != fuse.OK {
			log.Println("Error creating dir leading to path", d, code, fs.fileSystems[0])
			return fuse.EPERM
		}

		aTime := r.attr.AccessTime()
		mTime := r.attr.ModTime()
		fs.fileSystems[0].Utimens(d, &aTime, &mTime, nil)
		r.branch = 0
		fs.branchCache.Set(d, r)
	}
	return fuse.OK
}

func (fs *eltonFS) Create(name string, flags uint32, mode uint32, context *fuse.Context) (fuseFile nodefs.File, code fuse.Status) {
	writable := fs.fileSystems[0]

	code = fs.promoteDirsTo(name)
	if code != fuse.OK {
		return nil, code
	}
	fuseFile, code = writable.Create(name, flags, mode, context)
	if code.Ok() {
		fuseFile = fs.newEltonFsFile(fuseFile, 0)
		fs.removeDeletion(name)

		now := time.Now()
		a := fuse.Attr{
			Mode: fuse.S_IFREG | mode,
		}
		a.SetTimes(nil, &now, &now)
		fs.branchCache.Set(name, branchResult{&a, fuse.OK, 0})
	}
	return fuseFile, code
}

func (fs *eltonFS) GetAttr(name string, context *fuse.Context) (a *fuse.Attr, s fuse.Status) {
	if name == _DROP_CACHE {
		return &fuse.Attr{
			Mode: fuse.S_IFREG | 0777,
		}, fuse.OK
	}

	if name == fs.opts.Deldirname {
		return nil, fuse.ENOENT
	}

	isDel, s := fs.isDeleted(name)
	if !s.Ok() {
		return nil, s
	}

	if isDel {
		return nil, fuse.ENOENT
	}

	r := fs.getBranch(name)
	if r.branch < 0 {
		return nil, fuse.ENOENT

	}
	fi := *r.attr

	fi.Mode |= 0200
	return &fi, r.code
}

func (fs *eltonFS) GetXAttr(name string, attr string, context *fuse.Context) ([]byte, fuse.Status) {
	if name == _DROP_CACHE {
		return nil, fuse.ENODATA
	}
	r := fs.getBranch(name)
	if r.branch >= 0 {
		return fs.fileSystems[r.branch].GetXAttr(name, attr, context)
	}
	return nil, fuse.ENOENT
}

func (fs *eltonFS) OpenDir(directory string, context *fuse.Context) (stream []fuse.DirEntry, status fuse.Status) {
	dirBranch := fs.getBranch(directory)
	if dirBranch.branch < 0 {
		return nil, fuse.ENOENT
	}

	var wg sync.WaitGroup
	var deletions map[string]bool

	wg.Add(1)
	go func() {
		deletions = newDirnameMap(fs.fileSystems[0], fs.opts.Deldirname)
		wg.Done()
	}()

	entries := make([]map[string]uint32, len(fs.fileSystems))
	for i := range fs.fileSystems {
		entries[i] = make(map[string]uint32)
	}

	statuses := make([]fuse.Status, len(fs.fileSystems))
	for i, l := range fs.fileSystems {
		if i >= dirBranch.branch {
			wg.Add(1)
			go func(j int, pfs pathfs.FileSystem) {
				ch, s := pfs.OpenDir(directory, context)
				statuses[j] = s
				for _, v := range ch {
					entries[j][v.Name] = v.Mode
				}
				wg.Done()
			}(i, l)
		}
	}

	wg.Wait()
	if deletions == nil {
		_, code := fs.fileSystems[0].GetAttr(fs.opts.Deldirname, context)
		if code == fuse.ENOENT {
			deletions = map[string]bool{}
		} else {
			return nil, fuse.Status(syscall.EROFS)
		}
	}

	results := entries[0]

	for i, m := range entries {
		if statuses[i] != fuse.OK {
			continue
		}
		if i == 0 {
			continue
		}
		for k, v := range m {
			_, ok := results[k]
			if ok {
				continue
			}

			deleted := deletions[filePathHash(filepath.Join(directory, k))]
			if !deleted {
				results[k] = v
			}
		}
	}
	if directory == "" {
		delete(results, fs.opts.Deldirname)
	}

	stream = make([]fuse.DirEntry, 0, len(results))
	for k, v := range results {
		stream = append(stream, fuse.DirEntry{
			Name: k,
			Mode: v,
		})
	}
	return stream, fuse.OK
}

func (fs *eltonFS) recursivePromote(path string, pathResult branchResult, context *fuse.Context) (names []string, code fuse.Status) {
	names = []string{}
	if pathResult.branch > 0 {
		code = fs.Promote(path, pathResult, context)
	}

	if code.Ok() {
		names = append(names, path)
	}

	if code.Ok() && pathResult.attr != nil && pathResult.attr.IsDir() {
		var stream []fuse.DirEntry
		stream, code = fs.OpenDir(path, context)
		for _, e := range stream {
			if !code.Ok() {
				break
			}
			subnames := []string{}
			p := filepath.Join(path, e.Name)
			r := fs.getBranch(p)
			subnames, code = fs.recursivePromote(p, r, context)
			names = append(names, subnames...)
		}
	}

	if !code.Ok() {
		names = nil
	}
	return names, code
}

func (fs *eltonFS) renameDirectory(srcResult branchResult, srcDir string, dstDir string, context *fuse.Context) (code fuse.Status) {
	names := []string{}
	if code.Ok() {
		names, code = fs.recursivePromote(srcDir, srcResult, context)
	}
	if code.Ok() {
		code = fs.promoteDirsTo(dstDir)
	}

	if code.Ok() {
		writable := fs.fileSystems[0]
		code = writable.Rename(srcDir, dstDir, context)
	}

	if code.Ok() {
		for _, srcName := range names {
			relative := strings.TrimLeft(srcName[len(srcDir):], string(filepath.Separator))
			dst := filepath.Join(dstDir, relative)
			fs.removeDeletion(dst)

			srcResult := fs.getBranch(srcName)
			srcResult.branch = 0
			fs.branchCache.Set(dst, srcResult)

			srcResult = fs.branchCache.GetFresh(srcName).(branchResult)
			if srcResult.branch > 0 {
				code = fs.putDeletion(srcName)
			}
		}
	}
	return code
}

func (fs *eltonFS) Rename(src string, dst string, context *fuse.Context) (code fuse.Status) {
	srcResult := fs.getBranch(src)
	code = srcResult.code
	if code.Ok() {
		code = srcResult.code
	}

	if srcResult.attr.IsDir() {
		return fs.renameDirectory(srcResult, src, dst, context)
	}

	if code.Ok() && srcResult.branch > 0 {
		code = fs.Promote(src, srcResult, context)
	}
	if code.Ok() {
		code = fs.promoteDirsTo(dst)
	}
	if code.Ok() {
		code = fs.fileSystems[0].Rename(src, dst, context)
	}

	if code.Ok() {
		fs.removeDeletion(dst)
		fs.branchCache.DropEntry(dst)

		srcResult := fs.branchCache.GetFresh(src)
		if srcResult.(branchResult).branch > 0 {
			code = fs.putDeletion(src)
		}
	}
	return code
}

func (fs *eltonFS) DropBranchCache(names []string) {
	fs.branchCache.DropAll(names)
}

func (fs *eltonFS) DropDeletionCache() {
	fs.deletionCache.DropCache()
}

func (fs *eltonFS) DropSubFsCaches() {
	for _, fs := range fs.fileSystems {
		a, code := fs.GetAttr(_DROP_CACHE, nil)
		if code.Ok() && a.IsRegular() {
			f, _ := fs.Open(_DROP_CACHE, uint32(os.O_WRONLY), nil)
			if f != nil {
				f.Flush()
				f.Release()
			}
		}
	}
}

func (fs *eltonFS) Open(name string, flags uint32, context *fuse.Context) (fuseFile nodefs.File, status fuse.Status) {
	if name == _DROP_CACHE {
		if flags&fuse.O_ANYWRITE != 0 {
			log.Println("Forced cache drop on", fs)
			fs.DropBranchCache(nil)
			fs.DropDeletionCache()
			fs.DropSubFsCaches()
			fs.nodeFs.ForgetClientInodes()
		}
		return nodefs.NewDevNullFile(), fuse.OK
	}
	r := fs.getBranch(name)
	if r.branch < 0 {
		log.Println("Eltonfs: open of non-existent file: ", name)
		return nil, fuse.ENOENT
	}
	if flags&fuse.O_ANYWRITE != 0 && r.branch > 0 {
		code := fs.Promote(name, r, context)
		if code != fuse.OK {
			return nil, code
		}
		r.branch = 0
		now := time.Now()
		r.attr.SetTimes(nil, &now, nil)
		fs.branchCache.Set(name, r)
	}
	fuseFile, status = fs.fileSystems[r.branch].Open(name, uint32(flags), context)
	if fuseFile != nil {
		fuseFile = fs.newEltonFsFile(fuseFile, r.branch)
	}
	return fuseFile, status
}

func (fs *eltonFS) String() string {
	names := []string{}
	for _, fs := range fs.fileSystems {
		names = append(names, fs.String())
	}
	return fmt.Sprintf("Eltonfs(%v)", names)
}

func (fs *eltonFS) StatFs(name string) *fuse.StatfsOut {
	return fs.fileSystems[0].StatFs("")
}

type eltonFsFile struct {
	nodefs.File
	efs   *eltonFS
	node  *nodefs.Inode
	layer int
}

func (fs *eltonFsFile) String() string {
	return fmt.Sprintf("eltonFsFile(%s)", fs.File.String())
}

func (fs *eltonFS) newEltonFsFile(f nodefs.File, branch int) *eltonFsFile {
	return &eltonFsFile{
		File:  f,
		efs:   fs,
		layer: branch,
	}
}

func (fs *eltonFsFile) InnerFile() (file nodefs.File) {
	return fs.File
}

func (fs *eltonFsFile) Flush() (code fuse.Status) {
	code = fs.File.Flush()
	path := fs.efs.nodeFs.Path(fs.node)
	fs.efs.branchCache.GetFresh(path)
	return code
}

func (fs *eltonFsFile) SetInode(node *nodefs.Inode) {
	fs.node = node
}

func (fs *eltonFsFile) GetAttr(out *fuse.Attr) fuse.Status {
	code := fs.File.GetAttr(out)
	if code.Ok() {
		out.Mode |= 0200
	}
	return code
}
