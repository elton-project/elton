package elton

import (
	"database/sql"
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

type EltonFile struct {
	Name       string
	FileSize   uint64
	ModifyTime time.Time
}

type FileList []EltonFile

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

func (r *Registry) GetList() (FileList, error) {
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

	list := make(FileList, counter)
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

	var target, key string

	err = r.DB.QueryRow(`SELECT target, key FROM host WHERE name = ?`, name+"-"+version).Scan(&target, &key)
	if err == sql.ErrNoRows {
		return e, fmt.Errorf("not found: %s", name+"-"+version)
	} else if err != nil {
		return
	}

	e = EltonPath{Name: name + "-" + version, Host: target, Path: key}
	return
}

func (r *Registry) getLatestVersionHost(name string) (e EltonPath, err error) {
	var target, key string

	err = r.DB.QueryRow(`SELECT target, key FROM host WHERE name = (SELECT concat(name, "-", latest_version) FROM version WHERE name = ?)`, name).Scan(&target, &key)
	if err == sql.ErrNoRows {
		return e, fmt.Errorf("not found: %s", name)
	} else if err != nil {
		return
	}

	e = EltonPath{Name: name, Host: target, Path: key}
	return
}

func (r *Registry) GenerateNewVersion(name string) (e EltonPath, err error) {
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

	if _, err = tx.Exec(`INSERT INTO version (name) VALUES (?) ON DUPLICATE KEY UPDATE counter = counter + 1`, name); err != nil {
		return
	}

	var version uint64
	if err = tx.QueryRow(`SELECT counter FROM version WHERE name = ?`, name).Scan(&version); err != nil {
		return
	}

	e = EltonPath{Name: name + "-" + strconv.FormatUint(version, 10), Host: r.Balancer.GetServer(), Path: generateKey(name + "-" + strconv.FormatUint(version, 10))}
	return
}

func (m *Registry) CreateNewVersion(versionedName, host, key, name string) (err error) {
	tx, err := m.DB.Begin()

	if _, err = tx.Exec(`INSERT INTO hosts (name, target, key, perent_id) VALUES (?, ?, ?, (SELECT id FROM version WHERE name = ?))`, versionedName, host, key, name); err != nil {
		return
	}
	return
}

func (r *Registry) Close() {
	r.DB.Close()
}
