package elton

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type Balancer struct {
	Servers    []string
	ServersLen int
}

func NewBalancer(conf Config) (*Balancer, error) {
	var servers []string
	for _, server := range conf.Server {
		target := server.Host + ":" + server.Port

		if err := ping(target); err != nil {
			return nil, err
		}

		for i := 0; i < server.Weight; i++ {
			servers = append(servers, target)
		}
	}

	rand.Seed(time.Now().UnixNano())
	return &Balancer{Servers: servers, ServersLen: len(servers)}, nil
}

func (b *Balancer) GetServer() string {
	return b.Servers[rand.Intn(b.ServersLen)]
}

func ping(target string) error {
	res, err := http.Get("http://" + target + "/maint/ping")
	if err != nil || res.StatusCode != http.StatusOK {
		return fmt.Errorf("can not reach: %s, Error: %v", target, err)
	}
	return nil
}

//func (b *Balancer) heartbeat() {
//	for _, server := range b.Servers {
//		target := server.Host + ":" + server.Port

//		if err := ping(target); err != nil {
//			return nil, err
//		}
//	}
//}
