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

func TestSetHost(t *testing.T) {
	proxy := NewProxy(DBPATH)
	defer proxy.Close()

	//URL.Pathからとってるので '/:dir/:key/:version' の形ですね
	err := proxy.SetHost("/hoge/hideo.txt/0", "localhost:56789")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	err = proxy.SetHost("/hoge/hideo.txt/1", "localhost:56789")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestGetHost(t *testing.T) {
	proxy := NewProxy(DBPATH)
	defer proxy.Close()

	_, err := proxy.GetHost("hoge", "hideo.txt", "0")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	_, err = proxy.GetHost("hoge", "hideo.txt", "1")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	_, err = proxy.GetHost("hoge", "hideo.txt", "2")
	if err == nil {
		t.Fatalf("Expected Error")
	}
}

func TestGetNewVersion(t *testing.T) {
	proxy := NewProxy(DBPATH)
	defer proxy.Close()

	version, err := proxy.GetNewVersion("hoge", "hideo.txt")

	if version == "" {
		t.Fatalf("Error: %v", err)
	}
}

func TestDelete(t *testing.T) {
	proxy := NewProxy(DBPATH)
	defer proxy.Close()

	err := proxy.Delete("hoge", "hideo.txt")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestMigration(t *testing.T) {
	proxy := NewProxy(DBPATH)
	defer proxy.Close()
}
