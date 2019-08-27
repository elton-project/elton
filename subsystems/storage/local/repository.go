package localStorage

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/yuuki0xff/pathlib"
	"golang.org/x/xerrors"
	"io"
	"io/ioutil"
	"os"
	"time"
)

const maxMetadataSize = 16 << 20 // 16 KiB

type Repository struct {
	BasePath pathlib.Path
	KeyGen   KeyGenerator

	limit ObjectLimitV1
}
type Key struct {
	ID string
}
type Info struct {
	Hash          []byte
	HashAlgorithm string
	CreateTime    time.Time
	Size          uint64
}

// Save and load the cached object from file.
//
// Data format:
//     1 bytes: uint8:  Major version number  (Always 0x01)
//     8 bytes: uint64: Length of the header  (BigEndian)
//     n bytes: []byte: Header  (json marshalled)
//     8 bytes: uint64: Length of the body  (BigEndian)
//     n bytes: []byte: Body
type ObjectV1 struct {
	ObjectLimitV1

	// Content of object.
	Body []byte
	// Offset from first byte of body.
	Offset uint64
	// Metadata for the object.
	Info *Info
}

// ObjectLimitV1 is configuration of size limit for ObjectV1
type ObjectLimitV1 struct {
	MaxBodySize uint64
	MaxInfoSize uint64
}

func NewRepository(basePath pathlib.Path, keyGen KeyGenerator, maxSize uint64) *Repository {
	if keyGen == nil {
		keyGen = &RandomKeyGen{}
	}

	return &Repository{
		BasePath: basePath,
		KeyGen:   keyGen,
		limit: ObjectLimitV1{
			MaxBodySize: maxSize,
			MaxInfoSize: maxMetadataSize,
		},
	}
}
func (s *Repository) Create(body []byte, info Info) (Key, error) {
	if err := s.fillInfo(body, &info); err != nil {
		return Key{}, err
	}

	key := s.KeyGen.Generate()
	op := s.objectPath(key)
	tmp := s.tmpObjectPath(key)

	err := AtomicWrite(op, tmp, func(w io.Writer) error {
		obj := NewObjectV1(body, &info, s.limit)
		return obj.Save(w)
	})
	if err != nil {
		return Key{}, err
	}
	return key, nil
}
func (s *Repository) Get(key Key, offset, size uint64) ([]byte, *Info, error) {
	p := s.objectPath(key)

	f, err := p.Open()
	if err != nil {
		return nil, nil, NewObjectNotFoundError(key).Wrap(err)
	}
	defer f.Close()

	obj, err := LoadObjectV1(f, offset, size)
	if err != nil {
		return nil, nil, NewObjectNotFoundError(key).Wrap(err)
	}
	return obj.Body, obj.Info, nil
}
func (s *Repository) Exists(key Key) (bool, error) {
	p := s.objectPath(key)
	return p.Exists(), nil
}
func (s *Repository) Delete(key Key) (bool, error) {
	p := s.objectPath(key)
	err := p.Unlink()
	if err != nil {
		if os.IsNotExist(err) {
			// The object is already deleted.  Ignore this error.
			return false, nil
		}
		// Unexpected error.
		return false, err
	}
	// Deleted the object.
	return true, nil
}
func (s *Repository) objectPath(key Key) pathlib.Path {
	fileName := key.ID
	return s.BasePath.JoinPath("object", fileName)
}
func (s *Repository) tmpObjectPath(key Key) pathlib.Path {
	fileName := key.ID
	return s.BasePath.JoinPath("object.tmp", fileName)
}
func (s *Repository) fillInfo(body []byte, info *Info) error {
	if info.Hash == nil && info.HashAlgorithm == "" {
		hash := sha1.Sum(body)
		info.Hash = hash[:]
		info.HashAlgorithm = "SHA1"
	}
	if info.CreateTime.IsZero() {
		info.CreateTime = time.Now()
	}
	if info.Size == 0 {
		info.Size = uint64(len(body))
	}
	return nil
}

