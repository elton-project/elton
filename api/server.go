package api

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
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
	Migration() []string
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
		log.Printf("[elton server] No such file: %s", target)
		return "", errors.New("No such file: " + target)
	}

	log.Printf("[elton server] Read path: %s", target)
	return target, nil
}

func (s *server) Create(dir string, key string, version string, file multipart.File) error {
	if _, err := os.Stat(s.Dir + "/" + dir); os.IsNotExist(err) {
		os.Mkdir(s.Dir+"/"+dir, 0700)
	}

	out, err := os.Create(s.FormatPath(dir, key, version))
	if err != nil {
		log.Printf("[elton server] Can not create file: %s", s.FormatPath(dir, key, version))
		return errors.New("Can not create file: " + s.FormatPath(dir, key, version))
	}
	defer out.Close()

	log.Printf("[elton server] Create path: %s", s.FormatPath(dir, key, version))
	io.Copy(out, file)
	return nil
}

func (s *server) Delete(dir string, key string) error {
	if err := filepath.Walk(s.Dir+"/"+dir, func(path string, info os.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(), key) {
			log.Printf("[elton server] Delete path: %s", path)
			return os.Remove(path)
		}
		return nil
	}); err != nil {
		log.Printf("[elton server] Can not delete file: %s", err)
		return fmt.Errorf("Can not delete file: %s", err)
	}
	return nil
}

func (s *server) Migration() (paths []string) {
	filepath.Walk(s.Dir, func(p string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		paths = append(paths, strings.Replace(p, path.Clean(s.Dir)+"/", "", 1))
		log.Printf("[elton server] Migration path: %s", strings.Replace(p, path.Clean(s.Dir)+"/", "", 1))
		return nil
	})
	return
}

func (s *server) FormatPath(dir string, key string, version string) string {
	if version == "0" {
		return s.Dir + "/" + dir + "/" + key
	}
	return s.Dir + "/" + dir + "/" + key + "-" + version
}
