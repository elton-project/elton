package api

import (
	"os"
	"testing"
)

var DBPATH = "./elton_test.db"

func init() {
	if _, err := os.Stat(DBPATH); os.IsNotExist(err) {
		os.Remove(DBPATH)
	}
}

func TestProxySetHost(t *testing.T) {
	proxy := NewProxy(DBPATH)
	defer proxy.Close()

	//URL.Pathからとってるので '/:dir/:key/:version' の形ですね
	err := proxy.SetHost("/hoge/hideo.txt/0", "localhost:56789")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	err = proxy.SetHost("/hoge/hideo.txt/1", "localhost:67890")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestProxyGetHost(t *testing.T) {
	proxy := NewProxy(DBPATH)
	defer proxy.Close()

	host, err := proxy.GetHost("hoge", "hideo.txt", "0")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if host != "localhost:56789" {
		t.Fatalf("Error: expected '%s', got '%s'", "localhost:56789", host)
	}

	host, err = proxy.GetHost("hoge", "hideo.txt", "1")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if host != "localhost:67890" {
		t.Fatalf("Error: expected '%s', got '%s'", "localhost:67890", host)
	}

	_, err = proxy.GetHost("hoge", "hideo.txt", "2")
	if err == nil {
		t.Fatalf("Expected error")
	}
}

func TestProxyGetNewVersion(t *testing.T) {
	proxy := NewProxy(DBPATH)
	defer proxy.Close()

	version, err := proxy.GetNewVersion("hoge", "hideo.txt")

	if version == "" {
		t.Fatalf("Error: %v", err)
	}
}

func TestProxyDelete(t *testing.T) {
	proxy := NewProxy(DBPATH)
	defer proxy.Close()

	err := proxy.Delete("hoge", "hideo.txt")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestProxyMigration(t *testing.T) {
	proxy := NewProxy(DBPATH)
	defer proxy.Close()
}
