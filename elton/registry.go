package elton

import (
	"database/sql"
	"fmt"
	"strconv"

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

func (m *Registry) GenerateNewVersion(name string) (e EltonPath, err error) {
	tx, err := m.DB.Begin()
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

	if _, err = tx.Exec(`INSERT INTO version (name) VALUES (?) ON DUPLICATE KEY UPDATE counter = counter + 1`, name); err != nil {
		return
	}

	var version uint64
	if err = tx.QueryRow(`SELECT counter FROM version WHERE name = ?`, name).Scan(&version); err != nil {
		return
	}

	e = EltonPath{Name: name + "-" + strconv.FormatUint(version, 10), Host: m.Balancer.GetServer(), Path: generateKey(name + "-" + strconv.FormatUint(version, 10))}
	return
}

func (m *Registry) CreateNewVersion(versionedName, host, key, name string) (err error) {
	tx, err := m.DB.Begin()

	if _, err = tx.Exec(`INSERT INTO hosts (name, target, key, perent_id) VALUES (?, ?, ?, (SELECT id FROM version WHERE name = ?))`, versionedName, host, key, name); err != nil {
		return
	}
	return
}

func (m *Registry) Close() {
	m.DB.Close()
}
