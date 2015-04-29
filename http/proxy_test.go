package http

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	e "../elton"
)

var sts *httptest.Server

func init() {
	conf, _ := e.Load("../examples/server_config.tml")
	s := NewServer(conf)

	mux := http.NewServeMux()
	mux.HandleFunc("/maint/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	})
	mux.HandleFunc("/", s.DispatchHandler)
	sts = httptest.NewServer(mux)
}

func TestProxyPutHandler(t *testing.T) {
	conf, _ := e.Load("../examples/proxy_config.tml")
	uri := strings.Split(strings.Trim(sts.URL, "http://"), ":")
	conf.Proxy.Servers[0].Host = uri[0]
	conf.Proxy.Servers[0].Port = uri[1]

	p, _ := NewProxy(conf)
	defer p.Registry.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/maint/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	})
	mux.HandleFunc("/api/connect", p.ConnectHandler)
	mux.HandleFunc("/api/list", p.GetListHandler)
	mux.HandleFunc("/", p.DispatchHandler)

	pts := httptest.NewServer(mux)
	defer pts.Close()

	file, err := os.Open("../examples/CIMG1138.JPG")
	if err != nil {
		t.Fatalf("Can not open file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "CIMG1138.JPG")

	if err != nil {
		t.Fatalf("Can not create form file: %v", err)
	}
	_, err = io.Copy(part, file)

	_ = writer.WriteField("host", "hideo")

	err = writer.Close()
	if err != nil {
		t.Fatalf("Can not close writer: %v", err)
	}

	client := &http.Client{}
	req, _ := http.NewRequest("PUT", pts.URL+"/hideo/CIMG1138.JPG", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)

	if err != nil {
		t.Fatalf("Error by http.Do(). %v", err)
	}

	data, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		t.Fatalf("Error by ioutil.ReadAll(). %v", err)
	}

	if `{"version":1,"length":251}` != strings.Trim(string(data), "\n") {
		t.Fatalf("Data Error. %s", string(data))
	}
}

// func TestProxyGetHandler(t *testing.T) {
// 	InitProxy(DBPATH, []string{strings.Trim(sts.URL, "http://")})
// 	defer DestoryProxy()
// 	pmux := pat.New()
// 	pmux.Get("/:dir/:key/:version", http.HandlerFunc(ProxyGetHandler))
// 	pmux.Get("/:dir/:key", http.HandlerFunc(ProxyGetHandler))
// 	pts := httptest.NewServer(pmux)
// 	defer pts.Close()

// 	res0, err := http.Get(pts.URL + "/hoge/hideo.txt/0")
// 	if err != nil {
// 		t.Fatalf("Error by http.Get(). %v", err)
// 	}

// 	data, err := ioutil.ReadAll(res0.Body)
// 	defer res0.Body.Close()
// 	if err != nil {
// 		t.Fatalf("Error by ioutil.ReadAll(). %v", err)
// 	}

// 	if "hideo.txt" != strings.Trim(string(data), "\n") {
// 		t.Fatalf("Data Error. %s", string(data))
// 	}

// 	res1, err := http.Get(pts.URL + "/hoge/hideo.txt/1")
// 	if err != nil {
// 		t.Fatalf("Error by http.Get(). %v", err)
// 	}

// 	data, err = ioutil.ReadAll(res1.Body)
// 	defer res1.Body.Close()
// 	if err != nil {
// 		t.Fatalf("Error by ioutil.ReadAll(). %v", err)
// 	}

// 	if "hideo.txt-1" != strings.Trim(string(data), "\n") {
// 		t.Fatalf("Data Error. %v", string(data))
// 	}

// 	res3, err := http.Get(pts.URL + "/hideo/hideo.txt")
// 	if err != nil {
// 		t.Fatalf("Error by http.Get(). %v", err)
// 	}

// 	data, err = ioutil.ReadAll(res3.Body)
// 	defer res3.Body.Close()
// 	if err != nil {
// 		t.Fatalf("Error by ioutil.ReadAll(). %v", err)
// 	}

// 	if "hideo.txt" != strings.Trim(string(data), "\n") {
// 		t.Fatalf("Data Error. %v", string(data))
// 	}
// }

// func TestProxyDeleteHandler(t *testing.T) {
// 	InitProxy(DBPATH, []string{strings.Trim(sts.URL, "http://")})
// 	defer DestoryProxy()
// 	pmux := pat.New()
// 	pmux.Del("/:dir/:key", http.HandlerFunc(ProxyDeleteHandler))
// 	pts := httptest.NewServer(pmux)
// 	defer pts.Close()

// 	client := &http.Client{}
// 	req, _ := http.NewRequest("DELETE", pts.URL+"/hideo/hideo.txt", nil)
// 	_, err := client.Do(req)

// 	if err != nil {
// 		t.Fatalf("Error by http.Do(). %v", err)
// 	}

// 	if err = os.Remove("../examples/hideo"); err != nil {
// 		t.Fatalf("Can not delete directory: %v", err)
// 	}

// 	sts.Close()
// }
