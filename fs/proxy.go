package fs

import (
	//	"container/ring"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
)

//var hostRing *Ring

type Transport struct {
}

var db *bolt.DB

func ProxyInitialize(path string) {
	var err error
	db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func ProxyDestory() {
	db.Close()
}

func (t *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode == http.StatusOK {
		key := []byte(response.Request.URL.Path)
		host := response.Request.URL.Host

		db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte("hosts"))
			if err != nil {
				return fmt.Errorf("create backet: %s", err)
			}

			return bucket.Put(key[1:], []byte(host))
		})
	}

	return response, err
}

func ProxyGet(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)
	params := r.URL.Query()
	dir := params.Get(":dir")
	key := params.Get(":key")
	version := params.Get(":version")

	var host []byte
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("hosts"))
		host = bucket.Get([]byte(dir + "/" + key + "/" + version))
		log.Printf("host: %s", host)
		return nil
	})

	if err != nil || host == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	proxy := &httputil.ReverseProxy{Director: func(request *http.Request) {
		request.URL.Scheme = "http"
		request.URL.Host = string(host)
	}}
	proxy.ServeHTTP(w, r)
}

func ProxyPut(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)
	params := r.URL.Query()
	dir := params.Get(":dir")
	key := params.Get(":key")

	var version []byte
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("versions"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		n, err := strconv.ParseUint(string(bucket.Get([]byte(dir+"/"+key))), 10, 64)
		if err != nil {
			n = 0
		}

		version = []byte(strconv.FormatUint(n+1, 10))
		return bucket.Put([]byte(dir+"/"+key), version)
	})

	if err != nil || version == nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	proxy := &httputil.ReverseProxy{Director: func(request *http.Request) {
		request.URL.Scheme = "http"
		request.URL.Host = "localhost:12345"
		request.URL.Path += "/" + string(version)
	}}
	proxy.Transport = &Transport{}
	proxy.ServeHTTP(w, r)
}

func ProxyMigration(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)
	key := r.FormValue("key")
	log.Println(key)
}
