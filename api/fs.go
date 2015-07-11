package api

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type FileSystem struct {
	RootDir string
}

// ディスク容量をチェックする必要がある

func NewFileSystem(dir string) *FileSystem {
	return &FileSystem{RootDir: dir}
}

func (fs *FileSystem) Open(key string) (*os.File, error) {
	path := filepath.Join(fs.RootDir, key)
	return os.Open(path)
}

func (fs *FileSystem) Create(key string, src io.Reader) error {
	path := filepath.Join(fs.RootDir, key)
	if err := fs.mkDir(path); err != nil {
		return err
	}

	out, err := os.Create(path)
	if err != nil {
		log.Printf("Can not create file: %s", path)
		return err
	}
	defer out.Close()

	log.Printf("Create path: %s", path)
	_, err = io.Copy(out, src)
	return err
}

func (fs *FileSystem) Find(name string, version uint64) (path string, err error) {
	path := filename(name, version)
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

func (fs *FileSystem) GenerateKey(name string) string {
	hasher := md5.New()
	hasher.Write([]byte(name + time.Now().String()))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return string(hash[:2] + "/" + hash[2:])
}
