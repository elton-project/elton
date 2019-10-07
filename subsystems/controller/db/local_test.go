package controller_db

import (
	"errors"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"golang.org/x/xerrors"
	"io/ioutil"
	"os"
	"testing"
)

func withLocalDB(t *testing.T, fn func(stores Stores)) {
	dir, err := ioutil.TempDir("", "eltond")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(dir)

	stores, closer, err := CreateLocalDB(dir)
	if err != nil {
		t.Error(err)
		return
	}
	defer closer()

	fn(stores)
}

func TestLocalVS_Get(t *testing.T) {
	t.Run("should_error_when_access_not_found_volume", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
			notExistsID := &VolumeID{
				Id: "33221100",
			}
			info, err := vs.Get(notExistsID)
			assert.EqualError(t, err, "not found volume")
			assert.Nil(t, info)
		})
	})
	t.Run("should_success_when_access_exits_volume", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
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
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
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
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
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

func TestLocalVS_Create(t *testing.T) {
	t.Run("should_success_when_passed_valid_args", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
			vid, err := vs.Create(&VolumeInfo{
				Name: "foo",
			})
			assert.NoError(t, err)
			assert.NotEmpty(t, vid.GetId())
		})
	})
	t.Run("should_fail_when_creating_volume_with_duplicate_name", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
			_, err := vs.Create(&VolumeInfo{
				Name: "foo",
			})
			assert.NoError(t, err)

			_, err = vs.Create(&VolumeInfo{
				Name: "foo",
			})
			assert.EqualError(t, err, "duplicate volume name")
		})
	})
}

func TestLocalVS_Walk(t *testing.T) {
	t.Run("should_not_callback_when_emtpy", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
			err := vs.Walk(func(id *VolumeID, info *VolumeInfo) error {
				t.Error("callback function is called when walking the empty bucket")
				return nil
			})
			assert.Nil(t, err)
		})
	})
	t.Run("should_all_volumes_appeared_when_walking", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
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
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
			cs := stores.CommitStore()
			vid, err := vs.Create(&VolumeInfo{Name: "foo"})
			if !assert.Nil(t, err) || !assert.NotNil(t, vid) {
				return
			}

			ci, err := cs.Get(&CommitID{
				Id:     vid,
				Number: 0,
			})
			assert.EqualError(t, err, "not found commit")
			assert.Nil(t, ci)
		})
	})
	t.Run("should_success_when_access_exists_commit", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
			cs := stores.CommitStore()
			vid, err := vs.Create(&VolumeInfo{Name: "foo"})
			if !assert.Nil(t, err) || !assert.NotNil(t, vid) {
				return
			}

			cid, err := cs.Create(vid, &CommitInfo{
				CreatedAt:    ptypes.TimestampNow(),
				LeftParentID: nil,
				TreeID:       nil,
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

func TestLocalCS_Exists(t *testing.T) {
	t.Run("should_return_false_when_volume_not_found", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			cs := stores.CommitStore()
			ok, err := cs.Exists(&CommitID{
				Id: &VolumeID{Id: "not-found"},
			})
			assert.NoError(t, err)
			assert.False(t, ok)
		})
	})
	t.Run("should_return_false_when_commit_not_found", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
			cs := stores.CommitStore()
			// Create a volume.
			vid, err := vs.Create(&VolumeInfo{
				Name: "foo",
			})
			if !assert.NoError(t, err) {
				return
			}
			// Check whether the volume exists or not.
			ok, err := cs.Exists(&CommitID{
				Id:     vid,
				Number: 10,
			})
			assert.NoError(t, err)
			assert.False(t, ok)
		})
	})
	t.Run("should_return_true_when_commit_exists", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
			cs := stores.CommitStore()
			vid, err := vs.Create(&VolumeInfo{
				Name: "foo",
			})
			if !assert.NoError(t, err) {
				return
			}
			cid, err := cs.Create(vid, &CommitInfo{
				CreatedAt:    ptypes.TimestampNow(),
				LeftParentID: nil,
				TreeID:       nil,
			}, &Tree{})
			if !assert.NoError(t, err) {
				return
			}
			// Check whether the volume exists or not.
			ok, err := cs.Exists(cid)
			assert.NoError(t, err)
			assert.True(t, ok)
		})
	})
}

func TestLocalCS_Parents(t *testing.T) {
	t.Run("should_error_when_commit_is_not_found", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			cs := stores.CommitStore()
			left, right, err := cs.Parents(&CommitID{
				Id: &VolumeID{Id: "not-found"},
			})
			assert.EqualError(t, err, "not found commit")
			assert.Nil(t, left)
			assert.Nil(t, right)
		})
	})
	t.Run("should_return_nil_when_commit_is_first_commit", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
			cs := stores.CommitStore()

			vid, err := vs.Create(&VolumeInfo{
				Name: "foo",
			})
			if !assert.NoError(t, err) {
				return
			}
			cid, err := cs.Create(vid, &CommitInfo{
				CreatedAt: ptypes.TimestampNow(),
			}, &Tree{})
			if !assert.NoError(t, err) {
				return
			}
			assert.NotNil(t, cid)

			left, right, err := cs.Parents(cid)
			assert.NoError(t, err)
			assert.Nil(t, left)
			assert.Nil(t, right)
		})
	})
	t.Run("should_return_valid_commit_id", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
			cs := stores.CommitStore()

			vid, err := vs.Create(&VolumeInfo{
				Name: "foo",
			})
			if !assert.NoError(t, err) {
				return
			}
			cid, err := cs.Create(vid, &CommitInfo{
				CreatedAt: ptypes.TimestampNow(),
			}, &Tree{})
			if !assert.NoError(t, err) {
				return
			}
			cid2, err := cs.Create(vid, &CommitInfo{
				CreatedAt:    ptypes.TimestampNow(),
				LeftParentID: cid,
			}, &Tree{})
			if !assert.NoError(t, err) {
				return
			}
			assert.NotNil(t, cid2)

			left, right, err := cs.Parents(cid2)
			assert.NoError(t, err)
			assert.Equal(t, left, cid)
			assert.Nil(t, right)
		})
	})
}

