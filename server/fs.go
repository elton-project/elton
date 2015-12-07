package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type FileSystem struct {
	RootDir string
}

// TODO: ディスク容量をチェックする必要がある
// atimeとか見てduration 1week とかすると良いんじゃないかな

func NewFileSystem(dir string) *FileSystem {
	return &FileSystem{RootDir: dir}
}

func (fs *FileSystem) Create(name string, version uint64, body []byte) error {
	path := fs.filename(name, version)
	if err := fs.mkDir(path); err != nil {
		return err
	}

	return ioutil.WriteFile(path, body, 0600)
}

func (fs *FileSystem) Read(name string, version uint64) (body []byte, err error) {
	path := fs.filename(name, version)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		log.Printf("No such file: %s", path)
		return nil, err
	}

	return ioutil.ReadFile(path)
}

func (fs *FileSystem) Find(name string, version uint64) (path string, err error) {
	path = fs.filename(name, version)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		log.Printf("No such file: %s", path)
		return "", err
	}

	return path, nil
}

func (fs *FileSystem) Delete(name string) (err error) {
	path := filepath.Join(fs.RootDir, name)
	if err = os.Remove(name); err != nil {
		log.Printf("Can not delete file: %s", path)
	}
	return
}

func (fs *FileSystem) mkDir(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0700); err != nil {
			log.Printf("Can not create dir: %s", dir)
			return err
		}
	}

	return nil
}

func (fs *FileSystem) filename(name string, version uint64) string {
	return filepath.Join(fs.RootDir, name[:2], fmt.Sprintf("%s-%d", name[2:], version))
}
