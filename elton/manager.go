package elton

import (
	"database/sql"
	"rand"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Manager struct {
	conn       *sql.DB
	servers    []string
	serversLen int
}

func NewManager(conf Config) (*Manager, error) {
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

	var servers []string
	for _, server := range conf.Server {
		for i := 0; i < server.Weight; i++ {
			servers = append(servers, "http://"+server.Host+":"+server.Port)
		}
	}

	rand.Seed(time.Now().UnixNano())
	return &Manager{conn: conn, servers: servers, serversLen: len(servers)}, nil
}

func (m *Manager) GetServer() string {
	return m.servers[rand.Intn(h.serversLen)]
}

func (m *Manager) Close() {
	m.conn.Close()
}
