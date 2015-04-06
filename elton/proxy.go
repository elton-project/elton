package elton

import (
	"database/sql"
	"rand"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Proxy struct {
	db         *sql.DB
	servers    []string
	serversLen int
}

func NewProxy(conf Config) (*Proxy, error) {
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

	var servers []string
	for _, server := range conf.Server {
		for i := 0; i < server.Weight; i++ {
			servers = append(servers, "http://"+server.Host+":"+server.Port)
		}
	}

	rand.Seed(time.Now().UnixNano())
	return &Proxy{db: db, servers: servers, serversLen: len(servers)}, nil
}

func (p *Proxy) GetServer() string {
	return p.servers[rand.Intn(p.serversLen)]
}

func (p *Proxy) Close() {
	p.db.Close()
}
