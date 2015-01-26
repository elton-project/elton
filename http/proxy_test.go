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

	"github.com/bmizerany/pat"
)

var DBPATH = "../examples/elton_test.db"
var sts *httptest.Server

func init() {
	if _, err := os.Stat(DBPATH); !os.IsNotExist(err) {
		os.Remove(DBPATH)
	}

	InitServer("../examples", true)
	smux := pat.New()
	smux.Get("/api/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	smux.Get("/api/migration", http.HandlerFunc(ServerMigrationHandler))
	smux.Get("/:dir/:key/:version", http.HandlerFunc(ServerGetHandler))
	smux.Put("/:dir/:key/:version", http.HandlerFunc(ServerPutHandler))
	sts = httptest.NewServer(smux)
}

func TestProxyPutHandler(t *testing.T) {
	InitProxy(DBPATH, []string{strings.Trim(sts.URL, "http://")})
	defer DestoryProxy()
	Migration()
	pmux := pat.New()
	pmux.Put("/:dir/:key", http.HandlerFunc(ProxyPutHandler))
	pts := httptest.NewServer(pmux)
	defer pts.Close()

	file, err := os.Open("../examples/hoge/hideo.txt")
	if err != nil {
		t.Fatalf("Can not open file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "hideo.txt")
	if err != nil {
		t.Fatalf("Can not create form file: %v", err)
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		t.Fatalf("Can not close writer: %v", err)
	}

	client := &http.Client{}
	req, _ := http.NewRequest("PUT", pts.URL+"/hideo/hideo.txt", body)
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

	if `{"version":"1"}` != strings.Trim(string(data), "\n") {
		t.Fatalf("Data Error. %s", string(data))
	}
}

func TestProxyGetHandler(t *testing.T) {
	InitProxy(DBPATH, []string{strings.Trim(sts.URL, "http://")})
	defer DestoryProxy()
	pmux := pat.New()
	pmux.Get("/:dir/:key/:version", http.HandlerFunc(ProxyGetHandler))
	pmux.Get("/:dir/:key", http.HandlerFunc(ProxyGetHandler))
	pts := httptest.NewServer(pmux)
	defer pts.Close()

	res0, err := http.Get(pts.URL + "/hoge/hideo.txt/0")
	if err != nil {
		t.Fatalf("Error by http.Get(). %v", err)
	}

	data, err := ioutil.ReadAll(res0.Body)
	defer res0.Body.Close()
	if err != nil {
		t.Fatalf("Error by ioutil.ReadAll(). %v", err)
	}

	if "hideo.txt" != strings.Trim(string(data), "\n") {
		t.Fatalf("Data Error. %s", string(data))
	}

	res1, err := http.Get(pts.URL + "/hoge/hideo.txt/1")
	if err != nil {
		t.Fatalf("Error by http.Get(). %v", err)
	}

	data, err = ioutil.ReadAll(res1.Body)
	defer res1.Body.Close()
	if err != nil {
		t.Fatalf("Error by ioutil.ReadAll(). %v", err)
	}

	if "hideo.txt-1" != strings.Trim(string(data), "\n") {
		t.Fatalf("Data Error. %v", string(data))
	}

	res3, err := http.Get(pts.URL + "/hideo/hideo.txt")
	if err != nil {
		t.Fatalf("Error by http.Get(). %v", err)
	}

	data, err = ioutil.ReadAll(res3.Body)
	defer res3.Body.Close()
	if err != nil {
		t.Fatalf("Error by ioutil.ReadAll(). %v", err)
	}

	if "hideo.txt" != strings.Trim(string(data), "\n") {
		t.Fatalf("Data Error. %v", string(data))
	}

	sts.Close()
}

func TestProxyDeleteHandler(t *testing.T) {
}

func TestMigration(t *testing.T) {
}
