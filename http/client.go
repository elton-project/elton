package http

import (
	"io"
	"net/http"
	"net/http/httputil"
	"os"

	"git.t-lab.cs.teu.ac.jp/nashio/elton/api"
)

var client api.Server

func InitClient(dir string, host string) {
	client = api.NewServer(dir, host)
}

func ClientGetHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	dir := params.Get(":dir")
	key := params.Get(":key")
	version := params.Get(":version")

	target, err := client.Read(dir, key, version)
	if err != nil {
		res, err := http.Get("http://" + client.GetHost() + "/" + dir + "/" + key + "/" + version)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		if _, err := os.Stat(client.GetDir() + "/" + dir); os.IsNotExist(err) {
			os.Mkdir(client.GetDir()+"/"+dir, 0700)
		}

		target = client.FormatPath(dir, key, version)
		out, err := os.Create(target)

		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		defer out.Close()

		io.Copy(out, res.Body)
		defer res.Body.Close()
	}

	http.ServeFile(w, r, target)
}

func ClientPutHandler(w http.ResponseWriter, r *http.Request) {
	rp := &httputil.ReverseProxy{Director: func(request *http.Request) {
		request.URL.Scheme = "http"
		request.URL.Host = client.GetHost()
	}}
	rp.ServeHTTP(w, r)
}

func ClientDeleteHandler(w http.ResponseWriter, r *http.Request) {
	rp := &httputil.ReverseProxy{Director: func(request *http.Request) {
		request.URL.Scheme = "http"
		request.URL.Host = client.GetHost()
	}}
	rp.ServeHTTP(w, r)
}
