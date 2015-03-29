package http

import (
	//	"encoding/json"
	//	"io/ioutil"
	"fmt"
	"log"
	"net/http"
	//	"net/http/httputil"

	"github.com/fukata/golang-stats-api-handler"

	//	"git.t-lab.cs.teu.ac.jp/nashio/elton/api"
	"git.t-lab.cs.teu.ac.jp/nashio/elton/config"
)

type proxy struct {
	conf config.Config
}

type Proxy interface {
	Serve()
}

type Transport struct {
}

func NewProxy(conf config.Config) Proxy {
	return &proxy{conf: conf}
}

func (p *proxy) Serve() {
	http.HandleFunc("/maint/stats", stats_api.Handler)
	http.HandleFunc("/maint/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	})
	http.HandleFunc("/", dispatchHandler)
	log.Fatal(http.ListenAndServe(":"+p.conf.Server.Port, nil))
}

func dispatchHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getHandler(w, r)
	case "PUT":
		putHandler(w, r)
	case "DELETE":
		deleteHandler(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	fmt.Fprintf(w, key)
}

func putHandler(w http.ResponseWriter, r *http.Request) {
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
}
