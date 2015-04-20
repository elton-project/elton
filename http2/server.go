package http2

import (
	"log"
	"net/http"

	"github.com/bradfitz/http2"

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
	s.RegisterHandler(&srv)
	http2.ConfigureServer(&srv, new(http2.Server))

	log.Fatal(srv.ListenAndServeTLS("../examples/server.crt", "../examples/server.key"))
}
