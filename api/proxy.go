package api

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/boltdb/bolt"
)

type proxy struct {
	db *bolt.DB
}

type Proxy interface {
	GetHost(string, string, string) (string, error)
	GetNewVersion(string, string) (string, error)
	SetHost(string, string) error
	Migration()
	Close()
}

func NewProxy(path string) Proxy {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		log.Fatalf("Can not open db file: %s\n", err)
	}
	return &proxy{db}
}

func (p *proxy) GetHost(dir string, key string, version string) (string, error) {
	var host []byte
	err := p.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("hosts"))
		host = bucket.Get([]byte(dir + "/" + key + "/" + version))
		return nil
	})

	if err != nil || host == nil {
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
		return "", errors.New("Can not get new version")
	}

	return string(version), nil
}

func (p *proxy) SetHost(key string, host string) error {
	return p.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("hosts"))
		if err != nil {
			return fmt.Errorf("create backet: %s", err)
		}

		return bucket.Put([]byte(key)[1:], []byte(host))
	})
}

func (p *proxy) Migration() {
}

func (p *proxy) Close() {
	p.db.Close()
}
