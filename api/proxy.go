package api

import (
	"bytes"
	"container/ring"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"sync"

	"github.com/boltdb/bolt"
)

type proxy struct {
	db         *bolt.DB
	servers    []string
	serverRing *ring.Ring
	mutex      sync.Mutex
}

type Proxy interface {
	GetServers() []string
	GetServerHost() string
	GetHost(string, string, string) (string, error)
	GetNewVersion(string, string) (string, error)
	SetHost(string, string) error
	Delete(string, string) error
	Migration([]string, string) error
	Close()
}

func NewProxy(path string, servers []string) Proxy {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		log.Fatalf("Can not open db file: %s\n", err)
	}

	serverRing := ring.New(len(servers))
	for _, server := range servers {
		serverRing.Value = server
		serverRing = serverRing.Next()
	}

	return &proxy{db, servers, serverRing, sync.Mutex{}}
}

func (p *proxy) GetServers() []string {
	return p.servers
}

func (p *proxy) GetServerHost() (host string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	host = p.serverRing.Value.(string)
	p.serverRing = p.serverRing.Next()
	return
}

func (p *proxy) GetHost(dir string, key string, version string) (string, error) {
	var host []byte
	err := p.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("hosts"))
		host = bucket.Get([]byte(dir + "/" + key + "/" + version))
		if host == nil {
			return errors.New("Not found: " + dir + "/" + key + "/" + version)
		}
		return nil
	})

	if err != nil {
		return "", errors.New("No such file: " + dir + "/" + key + "/" + version)
	}

	return string(host), nil
}

func (p *proxy) GetNewVersion(dir string, key string) (string, error) {
	var version []byte
	err := p.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("versions"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		n, err := strconv.ParseUint(string(bucket.Get([]byte(dir+"/"+key))), 10, 64)
		if err != nil {
			n = 0
		}

		version = []byte(strconv.FormatUint(n+1, 10))
		return bucket.Put([]byte(dir+"/"+key), version)
	})

	if err != nil || version == nil {
		return "", err
	}

	return string(version), nil
}

func (p *proxy) SetHost(key string, host string) error {
	return p.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("hosts"))
		if err != nil {
			return fmt.Errorf("create backet: %s", err)
		}

		return bucket.Put([]byte(key), []byte(host))
	})
}

func (p *proxy) Delete(dir string, key string) error {
	return p.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("versions"))
		if err != nil {
			return fmt.Errorf("create backet: %s", err)
		}
		err = bucket.Delete([]byte(dir + "/" + key))
		if err != nil {
			return fmt.Errorf("Can not delete version: %s", dir+"/"+key)
		}
		bucket = tx.Bucket([]byte("hosts"))
		c := bucket.Cursor()

		prefix := []byte(dir + "/" + key)
		for k, _ := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			err = bucket.Delete([]byte(k))
			if err != nil {
				return fmt.Errorf("Can not delete host: %s", k)
			}
		}

		return nil
	})
}

func (p *proxy) Migration(path []string, host string) error {
	return p.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("hosts"))
		if err != nil {
			return fmt.Errorf("create backet: %s", err)
		}
		for _, p := range path {
			log.Printf("key: %s, host: %s", p, host)
			if match, _ := regexp.Match(`\d+\z`, []byte(p)); !match {
				p += "-0"
			}
			log.Printf("key: %s, host: %s", p, host)
			err = bucket.Put([]byte(p), []byte(host))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (p *proxy) Close() {
	p.db.Close()
}