func TestLocalCS_Latest(t *testing.T) {
	t.Run("should_error_when_volume_is_not_found", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			cs := stores.CommitStore()
			cid, err := cs.Latest(&VolumeID{
				Id: "not-found",
			})
			assert.EqualError(t, err, "not found commit")
			assert.Nil(t, cid)
		})
	})
	t.Run("should_error_when_volume_is_empty", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
			cs := stores.CommitStore()
			vid, err := vs.Create(&VolumeInfo{
				Name: "foo",
			})
			if !assert.NoError(t, err) {
				return
			}
			cid, err := cs.Latest(vid)
			assert.EqualError(t, err, "not found commit")
			assert.True(t, errors.Is(err, ErrNotFoundCommit))
			assert.Nil(t, cid)
		})
	})
	t.Run("should_return_valid_commit_id_when_volume_is_not_empty", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
			cs := stores.CommitStore()
			vid, err := vs.Create(&VolumeInfo{
				Name: "foo",
			})
			if !assert.NoError(t, err) {
				return
			}
			_, err = cs.Create(vid, &CommitInfo{}, &Tree{})
			if assert.NoError(t, err) {
				return
			}

			cid, err := cs.Latest(vid)
			assert.NoError(t, err)
			assert.NotNil(t, cid)
		})
	})
}

func TestLocalCS_Create(t *testing.T) {
	t.Run("should_error_when_volume_id_is_not_match", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
			cs := stores.CommitStore()

			vid1, err := vs.Create(&VolumeInfo{Name: "foo"})
			assert.NoError(t, err)
			vid2, err := vs.Create(&VolumeInfo{Name: "bar"})
			assert.NoError(t, err)

			cid, err := cs.Create(
				vid1,
				&CommitInfo{
					LeftParentID: &CommitID{
						Id: vid2,
					},
				},
				&Tree{},
			)
			assert.EqualError(t, err, "cross-volume commit")
			assert.Nil(t, cid)
		})
	})
	t.Run("should_error_when_volume_is_not_found", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			cs := stores.CommitStore()

			vid := &VolumeID{Id: "not-found"}
			cid, err := cs.Create(vid, &CommitInfo{}, &Tree{})
			assert.EqualError(t, err, "not found volume")
			assert.Nil(t, cid)
		})

	})
	t.Run("should_error_when_specified_parent_commit_id_is_not_found", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
			cs := stores.CommitStore()

			vid, err := vs.Create(&VolumeInfo{Name: "foo"})
			if !assert.NoError(t, err) {
				return
			}
			invalidCID := &CommitID{
				Id:     vid,
				Number: 100,
			}
			cid, err := cs.Create(vid, &CommitInfo{
				LeftParentID: invalidCID,
			}, &Tree{})
			assert.EqualError(t, err, "invalid parent commit")
			assert.Nil(t, cid)
		})
	})
}

func TestLocalCS_Tree(t *testing.T) {
	t.Run("should_error_when_access_not_exists_tree", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			cs := stores.CommitStore()
			tid, tree, err := cs.Tree(&CommitID{
				Id:     &VolumeID{Id: "not_found"},
				Number: 0,
			})
			assert.EqualError(t, err, "not found commit")
			assert.Nil(t, tid)
			assert.Nil(t, tree)
		})
	})
	t.Run("should_success_when_access_exists_tree", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			vs := stores.VolumeStore()
			cs := stores.CommitStore()
			vid, err := vs.Create(&VolumeInfo{
				Name: "vol",
			})
			if !assert.NotNil(t, vid) || !assert.Nil(t, err) {
				return
			}

			cid, err := cs.Create(vid, &CommitInfo{
				CreatedAt:    ptypes.TimestampNow(),
				LeftParentID: nil,
			}, &Tree{
				P2I: map[string]uint64{
					"/":    1,
					"/bin": 2,
				},
				I2F: map[uint64]*FileID{
					1: {Id: "root"},
					2: {Id: "bin"},
				},
			})
			if !assert.NotNil(t, cid) || !assert.Nil(t, err) {
				return
			}

			tid, tree, err := cs.Tree(cid)
			assert.Nil(t, err)
			assert.NotNil(t, tid)
			assert.NotNil(t, tree)
		})
	})
}

func TestLocalCS_TreeByTreeID(t *testing.T) {
	t.Run("should_error_when_access_not_exists_tree", func(t *testing.T) {
		withLocalDB(t, func(stores Stores) {
			cs := stores.CommitStore()
			tree, err := cs.TreeByTreeID(&TreeID{
				Id: "not_found",
			})
			assert.EqualError(t, err, "not found tree")
			assert.Nil(t, tree)
		})
	})
}
