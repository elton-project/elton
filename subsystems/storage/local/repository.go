package localStorage

import (
	"encoding/hex"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"time"
)

const FileMode = 0400

type Repository struct {
	Path string

	// ランダムなキーを生成するための、乱数ジェネレータ。
	random *rand.Rand
}
type Key struct {
	ID string
}

func (s *Repository) Create(body []byte) (key Key, err error) {
	key = s.generateKey()
	p := s.objectPath(key)

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
}
func (s *Repository) Get(key Key) ([]byte, error) {
	p := s.objectPath(key)
	return ioutil.ReadFile(p)
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
	fileName := key.ID
	return path.Join(s.Path, fileName)
}
func (s *Repository) generateKey() Key {
	if s.random == nil {
		now := time.Now().UnixNano()
		src := rand.NewSource(now)
		s.random = rand.New(src)
	}

	b := make([]byte, 24)
	for i := 0; i < 24; i += 8 {
		u := s.random.Uint64()
		b[i+0] = byte(u & 7)
		u >>= 8
		b[i+1] = byte(u & 7)
		u >>= 8
		b[i+2] = byte(u & 7)
		u >>= 8
		b[i+3] = byte(u & 7)
		u >>= 8
		b[i+4] = byte(u & 7)
		u >>= 8
		b[i+5] = byte(u & 7)
		u >>= 8
		b[i+6] = byte(u & 7)
		u >>= 8
		b[i+7] = byte(u & 7)
	}

	return Key{
		ID: hex.EncodeToString(b),
	}
}
