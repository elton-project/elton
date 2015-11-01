package main

import (
	"sync"
	"time"
)

type cacheEntry struct {
	data   interface{}
	expiry time.Time
}

type TimedCacheFetcher func(name string) (value interface{}, cacheable bool)

type TimedCache struct {
	fetch         TimedCacheFetcher
	ttl           time.Duration
	cacheMapMutex sync.RWMutex
	cacheMap      map[string]*cacheEntry
	PurgeTimer    *time.Timer
}

func NewTimedCache(fetcher TimedCacheFetcher, ttl time.Duration) *TimedCache {
	l := new(TimedCache)
	l.ttl = ttl
	l.fetch = fetcher
	l.cacheMap = make(map[string]*cacheEntry)
	return l
}

func (c *TimedCache) Get(name string) interface{} {
	c.cacheMapMutex.RLock()
	info, ok := c.cacheMap[name]
	c.cacheMapMutex.RUnlock()

	valid := ok && (c.ttl <= 0 || info.expiry.After(time.Now()))
	if valid {
		return info.data
	}
	return c.GetFresh(name)
}

func (c *TimedCache) Set(name string, val interface{}) {
	c.cacheMapMutex.Lock()
	defer c.cacheMapMutex.Unlock()

	c.cacheMap[name] = &cacheEntry{
		data:   val,
		expiry: time.Now().Add(c.ttl),
	}
}

func (c *TimedCache) DropEntry(name string) {
	c.cacheMapMutex.Lock()
	defer c.cacheMapMutex.Unlock()

	delete(c.cacheMap, name)
}

func (c *TimedCache) GetFresh(name string) interface{} {
	data, ok := c.fetch(name)
	if ok {
		c.Set(name, data)
	}
	return data
}

func (c *TimedCache) Purge() {
	keys := make([]string, 0, len(c.cacheMap))
	now := time.Now()

	c.cacheMapMutex.Lock()
	defer c.cacheMapMutex.Unlock()
	for k, v := range c.cacheMap {
		if now.After(v.expiry) {
			keys = append(keys, k)
		}
	}
	for _, k := range keys {
		delete(c.cacheMap, k)
	}
}

func (c *TimedCache) DropAll(names []string) {
	c.cacheMapMutex.Lock()
	defer c.cacheMapMutex.Unlock()

	if names == nil {
		c.cacheMap = make(map[string]*cacheEntry, len(c.cacheMap))
	} else {
		for _, nm := range names {
			delete(c.cacheMap, nm)
		}
	}
}
