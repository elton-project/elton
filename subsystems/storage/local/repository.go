package localStorage

import (
	"fmt"
	"os"
	"path"
)

const FileMode = 0400

type Repository struct {
	Path string
}
type Key struct {
	ID      string
	Version uint64
}

func (s *Repository) Create(key Key, body []byte) error {
	p := s.objectPath(key)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, FileMode)
	if err != nil {
		return err
	}
	_, err = f.Write(body)
	if err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
func (s *Repository) Exists(key Key) (bool, error) {
	p := s.objectPath(key)
	_, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
func (s *Repository) objectPath(key Key) string {
	fileName := fmt.Sprintf("%s---%d", key.ID, key.Version)
	return path.Join(s.Path, fileName)
}
