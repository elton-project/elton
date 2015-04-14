package elton

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type Balancer struct {
	servers    []string
	serversLen int
}

func NewBalancer(conf Config) (*Balancer, error) {
	var servers []string
	for _, server := range conf.Server {
		target := server.Host + ":" + server.Port

		res, err := http.Get("http://" + target + "/maint/ping")
		if err != nil || res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("can not reach: %s, Error: %v", server, err)
		}

		for i := 0; i < server.Weight; i++ {
			servers = append(servers, target)
		}
	}

	rand.Seed(time.Now().UnixNano())
	return &Balancer{servers: servers, serversLen: len(servers)}, nil
}

func (b *Balancer) GetServer() string {
	return b.servers[rand.Intn(b.serversLen)]
}
