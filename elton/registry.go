package elton

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type Registry struct {
	db *sql.DB
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

	return &Registry{db: db}, nil
}

func (m *Registry) GetHostWithVersion(name string, version string) (host string, path string, err error) {
	sql := `SELECT * FROM hideo`
}

func (m *Registry) GetHost(name string) (host string, path string, err error) {
}

func (m *Registry) Close() {
	m.db.Close()
}
