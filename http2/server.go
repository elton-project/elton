package http2

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bradfitz/http2"
	"github.com/fukata/golang-stats-api-handler"
)

type Server struct {
	port   string
	dir    string
	weight int
}

func NewServer(port string, dir string, weight int) *Server {
	return &Server{port: port, dir: dir, weight: weight}
}

func (s *Server) Serve() {
	var srv http.Server
	srv.Addr = ":" + s.port
	mux := http.NewServeMux()
	mux.HandleFunc("/maint/stats", stats_api.Handler)
	mux.HandleFunc("/maint/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	})
	mux.HandleFunc("/", s.dispatchHandler)
	srv.Handler = mux

	http2.ConfigureServer(&srv, new(http2.Server))
	log.Fatal(srv.ListenAndServeTLS("../examples/server.crt", "../examples/server.key"))
}

func (s *Server) dispatchHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getHandler(w, r)
	case "PUT":
		s.putHandler(w, r)
	case "DELETE":
		s.deleteHandler(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (s *Server) getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	fmt.Fprintf(w, key)
}

func (s *Server) putHandler(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) deleteHandler(w http.ResponseWriter, r *http.Request) {
}
