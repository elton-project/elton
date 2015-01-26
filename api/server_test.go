package api

import (
	"mime/multipart"
	"os"
	"reflect"
	"testing"
)

var dir = "../examples"
var host = "localhost:12345"
var s Server

func init() {
	s = NewServer(dir, host)
}

func TestServerGetDir(t *testing.T) {
	d := s.GetDir()
	if dir != d {
		t.Fatalf("Error: expected '%s', got '%s'", dir, d)
	}
}

func TestServerGetHost(t *testing.T) {
	h := s.GetHost()
	if host != h {
		t.Fatalf("Error: expected '%s', got '%s'", host, h)
	}
}

func TestServerRead(t *testing.T) {
	path, err := s.Read("hoge", "hideo.txt", "0")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if path != "../examples/hoge/hideo.txt" {
		t.Fatalf("Error: expected '%s', got '%s'", "../examples/hoge/hideo.txt", path)
	}

	path, err = s.Read("hoge", "hideo.txt", "1")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if path != "../examples/hoge/hideo.txt-1" {
		t.Fatalf("Error: expected '%s', got '%s'", "../examples/hoge/hideo.txt-1", path)
	}

	_, err = s.Read("hoge", "hideo.txt", "2")
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestServerCreate(t *testing.T) {
	fh := multipart.FileHeader{Filename: "../examples/hoge/hideo.txt"}

	file, _ := fh.Open()
	err := s.Create("hideo", "fuga.txt", "1", file)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if _, err := os.Stat("../examples/hideo/fuga.txt-1"); os.IsNotExist(err) {
		t.Fatalf("Error: expected '%s' file created, but does not found", "../examples/hoge/hideo.txt-2")
	}

	fh = multipart.FileHeader{Filename: "../examples/hoge/hideo.txt"}

	file, _ = fh.Open()
	err = s.Create("hideo", "fuga.txt", "2", file)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if _, err := os.Stat("../examples/hideo/fuga.txt-2"); os.IsNotExist(err) {
		t.Fatalf("Error: expected '%s' file created, but does not found", "../examples/hoge/hideo.txt-2")
	}
}

func TestServerMigration(t *testing.T) {
	path := s.Migration()
	truePath := []string{"elton_test.db", "hideo/fuga.txt-1", "hideo/fuga.txt-2", "hoge/hideo.txt", "hoge/hideo.txt-1"}

	if !reflect.DeepEqual(path, truePath) {
		t.Fatalf("Error: expected '%v', got '%v'", truePath, path)
	}
}

func TestServerDelete(t *testing.T) {
	err := s.Delete("hideo", "fuga.txt")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if err = os.Remove("../examples/hideo"); err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestServerPormatPath(t *testing.T) {
	path := s.FormatPath("hoge", "hideo.txt", "0")
	if path != "../examples/hoge/hideo.txt" {
		t.Fatalf("Error: expected '%s', got '%s'", "../examples/hoge/hideo.txt", path)
	}

	path = s.FormatPath("hoge", "hideo.txt", "1")
	if path != "../examples/hoge/hideo.txt-1" {
		t.Fatalf("Error: expected '%s', got '%s'", "../examples/hoge/hideo.txt-1", path)
	}
}
