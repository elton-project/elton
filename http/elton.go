package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"path"
	"strconv"

	elton "../elton"
	stats_api "github.com/fukata/golang-stats-api-handler"
	"github.com/gorilla/mux"
)

var client *http.Client = &http.Client{
	Transport: &http.Transport{MaxIdleConnsPerHost: 32},
}

type Elton struct {
	Conf     elton.Config
	Registry *elton.Registry
}

type Transport struct {
	Registry *elton.Registry
	Object   elton.ObjectInfo
}

func NewEltonServer(conf elton.Config) (*Elton, error) {
	registry, err := elton.NewRegistry(conf)
	if err != nil {
		return nil, err
	}

	return &Elton{Conf: conf, Registry: registry}, nil
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

	router.HandleFunc(
		"/maint/stats",
		stats_api.Handler,
	).Methods("GET")
	// router.HandleFunc(
	// 	"/maint/db",
	// 	e.BackupHandleFunc,
	// ).Methods("GET")

	router.HandleFunc(
		"/api/objectid",
		e.GenerateObjectsIDHandler,
	).Methods("PUT").Headers("X-Elton-Hostname", "")
	router.HandleFunc(
		"/api/objects/new",
		e.PutObjectsInfoHandler,
	).Methods("PUT").Headers("X-Elton-Hostname", "")

	router.HandleFunc(
		"/{host:(elton[1-9][0-9]+)}/{id}}",
		e.GetObjectHandler,
	).Methods("GET").Headers("X-Elton-Hostname", "")
	router.HandleFunc(
		"/{delegate:(elton[1-9][0-9]+)}/{id}/{version:([1-9][0-9]+)}",
		e.GetObjectHandler,
	).Methods("GET").Headers("X-Elton-Hostname", "")
	router.HandleFunc(
		"/{delegate:(elton[1-9][0-9]+)}/{id}",
		e.DeleteObjectHandler,
	).Methods("DELETE").Headers("X-Elton-Hostname", "")

	srv.Handler = router
}

// ex. Put /api/objectid -H X-Elton-Hostname: "localhost:2345" -d {"name":["a.txt","foo/b.txt","bar/c.txt"]}
//     Response [{"objectid":"ab9d90d90ad0a9d","delegate":"elton1"},{"objectid":"ab9d90d99ad0a9d","delegate":"elton1"},{"objectid":"elton1/ab9d90d90ad0d9d","delegate":"elton1"}]
func (e *Elton) GenerateObjectsIDHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v", r)

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var names elton.FileName
	if err := json.Unmarshal([]byte(body), &names); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	list, err := e.Registry.GenerateObjectsInfo(names)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	result, _ := json.Marshal(&list)
	fmt.Fprintf(w, string(result))
}

// ex. Put /api/objects/new -H X-Elton-Hostname: "localhost:2345" -d [{"objectid":"ab9d90d90ad0a9d","delegate":"elton1"},{"objectid":"ab9d90d99ad0a9d","delegate":"elton1"},{"objectid":"ab9d90d90ad0d9d","delegate":"elton1"}]
//     Response [{"objectid":"ab9d90d90ad0a9d","version":1,"delegate":"elton1"},{"objectid":"ab9d90d99ad0a9d","version":1,"delegate":"elton1"},{"objectid":"ab9d90d90ad0d9d","version":1,"delegate":"elton1"}]
func (e *Elton) PutObjectsInfoHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v", r)

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var objects []elton.ObjectInfo
	if err := json.Unmarshal([]byte(body), &objects); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	list, err := e.getNewVersions(objects)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err = e.Registry.SetObjectsInfo(list, r.Header.Get("X-Elton-Hostname")); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	result, _ := json.Marshal(&list)
	fmt.Fprintf(w, string(result))
}

func (e *Elton) getNewVersions(objects []elton.ObjectInfo) (result []elton.ObjectInfo, err error) {
	fobjs, oobjs, err := e.Registry.GetNewVersions(objects)
	if err != nil {
		return nil, err
	}

	result = append(result, fobjs...)

	if len(oobjs) > 0 {
		for _, master := range e.Conf.Masters {
			objs := make([]elton.ObjectInfo, 0)
			for _, obj := range objects {
				if master.Name == obj.Delegate {
					objs = append(objs, obj)
				}
			}

			objs, err = e.request(master.HostName, objs)
			if err != nil {
				return nil, err
			}

			result = append(result, objs...)
		}
	}

	return result, nil
}

func (e *Elton) request(hostname string, objects []elton.ObjectInfo) (objs []elton.ObjectInfo, err error) {
	jsonStr, err := json.Marshal(objects)
	req, _ := http.NewRequest(
		"PUT",
		fmt.Sprintf("http://%s", path.Join(hostname, "api", "objects", "new")),
		bytes.NewBuffer(jsonStr),
	)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	} else if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Response StatusCode: %d", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(body), &objs)
	if err != nil {
		return nil, err
	}

	return objs, nil
}

// ex. Get /elton1/ab9d90d90ad0a9d/1
//     Response [ab9d90d90ad0a9d/1]
func (e *Elton) GetObjectHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v", r)

	vars := mux.Vars(r)
	delegate := vars["delegate"]
	oid := vars["id"]
	version, err := strconv.ParseUint(vars["version"], 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	host, err := e.Registry.GetObjectHost(oid, version)
	if err != nil {
		for _, master := range e.Conf.Masters {
			if master.Name == delegate {
				host = master.HostName
				break
			}
		}

		if host == "" {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
	}

	rp := &httputil.ReverseProxy{
		Director: func(request *http.Request) {
			request.URL.Host = host
		},
	}

	rp.Transport = &Transport{
		Registry: e.Registry,
		Object: elton.ObjectInfo{
			ObjectID: oid,
			Version:  version,
		},
	}
	rp.ServeHTTP(w, r)
}

func (e *Elton) DeleteObjectHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v", r)
	vars := mux.Vars(r)
	oid := vars["id"]

	if err := e.Registry.DeleteObject(oid); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// func (e *Elton) BackupHandleFunc(w http.ResponseWriter, req *http.Request) {
// 	if err := e.Registry.DB.View(func(tx *bolt.Tx) error {
// 		w.Header().Set("Content-Type", "application/octet-stream")
// 		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, e.Conf.Database.DBPath))
// 		w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
// 		_, err := tx.WriteTo(w)
// 		return err
// 	}); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 	}
// }

func (t *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode == http.StatusOK {
		if err = t.Registry.SetObjectsInfo([]elton.ObjectInfo{t.Object}, request.Header.Get("X-Elton-Hostname")); err != nil {
			return nil, err
		}
	}

	return response, err
}
