package elton

import (
	"container/ring"
	"sync"

	"git.t-lab.cs.teu.ac.jp/nashio/elton/config"
)

type Elton struct {
	db         *DBManager
	serverRing *ring.Ring
	mutex      sync.Mutex
}

func NewElton(conf config.Config) (*Elton, error) {
	db, err := NewDBManager(conf)
	r := ring.New(len(conf.Server))
	if err != nil {
		return nil, err
	}
	return &Elton{db: db, serverRing: r, mutex: sync.Mutex{}}, nil
}

func (e *Elton) Close() {
}
