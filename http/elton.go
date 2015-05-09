package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"path"
	"strconv"
	"strings"

	"github.com/fukata/golang-stats-api-handler"

	e "../elton"
)

type Elton struct {
	Conf     e.Config
	Registry *e.Registry
	FS       *e.FileSystem
}

type getTransport struct {
	Elton *Elton
}

type EltonTransport struct {
	Registry *e.Registry
	Target   string
}

type FileList struct {
	Host  string        `json:"host"`
	Files []e.EltonFile `json:"files"`
}

func NewElton(conf e.Config) (*Elton, error) {
	registry, err := e.NewRegistry(conf)
	if err != nil {
		return nil, err
	}

	return &Elton{Conf: conf, Registry: registry, FS: e.NewFileSystem(conf.Elton.Dir)}, nil
}

func (e *Elton) Serve() {
	defer e.Registry.Close()

	var srv http.Server
	srv.Addr = ":" + e.Conf.Proxy.Port
	e.RegisterHandler(&srv)

	log.Fatal(srv.ListenAndServe())
}

func (e *Elton) RegisterHandler(srv *http.Server) {
	mux := http.NewServeMux()
	mux.HandleFunc("/maint/stats", stats_api.Handler)

	mux.HandleFunc("/api/elton/", e.GetFileApiHandler)
	mux.HandleFunc("/elton/", e.DispatchFileHandler)
	srv.Handler = mux
}

func (e *Elton) DispatchFileHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		e.getFileHandler(w, r)
	case "PUT":
		e.putFileHandler(w, r)
	case "DELETE":
		e.deleteFileHandler(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (e *Elton) getFileHandler(w http.ResponseWriter, r *http.Request) {
	name, version := parsePath(r.URL.Path)
	data, err := e.Registry.GetHost(name, version)
	if err != nil || data.Host == "" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	localPath := filepath.Join(e.FS.RootDir, data.Path)
	if err := e.FS.Find(localPath); err != nil {
		// must isLocalHost(data.Host) check
		rp := &httputil.ReverseProxy{Director: func(request *http.Request) {
			request.URL.Scheme = "http"
			request.URL.Host = data.Host
			request.URL.Path = data.Path
		}}
		rp.Transport = &getTransport{Elton: e}
		rp.ServeHTTP(w, r)
		return
	}

	http.ServeFile(w, r, localPath)
}

func (e *Elton) putFileHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path

	version, err := p.Registry.GenerateNewVersion(name)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	key, err := s.FS.Create(name+"-"+version, file)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err = RegisterNewVersion(name, key, e.Conf.Host, r.ContentLength); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	result, _ := json.Marshal(&Result{Name: name, Version: version, Length: r.ContentLength})
	fmt.Fprintf(w, string(result))
}

func (e *Elton) deleteFileHandler(w http.ResponseWriter, r *http.Request) {
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

func (t *getTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	name, version := parsePath(request.URL.Path)
	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	localPath := filepath.Join(t.FS.RootDir, name)
	if response.StatusCode == http.StatusOK {
		t.Elton.FS.Create(name+"-"+version, response.Body.Close())
	} else {

	}

	return response, err
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

func parsePath(path string) (string, string) {
	paths := strings.SplitN(path, "elton", 2)
	name := paths[1]
	//	name := strings.TrimPrefix(path, "/elton")
	list := strings.Split(name, "-")
	version, err := strconv.ParseUint(list[len(list)-1], 64, 10)
	if err != nil {
		return name, "0"
	}
	return name, strconv.FormatUint(version, 10)
}
