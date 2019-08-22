package localStorage

import (
	"fmt"
	"github.com/yuuki0xff/pathlib"
	"io"
	"io/ioutil"
	"os"
)

type Repository struct {
	BasePath pathlib.Path
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

	err = p.WriteBytes(body)
	return
}
func (s *Repository) Get(key Key, offset, size uint64) ([]byte, error) {
	p := s.objectPath(key)

	f, err := p.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if size == 0 {
		// Without size limit.  Use ReadAll() function.
		if _, err := f.Seek(int64(offset), 0); err != nil {
			return nil, err
		}
		return ioutil.ReadAll(f)
	} else {
		// With size limit.  Allocate buffer and use ReadAt().
		if s.MaxSize > 0 && s.MaxSize < size {
			size = s.MaxSize
		}
		buf := make([]byte, size)
		n, err := f.ReadAt(buf, int64(offset))
		if err != nil && err != io.EOF {
			return nil, err
		}
		return buf[:n], nil
	}
}
func (s *Repository) Exists(key Key) (bool, error) {
	p := s.objectPath(key)
	return p.Exists(), nil
}
func (s *Repository) Delete(key Key) (bool, error) {
	p := s.objectPath(key)
	err := p.Unlink()
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
func (s *Repository) objectPath(key Key) pathlib.Path {
	fileName := key.ID
	return s.BasePath.JoinPath(fileName)
}
