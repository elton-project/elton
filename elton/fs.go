package elton

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type FileSystem struct {
	RootDir string
}

func NewFileSystem(dir string) *FileSystem {
	return &FileSystem{RootDir: dir}
}

func (fs *FileSystem) Create(name, version string, file *os.File) (string, error) {
	key := generateKey(name)
	path := filepath.Join(fs.RootDir, key)
	fs.mkDir(path)

	out, err := os.Create(path)
	if err != nil {
		log.Printf("[elton server] Can not create file: %s", path)
		return "", errors.New("Can not create file: " + path)
	}
	defer out.Close()

	log.Printf("[elton server] Create path: %s", path)
	io.Copy(out, file)
	return key, nil
}

func (fs *FileSystem) Find(name string) error {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return errors.New("No such file: " + name)
	}
	return nil
}

func (fs *FileSystem) mkDir(path string) {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0700)
	}
}

func generateKey(name string) string {
	hasher := md5.New()
	hasher.Write([]byte(name + time.Now().String()))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return string(hash[:2] + "/" + hash[2:])
}
