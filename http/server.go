package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/fukata/golang-stats-api-handler"
)

type server struct {
	port   string
	dir    string
	weight int
}

type Server interface {
	Serve()
	dispatchHandler(w http.ResponseWriter, r *http.Request)
	getHandler(w http.ResponseWriter, r *http.Request)
	putHandler(w http.ResponseWriter, r *http.Request)
	deleteHandler(w http.ResponseWriter, r *http.Request)
}

func NewServer(port string, dir string, weight int) Server {
	return &server{port: port, dir: dir, weight: weight}
}

func (s *server) Serve() {
	http.HandleFunc("/maint/stats", stats_api.Handler)
	http.HandleFunc("/maint/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	})
	http.HandleFunc("/", s.dispatchHandler)
	log.Fatal(http.ListenAndServe(":"+s.port, nil))
}

func (s *server) dispatchHandler(w http.ResponseWriter, r *http.Request) {
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

func (s *server) getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	fmt.Fprintf(w, key)
}

func (s *server) putHandler(w http.ResponseWriter, r *http.Request) {
}

func (s *server) deleteHandler(w http.ResponseWriter, r *http.Request) {
}
