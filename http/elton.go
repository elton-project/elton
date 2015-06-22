package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	elton "../elton"
	"github.com/gorilla/mux"
)

var client *http.Client = &http.Client{
	Transport: &http.Transport{MaxIdleConnsPerHost: 32},
}

type Elton struct {
	Conf     elton.Config
	Registry *elton.Registry
	FS       *elton.FileSystem
	Backup   bool
}

// type Result struct {
// 	Name    string
// 	Version string
// 	Length  int64
// }

// type Transport struct {
// 	Elton  *Elton
// 	Backup bool
// }

func NewEltonServer(conf elton.Config, backup bool) (*Elton, error) {
	fs := elton.NewFileSystem(conf.Elton.Dir)
	if backup {
		return &Elton{Conf: conf, FS: fs, Backup: backup}, nil
	}

	registry, err := elton.NewRegistry(conf)
	if err != nil {
		return nil, err
	}

	return &Elton{Conf: conf, Registry: registry, FS: fs, Backup: false}, nil
}

func (e *Elton) Serve() {
	defer e.Registry.Close()

	var srv http.Server
	srv.Addr = e.Conf.Elton.HostName
	e.RegisterHandler(&srv)

	log.Fatal(srv.ListenAndServe())
}

func (e *Elton) RegisterHandler(srv *http.Server) {
	router := mux.NewRouter()
	//	router.HandleFunc("/api/stats", stats_api.Handler).Method("GET")
	router.HandleFunc(
		"/api/objectid",
		e.GetObjectsIDHandler,
	).Methods("GET").Queries("objects", "{objects}")
	router.HandleFunc(
		"/api/objects/new",
		e.PutObjectsInfoHandler,
	).Methods("PUT")

	router.HandleFunc(
		"/{host:(elton[1-9][0-9]+)}/{id}/{version:([1-9][0-9]+)}",
		e.GetObjectHandler,
	).Methods("GET")
	router.HandleFunc(
		"/{host:(elton[1-9][0-9]+)}/{id}",
		e.DeleteObjectHandler,
	).Methods("DELETE")

	srv.Handler = router
}

// ex. Get /api/objectid?objects={"name":["a.txt","foo/b.txt","bar/c.txt"]}
//     Response [{"objectid":"elton1/ab9d90d90ad0a9d"},{"objectid":"elton1/ab9d90d99ad0a9d"},{"objectid":"elton1/ab9d90d90ad0d9d"}]
func (e *Elton) GetObjectsIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	objects := vars["objects"]

	var names elton.FileName
	if err := json.Unmarshal(objects, &names); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	list, err := e.Registry.GetObjectsID(names)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	result, _ := json.Marshal(&list)
	fmt.Fprintf(w, string(result))
}

// ex. Put /api/objects/new -F objects=[{"objectid":"elton1/ab9d90d90ad0a9d","delegate":"192.168.1.1:12345"},{"objectid":"elton1/ab9d90d99ad0a9d","delegate":"192.168.1.1:12345"},{"objectid":"elton1/ab9d90d90ad0d9d","delegate":"192.168.1.1:12345"}]
//     Response [{"objectid":"elton1/ab9d90d90ad0a9d","version":1,"delegate":"192.168.1.1:12345"},{"objectid":"elton1/ab9d90d99ad0a9d","version":1,"delegate":"192.168.1.1:12345"},{"objectid":"elton1/ab9d90d90ad0d9d","version":1,"delegate":"192.168.1.1:12345"}]
func (e *Elton) PutObjectsInfoHandler(w http.ResponseWriter, r *http.Request) {
	var objects []ObjectInfo
	if err := json.Unmarshal(r.FormValue("objects"), &objectsInfo); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	list, err := e.Registry.SetObjectsInfo(objects)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	result, _ := json.Marshal(&list)
	fmt.Fprintf(w, string(result))
}

func (e *Elton) GetObjectHandler(w http.ResponseWriter, r *http.Request) {

}

func (e *Elton) DeleteObjectHandler(w http.ResponseWriter, r *http.Request) {
}

// func (e *Elton) newReverseProxy(host, name string) *httputil.ReverseProxy {
// 	rp := &httputil.ReverseProxy{Director: func(request *http.Request) {
// 		request.URL.Scheme = "http"
// 		request.URL.Host = host
// 		request.URL.Path = path.Join("/api", "elton", name)
// 	}}
// 	rp.Transport = &Transport{Elton: e}
// 	return rp
// }

// func (e *Elton) doBackup(key string) {
// 	file, err := e.FS.Open(key)
// 	if err != nil {
// 		log.Printf("Can not backup: %v", err)
// 		return
// 	}
// 	defer file.Close()

// 	body := new(bytes.Buffer)
// 	writer := multipart.NewWriter(body)

// 	part, err := writer.CreateFormFile("file", key)
// 	if err != nil {
// 		log.Printf("Can not backup: %v", err)
// 		return
// 	}

// 	if _, err = io.Copy(part, file); err != nil {
// 		log.Printf("Can not backup: %v", err)
// 		return
// 	}
// 	writer.Close()

// 	req, _ := http.NewRequest("PUT", "http://"+path.Join(e.Conf.Backup.HostName+":"+e.Conf.Backup.Port, "api", "elton", key), body)
// 	req.Header.Add("Content-Type", writer.FormDataContentType())
// 	res, err := client.Do(req)
// 	if err != nil || res.StatusCode != http.StatusOK {
// 		log.Printf("Can not backup: %s", res.StatusCode)
// 		return
// 	}
// 	defer res.Body.Close()

// 	if err = e.Registry.RegisterBackup(key); err != nil {
// 		log.Printf("Can not backup: %v", err)
// 		return
// 	}
// }

// func (t *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
// 	name := strings.TrimPrefix(request.URL.Path, "/api/elton/")

// 	response, err := http.DefaultTransport.RoundTrip(request)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if response.StatusCode == http.StatusOK {
// 		if err := t.Elton.FS.Create(name, response.Body); err != nil {
// 			return nil, err
// 		}

// 		file, err := t.Elton.FS.Open(name)
// 		if err != nil {
// 			return nil, err
// 		}
// 		response.Body = ioutil.NopCloser(file)
// 	}

// 	return response, err
// }

// func parsePath(path string) (string, string) {
// 	name := strings.TrimPrefix(path, "/elton/")
// 	list := strings.Split(name, "-")
// 	version, err := strconv.ParseUint(list[len(list)-1], 10, 64)
// 	if err != nil {
// 		return name, "0"
// 	}
// 	return list[:len(list)-1][0], strconv.FormatUint(version, 10)
// }
