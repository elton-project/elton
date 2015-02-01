package api

import (
	"os"
	"testing"
)

var DBPATH = "../examples/elton_test.db"
var SERVERS = []string{"localhost:12345", "localhost:13579"}

func init() {
	if _, err := os.Stat(DBPATH); !os.IsNotExist(err) {
		os.Remove(DBPATH)
	}
}

func TestProxySetHost(t *testing.T) {
	proxy := NewProxy(DBPATH, SERVERS)
	defer proxy.Close()

	err := proxy.SetHost("hoge/hideo.txt/0", "localhost:56789")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	err = proxy.SetHost("hoge/hideo.txt/1", "localhost:67890")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestProxyGetLatestVersion(t *testing.T) {
	proxy := NewProxy(DBPATH, SERVERS)
	defer proxy.Close()

	version, err := proxy.GetLatestVersion("hoge", "hideo.txt")

	if version == "" {
		t.Fatalf("Error: %v", err)
	}
}

func TestProxyGetHost(t *testing.T) {
	proxy := NewProxy(DBPATH, SERVERS)
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
	proxy := NewProxy(DBPATH, SERVERS)
	defer proxy.Close()

	version, err := proxy.GetNewVersion("hoge", "hideo.txt")

	if version == "" {
		t.Fatalf("Error: %v", err)
	}
}

func TestProxyDelete(t *testing.T) {
	proxy := NewProxy(DBPATH, SERVERS)
	defer proxy.Close()

	err := proxy.Delete("hoge", "hideo.txt")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestProxyMigration(t *testing.T) {
	proxy := NewProxy(DBPATH, SERVERS)
	defer proxy.Close()

	err := proxy.Migration([]string{"aaa/bbb", "aaa/bbb-1", "aaa/bb10-1-100"}, "localhost:12345")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}
