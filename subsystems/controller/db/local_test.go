package controller_db

import (
	"github.com/stretchr/testify/assert"
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"io/ioutil"
	"os"
	"testing"
)

func withLocalDB(t *testing.T, fn func(vs VolumeStore, cs CommitStore)) {
	dir, err := ioutil.TempDir("", "eltond")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(dir)

	vs, cs, closer, err := CreateLocalDB(dir)
	if err != nil {
		t.Error(err)
		return
	}
	defer closer()

	fn(vs, cs)
}

func TestLocalVS_Get(t *testing.T) {
	t.Run("should_error_when_access_not_found_volume", func(t *testing.T) {
		withLocalDB(t, func(vs VolumeStore, cs CommitStore) {
			notExistsID := &VolumeID{
				Id: "33221100",
			}
			info, err := vs.Get(notExistsID)
			assert.Error(t, err)
			assert.Nil(t, info)
		})
	})
	t.Run("should_success_when_access_exits_volume", func(t *testing.T) {
		withLocalDB(t, func(vs VolumeStore, cs CommitStore) {
			info := &VolumeInfo{
				Name: "dummy",
			}
			id, err := vs.Create(info)
			if !assert.NotNil(t, id) || !assert.Nil(t, err) {
				return
			}

			info2, err := vs.Get(id)
			assert.NotNil(t, info2)
			assert.Nil(t, err)
			assert.Equal(t, info, info2)
		})
	})
}
