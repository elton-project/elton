package localStorage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

const FileMode = 0400

type Repository struct {
	Path string
	// Maximum size of the object.
	// If MaxSize is 0, the size limit is disabled.
	MaxSize uint64
	KeyGen  KeyGenerator
}
type Key struct {
	ID string
}

func (s *Repository) Create(body []byte) (key Key, err error) {
	key = s.KeyGen.Generate()
	p := s.objectPath(key)

	if s.MaxSize > 0 && uint64(len(body)) > s.MaxSize {
		// TODO: 独自のエラー型を定義し、それを返す。
		err = fmt.Errorf("body too large")
		return
	}

	var f *os.File
	f, err = os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, FileMode)
	if err != nil {
		return
	}
	_, err = f.Write(body)
	if err != nil {
		f.Close()
		return
	}
	err = f.Close()
	return
}
func (s *Repository) Get(key Key, offset, size uint64) ([]byte, error) {
	p := s.objectPath(key)
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	if _, err := f.Seek(int64(offset), 0); err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
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
func (s *Repository) Delete(key Key) (bool, error) {
	p := s.objectPath(key)
	err := os.Remove(p)
	if err != nil {
		if os.IsNotExist(err) {
			// The object is already deleted.  Ignore this error.
			return false, nil
		}
		// Unexpected error.
		return false, err
	}
	// Deleted the object.
	return true, nil
}
func (s *Repository) objectPath(key Key) string {
	fileName := key.ID
	return path.Join(s.Path, fileName)
}
