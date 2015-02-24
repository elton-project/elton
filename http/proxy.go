package http

import (
	"encoding/json"
	"io/ioutil"
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
		res, err := http.Get("http://" + server + "/api/ping")
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

	if version == "" {
		version, _ = proxy.GetLatestVersion(dir, key)
	}

	if version == "" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	host, err := proxy.GetHost(dir, key, version)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	rp := &httputil.ReverseProxy{Director: func(request *http.Request) {
		request.URL.Scheme = "http"
		request.URL.Host = string(host)
		request.URL.Path = "/" + dir + "/" + key + "/" + version
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

	client := &http.Client{}
	//Client側にも送らないとね
	for _, server := range proxy.GetServers() {
		req, _ := http.NewRequest("DELETE", "http://"+server+r.URL.Path, nil)
		go client.Do(req)
	}
}

func Migration() {
	for _, server := range proxy.GetServers() {
		var path Path
		res, err := http.Get("http://" + server + "/api/migration")
		if err != nil {
			log.Printf("Error: can not reach: %s, error: %v", server, err)
			return
		}
		content, _ := ioutil.ReadAll(res.Body)
		json.Unmarshal(content, &path)
		defer res.Body.Close()
		proxy.Migration(path.Path, res.Request.URL.Host)
	}
}

func (t *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode == http.StatusOK {
		key := []byte(response.Request.URL.Path)
		host := response.Request.URL.Host
		err = proxy.SetHost(string(key[1:]), host)
		if err != nil {
			return nil, err
		}
	}

	return response, err
}

func DestoryProxy() {
	proxy.Close()
}
