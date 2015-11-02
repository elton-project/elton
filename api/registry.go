package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

type Registry struct {
	DB   *bolt.DB
	Name string
}

type ObjectInfo struct {
	ObjectID string `json:"objectid, omitempty"`
	Version  uint64 `json:"version, omitempty"`
	Delegate string `json:"delegate, omitempty"`
}

type ObjectName struct {
	Name []string `json:"name, omitempty"`
}

func NewRegistry(conf Config) (*Registry, error) {
	db, err := bolt.Open(conf.Database.DBPath, 0600, nil)
	if err != nil {
		return nil, err
	}

	if err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("versions"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("hosts"))
		return err
	}); err != nil {
		return nil, err
	}

	return &Registry{DB: db, Name: conf.Master.Name}, nil
}

func (r *Registry) GenerateObjectInfo(name string) (obj ObjectInfo, err error) {
	oid := r.generateObjectID(name)
	if err = r.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("versions"))
		return bucket.Put([]byte(oid), []byte("1"))
	}); err != nil {
		return
	}

	obj = ObjectInfo{
		ObjectID: oid,
		Version:  1,
		Delegate: r.Name,
	}

	return
}

func (r *Registry) generateObjectID(name string) string {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s%d", name, time.Now().Nanosecond())))

	return string(hex.EncodeToString(hasher.Sum(nil)))
}

func (r *Registry) GetNewVersion(obj ObjectInfo) (object ObjectInfo, err error) {
	var version uint64
	if err = r.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("versions"))
		n, err := strconv.ParseUint(
			string(bucket.Get([]byte(obj.ObjectID))),
			10,
			64)
		if err != nil {
			return err
		}

		version = n + 1
		return bucket.Put(
			[]byte(obj.ObjectID),
			[]byte(strconv.FormatUint(version, 10)))
	}); err != nil {
		return
	}

	object = ObjectInfo{
		ObjectID: obj.ObjectID,
		Version:  version,
		Delegate: obj.Delegate,
	}

	return
}

func (r *Registry) SetObjectInfo(obj ObjectInfo, host string) error {
	return r.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("hosts"))
		return bucket.Put(
			[]byte(fmt.Sprintf("%s/%s", obj.ObjectID, obj.Version)),
			[]byte(host))
	})
}

func (r *Registry) GetObjectHost(oid string, version uint64) (host string, err error) {
	if version == 0 {
		version, err = r.getVersion(oid)
		if err != nil {
			return "", err
		}
	}

	err = r.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("hosts"))
		host = string(bucket.Get([]byte(fmt.Sprintf(
			"%s/%d",
			oid,
			version))))
		if host == "" {
			return fmt.Errorf("Not found: %s/%s", oid, version)
		}

		return nil
	})

	return
}

func (r *Registry) getVersion(oid string) (version uint64, err error) {
	err = r.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("versions"))
		version, err = strconv.ParseUint(string(bucket.Get([]byte(oid))), 10, 64)
		if err != nil {
			return fmt.Errorf("Not found: %s", oid)
		}

		return nil
	})

	return
}

func (r *Registry) DeleteObjectVersions(oid string) (err error) {
	return r.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("hosts"))
		cursor := bucket.Cursor()
		for k, _ := cursor.Seek([]byte(oid)); bytes.HasPrefix(k, []byte(oid)); k, _ = cursor.Next() {
			err = bucket.Delete(k)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *Registry) DeleteObjectInfo(oid string) (err error) {
	return r.DB.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("versions")).Delete([]byte(oid))
	})
}

func (r *Registry) Close() {
	r.DB.Close()
}
