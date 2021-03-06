package controller_db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems/idgen"
	"go.etcd.io/bbolt"
	"golang.org/x/xerrors"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

// File name of database file.
const localDbFileName = "db.bbolt"

// Meta bucket: It keeps properties.
// Key: PropertyID
// Value: Property
var localMetaBucket = []byte("meta")

// Volume bucket: It keeps VolumeInfo.
// - Key: VolumeID
// - Value: VolumeInfo (JSON encoded)
var localVolumeBucket = []byte("volume")

// Volume Name bucket: It is lookup table from the volume name to VolumeID.
// - Key: VolumeName
// - Value: VolumeID
var localVolumeNameBucket = []byte("volume-name")

// Commit bucket: It keeps Commit information.
// - Key: CommitID
// - Value: CommitInfo
var localCommitBucket = []byte("commit")

// Latest Commit bucket: It keeps the latest CommitID in each volume.
// - Key: VolumeID
// - Value: CommitID
var localLatestCommitBucket = []byte("latest-commit")

// Node bucket: It keeps node information.
// - Key: NodeID
// - Value: Node (JSON encoded)
var localNodeBucket = []byte("node")

// CreateLocalDB creates database accessors.  It saves data on local file system.
func CreateLocalDB(dir string) (stores Stores, closer func() error, err error) {
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		err = IErrInitialize.Wrap(err)
		return
	}

	db := &localDB{
		Path: path.Join(dir, localDbFileName),
	}
	err = db.Open()
	if err != nil {
		return
	}

	closer = db.Close
	stores = &localStores{
		localMS: localMS{DB: db},
		localVS: localVS{DB: db},
		localCS: localCS{DB: db},
		localNS: localNS{DB: db},
	}
	return
}

type localStores struct {
	localMS
	localVS
	localCS
	localNS
}

func (s *localStores) MetaStore() MetaStore     { return &s.localMS }
func (s *localStores) VolumeStore() VolumeStore { return &s.localVS }
func (s *localStores) CommitStore() CommitStore { return &s.localCS }
func (s *localStores) NodeStore() NodeStore     { return &s.localNS }

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
func (localEncoder) VolumeName(info *VolumeInfo) []byte {
	return []byte(info.GetName())
}
func (localEncoder) VolumeInfo(info *VolumeInfo) []byte {
	return mustMarshall(info)
}
func (localEncoder) CommitIDPrefix(id *VolumeID) []byte {
	s := fmt.Sprintf("%s/", id.GetId())
	return []byte(s)
}
func (localEncoder) CommitID(id *CommitID) []byte {
	s := fmt.Sprintf("%s/%d", id.GetId().GetId(), id.GetNumber())
	return []byte(s)
}
func (localEncoder) CommitInfo(info *CommitInfo) []byte {
	return mustMarshall(info)
}
func (localEncoder) Tree(tree *Tree) []byte {
	return mustMarshall(tree)
}
func (localEncoder) PropertyID(id *PropertyID) []byte {
	return []byte(id.Id)
}
func (localEncoder) Property(prop *Property) []byte {
	return mustMarshall(prop)
}
func (localEncoder) NodeID(id *NodeID) []byte {
	return []byte(id.GetId())
}
func (localEncoder) Node(node *Node) []byte {
	return mustMarshall(node)
}

type localDecoder struct{}

