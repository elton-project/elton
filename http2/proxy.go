package http2

import (
	"log"
	"net/http"

	"github.com/bradfitz/http2"

	e "../elton"
	eh "../http"
)

type Proxy2 struct {
	*eh.Proxy
}

func NewProxy(conf e.Config) (*Proxy2, error) {
	p, err := eh.NewProxy(conf)
	if err != nil {
		return nil, err
	}

	return &Proxy2{p}, nil
}

func (p *Proxy2) Serve() {
	defer p.EltonProxy.Close()

	var srv http.Server
	srv.Addr = ":" + p.Conf.Proxy.Port
	p.RegisterHandler(&srv)
	http2.ConfigureServer(&srv, new(http2.Server))

	log.Fatal(srv.ListenAndServeTLS("../examples/server.crt", "../examples/server.key"))
}
