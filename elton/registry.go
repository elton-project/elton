package elton

import (
	"database/sql"
	"fmt"
	"path"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Registry struct {
	DB       *sql.DB
	balancer *Balancer
	Clients  []string
	mux      *sync.Mutex
}

type EltonPath struct {
	Host    string
	Path    string
	Version string
}

type EltonFile struct {
	Name       string    `json:"name"`
	FileSize   uint64    `json:"size"`
	ModifyTime time.Time `json:"modifytime,omitempty"`
}

var dnsTemplate = `%s:%s@tcp(%s:%s)/%s?charset=utf8&autocommit=false&parseTime=true`

func NewRegistry(conf Config) (*Registry, error) {
	dns := fmt.Sprintf(dnsTemplate, conf.Proxy.Database.User, conf.Proxy.Database.Pass, conf.Proxy.Database.Host, conf.Proxy.Database.Port, conf.Proxy.Database.DBName)
	db, err := sql.Open("mysql", dns)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(len(conf.Proxy.Servers))

	balancer, err := NewBalancer(conf)
	if err != nil {
		return nil, err
	}

	clients := make([]string, len(conf.Proxy.Servers))
	for i, server := range conf.Proxy.Servers {
		clients[i] = server.Host + ":" + server.Port
	}

	return &Registry{DB: db, balancer: balancer, Clients: clients, mux: new(sync.Mutex)}, nil
}

func (r *Registry) AddClient(client string) {
	r.mux.Lock()
	r.Clients = append(r.Clients, client)
	r.mux.Unlock()
}

func (r *Registry) GetList() ([]EltonFile, error) {
	var counter uint64
	err := r.DB.QueryRow(`SELECT COUNT(*) FROM host`).Scan(&counter)
	if err != nil {
		return nil, err
	}

	rows, err := r.DB.Query(`SELECT name, size, created_at FROM host`)
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
	defer func() {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("not found: %s", name)
		}
	}()

	versionedName := name + "/" + version
	var target, key string
	if err = r.DB.QueryRow(`SELECT target, eltonkey FROM host WHERE name = ?`, versionedName).Scan(&target, &key); err != nil {
		return
	}

	e = EltonPath{Host: target, Path: key, Version: version}
	return
}

func (r *Registry) GetLatestVersionHost(name string) (e EltonPath, err error) {
	defer func() {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("not found: %s", name)
		}
	}()

	var target, key, version string
	if err = r.DB.QueryRow(`SELECT latest_version FROM version WHERE name = ?`, name).Scan(&version); err != nil {
		return
	}

	versionedName := name + "/" + version
	if err = r.DB.QueryRow(`SELECT target, eltonkey FROM host WHERE name = ?`, versionedName).Scan(&target, &key); err != nil {
		return
	}

	e = EltonPath{Host: target, Path: key, Version: version}
	return
}

func (r *Registry) GenerateNewVersion(host, name string) (version string, err error) {
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

	newVersionStmt, _ := tx.Prepare(`INSERT INTO version (name) VALUES (?) ON DUPLICATE KEY UPDATE counter = counter + 1`)
	versionStmt, _ := tx.Prepare(`SELECT counter FROM version WHERE name = ?`)
	newTargetStmt, _ := tx.Prepare(`INSERT INTO host (name, target, eltonkey, perent_id) VALUES (?, ?, ?, (SELECT id FROM version WHERE name = ?))`)

	e = make([]EltonFile, len(files))
	for i, file := range files {
		if _, err = newVersionStmt.Exec(file.Name); err != nil {
			return
		}

		var version string
		if err = versionStmt.QueryRow(file.Name).Scan(&version); err != nil {
			return
		}

		versionedName := file.Name + "/" + version
		if _, err = newTargetStmt.Exec(versionedName, host, file.Name, file.Name); err != nil {
			return
		}

		e[i] = EltonFile{Name: versionedName, FileSize: file.FileSize}
	}
	return
}

func (r *Registry) RegisterNewVersion(name, key, target string, size int64) (err error) {
	tx, err := r.BD.Begin()
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

	if _, err = tx.Exec(`UPDATE host SET target = ?, eltonkey = ?, size = ?, delegate = TRUE WHERE name = ?`, target, key, size, name); err != nil {
		return
	}

	version, err := strconv.ParseUint(path.Base(name), 10, 64)
	if err != nil {
		return
	}

	if _, err = tx.Exec(`UPDATE version SET latest_version = ? WHERE name = ?`, version, name); err != nil {
		return
	}

	return
}

func (r *Registry) DeleteVersion(name, version string) (err error) {
	if version == "" {
		return r.deleteAllVersion(name)
	}

	versionedName := name + "/" + version
	_, err = r.DB.Exec(`DELETE FROM host WHERE name = ?`, versionedName)
	return
}

func (r *Registry) deleteAllVersion(name string) (err error) {
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

	if _, err = tx.Exec(`DELETE FROM host WHERE name like '?%'`, name); err != nil {
		return
	}

	if _, err = tx.Exec(`DELETE FROM version WHERE name = ?`, name); err != nil {
		return
	}
	return
}

func (r *Registry) Close() {
	r.DB.Close()
}
