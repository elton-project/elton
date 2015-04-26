package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/fukata/golang-stats-api-handler"

	e "../elton"
)

type Server struct {
	Port string
	FS   *e.FileSystem
}

type Result struct {
	Name    string
	Key     string
	Target  string
	Version string
	Length  int64
}

func NewServer(port string, dir string) *Server {
	fs := e.NewFileSystem(dir)
	return &Server{Port: port, FS: fs}
}

func (s *Server) Serve() {
	var srv http.Server
	srv.Addr = ":" + s.Port
	s.RegisterHandler(&srv)

	log.Fatal(srv.ListenAndServe())
}

func (s *Server) RegisterHandler(srv *http.Server) {
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
	name := r.URL.Path

	if err := s.FS.Find(name); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, name)
}

func (s *Server) putHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path
	version := r.PostFormValue("version")
	host := r.PostFormValue("host")

	file, _, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	key, err := s.FS.Create(name, version, file.(*os.File))
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	result, _ := json.Marshal(&Result{Name: name, Key: key, Target: host, Version: version, Length: r.ContentLength})
	fmt.Fprintf(w, string(result))
}

func (s *Server) deleteHandler(w http.ResponseWriter, r *http.Request) {
}
