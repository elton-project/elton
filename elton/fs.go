package elton

import (
	"crypto/md5"
	"encoding/hex"
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

func (fs *FileSystem) Create(name string, src io.Reader) (string, error) {
	key := generateKey(name)
	path := filepath.Join(fs.RootDir, key)
	err := fs.mkDir(path)
	if err != nil {
		return "", err
	}

	out, err := os.Create(path)
	if err != nil {
		log.Printf("Can not create file: %s", path)
		return "", err
	}
	defer out.Close()

	log.Printf("Create path: %s", path)
	_, err = io.Copy(out, src)

	return key, err
}

func (fs *FileSystem) Find(name string) (path string, err error) {
	path = filepath.Join(fs.RootDir, name)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		log.Printf("No such file: %s", path)
		return "", err
	}
	return path, nil
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

func generateKey(name string) string {
	hasher := md5.New()
	hasher.Write([]byte(name + time.Now().String()))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return string(hash[:2] + "/" + hash[2:])
}
