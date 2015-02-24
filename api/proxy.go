package api

import (
	"bytes"
	"container/ring"
	"errors"
	"log"
	"regexp"
	"strconv"
	"strings"
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
	GetLatestVersion(string, string) (string, error)
	GetNewVersion(string, string) (string, error)
	SetHost(string, string) error
	Delete(string, string) error
	Migration([]string, string) error
	Close()
}

func NewProxy(path string, servers []string) Proxy {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		log.Fatalf("[elton proxy] Can not open db file: %s", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("counter"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("hosts"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("versions"))
		return err
	})

	if err != nil {
		log.Fatalf("[elton proxy] Can not create bucket file: %v", err)
	}

	serverRing := ring.New(len(servers))
	for _, server := range servers {
		log.Printf("[elton proxy] Add server: %s", server)
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

func (p *proxy) GetHost(dir string, key string, version string) (host string, err error) {
	err = p.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("hosts"))

		host = string(bucket.Get([]byte(dir + "/" + key + "/" + version)))
		if host == "" {
			return errors.New("Not found: " + dir + "/" + key + "/" + version)
		}
		return nil
	})

	log.Printf("[elton proxy] Get key: %s, host: %s", dir+"/"+key+"/"+version, host)
	return
}

func (p *proxy) GetLatestVersion(dir string, key string) (version string, err error) {
	err = p.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("versions"))

		version = string(bucket.Get([]byte(dir + "/" + key)))
		if version == "" {
			return errors.New("Not found: " + dir + "/" + key)
		}
		return nil
	})

	log.Printf("[elton proxy] Get key: %s, version: %s", dir+"/"+key, version)
	return
}

func (p *proxy) GetNewVersion(dir string, key string) (version string, err error) {
	err = p.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("counter"))

		n, err := strconv.ParseUint(string(bucket.Get([]byte(dir+"/"+key))), 10, 64)
		if err != nil {
			n = 0
		}

		version = strconv.FormatUint(n+1, 10)
		return bucket.Put([]byte(dir+"/"+key), []byte(version))
	})

	log.Printf("[elton proxy] Get key: %s, version: %s", dir+"/"+key, version)
	return
}

func (p *proxy) SetHost(key string, host string) error {
	return p.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("versions"))

		keys := strings.Split(string(key), "/")
		version := keys[len(keys)-1]
		err := bucket.Put([]byte(keys[0]+"/"+keys[1]), []byte(version))
		if err != nil {
			return err
		}

		bucket = tx.Bucket([]byte("hosts"))

		log.Printf("[elton proxy] Set key: %s, host: %s", key, host)
		return bucket.Put([]byte(key), []byte(host))
	})
}

func (p *proxy) Delete(dir string, key string) error {
	return p.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("versions"))

		err := bucket.Delete([]byte(dir + "/" + key))
		if err != nil {
			return err
		}

		bucket = tx.Bucket([]byte("hosts"))
		c := bucket.Cursor()

		prefix := []byte(dir + "/" + key)
		for k, _ := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			err = bucket.Delete([]byte(k))
			if err != nil {
				return err
			}
		}

		bucket = tx.Bucket([]byte("counter"))

		log.Printf("[elton proxy] Delete key: %s", dir+"/"+key)
		return bucket.Delete([]byte(dir + "/" + key))
	})
}

func (p *proxy) Migration(path []string, host string) error {
	return p.db.Update(func(tx *bolt.Tx) error {
		versionsBucket := tx.Bucket([]byte("versions"))
		hostsBucket := tx.Bucket([]byte("hosts"))
		counterBucket := tx.Bucket([]byte("counter"))

		for _, p := range path {
			regex := regexp.MustCompile(`-(\d+)\z`)
			if regex.MatchString(p) {
				p = regex.ReplaceAllString(p, "/$1")
			} else {
				p += "/0"
			}

			keys := strings.Split(string(p), "/")
			version := keys[len(keys)-1]
			log.Printf("[elton proxy] Migration path: %s, host: %s", p, host)
			err := versionsBucket.Put([]byte(keys[0]+"/"+keys[1]), []byte(version))
			if err != nil {
				return err
			}

			err = hostsBucket.Put([]byte(p), []byte(host))
			if err != nil {
				return err
			}

			err = counterBucket.Put([]byte(p), []byte(version))
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
