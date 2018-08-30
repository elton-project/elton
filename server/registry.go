package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

// masterサーバのデータベース
type Registry struct {
	// boltDBのデータベースオブジェクト。
	//
	// versionsバケット:
	// 各オブジェクトの最新のバージョン番号を格納する。
	// バージョン番号の最小値は1。
	//
	//   versions[objectID] = latestVersion
	//   objectID: SHA256のハッシュ値を16進数表記に変換した文字列
	//   latestVersion: 10進数の文字列
	//
	//
	// hostsバケット:
	// 指定したバージョンのオブジェクトを保持しているノードのホスト名を格納する。
	// TODO: なんのために使っているのか、よくわからない
	//
	//   hosts[objectID, version] = hostName
	//   hostName: オブジェクトを保持しているノードのアドレス。書式は"host:port"。
	//             hostはelton serverの名前であり、DNSで引けるとは限らない。
	DB   *bolt.DB
	// masterサーバの名前。
	Name string
}

type ObjectInfo struct {
	// SHA256のハッシュ値を16進数表記に変換した文字列
	ObjectID string `json:"objectid, omitempty"`
	Version  uint64 `json:"version, omitempty"`
	// masterサーバの名前
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

// 新しいObjectInfoを作成する。
// name引数の意味はあまりない。正直適当な値でも問題なさそう。
//
// TODO: name引数を削除する。
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

// nameに対応するobjectIDを返す。
// nameが同じ場合、1<<30の確率でIDが衝突する。
// しかし、処理のタイミング次第ではもっと高い確率で衝突する恐れがある。
//
// TODO: atomic.AddUint32()を使うなどして、シーケンシャルな数値を使う手法に変えられないか検討する。
func (r *Registry) generateObjectID(name string) string {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s%d", name, time.Now().Nanosecond())))

	return string(hex.EncodeToString(hasher.Sum(nil)))
}

// objの新しいバージョンを作成して返す。
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
			[]byte(fmt.Sprintf("%s/%d", obj.ObjectID, obj.Version)),
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
			return fmt.Errorf("Not found: %s/%d", oid, version)
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
