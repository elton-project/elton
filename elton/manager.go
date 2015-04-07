package elton

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type Manager struct {
	db *sql.DB
}

func NewManager(conf Config) (*Manager, error) {
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

	return &Manager{db: db}, nil
}

func (m *Manager) GetHostWithVersion(name string, version string) (host string, path string, err error) {
	sql := `SELECT * FROM `
}

func (m *Manager) GetHost(name string) (host string, path string, err error) {
}

func (m *Manager) Close() {
	m.db.Close()
}
