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

func TestServerGetHandler(t *testing.T) {
	InitServer("../examples", false)
	mux := pat.New()
	mux.Get("/:dir/:key/:version", http.HandlerFunc(ServerGetHandler))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	res0, err := http.Get(ts.URL + "/hoge/hideo.txt/0")
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

	res1, err := http.Get(ts.URL + "/hoge/hideo.txt/1")
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
}

func TestServerPutHandler(t *testing.T) {
	InitServer("../examples", false)
	mux := pat.New()
	mux.Put("/:dir/:key/:version", http.HandlerFunc(ServerPutHandler))
	ts := httptest.NewServer(mux)
	defer ts.Close()

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
	req, _ := http.NewRequest("PUT", ts.URL+"/hideo/hideo.txt/0", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res0, err := client.Do(req)

	if err != nil {
		t.Fatalf("Error by http.Do(). %v", err)
	}

	data, err := ioutil.ReadAll(res0.Body)
	defer res0.Body.Close()
	if err != nil {
		t.Fatalf("Error by ioutil.ReadAll(). %v", err)
	}

	if `{"version":"0"}` != strings.Trim(string(data), "\n") {
		t.Fatalf("Data Error. %s", string(data))
	}

	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	part, err = writer.CreateFormFile("file", "hideo.txt")
	if err != nil {
		t.Fatalf("Can not create form file: %v", err)
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		t.Fatalf("Can not close writer: %v", err)
	}

	req, _ = http.NewRequest("PUT", ts.URL+"/hideo/hideo.txt/1", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res1, err := client.Do(req)

	if err != nil {
		t.Fatalf("Error by http.Do(). %v", err)
	}

	data, err = ioutil.ReadAll(res1.Body)
	defer res1.Body.Close()
	if err != nil {
		t.Fatalf("Error by ioutil.ReadAll(). %v", err)
	}

	if `{"version":"1"}` != strings.Trim(string(data), "\n") {
		t.Fatalf("Data Error. %v", string(data))
	}
}

func TestServerDeleteHandler(t *testing.T) {
	InitServer("../examples", false)
	mux := pat.New()
	mux.Del("/:dir/:key", http.HandlerFunc(ServerDeleteHandler))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("DELETE", ts.URL+"/hideo/hideo.txt", nil)
	_, err := client.Do(req)

	if err != nil {
		t.Fatalf("Error by http.Do(). %v", err)
	}

	if err = os.Remove("../examples/hideo"); err != nil {
		t.Fatalf("Can not delete directory: %v", err)
	}
}

func TestServerMigrationHandler(t *testing.T) {
	InitServer("../examples", true)
	mux := pat.New()
	mux.Get("/api/migration", http.HandlerFunc(ServerMigrationHandler))
	ts := httptest.NewServer(mux)

	res0, err := http.Get(ts.URL + "/api/migration")
	if err != nil {
		t.Fatalf("Error by http.Get(). %v", err)
	}

	data, err := ioutil.ReadAll(res0.Body)
	defer res0.Body.Close()
	if err != nil {
		t.Fatalf("Error by ioutil.ReadAll(). %v", err)
	}

	if `{"path":["elton_test.db","hoge/hideo.txt","hoge/hideo.txt-1"]}` != strings.Trim(string(data), "\n") {
		t.Fatalf("Data Error. %s", string(data))
	}

	ts.Close()

	InitServer("../examples", false)
	ts = httptest.NewServer(mux)

	res1, err := http.Get(ts.URL + "/api/migration")
	if err != nil {
		t.Fatalf("Error by http.Get(). %v", err)
	}

	data, err = ioutil.ReadAll(res1.Body)
	defer res1.Body.Close()
	if err != nil {
		t.Fatalf("Error by ioutil.ReadAll(). %v", err)
	}

	if "" != strings.Trim(string(data), "\n") {
		t.Fatalf("Data Error. %s", string(data))
	}

	ts.Close()
}
