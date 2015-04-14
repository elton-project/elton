package elton

import ()

type EltonServer struct {
}

type EltonProxy struct {
	registry *Registry
	balancer *Balancer
}

func NewEltonServer(conf Config) (*EltonServer, error) {
	return nil, nil
}

func NewEltonProxy(conf Config) (*EltonProxy, error) {
	m, err := NewRegistry(conf)
	if err != nil {
		return nil, err
	}

	b, err := NewBalancer(conf)
	if err != nil {
		return nil, err
	}

	return &EltonProxy{registry: m, balancer: b}, nil
}

func (p *EltonProxy) Close() {
	p.registry.Close()
}

func (e *EltonServer) Close() {
}
