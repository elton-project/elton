package localStorage

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/yuuki0xff/pathlib"
	"golang.org/x/xerrors"
	"io/ioutil"
	"os"
	"testing"
)

func withTempRepo(fn func(repo *Repository)) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	repo := Repository{
		BasePath: pathlib.New(dir),
		MaxSize:  5,
		KeyGen:   &RandomKeyGen{},
	}
	fn(&repo)
}

func TestRepository_Create(t *testing.T) {
	withTempRepo(func(repo *Repository) {
		_, err := repo.Create([]byte("test"))
		assert.NoError(t, err)

		_, err = repo.Create([]byte("testt"))
		assert.NoError(t, err)

		_, err = repo.Create([]byte("testte"))
		assert.Error(t, err, "body too big")
	})
}
func TestRepository_Get(t *testing.T) {
	withTempRepo(func(repo *Repository) {
		key, err := repo.Create([]byte("test"))
		assert.NoError(t, err)

		// Get with the valid key.
		b, err := repo.Get(key, 0, 0)
		assert.Equal(t, b, []byte("test"))
		assert.Nil(t, err)

		// Get with invalid key.
		b, err = repo.Get(Key{"invalid"}, 0, 0)
		assert.Nil(t, b)
		fmt.Println(err)
		assert.True(t, xerrors.Is(err, &ObjectNotFoundError{}))

		// Try to get an object with the not found key>
		b, err = repo.Get(Key{"not found"}, 0, 0)
		assert.Nil(t, b)
		fmt.Println(err)
		assert.True(t, xerrors.Is(err, &ObjectNotFoundError{}))
	})
}
func TestExists(t *testing.T) {
	withTempRepo(func(repo *Repository) {
		key, err := repo.Create([]byte("test"))
		assert.NoError(t, err)

		ok, err := repo.Exists(key)
		assert.NoError(t, err)
		assert.True(t, ok)

		ok, err = repo.Exists(Key{"not found"})
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}