func (localDecoder) VolumeID(data []byte) *VolumeID {
	if data == nil {
		return nil
	}
	id := &VolumeID{
		Id: string(data),
	}
	return id
}
func (localDecoder) VolumeInfo(data []byte) *VolumeInfo {
	if data == nil {
		return nil
	}
	info := &VolumeInfo{}
	mustUnmarshal(data, info)
	return info
}
func (localDecoder) CommitID(data []byte) *CommitID {
	if data == nil {
		return nil
	}
	s := string(data)
	components := strings.SplitN(s, "/", 2)
	n, err := strconv.ParseUint(components[1], 10, 64)
	if err != nil {
		panic(xerrors.Errorf("Invalid Id (%s): %w", s, err))
	}
	return &CommitID{
		Id:     &VolumeID{Id: components[0]},
		Number: n,
	}
}
func (localDecoder) CommitInfo(data []byte) *CommitInfo {
	if data == nil {
		return nil
	}
	info := &CommitInfo{}
	mustUnmarshal(data, info)
	return info
}
func (localDecoder) Tree(data []byte) *Tree {
	if data == nil {
		return nil
	}
	tree := &Tree{}
	mustUnmarshal(data, tree)
	return tree
}
func (localDecoder) Property(data []byte) *Property {
	if data == nil {
		return nil
	}
	prop := &Property{}
	mustUnmarshal(data, prop)
	return prop
}
func (localDecoder) NodeID(data []byte) *NodeID {
	if data == nil {
		return nil
	}
	return &NodeID{
		Id: string(data),
	}
}
func (localDecoder) Node(data []byte) *Node {
	if data == nil {
		return nil
	}
	node := &Node{}
	mustUnmarshal(data, node)
	return node
}

type localGenerator struct{}

