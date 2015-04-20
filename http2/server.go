package http2

import (
	"log"
	"net/http"

	"github.com/bradfitz/http2"
	"github.com/fukata/golang-stats-api-handler"

	eh "../http"
)

type Server2 struct {
	*eh.Server
}

func NewServer(port string, dir string, weight int) *Server2 {
	return &Server2{eh.NewServer(port, dir, weight)}
}

func (s *Server2) Serve() {
	var srv http.Server
	srv.Addr = ":" + s.Port
	mux := http.NewServeMux()
	mux.HandleFunc("/maint/stats", stats_api.Handler)
	mux.HandleFunc("/maint/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	})
	mux.HandleFunc("/", s.DispatchHandler)
	srv.Handler = mux

	http2.ConfigureServer(&srv, new(http2.Server))
	log.Fatal(srv.ListenAndServeTLS("../examples/server.crt", "../examples/server.key"))
}
