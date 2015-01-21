package fs

import (
	"nashio-lab.info/elton/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var DBPATH = "../elton.db"

func init() {
	if _, err := os.Stat(DBPATH); os.IsNotExist(err) {
		os.Remove(DBPATH)
	}
}

func TestProxyPut(t *testing.T) {
	fs.ProxyInitialize(DBPATH)
	defer fs.ProxyDestory()

	proxy := httptest.NewServer(http.HandlerFunc(fs.ProxyPut))
	defer proxy.Close()
	server := httptest.NewServer(http.HandlerFunc(fs.ServerPut))
	defer server.Close()

}

func TestProxyGet(t *testing.T) {
	fs.ProxyInitialize(DBPATH)
	defer fs.ProxyDestory()

	proxy := httptest.NewServer(http.HandlerFunc(fs.ProxyGet))
	defer proxy.Close()
	server := httptest.NewServer(http.HandlerFunc(fs.ServerGet))
	defer server.Close()
}
