package elton

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"git.t-lab.cs.teu.ac.jp/nashio/elton/config"
)

type DBManager struct {
	conn *sql.DB
}

func NewDBManager(conf config.Config) (*DBManager, error) {
	dns := conf.DB.User + ":" + conf.DB.Pass + "@tcp(" + conf.DB.Host + ":" + conf.DB.Port + ")/" + conf.DB.DBName + "?charset=utf8&autocommit=false"
	conn, err := sql.Open("mysql", dns)
	if err != nil {
		return nil, err
	}

	conn.SetMaxIdleConns(len(conf.Server))

	err = conn.Ping()
	if err != nil {
		return nil, err
	}

	return &DBManager{conn: conn}, nil
}

func (db *DBManager) Close() {
	db.conn.Close()
}
