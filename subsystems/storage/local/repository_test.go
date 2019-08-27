package localStorage

import (
	"crypto/sha1"
	"github.com/stretchr/testify/assert"
	"github.com/yuuki0xff/pathlib"
	"golang.org/x/xerrors"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func withTempRepo(maxSize uint64, fn func(repo *Repository)) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	repo := NewRepository(pathlib.New(dir), nil, maxSize)
	fn(repo)
}
func withTempRepoAndObject(maxSize uint64, objs [][]byte, fn func(repo *Repository, keys []Key)) {
	withTempRepo(maxSize, func(repo *Repository) {
		var keys []Key
		for _, obj := range objs {
			hash := sha1.Sum(obj)
			info := Info{
				Hash:          hash[:],
				HashAlgorithm: "SHA1",
				CreateTime:    time.Now(),
				Size:          uint64(len(obj)),
			}
			key, err := repo.Create(obj, info)
			if err != nil {
				panic(err)
			}
			keys = append(keys, key)
		}

		fn(repo, keys)
	})
}

func TestRepository_Create(t *testing.T) {
	body := []byte("test")
	hash := sha1.Sum(body)
	info := Info{
		Hash:          hash[:],
		HashAlgorithm: "SHA1",
		CreateTime:    time.Now(),
		Size:          uint64(len(body)),
	}

	t.Run("normal-case", func(t *testing.T) {
		withTempRepo(10, func(repo *Repository) {
			_, err := repo.Create(body, info)
			assert.NoError(t, err)
		})
	})
	t.Run("max-length", func(t *testing.T) {
		withTempRepo(uint64(len(body)), func(repo *Repository) {
			_, err := repo.Create(body, info)
			assert.NoError(t, err)
		})
	})
	t.Run("size-over", func(t *testing.T) {
		withTempRepo(1, func(repo *Repository) {
			_, err := repo.Create(body, info)
			assert.Error(t, err, "body too big")
		})
	})
	t.Run("mismatch-size", func(t *testing.T) {
		withTempRepo(10, func(repo *Repository) {
			info2 := info
			info2.Size = 1
			_, err := repo.Create(body, info2)
			assert.Error(t, err, "hoge")
		})
	})
	t.Run("mismatch-hash", func(t *testing.T) {
		withTempRepo(10, func(repo *Repository) {
			info2 := info
			info2.Hash = []byte("bla bla")
			_, err := repo.Create(body, info2)
			assert.Error(t, err, "hoge")
		})
	})
}
func TestRepository_Get(t *testing.T) {
	objs := [][]byte{
		[]byte("test"),
	}
	t.Run("with-the-valid-key", func(t *testing.T) {
		withTempRepoAndObject(10, objs, func(repo *Repository, keys []Key) {
			body, info, err := repo.Get(keys[0], 0, 0)
			assert.Nil(t, err)
			assert.Equal(t, body, []byte("test"))
			if assert.NotNil(t, info){
				// TODO: more test cases
			}
		})
	})
	t.Run("not-found", func(t *testing.T) {
		withTempRepoAndObject(10, objs, func(repo *Repository, keys []Key) {
			body, info, err := repo.Get(Key{"not found"}, 0, 0)
			assert.True(t, xerrors.Is(err, &ObjectNotFoundError{}))
			assert.Nil(t, body)
			assert.Nil(t, info)
		})
	})
}
func TestExists(t *testing.T) {
	objs := [][]byte{
		[]byte("test"),
	}
	t.Run("found", func(t *testing.T) {
		withTempRepoAndObject(10, objs, func(repo *Repository, keys []Key) {
			ok, err := repo.Exists(keys[0])
			assert.NoError(t, err)
			assert.True(t, ok)
		})
	})
	t.Run("not-found", func(t *testing.T) {
		withTempRepoAndObject(10, objs, func(repo *Repository, keys []Key) {
			ok, err := repo.Exists(Key{"not found"})
			assert.NoError(t,err)
			assert.False(t, ok)
		})
	})
}
