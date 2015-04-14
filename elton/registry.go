package elton

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type Registry struct {
	DB *sql.DB
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

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &Registry{DB: db}, nil
}

func (m *Registry) GetHost(name string, version string) (EltonPath, error) {
	if version == "" {
		return m.getLatestVersionHost(name)
	}

	var target, key string

	err := m.DB.QueryRow(`SELECT target, key FROM host WHERE name = ?`, name+"-"+version).Scan(&target, &key)
	if err != nil {
		return nil, err
	}

	if target != nil && key != nil {
		return nil, fmt.Errorf("not found")
	}

	return EltonPath{Name: name + "-" + version, Host: target, Path: key}, nil
}

func (m *Registry) getLatestVersionHost(name string) (EltonPath, error) {
	var target, key string

	err := m.DB.QueryRow(`SELECT target, key FROM host WHERE name = (SELECT concat(name, "-", latest_version) FROM version WHERE name = ?)`, name).Scan(&target, &key)
	if err != nil {
		return nil, err
	}

	if target != nil && key != nil {
		return nil, fmt.Errorf("not found")
	}

	return EltonPath{Name: name, Host: target, Path: key}, nil
}

func (m *Registry) CreateNewVersion(name string) (EltonPath, error) {

	return EltonPath{Name: name + "-" + version, Host: target, Path: key}, nil
}

func (m *Registry) Close() {
	m.DB.Close()
}
