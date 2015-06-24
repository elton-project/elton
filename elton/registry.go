package elton

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

type FileName struct {
	Name []string `json:"name, omitempty"`
}

func NewRegistry(conf Config) (*Registry, error) {
	db, err := bolt.Open(conf.Database.DBPath, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("versions"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("hosts"))
		return err
	})

	if err != nil {
		return nil, err
	}

	return &Registry{DB: db, Name: conf.Elton.Name}, nil
}

func (r *Registry) GenerateObjectsInfo(names FileName) ([]ObjectInfo, error) {
	result := make([]ObjectInfo, len(names.Name))
	for i, n := range names.Name {
		oid := r.generateObjectID(n)

		if err := r.DB.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("versions"))
			return bucket.Put([]byte(oid), []byte("0"))
		}); err != nil {
			return nil, err
		}

		result[i] = ObjectInfo{
			ObjectID: oid,
			Delegate: r.Name,
		}
	}

	return result, nil
}

func (r *Registry) generateObjectID(name string) string {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s%d", name, time.Now().Nanosecond())))

	return string(hex.EncodeToString(hasher.Sum(nil)))
}

func (r *Registry) GetNewVersions(objects []ObjectInfo) (foundObjs []ObjectInfo, otherObjs []ObjectInfo, err error) {
	for _, obj := range objects {
		if obj.Delegate != r.Name {
			otherObjs = append(otherObjs, obj)
			continue
		}

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

		foundObjs = append(foundObjs, ObjectInfo{
			ObjectID: obj.ObjectID,
			Delegate: obj.Delegate,
			Version:  version,
		})
	}

	return
}

func (r *Registry) SetObjectsInfo(objects []ObjectInfo, host string) error {
	for _, obj := range objects {
		if err := r.DB.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("hosts"))
			return bucket.Put(
				[]byte(fmt.Sprintf("%s/%s", obj.ObjectID, obj.Version)),
				[]byte(host))
		}); err != nil {
			return err
		}
	}

	return nil
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

func (r *Registry) DeleteObject(oid string) (err error) {
	return r.DB.Update(func(tx *bolt.Tx) error {
		hosts := tx.Bucket([]byte("hosts"))
		cursor := hosts.Cursor()
		for k, _ := cursor.Seek([]byte(oid)); bytes.HasPrefix(k, []byte(oid)); k, _ = cursor.Next() {
			err = hosts.Delete(k)
			if err != nil {
				return err
			}
		}

		return tx.Bucket([]byte("versions")).Delete([]byte(oid))
	})
}

func (r *Registry) Close() {
	r.DB.Close()
}
