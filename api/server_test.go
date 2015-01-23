package api

import (
	"os"
	"testing"
)

func TestGetDir(t *testing.T) {
	dir := "elton"
	host := "localhost:12345"
	server := NewServer(dir, host)

	d := server.GetDir()
	if dir != d {
		t.Fatalf("Error: expected '%s', get '%s'", dir, d)
	}
}

func TestGetHost(t *testing.T) {
	dir := "elton"
	host := "localhost:12345"
	server := NewServer(dir, host)

	h := server.GetHost()
	if host != h {
		t.Fatalf("Error: expected '%s', get '%s'", host, h)
	}
}

func TestRead(t *testing.T) {
	dir := "../"
	host := "localhost:12345"
	server := NewServer(dir, host)

	path, err := server.Read("", "", "")
	if dir != d {
		t.Fatalf("Error: expected '%s', get '%s'", dir, d)
	}
}

func TestGetDir(t *testing.T) {

}

func TestGetDir(t *testing.T) {

}

func TestGetDir(t *testing.T) {

}
