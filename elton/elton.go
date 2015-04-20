package elton

import ()

type EltonServer struct {
}

type EltonProxy struct {
	Registry *Registry
}

func NewEltonServer(conf Config) (*EltonServer, error) {
	return nil, nil
}

func NewEltonProxy(conf Config) (*EltonProxy, error) {
	m, err := NewRegistry(conf)
	if err != nil {
		return nil, err
	}

	return &EltonProxy{Registry: m}, nil
}

func (p *EltonProxy) Close() {
	p.Registry.Close()
}

func (e *EltonServer) Close() {
}
