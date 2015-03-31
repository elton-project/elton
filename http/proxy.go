package http

import (
	//	"encoding/json"
	//	"io/ioutil"
	"fmt"
	"log"
	"net/http"
	//	"net/http/httputil"

	"github.com/fukata/golang-stats-api-handler"

	"git.t-lab.cs.teu.ac.jp/nashio/elton/config"
	e "git.t-lab.cs.teu.ac.jp/nashio/elton/elton"
)

type proxy struct {
	conf  config.Config
	elton e.DBManager
}

type Proxy interface {
	Serve()
	dispatchHandler(w http.ResponseWriter, r *http.Request)
	getHandler(w http.ResponseWriter, r *http.Request)
	putHandler(w http.ResponseWriter, r *http.Request)
	deleteHandler(w http.ResponseWriter, r *http.Request)
}

type Transport struct {
}

func NewProxy(conf config.Config) (Proxy, error) {
	elt, err := e.NewDBManager(conf)
	if err != nil {
		return new(proxy), err
	}

	return &proxy{conf: conf, elton: elt}
}

func (p *proxy) Serve() {
	defer p.elton.Close()

	for _, server := range p.conf.Server {
		res, err := http.Get("http://" + server.Host + ":" + server.Port + "/api/ping")
		if err != nil || res.StatusCode != http.StatusOK {
			log.Fatalf("can not reach: %s, Error: %v", server, err)
		}
	}

	http.HandleFunc("/maint/stats", stats_api.Handler)
	http.HandleFunc("/maint/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	})
	http.HandleFunc("/", p.dispatchHandler)
	log.Fatal(http.ListenAndServe(":"+p.conf.Proxy.Port, nil))
}

func (p *proxy) dispatchHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		p.getHandler(w, r)
	case "PUT":
		p.putHandler(w, r)
	case "DELETE":
		p.deleteHandler(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (p *proxy) getHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path
	fmt.Fprintf(w, name)
}

func (p *proxy) putHandler(w http.ResponseWriter, r *http.Request) {
	//	name := r.URL.Path

	//	version, err := p.getNewVersion(name)
}

func (p *proxy) deleteHandler(w http.ResponseWriter, r *http.Request) {
}
