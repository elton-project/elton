package elton

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Registry struct {
	DB       *sql.DB
	Balancer *Balancer
}

type EltonPath struct {
	Host    string
	Path    string
	Version string
}

type EltonFile struct {
	Name       string
	FileSize   uint64
	ModifyTime time.Time
}

var dnsTemplate = `%s:%s@tcp(%s:%s)/%s?charset=utf8&autocommit=false`

func NewRegistry(conf Config) (*Registry, error) {
	dns := fmt.Sprintf(dnsTemplate, conf.DB.User, conf.DB.Pass, conf.DB.Host, conf.DB.Port, conf.DB.DBName)
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

func (r *Registry) GetList() ([]EltonFile, error) {
	var counter uint64
	err := r.DB.QueryRow(`SELECT COUNT(*) FROM host`).Scan(&counter)
	if err != nil {
		return nil, err
	}

	rows, err := r.DB.Query(`SELECT name, size, create_at FROM host`)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	list := make([]EltonFile, counter)
	for i := 0; rows.Next(); i++ {
		var name string
		var size uint64
		var modifyTime time.Time
		if err := rows.Scan(&name, &size, &modifyTime); err != nil {
			return nil, err
		}
		list[i] = EltonFile{Name: name, FileSize: size, ModifyTime: modifyTime}
	}
	return list, nil
}

func (r *Registry) GetHost(name string, version string) (e EltonPath, err error) {
	if version == "" {
		return r.getLatestVersionHost(name)
	}

	defer func() {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("not found: %s", name)
		}
	}()

	versionedName := name + "-" + version
	var target, key string
	if err = r.DB.QueryRow(`SELECT target, key FROM host WHERE name = ?`, versionedName).Scan(&target, &key); err != nil {
		return
	}

	e = EltonPath{Host: target, Path: key, Version: version}
	return
}

func (r *Registry) getLatestVersionHost(name string) (e EltonPath, err error) {
	defer func() {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("not found: %s", name)
		}
	}()

	var target, key, version string
	if err = r.DB.QueryRow(`SELECT version FROM version WHERE name = ?`, name).Scan(&version); err != nil {
		return
	}

	versionedName := name + "-" + version
	if err = r.DB.QueryRow(`SELECT target, key FROM host WHERE name = ?`, versionedName).Scan(&target, &key); err != nil {
		return
	}

	e = EltonPath{Host: target, Path: key, Version: version}
	return
}

func (r *Registry) GenerateNewVersion(name, host string) (e EltonPath, err error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	if _, err = tx.Exec(`INSERT INTO version (name) VALUES (?) ON DUPLICATE KEY UPDATE latest_version = latest_version + 1`, name); err != nil {
		return
	}

	var version string
	if err = tx.QueryRow(`SELECT latest_version FROM version WHERE name = ?`, name).Scan(&version); err != nil {
		return
	}

	versionedName := name + "-" + version
	if _, err = tx.Exec(`INSERT INTO host (name, target, key, perent_id) VALUES (?, ?, ?, (SELECT id FROM version WHERE name = ?))`, versionedName, host, name, name); err != nil {
		return
	}

	e = EltonPath{Path: versionedName, Host: r.Balancer.GetServer(), Version: version}
	return
}

func (m *Registry) RegisterNewVersion(name, key, target string, size int64) (err error) {
	_, err = m.DB.Exec(`UPDATE host SET (target, key, size, delegate) VALUES (?, ?, ?, true) WHERE name = ?`, target, key, size, name)
	return
}

func (r *Registry) Close() {
	r.DB.Close()
}
