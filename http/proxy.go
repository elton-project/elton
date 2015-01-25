package http

import (
	"log"
	"net/http"
	"net/http/httputil"

	"git.t-lab.cs.teu.ac.jp/nashio/elton/api"
)

type Transport struct {
}

var proxy api.Proxy

func InitProxy(path string, servers []string) {
	for _, server := range servers {
		res, err := http.PostForm("http://"+server+"/api/ping", nil)
		if err != nil || res.StatusCode != http.StatusOK {
			log.Fatalf("can not reach: %s, Error: %v", server, err)
		}
	}
	proxy = api.NewProxy(path, servers)
}

func ProxyGetHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	dir := params.Get(":dir")
	key := params.Get(":key")
	version := params.Get(":version")

	host, err := proxy.GetHost(dir, key, version)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	rp := &httputil.ReverseProxy{Director: func(request *http.Request) {
		request.URL.Scheme = "http"
		request.URL.Host = string(host)
	}}
	rp.ServeHTTP(w, r)
}

func ProxyPutHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	dir := params.Get(":dir")
	key := params.Get(":key")

	version, err := proxy.GetNewVersion(dir, key)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	rp := &httputil.ReverseProxy{Director: func(request *http.Request) {
		request.URL.Scheme = "http"
		request.URL.Host = proxy.GetServerHost()
		request.URL.Path += "/" + string(version)
	}}
	rp.Transport = &Transport{}
	rp.ServeHTTP(w, r)
}

func ProxyDeleteHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	dir := params.Get(":dir")
	key := params.Get(":key")

	err := proxy.Delete(dir, key)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	//Client側にも送らないとね
	for _, server := range proxy.GetServers() {
		rp := &httputil.ReverseProxy{Director: func(request *http.Request) {
			request.URL.Scheme = "http"
			request.URL.Host = server
		}}
		rp.ServeHTTP(w, r)
	}
}

func ProxyMigrationHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)
	key := r.FormValue("key")
	log.Println(key)
}

func (t *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode == http.StatusOK {
		key := response.Request.URL.Path
		host := response.Request.URL.Host
		err = proxy.SetHost(key, host)
		if err != nil {
			return nil, err
		}
	}

	return response, err
}

func DestoryProxy() {
	proxy.Close()
}
