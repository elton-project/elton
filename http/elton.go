package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"path"
	"strconv"
	"strings"

	"github.com/fukata/golang-stats-api-handler"

	e "../elton"
)

var client *http.Client = &http.Client{
	Transport: &http.Transport{MaxIdleConnsPerHost: 32},
}

type Elton struct {
	Conf     e.Config
	Registry *e.Registry
	FS       *e.FileSystem
	Backup   bool
}

type Result struct {
	Name    string
	Version string
	Length  int64
}

type Transport struct {
	Elton  *Elton
	Backup bool
}

func NewEltonServer(conf e.Config, backup bool) (*Elton, error) {
	fs := e.NewFileSystem(conf.Elton.Dir)
	if backup {
		return &Elton{Conf: conf, FS: fs, Backup: backup}, nil
	}

	registry, err := e.NewRegistry(conf)
	if err != nil {
		return nil, err
	}

	return &Elton{Conf: conf, Registry: registry, FS: fs, Backup: false}, nil
}

func (e *Elton) Serve() {
	defer e.Registry.Close()

	var srv http.Server
	srv.Addr = ":" + e.Conf.Elton.Port
	e.RegisterHandler(&srv)

	log.Fatal(srv.ListenAndServe())
}

func (e *Elton) RegisterHandler(srv *http.Server) {
	mux := http.NewServeMux()
	mux.HandleFunc("/maint/stats", stats_api.Handler)

	mux.HandleFunc("/api/elton/", e.DispatchFileAPIHandler)
	if !e.Backup {
		mux.HandleFunc("/api/list", e.GetListHandler)
		mux.HandleFunc("/elton/", e.DispatchFileHandler)
	}

	srv.Handler = mux
}

func (e *Elton) DispatchFileAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		e.getFileAPIHandler(w, r)
	case r.Method == "PUT" && e.Backup:
		e.putFileAPIHandler(w, r)
		//	case r.Method=="DELETE":
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
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

func (e *Elton) GetListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	list, err := e.Registry.GetList()
	if err != nil {
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

func (e *Elton) getFileAPIHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/elton")

	localPath, err := e.FS.Find(name)
	if err != nil {
		if e.Backup {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		rp := e.newReverseProxy(e.Conf.Backup.HostName+":"+e.Conf.Backup.Port, name)
		rp.ServeHTTP(w, r)
		return
	}

	http.ServeFile(w, r, localPath)
}

func (e *Elton) putFileAPIHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/elton")

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err = e.FS.Create(name, file); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (e *Elton) getFileHandler(w http.ResponseWriter, r *http.Request) {
	name, version := parsePath(r.URL.Path)

	result, err := e.Registry.GetHost(name, version)
	if err != nil || result.Host == "" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	localPath, err := e.FS.Find(result.Path)
	if err != nil {
		if result.Host == (e.Conf.Elton.HostName + ":" + e.Conf.Elton.Port) {
			result.Host = e.Conf.Backup.HostName + ":" + e.Conf.Backup.Port
		}

		rp := e.newReverseProxy(result.Host, result.Path)
		rp.ServeHTTP(w, r)
	} else {
		http.ServeFile(w, r, localPath)
	}
}

func (e *Elton) putFileHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/elton")

	version, err := e.Registry.GenerateNewVersion(name)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer file.Close()

	key := e.FS.GenerateKey(name + "-" + version)
	if err := e.FS.Create(key, file); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err = e.Registry.RegisterNewVersion(name, version, key, e.Conf.Elton.HostName+":"+e.Conf.Elton.Port, r.ContentLength); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	go e.doBackup(key)

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

func (e *Elton) newReverseProxy(host, name string) *httputil.ReverseProxy {
	rp := &httputil.ReverseProxy{Director: func(request *http.Request) {
		request.URL.Scheme = "http"
		request.URL.Host = host
		request.URL.Path = path.Join("/api", "elton", name)
	}}
	rp.Transport = &Transport{Elton: e}
	return rp
}

func (e *Elton) doBackup(key string) {
	file, err := e.FS.Open(key)
	if err != nil {
		log.Printf("Can not backup: %v", err)
		return
	}
	defer file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", key)
	if err != nil {
		log.Printf("Can not backup: %v", err)
		return
	}

	if _, err = io.Copy(part, file); err != nil {
		log.Printf("Can not backup: %v", err)
		return
	}
	writer.Close()

	req, _ := http.NewRequest("PUT", "http://"+path.Join(e.Conf.Backup.HostName+":"+e.Conf.Backup.Port, "api", "elton", key), body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		log.Printf("Can not backup: %v", res.StatusCode)
		return
	}
	defer res.Body.Close()

	if err = e.Registry.RegisterBackup(key); err != nil {
		log.Printf("Can not backup: %v", err)
		return
	}
}

func (t *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	name := strings.TrimPrefix(request.URL.Path, "/api/elton")

	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode == http.StatusOK {
		if err := t.Elton.FS.Create(name, response.Body); err != nil {
			return nil, err
		}

		file, err := t.Elton.FS.Open(name)
		if err != nil {
			return nil, err
		}
		response.Body = ioutil.NopCloser(file)
	}

	return response, err
}

func parsePath(path string) (string, string) {
	name := strings.TrimPrefix(path, "/elton")
	list := strings.Split(name, "-")
	version, err := strconv.ParseUint(list[len(list)-1], 10, 64)
	if err != nil {
		return name, "0"
	}
	return list[:len(list)-1][0], strconv.FormatUint(version, 10)
}
