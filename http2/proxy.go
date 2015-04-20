package http2

import (
	"log"
	"net/http"

	"github.com/bradfitz/http2"
	"github.com/fukata/golang-stats-api-handler"

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
	mux := http.NewServeMux()
	mux.HandleFunc("/maint/stats", stats_api.Handler)
	mux.HandleFunc("/maint/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	})
	mux.HandleFunc("/", p.DispatchHandler)
	srv.Handler = mux

	http2.ConfigureServer(&srv, new(http2.Server))
	log.Fatal(srv.ListenAndServeTLS("../examples/server.crt", "../examples/server.key"))
}