func (localGenerator) next() uint64 {
	uniqId, err := idgen.Gen.NextID()
	if err != nil {
		panic(err)
	}
	return uniqId
}
func (g localGenerator) nextString() string {
	return fmt.Sprintf("%x", g.next())
}
func (g localGenerator) VolumeID() *VolumeID {
	return &VolumeID{
		Id: g.nextString(),
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
		return IErrOpen.Wrap(err)
	}
	s.db = db

	return s.createAllBuckets()
}
func (s *localDB) Close() error {
	if s.db != nil {
		err := s.db.Close()
		s.db = nil
		if err != nil {
			return IErrClose.Wrap(err)
		}
	}
	return nil
}
func (s *localDB) createAllBuckets() error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(localMetaBucket); err != nil {
			return xerrors.Errorf("meta bucket cannot create: %w", err)
		}
		if _, err := tx.CreateBucketIfNotExists(localVolumeBucket); err != nil {
			return xerrors.Errorf("volume bucket cannot create: %w", err)
		}

		if _, err := tx.CreateBucketIfNotExists(localVolumeNameBucket); err != nil {
			return xerrors.Errorf("volume-name bucket cannot create: %w", err)
		}

		if _, err := tx.CreateBucketIfNotExists(localCommitBucket); err != nil {
			return xerrors.Errorf("commit bucket cannot create: %w", err)
		}

		if _, err := tx.CreateBucketIfNotExists(localLatestCommitBucket); err != nil {
			return xerrors.Errorf("latest commit bucket cannot create: %w", err)
		}

		if _, err := tx.CreateBucketIfNotExists(localNodeBucket); err != nil {
			return xerrors.Errorf("node bucket cannot create: %w", err)
		}
		return nil
	})
	if err != nil {
		return IErrInitialize.Wrap(err)
	}
	return nil
}
func (s *localDB) runTx(writable bool, bucket []byte, callback localTxFn) error {
	innerFn := func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return IErrDatabase.Wrap(fmt.Errorf("not found %s bucket", string(bucket)))
		}
		return callback(b)
	}

	if writable {
		return s.db.Update(innerFn)
	} else {
		return s.db.View(innerFn)
	}
}
func (s *localDB) View(callback func(tx *bbolt.Tx) error) error {
	return s.db.View(callback)
}
func (s *localDB) Update(callback func(tx *bbolt.Tx) error) error {
	return s.db.Update(callback)
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
func (s *localDB) MetaView(callback localTxFn) error {
	return s.runTx(false, localMetaBucket, callback)
}
func (s *localDB) MetaUpdate(callback localTxFn) error {
	return s.runTx(true, localMetaBucket, callback)
}
func (s *localDB) NodeView(callback localTxFn) error {
	return s.runTx(false, localNodeBucket, callback)
}
func (s *localDB) NodeUpdate(callback localTxFn) error {
	return s.runTx(true, localNodeBucket, callback)
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
		return ErrNotFoundVolume.Wrap(fmt.Errorf("id=%s", id))
	})
	return
}
func (vs *localVS) GetByName(name string) (id *VolumeID, vi *VolumeInfo, err error) {
	err = vs.DB.View(func(tx *bbolt.Tx) error {
		tmpVI := &VolumeInfo{
			Name: name,
		}
		vnb := tx.Bucket(localVolumeNameBucket)

		// Get VolumeID.
		data := vnb.Get(vs.Enc.VolumeName(tmpVI))
		if data == nil {
			return ErrNotFoundVolume.Wrap(fmt.Errorf("name=%s", name))
		}
		id = vs.Dec.VolumeID(data)
		return nil
	})

	if err == nil && id != nil {
		// Get VolumeID.
		vi, err = vs.Get(id)
	}
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
	return vs.DB.Update(func(tx *bbolt.Tx) error {
		vb := tx.Bucket(localVolumeBucket)
		vnb := tx.Bucket(localVolumeNameBucket)
		lcb := tx.Bucket(localLatestCommitBucket)
		cb := tx.Bucket(localCommitBucket)

		// Get volume info.
		data := vb.Get(vs.Enc.VolumeID(id))
		if len(data) == 0 {
			return ErrNotFoundVolume.Wrap(fmt.Errorf("id=%s", id))
		}
		info := vs.Dec.VolumeInfo(data)
		log.Printf("[INFO] Deleting %s volume", info.GetName())

		// Delete volume info.
		if err := vb.Delete(vs.Enc.VolumeID(id)); err != nil {
			return IErrDelete.Wrap(err)
		}
		if err := vnb.Delete(vs.Enc.VolumeName(info)); err != nil {
			return IErrDelete.Wrap(err)
		}

		// Get latest commit.
		data = lcb.Get(vs.Enc.VolumeID(id))
		if len(data) == 0 {
			// Volume is empty.  We don't need to delete commits and trees.
			return nil
		}
		if err := lcb.Delete(vs.Enc.VolumeID(id)); err != nil {
			return IErrDelete.Wrap(err)
		}

		// Delete commits and trees.
		prefix := vs.Enc.CommitIDPrefix(id)
		if err := bboltPrefixScan(cb, prefix, func(k, v []byte) error {
			log.Printf("[INFO] Deleting commit %s", k)
			return cb.Delete(k)
		}); err != nil {
			return IErrDelete.Wrap(err)
		}
		return nil
	})
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
	err = vs.DB.Update(func(tx *bbolt.Tx) error {
		vb := tx.Bucket(localVolumeBucket)
		vnb := tx.Bucket(localVolumeNameBucket)
		cb := tx.Bucket(localCommitBucket)
		lcb := tx.Bucket(localLatestCommitBucket)

		// Duplication check.
		if vb.Get(vs.Enc.VolumeID(id)) != nil {
			return ErrDupVolumeID.Wrap(fmt.Errorf("id=%s", id))
		}
		if vnb.Get(vs.Enc.VolumeName(info)) != nil {
			return ErrDupVolumeName.Wrap(fmt.Errorf("name=%s", info.GetName()))
		}

		// Save volume info.
		if err := vb.Put(
			vs.Enc.VolumeID(id),
			vs.Enc.VolumeInfo(info),
		); err != nil {
			return err
		}
		if err := vnb.Put(
			vs.Enc.VolumeName(info),
			vs.Enc.VolumeID(id),
		); err != nil {
			return err
		}

		// Create empty commit.
		newCID := vs.Gen.CommitID(id)
		if err := cb.Put(
			vs.Enc.CommitID(newCID),
			vs.Enc.CommitInfo(firstCommit()),
		); err != nil {
			return err
		}
		return lcb.Put(
			vs.Enc.VolumeID(id),
			vs.Enc.CommitID(newCID),
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
		return ErrNotFoundCommit.Wrap(fmt.Errorf("id=%s", id))
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
	err = cs.DB.CommitView(func(b *bbolt.Bucket) error {
		data := b.Get(cs.Enc.CommitID(id))
		if len(data) > 0 {
			info := cs.Dec.CommitInfo(data)
			left = info.GetLeftParentID()
			right = info.GetRightParentID()
			return nil
		}
		return ErrNotFoundCommit.Wrap(fmt.Errorf("id=%s", id))
	})
	return
}
func (cs *localCS) Latest(vid *VolumeID) (latest *CommitID, err error) {
	err = cs.DB.View(func(tx *bbolt.Tx) error {
		key := cs.Enc.VolumeID(vid)
		data := tx.Bucket(localLatestCommitBucket).Get(key)
		if len(data) > 0 {
			latest = cs.Dec.CommitID(data)
			return nil
		}
		return ErrNotFoundCommit.Wrap(fmt.Errorf("no commit in volume"))
	})
	return
}
func (cs *localCS) Create(vid *VolumeID, info *CommitInfo, tree *Tree) (cid *CommitID, err error) {
	newCID := cs.Gen.CommitID(vid)
	info.Tree = tree

	left := info.GetLeftParentID()
	right := info.GetRightParentID()

	// Validate arguments.
	if left.GetId().GetId() != "" {
		// Request to create normal commit.
		if bytes.Compare(
			cs.Enc.VolumeID(vid),
			cs.Enc.VolumeID(left.GetId()),
		) != 0 {
			err = ErrCrossVolumeCommit.Wrap(fmt.Errorf("mismatch VolumeID and CommitInfo.LeftParentID"))
			return
		}
	}
	if right.GetId().GetId() != "" {
		// Request to create merge commit.
		if bytes.Compare(
			cs.Enc.VolumeID(vid),
			cs.Enc.VolumeID(right.GetId()),
		) != 0 {
			err = ErrCrossVolumeCommit.Wrap(fmt.Errorf("mismatch VolumeID and CommitInfo.RightParentID"))
			return
		}
		if bytes.Compare(
			cs.Enc.VolumeID(left.GetId()),
			cs.Enc.VolumeID(right.GetId()),
		) != 0 {
			err = ErrCrossVolumeCommit.Wrap(fmt.Errorf("mismatch VolumeID of CommitInfo.LeftParentID and CommitInfo.RightParentID"))
			return
		}
	}

	// Validate tree.
	if err2 := tree.FastValidate(); err2 != nil {
		err = ErrInvalidTree.Wrap(err2)
		return
	}

	err = cs.DB.Update(func(tx *bbolt.Tx) error {
		// Check whether the volume is exist.
		if tx.Bucket(localVolumeBucket).Get(cs.Enc.VolumeID(vid)) == nil {
			return ErrNotFoundVolume.Wrap(fmt.Errorf("id=%s", vid))
		}

		// Check whether parent commits are valid.
		lastCID := tx.Bucket(localLatestCommitBucket).Get(cs.Enc.VolumeID(vid))
		if !(lastCID != nil && left != nil) {
			// Invalid combination.
			return ErrInvalidParentCommit.Wrap(fmt.Errorf(
				"last commit=%s, left=%s, right=%s",
				cs.Dec.CommitID(lastCID), left, right,
			))
		}
		if tx.Bucket(localCommitBucket).Get(cs.Enc.CommitID(left)) == nil {
			// Specified left parent is not found.
			return ErrInvalidParentCommit.Wrap(fmt.Errorf("left parent commit is not found: %s", left))
		}
		if right != nil && tx.Bucket(localCommitBucket).Get(cs.Enc.CommitID(right)) == nil {
			// Right parent is specified.  But it is not found.
			return ErrInvalidParentCommit.Wrap(fmt.Errorf("right parent commit is not found: %s", right))
		}

		if err := tx.Bucket(localCommitBucket).Put(
			cs.Enc.CommitID(newCID),
			cs.Enc.CommitInfo(info),
		); err != nil {
			return err
		}

		binVid := cs.Enc.VolumeID(vid)
		latest := cs.Dec.CommitID(tx.Bucket(localLatestCommitBucket).Get(binVid))
		if latest.Equals(info.LeftParentID) {
			// New commit is based on the latest commit.  Should update latest CommitID.
			return tx.Bucket(localLatestCommitBucket).Put(binVid, cs.Enc.CommitID(newCID))
		}
		return nil
	})
	if err == nil {
		cid = newCID
	}
	return
}
func (cs *localCS) Tree(id *CommitID) (tree *Tree, err error) {
	var ci *CommitInfo
	ci, err = cs.Get(id)
	if err != nil {
		return
	}
	tree = ci.GetTree()
	return
}

type localMS struct {
	DB  *localDB
	Enc localEncoder
	Dec localDecoder
}

func (ms *localMS) Get(id *PropertyID) (prop *Property, err error) {
	err = ms.DB.MetaView(func(b *bbolt.Bucket) error {
		data := b.Get(ms.Enc.PropertyID(id))
		if len(data) > 0 {
			prop = ms.Dec.Property(data)
			return nil
		}
		return ErrNotFoundProp.Wrap(fmt.Errorf("id=%s", id))
	})
	return
}
func (ms *localMS) Set(id *PropertyID, prop *Property, mustCreate bool) (old *Property, err error) {
	err = ms.DB.MetaUpdate(func(b *bbolt.Bucket) error {
		data := b.Get(ms.Enc.PropertyID(id))
		if len(data) > 0 {
			if mustCreate {
				return ErrAlreadyExists.Wrap(fmt.Errorf("id=%s", id))
			}

			old = ms.Dec.Property(data)
			if !old.GetAllowReplace() {
				old = nil
				return ErrNotAllowedReplace.Wrap(fmt.Errorf("id=%s", id))
			}
		}

		return b.Put(
			ms.Enc.PropertyID(id),
			ms.Enc.Property(prop),
		)
	})
	return
}

type localNS struct {
	DB  *localDB
	Enc localEncoder
	Dec localDecoder
	Gen localGenerator
}

func (ns *localNS) Register(id *NodeID, node *Node) error {
	return ns.DB.NodeUpdate(func(b *bbolt.Bucket) error {
		key := ns.Enc.NodeID(id)
		if b.Get(key) != nil {
			return ErrNodeAlreadyExists.Wrap(fmt.Errorf("id=%s", id))
		}
		return b.Put(
			key,
			ns.Enc.Node(node),
		)
	})
}
func (ns *localNS) Unregister(id *NodeID) error {
	return ns.DB.NodeUpdate(func(b *bbolt.Bucket) error {
		key := ns.Enc.NodeID(id)
		return b.Delete(key)
	})
}
func (ns *localNS) Update(id *NodeID, callback func(node *Node) error) error {
	return ns.DB.NodeUpdate(func(b *bbolt.Bucket) error {
		key := ns.Enc.NodeID(id)
		data := b.Get(key)
		if data == nil {
			return ErrNotFoundNode.Wrap(fmt.Errorf("id=%s", id))
		}
		node := ns.Dec.Node(data)

		err := callback(node)
		if err != nil {
			return err
		}
		return b.Put(
			key,
			ns.Enc.Node(node),
		)
	})
}
func (ns *localNS) List(walker func(id *NodeID, node *Node) error) error {
	return ns.DB.NodeView(func(b *bbolt.Bucket) error {
		return b.ForEach(func(k, v []byte) error {
			id := ns.Dec.NodeID(k)
			node := ns.Dec.Node(v)
			return walker(id, node)
		})
	})
}

func bboltPrefixScan(b *bbolt.Bucket, prefix []byte, fn func(k, v []byte) error) error {
	c := b.Cursor()
	for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}
func firstCommit() *CommitInfo {
	now := ptypes.TimestampNow()
	return &CommitInfo{
		CreatedAt:     now,
		LeftParentID:  nil,
		RightParentID: nil,
		Tree: &Tree{
			RootIno: 1,
			Inodes: map[uint64]*File{
				1: {
					ContentRef: nil,
					FileType:   FileType_Directory,
					Mode:       0755,
					Owner:      0,
					Group:      0,
					Atime:      now,
					Mtime:      now,
					Ctime:      now,
					Major:      0,
					Minor:      0,
					Entries:    nil,
				},
			},
		},
	}
}
