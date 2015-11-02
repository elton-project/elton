package main

import (
	"log"
	"sync"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

func newDirnameMap(fs pathfs.FileSystem, dir string) map[string]bool {
	stream, code := fs.OpenDir(dir, nil)
	if code == fuse.ENOENT {
		return map[string]bool{}
	}

	if !code.Ok() {
		log.Printf("newDirnameMap(%v): %v %v", fs, dir, code)
		return nil
	}

	result := make(map[string]bool)
	for _, e := range stream {
		if e.Mode&fuse.S_IFREG != 0 {
			result[e.Name] = true
		}
	}
	return result
}

type dirCache struct {
	dir           string
	ttl           time.Duration
	fs            pathfs.FileSystem
	lock          sync.RWMutex
	names         map[string]bool
	updateRunning bool
}

func (c *dirCache) setMap(newMap map[string]bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.names = newMap
	c.updateRunning = false
	_ = time.AfterFunc(c.ttl,
		func() { c.DropCache() })
}

func (c *dirCache) DropCache() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.names = nil
}

func (c *dirCache) maybeRefresh() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.updateRunning {
		return
	}
	c.updateRunning = true
	go func() {
		newmap := newDirnameMap(c.fs, c.dir)
		c.setMap(newmap)
	}()
}

func (c *dirCache) RemoveEntry(name string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.names == nil {
		go c.maybeRefresh()
		return
	}

	delete(c.names, name)
}

func (c *dirCache) AddEntry(name string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.names == nil {
		go c.maybeRefresh()
		return
	}

	c.names[name] = true
}

func newDirCache(fs pathfs.FileSystem, dir string, ttl time.Duration) *dirCache {
	dc := new(dirCache)
	dc.dir = dir
	dc.fs = fs
	dc.ttl = ttl
	return dc
}

func (c *dirCache) HasEntry(name string) (mapPresent bool, found bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.names == nil {
		go c.maybeRefresh()
		return false, false
	}

	return true, c.names[name]
}
