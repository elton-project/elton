package controller_db

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"golang.org/x/xerrors"
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

func TestLocalVS_Exists(t *testing.T) {
	t.Run("should_return_true_when_access_exist_volume", func(t *testing.T) {
		withLocalDB(t, func(vs VolumeStore, cs CommitStore) {
			id, err := vs.Create(&VolumeInfo{
				Name: "dummy",
			})
			if !assert.Nil(t, err) || !assert.NotNil(t, id) {
				return
			}

			ok, err := vs.Exists(id)
			assert.Nil(t, err)
			assert.True(t, ok)
		})
	})
	t.Run("should_return_false_when_access_not_exists_volume", func(t *testing.T) {
		withLocalDB(t, func(vs VolumeStore, cs CommitStore) {
			notExistsID := &VolumeID{
				Id: "33221100",
			}
			ok, err := vs.Exists(notExistsID)
			assert.Nil(t, err)
			assert.False(t, ok)
		})
	})
}
func TestLocalVS_Delete(t *testing.T) {
	// TODO
}

func TestLocalVS_Walk(t *testing.T) {
	t.Run("should_not_callback_when_emtpy", func(t *testing.T) {
		withLocalDB(t, func(vs VolumeStore, cs CommitStore) {
			err := vs.Walk(func(id *VolumeID, info *VolumeInfo) error {
				t.Error("callback function is called when walking the empty bucket")
				return nil
			})
			assert.Nil(t, err)
		})
	})
	t.Run("should_all_volumes_appeared_when_walking", func(t *testing.T) {
		withLocalDB(t, func(vs VolumeStore, cs CommitStore) {
			_, err := vs.Create(&VolumeInfo{Name: "vol-1"})
			assert.Nil(t, err)
			_, err = vs.Create(&VolumeInfo{Name: "vol-2"})
			assert.Nil(t, err)
			_, err = vs.Create(&VolumeInfo{Name: "vol-3"})
			assert.Nil(t, err)

			volumes := map[string]bool{
				"vol-1": false,
				"vol-2": false,
				"vol-3": false,
			}
			err = vs.Walk(func(id *VolumeID, info *VolumeInfo) error {
				name := info.GetName()
				isAppeared, ok := volumes[name]
				if !assert.True(t, ok) || !assert.False(t, isAppeared) {
					return xerrors.New("something wrong")
				}

				volumes[name] = true
				return nil
			})
			assert.Nil(t, err)
		})
	})
}

func TestLocalCS_Get(t *testing.T) {
	t.Run("should_error_when_access_not_exists_commit", func(t *testing.T) {
		withLocalDB(t, func(vs VolumeStore, cs CommitStore) {
			vid, err := vs.Create(&VolumeInfo{Name: "foo"})
			if !assert.Nil(t, err) || !assert.NotNil(t, vid) {
				return
			}

			ci, err := cs.Get(&CommitID{
				Id:     vid,
				Number: 0,
			})
			assert.Error(t, err)
			assert.Nil(t, ci)
		})
	})
	t.Run("should_success_when_access_exists_commit", func(t *testing.T) {
		withLocalDB(t, func(vs VolumeStore, cs CommitStore) {
			vid, err := vs.Create(&VolumeInfo{Name: "foo"})
			if !assert.Nil(t, err) || !assert.NotNil(t, vid) {
				return
			}

			cid, err := cs.Create(vid, &CommitInfo{
				CreatedAt: ptypes.TimestampNow(),
				ParentID:  nil,
				TreeID:    nil,
			}, &Tree{
				P2I: map[string]uint64{
					"/":            1,
					"/bin":         2,
					"/bin/busybox": 3,
				},
				I2F: map[uint64]*FileID{
					1: {Id: "root"},
					2: {Id: "bin"},
					3: {Id: "busybox"},
				},
			})
			if !assert.Nil(t, err) || !assert.NotNil(t, cid) {
				return
			}

			ci, err := cs.Get(cid)
			assert.Nil(t, err)
			assert.NotNil(t, ci)
		})
	})
}
