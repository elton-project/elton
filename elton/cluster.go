package elton

import (
	"sync"
)

type Cluster struct {
	cluster map[int][]string
	mux     sync.Mutex
}

const (
	Local = itoa + 1
	Global
)

func NewCluster() *Cluster {
	return new(Cluster)
}

func (c *Cluster) SetHost(host string, t int) {
	c.mux.Lock()
	c.cluster[t] = append(c.cluster[t], host)
	c.mux.Unlock()
}

func (c *Cluster) GetGlobalHost() []string {
	return c.cluster[Global]
}

func (c *Cluster) GetLocalHost() []string {
	return c.cluster[Local]
}

func (c *Cluster) DeleteHost(host string, t int) {
	c.mux.Lock()
	c.cluster[t] = delete(c.cluster[t], host)
	c.mux.Unlock()
}
