package api

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type server struct {
	Dir  string
	Host string
}

type Server interface {
	GetDir() string
	GetHost() string
	Read(string, string, string) (string, error)
	Create(string, string, string, multipart.File) error
	Delete(string, string) error
	Migration()
	FormatPath(string, string, string) string
}

func NewServer(dir string, host string) Server {
	return &server{dir, host}
}

func (s *server) GetDir() string {
	return s.Dir
}

func (s *server) GetHost() string {
	return s.Host
}

func (s *server) Read(dir string, key string, version string) (string, error) {
	target := s.FormatPath(dir, key, version)

	if _, err := os.Stat(target); os.IsNotExist(err) {
		return "", errors.New("No such file: " + target)
	}
	return target, nil
}

func (s *server) Create(dir string, key string, version string, file multipart.File) error {
	if _, err := os.Stat(s.Dir + "/" + dir); os.IsNotExist(err) {
		os.Mkdir(s.Dir+"/"+dir, 0700)
	}

	out, err := os.Create(s.formatPath(dir, key, version))
	if err != nil {
		return errors.New("Can not create file: " + s.formatPath(dir, key, version))
	}
	defer out.Close()

	io.Copy(out, file)
	return nil
}

func (s *server) Delete(dir string, key string) error {
	if err := filepath.Walk(s.Dir+"/"+dir, func(path string, info os.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(), key) {
			log.Println(path)
			return os.Remove(path)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("Can not delete file: %s", err)
	}
	return nil
}

func (s *server) Migration() {
	filepath.Walk(s.Dir, func(p string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		p = strings.Replace(p, path.Clean(s.Dir)+"/", "", 1)

		for {
			res, _ := http.PostForm(
				s.Host+"/api/migration",
				url.Values{"key": {p}},
			)
			if res.StatusCode == http.StatusOK {
				break
			}
		}

		return nil
	})
}

func (s *server) FormatPath(dir string, key string, version string) string {
	if version == "0" {
		return s.Dir + "/" + dir + "/" + key
	}
	return s.Dir + "/" + dir + "/" + key + "-" + version

}