func NewObjectV1(body []byte, info *Info, limit ObjectLimitV1) *ObjectV1 {
	return &ObjectV1{
		ObjectLimitV1: limit,
		Body:          body,
		Info:          info,
	}
}
func LoadObjectV1(rs io.ReadSeeker, offset, size uint64) (*ObjectV1, error) {
	r := &ObjectV1{}
	err := WithMustReadSeeker(rs, func(rs io.ReadSeeker) error {
		var version uint8
		binary.Read(rs, binary.BigEndian, &version)
		if version != r.Version() {
			return xerrors.New("mismatch version")
		}

		var headerLen uint64
		binary.Read(rs, binary.BigEndian, &headerLen)
		if r.MaxInfoSize < headerLen {
			return NewMetadataTooLargeError().Wrap(nil)
		}
		jsInfo := make([]byte, headerLen)
		rs.Read(jsInfo)
		if err := json.Unmarshal(jsInfo, &r.Info); err != nil {
			return err
		}
		jsInfo = nil

		var bodyLen uint64
		binary.Read(rs, binary.BigEndian, &bodyLen)

		if size == 0 {
			// Without size limit.  Use ReadAll() function.
			rs.Seek(int64(offset), 1)
			r.Body, _ = ioutil.ReadAll(rs)
			return nil
		} else {
			// With size limit.  Allocate buffer and use ReadAt().
			if r.MaxBodySize < size {
				size = r.MaxBodySize
			}
			buf := make([]byte, size)
			rs.Seek(int64(offset), io.SeekCurrent)
			n, _ := rs.Read(buf)
			r.Body = buf[:n]
			return nil
		}
	})

	if err != nil {
		return nil, err
	}
	return r, nil
}
func (r *ObjectV1) Save(w io.Writer) error {
	if r.Info == nil || r.Body == nil {
		return xerrors.New("illegal argument on ObjectV1.Save()")
	}
	if uint64(len(r.Body)) > r.MaxBodySize {
		return NewObjectTooLargeError(uint64(len(r.Body)), r.MaxBodySize).Wrap(nil)
	}
	if r.Info.Size != uint64(len(r.Body)) {
		return NewInvalidObject("mismatch Body length and Info.Size").Wrap(nil)
	}
	if r.Offset != 0 {
		return NewInvalidObject("Info.Offset must be zero when saving").Wrap(nil)
	}
	if err := r.checkHash(); err != nil {
		return err
	}

	jsInfo, err := json.Marshal(r.Info)
	if err != nil {
		return err
	}
	if uint64(len(jsInfo)) > r.MaxInfoSize {
		return NewMetadataTooLargeError().Wrap(nil)
	}

	return WithMustWriter(w, func(w io.Writer) error {
		binary.Write(w, binary.BigEndian, uint8(r.Version()))
		binary.Write(w, binary.BigEndian, uint64(len(jsInfo)))
		w.Write(jsInfo)
		binary.Write(w, binary.BigEndian, uint64(len(r.Body)))
		w.Write(r.Body)
		return nil
	})
}
func (r *ObjectV1) Version() uint8 {
	return 1
}
func (r *ObjectV1) checkHash() error {
	hash, err := r.hash()
	if err != nil {
		return err
	}
	if bytes.Compare(r.Info.Hash, hash) != 0 {
		return NewInvalidObject("hash value does not match").Wrap(nil)
	}
	return nil
}
func (r *ObjectV1) hash() ([]byte, error) {
	switch r.Info.HashAlgorithm {
	case "SHA1":
		hash := sha1.Sum(r.Body)
		return hash[:], nil
	default:
		msg := fmt.Sprintf("not supported hash type: %s", r.Info.HashAlgorithm)
		return nil, NewInvalidObject(msg).Wrap(nil)
	}
}
