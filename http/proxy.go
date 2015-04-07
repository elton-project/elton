package http

import (
	//	"encoding/json"
	//	"io/ioutil"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/fukata/golang-stats-api-handler"

	e "git.t-lab.cs.teu.ac.jp/nashio/elton/elton"
)

type Proxy struct {
	conf e.Config
	ep   *e.EltonProxy
}

type Transport struct {
}

func NewProxy(conf e.Config) (*Proxy, error) {
	ep, err := e.NewEltonProxy(conf)
	if err != nil {
		return nil, err
	}

	return &Proxy{conf: conf, ep: ep}, nil
}

func (p *Proxy) Serve() {
	defer p.ep.Close()

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

func (p *Proxy) dispatchHandler(w http.ResponseWriter, r *http.Request) {
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

func (p *Proxy) getHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	name := r.URL.Path
	version := params.Get("version")

	var host, path string
	var err error
	if version == "" {
		host, path, err := p.ep.manager.GetHost(name)
	} else {
		host, path, err := p.ep.manager.GetHostWithVersion(name, version)
	}

	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	rp := &httputil.ReverseProxy{Director: func(request *http.Reqeust) {
		request.URL.Scheme = "http"
		request.URL.Host = host
		request.URL.Path = path
	}}
	rp.ServeHTTP(w, r)
}

func (p *Proxy) putHandler(w http.ResponseWriter, r *http.Request) {
	//	name := r.URL.Path

	//	version, err := p.getNewVersion(name)
}

func (p *Proxy) deleteHandler(w http.ResponseWriter, r *http.Request) {
}
