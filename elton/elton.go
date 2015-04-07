package elton

import ()

type EltonServer struct {
}

type EltonProxy struct {
	manager  *Manager
	balancer *Balancer
}

func NewEltonServer(conf Config) (*EltonServer, error) {
	db, err := NewDBManager(conf)
	r := ring.New(len(conf.Server))
	if err != nil {
		return nil, err
	}
	return &Elton{db: db, serverRing: r, mutex: sync.Mutex{}}, nil
}

func NewEltonProxy(conf Config) (*EltonProxy, error) {
	m, err := NewManager(conf)
	if err != nil {
		return nil, err
	}

	b, err := NewBalancer(conf)
	if err != nil {
		return nil, err
	}

	return &EltonProxy{manager: m, balancer: b}, nil
}

func (p *EltonProxy) Close() {
	p.manager.Close()
}

func (e *Elton) Close() {
}
