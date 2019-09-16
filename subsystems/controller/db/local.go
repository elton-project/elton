package controller_db

import (
	"encoding/json"
	"fmt"
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems/idgen"
	"go.etcd.io/bbolt"
	"golang.org/x/xerrors"
	"os"
	"path"
)

var localVolumeBucket = []byte("volume")
var localCommitBucket = []byte("commit")

func CreateLocalDB(dir string) (vs VolumeStore, cs CommitStore, closer func() error, err error) {
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return
	}

	db := &localDB{
		Path: path.Join(dir, "db.bbolt"),
	}
	err = db.Open()
	if err != nil {
		err = xerrors.Errorf("db error: %w")
		return
	}

	closer = db.Close
	vs = &localVS{
		DB: db,
	}
	cs = &localCS{
		DB: db,
	}
	return
}

func mustMarshall(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
func mustUnmarshal(data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	if err != nil {
		panic(err)
	}
}

type localEncoder struct{}

func (localEncoder) VolumeID(id *VolumeID) []byte {
	return []byte(id.GetId())
}
func (localEncoder) VolumeInfo(info *VolumeInfo) []byte {
	return mustMarshall(info)
}
func (localEncoder) CommitID(id *CommitID) []byte {
	s := fmt.Sprintf("%s/%d", id.GetId().GetId(), id.GetNumber())
	return []byte(s)
}
func (localEncoder) CommitInfo(info *CommitInfo) []byte {
	return mustMarshall(info)
}

type localDecoder struct{}

func (localDecoder) VolumeID(data []byte) *VolumeID {
	id := &VolumeID{
		Id: string(data),
	}
	return id
}
func (localDecoder) VolumeInfo(data []byte) *VolumeInfo {
	info := &VolumeInfo{}
	mustUnmarshal(data, info)
	return info
}
func (localDecoder) CommitInfo(data []byte) *CommitInfo {
	info := &CommitInfo{}
	mustUnmarshal(data, info)
	return info
}

type localGenerator struct{}

func (localGenerator) next() uint64 {
	uniqId, err := idgen.Gen.NextID()
	if err != nil {
		panic(err)
	}
	return uniqId
}
func (g localGenerator) VolumeID() *VolumeID {
	return &VolumeID{
		Id: fmt.Sprintf("%x", g.next()),
	}
}
func (g localGenerator) CommitID(id *VolumeID) *CommitID {
	return &CommitID{
		Id:     id,
		Number: g.next(),
	}
}

type localTxFn func(b *bbolt.Bucket) error
type localDB struct {
	// Path to database file.
	Path string

	db *bbolt.DB
}

func (s *localDB) Open() error {
	db, err := bbolt.Open(s.Path, 0600, bbolt.DefaultOptions)
	if err != nil {
		return err
	}
	s.db = db

	return s.createAllBuckets()
}
func (s *localDB) Close() error {
	if s.db != nil {
		err := s.db.Close()
		s.db = nil
		return err
	}
	return nil
}
func (s *localDB) createAllBuckets() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(localVolumeBucket); err != nil {
			return xerrors.Errorf("volume bucket cannot create: %w", err)
		}

		if _, err := tx.CreateBucketIfNotExists(localCommitBucket); err != nil {
			return xerrors.Errorf("commit bucket cannot create: %w", err)
		}
		return nil
	})
}
func (s *localDB) runTx(writable bool, bucket []byte, callback localTxFn) error {
	innerFn := func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return xerrors.Errorf("not found an bucket: %s", string(bucket))
		}
		return callback(b)
	}

	if writable {
		return s.db.Update(innerFn)
	} else {
		return s.db.View(innerFn)
	}
}
func (s *localDB) VolumeView(callback localTxFn) error {
	return s.runTx(false, localVolumeBucket, callback)
}
func (s *localDB) VolumeUpdate(callback func(b *bbolt.Bucket) error) error {
	return s.runTx(true, localVolumeBucket, callback)
}
func (s *localDB) CommitView(callback localTxFn) error {
	return s.runTx(false, localCommitBucket, callback)
}
func (s *localDB) CommitUpdate(callback localTxFn) error {
	return s.runTx(true, localCommitBucket, callback)
}

type localVS struct {
	DB  *localDB
	Enc localEncoder
	Dec localDecoder
	Gen localGenerator
}

func (vs *localVS) Get(id *VolumeID) (vi *VolumeInfo, err error) {
	err = vs.DB.VolumeView(func(b *bbolt.Bucket) error {
		data := b.Get(vs.Enc.VolumeID(id))
		if len(data) > 0 {
			vi = vs.Dec.VolumeInfo(data)
			return nil
		}
		return xerrors.New("not found volume")
	})
	return
}
func (vs *localVS) Exists(id *VolumeID) (ok bool, err error) {
	err = vs.DB.VolumeView(func(b *bbolt.Bucket) error {
		data := b.Get(vs.Enc.VolumeID(id))
		ok = len(data) > 0
		return nil
	})
	return
}
func (vs *localVS) Delete(id *VolumeID) error {
	return xerrors.New("not implemented")
}
func (vs *localVS) Walk(callback func(id *VolumeID, info *VolumeInfo) error) error {
	return vs.DB.VolumeView(func(b *bbolt.Bucket) error {
		return b.ForEach(func(k, v []byte) error {
			id := vs.Dec.VolumeID(k)
			info := vs.Dec.VolumeInfo(v)
			return callback(id, info)
		})
	})
}
func (vs *localVS) Create(info *VolumeInfo) (id *VolumeID, err error) {
	id = vs.Gen.VolumeID()
	err = vs.DB.VolumeUpdate(func(b *bbolt.Bucket) error {
		return b.Put(
			vs.Enc.VolumeID(id),
			vs.Enc.VolumeInfo(info),
		)
	})
	return
}

type localCS struct {
	DB  *localDB
	Enc localEncoder
	Dec localDecoder
	Gen localGenerator
}

func (cs *localCS) Get(id *CommitID) (ci *CommitInfo, err error) {
	err = cs.DB.CommitView(func(b *bbolt.Bucket) error {
		data := b.Get(cs.Enc.CommitID(id))
		if len(data) > 0 {
			ci = cs.Dec.CommitInfo(data)
			return nil
		}
		return xerrors.New("not found commit")
	})
	return
}
func (cs *localCS) Exists(id *CommitID) (ok bool, err error) {
	err = cs.DB.CommitView(func(b *bbolt.Bucket) error {
		data := b.Get(cs.Enc.CommitID(id))
		ok = len(data) > 0
		return nil
	})
	return
}
func (cs *localCS) Parents(id *CommitID) (left *CommitID, right *CommitID, err error) {
	// todo
	err = xerrors.New("todo")
	return
}
func (cs *localCS) Latest() (latest *CommitID, err error) {
	// todo
	err = xerrors.New("todo")
	return
}
func (cs *localCS) Create(vid *VolumeID, info *CommitInfo) (id *CommitID, err error) {
	id = cs.Gen.CommitID(vid)
	err = cs.DB.CommitUpdate(func(b *bbolt.Bucket) error {
		return b.Put(
			cs.Enc.CommitID(id),
			cs.Enc.CommitInfo(info),
		)
	})
	return
}
