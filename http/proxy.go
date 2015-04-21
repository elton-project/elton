package http

import (
	"encoding/json"
	//	"io/ioutil"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/fukata/golang-stats-api-handler"

	e "../elton"
)

type Proxy struct {
	Conf       e.Config
	EltonProxy *e.EltonProxy
}

type Transport struct {
	name          string
	version       uint64
	versionedName string
}

func NewProxy(conf e.Config) (*Proxy, error) {
	ep, err := e.NewEltonProxy(conf)
	if err != nil {
		return nil, err
	}

	return &Proxy{Conf: conf, EltonProxy: ep}, nil
}

func (p *Proxy) Serve() {
	defer p.EltonProxy.Close()

	var srv http.Server
	srv.Addr = ":" + p.Conf.Proxy.Port
	p.RegisterHandler(&srv)

	log.Fatal(srv.ListenAndServe())
}

func (p *Proxy) RegisterHandler(srv *http.Server) {
	mux := http.NewServeMux()
	mux.HandleFunc("/maint/stats", stats_api.Handler)
	mux.HandleFunc("/maint/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	})
	mux.HandleFunc("/api/list", p.GetListHandler)
	mux.HandleFunc("/", p.DispatchHandler)
	srv.Handler = mux
}

func (p *Proxy) GetListHandler(w http.ResponseWriter, r *http.Request) {
	list, err := p.EltonProxy.Registry.GetList()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	jsonString, err := json.Marshal(list)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, string(jsonString))
}

func (p *Proxy) DispatchHandler(w http.ResponseWriter, r *http.Request) {
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

	data, err := p.EltonProxy.Registry.GetHost(name, version)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	rp := &httputil.ReverseProxy{Director: func(request *http.Request) {
		request.URL.Scheme = "http"
		request.URL.Host = data.Host
		request.URL.Path = data.Path
	}}
	rp.ServeHTTP(w, r)
}

func (p *Proxy) putHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path

	data, err := p.EltonProxy.Registry.GenerateNewVersion(name)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	rp := &httputil.ReverseProxy{Director: func(request *http.Request) {
		request.URL.Scheme = "http"
		request.URL.Host = data.Host
		request.URL.Path += data.Path
	}}
	rp.Transport = &Transport{name: name, versionedName: data.Name}
	rp.ServeHTTP(w, r)
}

func (p *Proxy) deleteHandler(w http.ResponseWriter, r *http.Request) {
}

func (t *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)
	//	if err != nil {
	//		return nil, err
	//	}

	//	if response.StatusCode == http.StatusOK {
	//		host := response.Request.URL.Host
	//		key := []byte(response.Request.URL.Path)

	//		err = p.ep.Registry.CreateNewVersion(t.versionedName, host, string(key[1:]), t.name)
	//		err = proxy.SetHost(string(key[1:]), host)
	//		if err != nil {
	//			return nil, err
	//		}
	//	}

	return response, err
}
