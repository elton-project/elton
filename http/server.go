package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/fukata/golang-stats-api-handler"
)

type Server struct {
	Port   string
	Dir    string
	Weight int
}

func NewServer(port string, dir string, weight int) *Server {
	return &Server{Port: port, Dir: dir, Weight: weight}
}

func (s *Server) Serve() {
	http.HandleFunc("/maint/stats", stats_api.Handler)
	http.HandleFunc("/maint/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	})
	http.HandleFunc("/", s.DispatchHandler)
	log.Fatal(http.ListenAndServe(":"+s.Port, nil))
}

func (s *Server) DispatchHandler(w http.ResponseWriter, r *http.Request) {
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
