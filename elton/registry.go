package elton

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

type Registry struct {
	DB *bolt.DB
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

	return &Registry{DB: db}, nil
}

func (r *Registry) GetObjectsID(names []FileName) ([]ObjectInfo, error) {
	result := make([]ObjectInfo, len(names))
	for i, n := range names {
		oid := generateObjectID(n.Name)

		if err := r.DB.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("versions"))
			return bucket.Put([]byte(oid), []byte(0))
		}); err != nil {
			return nil, err
		}

		result[i].ObjectID = fmt.Sprintf("%s/%s", r.Conf.Elton.Name, oid)
	}

	return result, nil
}

func (r *Registry) SetObjectsInfo(objects []ObjectInfo) ([]ObjectInfo, error) {
	result := make([]ObjectInfo, len(objects))
	for i, obj := range objects {
		var version uint64
		if err := r.DB.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("versions"))
			n, err := strconv.ParseUint(
				string(bucket.Get([]byte(obj.ObjectID))),
				10,
				64,
			)
			if err != nil {
				return err
			}

			version = n + 1
			err = bucket.Put(
				[]byte(obj.ObjectID),
				[]byte(strconv.FormatUint(version, 10)),
			)

			return tx.Bucket([]byte("hosts")).Put(
				[]byte(obj.ObjectID),
				[]byte(obj.Delegate),
			)
		}); err != nil {
			return nil, err
		}

		result[i] = ObjectInfo{
			ObjectID: obj.ObjectID,
			Delegate: obj.Delegate,
			Version:  version,
		}
	}

	return result, nil
}

func (r *Registry) generateObjectID(name string) string {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s%d", name, time.Now().Nanoseconds())))

	return string(hex.EncodeToString(hasher.Sum(nil)))
}

// func (r *Registry) GetList() ([]FileInfo, error) {
// 	rows, err := r.DB.Query(`SELECT version.name, host.eltonkey, host.size, host.created_at FROM version INNER JOIN host ON host.name = CONCAT(version.name, '-', version.latest_version)`)
// 	defer rows.Close()
// 	if err != nil {
// 		return nil, err
// 	}

// 	files := make([]FileInfo, 0)
// 	for i := 0; rows.Next(); i++ {
// 		var name, key string
// 		var size uint64
// 		var createdTime time.Time
// 		if err := rows.Scan(&name, &key, &size, &createdTime); err != nil {
// 			return nil, err
// 		}
// 		files = append(files, FileInfo{Name: name, Key: key, Size: size, Time: createdTime})
// 	}
// 	return files, nil
// }

// func (r *Registry) GetHost(name string, version string) (e EltonPath, err error) {
// 	log.Println(version)
// 	if version == "0" {
// 		return r.GetLatestVersionHost(name)
// 	}

// 	defer func() {
// 		if err == sql.ErrNoRows {
// 			err = fmt.Errorf("not found: %s", name)
// 		}
// 	}()

// 	versionedName := name + "-" + version
// 	var target, key string
// 	if err = r.DB.QueryRow(`SELECT target, eltonkey FROM host WHERE name = ?`, versionedName).Scan(&target, &key); err != nil {
// 		return
// 	}

// 	e = EltonPath{Host: target, Path: key, Version: version}
// 	return
// }

// func (r *Registry) GetLatestVersionHost(name string) (e EltonPath, err error) {
// 	defer func() {
// 		if err == sql.ErrNoRows {
// 			err = fmt.Errorf("not found: %s", name)
// 		}
// 	}()

// 	var target, key, version string
// 	if err = r.DB.QueryRow(`SELECT latest_version FROM version WHERE name = ?`, name).Scan(&version); err != nil {
// 		return
// 	}

// 	versionedName := name + "-" + version
// 	if err = r.DB.QueryRow(`SELECT target, eltonkey FROM host WHERE name = ?`, versionedName).Scan(&target, &key); err != nil {
// 		return
// 	}

// 	e = EltonPath{Host: target, Path: key, Version: version}
// 	return
// }

// func (r *Registry) GenerateNewVersion(name string) (version string, err error) {
// 	tx, err := r.DB.Begin()
// 	if err != nil {
// 		return
// 	}

// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 			return
// 		}
// 		err = tx.Commit()
// 	}()

// 	if _, err = tx.Exec(`INSERT INTO version (name) VALUES (?) ON DUPLICATE KEY UPDATE counter = counter + 1`, name); err != nil {
// 		return
// 	}

// 	if err = tx.QueryRow(`SELECT counter FROM version WHERE name = ?`, name).Scan(&version); err != nil {
// 		return
// 	}

// 	return
// }

// func (r *Registry) RegisterNewVersion(name, version, key, target string, size int64) (err error) {
// 	tx, err := r.DB.Begin()
// 	if err != nil {
// 		return
// 	}

// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 			return
// 		}
// 		err = tx.Commit()
// 	}()

// 	if _, err = tx.Exec(`INSERT INTO host (name, target, eltonkey, size, perent_id) VALUES (?, ?, ?, ?, (SELECT id FROM version WHERE name = ?))`, name+"-"+version, target, key, size, name); err != nil {
// 		return
// 	}

// 	if _, err = tx.Exec(`UPDATE version SET latest_version = ? WHERE name = ?`, version, name); err != nil {
// 		return
// 	}

// 	return
// }

// func (r *Registry) RegisterBackup(key string) (err error) {
// 	_, err = r.DB.Exec(`UPDATE host SET backup = TRUE WHERE eltonkey = ?`, key)
// 	return
// }

// func (r *Registry) DeleteVersion(name, version string) (err error) {
// 	if version == "" {
// 		return r.deleteAllVersion(name)
// 	}

// 	versionedName := name + "-" + version
// 	_, err = r.DB.Exec(`DELETE FROM host WHERE name = ?`, versionedName)
// 	return
// }

// func (r *Registry) deleteAllVersion(name string) (err error) {
// 	tx, err := r.DB.Begin()
// 	if err != nil {
// 		return
// 	}

// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 			return
// 		}
// 		err = tx.Commit()
// 	}()

// 	if _, err = tx.Exec(`DELETE FROM host WHERE name like '?%'`, name); err != nil {
// 		return
// 	}

// 	if _, err = tx.Exec(`DELETE FROM version WHERE name = ?`, name); err != nil {
// 		return
// 	}
// 	return
// }

func (r *Registry) Close() {
	r.DB.Close()
}
