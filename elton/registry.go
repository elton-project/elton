package elton

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Registry struct {
	DB       *sql.DB
	Balancer *Balancer
}

type EltonPath struct {
	Name string
	Host string
	Path string
}

func NewRegistry(conf Config) (*Registry, error) {
	dns := conf.DB.User + ":" + conf.DB.Pass + "@tcp(" + conf.DB.Host + ":" + conf.DB.Port + ")/" + conf.DB.DBName + "?charset=utf8&autocommit=false"

	db, err := sql.Open("mysql", dns)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(len(conf.Server))

	balancer, err := NewBalancer(conf)
	if err != nil {
		return nil, err
	}

	return &Registry{DB: db, Balancer: balancer}, nil
}

func (m *Registry) GetHost(name string, version string) (e EltonPath, err error) {
	if version == "" {
		return m.getLatestVersionHost(name)
	}

	var target, key string

	err = m.DB.QueryRow(`SELECT target, key FROM host WHERE name = ?`, name+"-"+version).Scan(&target, &key)
	if err == sql.ErrNoRows {
		return e, fmt.Errorf("not found: %s", name+"-"+version)
	} else if err != nil {
		return
	}

	e = EltonPath{Name: name + "-" + version, Host: target, Path: key}
	return
}

func (m *Registry) getLatestVersionHost(name string) (e EltonPath, err error) {
	var target, key string

	err = m.DB.QueryRow(`SELECT target, key FROM host WHERE name = (SELECT concat(name, "-", latest_version) FROM version WHERE name = ?)`, name).Scan(&target, &key)
	if err == sql.ErrNoRows {
		return e, fmt.Errorf("not found: %s", name)
	} else if err != nil {
		return
	}

	e = EltonPath{Name: name, Host: target, Path: key}
	return
}

func (m *Registry) CreateNewVersion(name string) (e EltonPath, err error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return
	}

	//	defer func() {
	//		if err != nil {
	//			tx.Rollback()
	//			return
	//		}
	//		err = tx.Commit()
	//	}()
	var version uint64
	_, err = tx.Exec(`INSERT INTO version (name) VALUES (?) ON DUPLICATE KEY UPDATE counter = counter + 1`, name)
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.QueryRow(`SELECT counter FROM version WHERE name = ?`, name).Scan(&version)
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()

	e = EltonPath{Name: name + "-" + strconv.FormatUint(version, 10), Host: m.Balancer.GetServer(), Path: generateKey(name + "-" + strconv.FormatUint(version, 10))}
	return
}

func (m *Registry) Close() {
	m.DB.Close()
}

func generateKey(name string) string {
	hasher := md5.New()
	hasher.Write([]byte(name + time.Now().String()))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return string(hash[:2] + "/" + hash[2:])
}
