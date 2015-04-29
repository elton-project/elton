package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"path"

	"github.com/fukata/golang-stats-api-handler"

	e "../elton"
)

type Proxy struct {
	Conf     e.Config
	Registry *e.Registry
}

type EltonTransport struct {
	Registry *e.Registry
	Target   string
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
	mux.HandleFunc("/api/list", p.DispatchListHandler)
	mux.HandleFunc("/", p.DispatchFileHandler)
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

func (p *Proxy) DispatchListHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		p.getListHandler(w, r)
	case "PUT":
		p.putListHandler(w, r)
	case "DELETE":
		p.deleteListHandler(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (p *Proxy) getListHandler(w http.ResponseWriter, r *http.Request) {
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

func (p *Proxy) putListHandler(w http.ResponseWriter, r *http.Request) {
}

func (p *Proxy) deleteListHandler(w http.ResponseWriter, r *http.Request) {
}

func (p *Proxy) DispatchFileHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		p.getFileHandler(w, r)
	case "PUT":
		p.putFileHandler(w, r)
	case "DELETE":
		p.deleteFileHandler(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (p *Proxy) getFileHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	name := r.URL.Path
	if params.Get("latest") == "true" {
		data, err := p.Registry.GetLatestVersionHost(name)
	} else {
		version := path.Base(name)
		data, err := p.Registry.GetHost(name, version)
	}

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

func (p *Proxy) putFileHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path

	data, err := p.Registry.GenerateNewVersion(name, host)
	if err != nil {
		log.Printf("L.132: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	rp := &httputil.ReverseProxy{Director: func(request *http.Request) {
		request.URL.Scheme = "http"
		request.URL.Host = data.Host
		request.URL.Path = data.Path + "/" + data.Version
	}}

	rp.Transport = &EltonTransport{Registry: p.Registry, Target: data.Host}
	rp.ServeHTTP(w, r)
}

func (p *Proxy) deleteFileHandler(w http.ResponseWriter, r *http.Request) {
	// name := r.URL.Path
	// version := r.PostFormValue("version")

	// client := &http.Client{}
	// for _, c := range p.Registry.Clients {
	// 	if _, err := client.Do(r); err != nil {
	// 		log.Printf("Can not Delete file: %s.", c)
	// 		log.Printf("Error by: %v", err)
	// 	}
	// }
}

func (t *EltonTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		log.Printf("L.164: %v", err)
		return nil, err
	}

	if response.StatusCode == http.StatusOK {
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("L.172: %v", err)
			return nil, err
		}

		var result Result
		json.Unmarshal(body, &result)

		if err = t.Registry.RegisterNewVersion(result.Name, result.Key, t.Target, result.Length); err != nil {
			log.Printf("L.180: %v", err)
			return nil, err
		}
	}

	return response, err
}
