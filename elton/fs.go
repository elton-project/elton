package elton

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

type FileSystem struct {
	RootDir string
	Backup  []BackupConfig
}

func NewFileSystem(dir string, backup []BackupConfig) *FileSystem {
	return &FileSystem{RootDir: dir, Backup: backup}
}

func (fs *FileSystem) Create(name string, file multipart.File) (string, error) {
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
	defer file.Close()

	return key, nil
}

func (fs *FileSystem) Create(name string, body io.ReadCloser) (string, error) {
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
	io.Copy(out, body)
	defer body.Close()

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
