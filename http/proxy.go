package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/fukata/golang-stats-api-handler"

	e "../elton"
)

type Proxy struct {
	Conf     e.Config
	Registry *e.Registry
}

type EltonTransport struct {
	Registry *e.Registry
}

func NewProxy(conf e.Config) (*Proxy, error) {
	registry, err := e.NewRegistry(conf)
	if err != nil {
		return nil, err
	}

	return &Proxy{Conf: conf, Registry: registry}, nil
}

func (p *Proxy) Serve() {
	defer p.Registry.Close()

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
	mux.HandleFunc("/api/connect", p.ConnectHandler)
	mux.HandleFunc("/api/list", p.GetListHandler)
	mux.HandleFunc("/", p.DispatchHandler)
	srv.Handler = mux
}

func (p *Proxy) ConnectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	host := r.PostFormValue("host")
	p.Registry.AddClient(host)
}

func (p *Proxy) GetListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	list, err := p.Registry.GetList()
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	result, err := json.Marshal(list)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, string(result))
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

	data, err := p.Registry.GetHost(name, version)
	if err != nil {
		log.Println(err)
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
	host := r.PostFormValue("host")

	data, err := p.Registry.GenerateNewVersion(name, host)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	rp := &httputil.ReverseProxy{Director: func(request *http.Request) {
		request.URL.Scheme = "http"
		request.URL.Host = data.Host
		request.URL.Path = data.Path
		request.PostForm.Add("version", data.Version)
		request.PostForm.Set("host", data.Host)
	}}
	rp.Transport = &EltonTransport{Registry: p.Registry}
	rp.ServeHTTP(w, r)
}

func (p *Proxy) deleteHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path
	version := r.PostFormValue("version")

	for _, client := range p.Registry.Clients {

	}
	data, err := p.Registry.GetHost(name, version)
	if err != nil {
		log.Println(err)
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

func (t *EltonTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if response.StatusCode == http.StatusOK {
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		var result Result
		json.Unmarshal(body, &result)

		if err = t.Registry.RegisterNewVersion(result.Name, result.Key, result.Target, result.Length); err != nil {
			log.Println(err)
			return nil, err
		}
	}

	return response, err
}
