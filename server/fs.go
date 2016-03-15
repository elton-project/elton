package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

type FileSystem struct {
	RootDir    string
	PurgeTimer *time.Timer
}

// TODO: DURATION_TIMEが一週間のマジックナンバーなのをなおす
const durationTime time.Duration = time.Duration(7 * 24 * uint64(time.Hour))

func NewFileSystem(dir string, backup bool) *FileSystem {
	fs := &FileSystem{RootDir: dir}

	if !backup {
		// TODO: 1時間間隔でPurgeチェックがマジックナンバーになってる
		fs.PurgeTimer = time.AfterFunc(
			time.Hour,
			func() {
				purge(dir)
			},
		)
	}

	return fs
}

func purge(dir string) {
	now := time.Now()

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		stat := info.Sys().(*syscall.Stat_t)
		atime := time.Unix(stat.Atim.Unix())

		if (atime.Add(durationTime)).Sub(now) < 0 {
			return os.Remove(path)
		}

		return nil
	})
}

func (fs *FileSystem) WriteFile(name string, version uint64, body []byte) error {
	path := fs.filename(name, version)
	if err := fs.mkDir(path); err != nil {
		return err
	}

	return ioutil.WriteFile(path, body, 0600)
}

func (fs *FileSystem) Create(name string, version uint64) (fp *os.File, err error) {
	path := fs.filename(name, version)
	if err = fs.mkDir(path); err != nil {
		return
	}

	return os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}

func (fs *FileSystem) Open(name string, version uint64) (fp *os.File, err error) {
	path, err := fs.Find(name, version)
	if err != nil {
		return
	}

	return os.Open(path)
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
