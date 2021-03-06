// XDR - eXternal Data Representation
package eltonfs_rpc

import (
	"bytes"
	"encoding/binary"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"golang.org/x/xerrors"
	"math"
	"reflect"
	"sort"
	"strconv"
	"time"
	"unsafe"
)

const XDRTag = "xdr"
const XDRStructIDField = "XXX_XDR_ID"
const XDRStructIDTag = "xdrid"
const MaxAllowedPacketSize = 128 * (1 << 20) // 128 MiB

type XDREncoder interface {
	Uint8(n uint8)
	Bool(b bool)
	Uint64(n uint64)
	String(s string)
	Bytes(b []byte)
	Timestamp(t time.Time)
	Slice(s interface{})
	Map(m interface{})
	Struct(s interface{})
	RawPacket(nsid NSID, flags PacketFlag, data interface{})
	Auto(v interface{})
}

type XDRDecoder interface {
	Uint8() uint8
	Bool() bool
	Uint64() uint64
	String() string
	Bytes() []byte
	Timestamp() time.Time
	Slice(emptySlice interface{}) (slice interface{})
	Map(emptyMapping interface{}) (mapping interface{})
	Struct(emptyStruct interface{}) (aStruct interface{})
	RawPacket() *rawPacket
}

func NewXDREncoder(writer utils.MustWriter) XDREncoder {
	return &binEncoder{writer}
}
func NewXDRDecoder(reader utils.MustReader) XDRDecoder {
	return &binDecoder{reader}
}

type binEncoder struct {
	w utils.MustWriter
}
type binDecoder struct {
	r utils.MustReader
}

func (e *binEncoder) Uint8(n uint8) {
	e.w.MustWriteAll([]byte{n})
}
func (e *binEncoder) Bool(b bool) {
	if b {
		e.Uint8(1)
	} else {
		e.Uint8(0)
	}
}
func (e *binEncoder) Uint64(n uint64) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	e.w.MustWriteAll(b)
}
func (e *binEncoder) String(s string) {
	e.Bytes([]byte(s))
}
func (e *binEncoder) Bytes(b []byte) {
	length := uint64(len(b))
	e.Uint64(length)
	e.w.MustWriteAll(b)
}
func (e *binEncoder) Timestamp(t time.Time) {
	e.Uint64(uint64(t.Unix()))
	e.Uint64(uint64(t.Nanosecond()))
}
func (e *binEncoder) Slice(s interface{}) {
	kind := reflect.TypeOf(s).Kind()
	if kind != reflect.Slice {
		err := xerrors.Errorf("not a slice: %s", kind)
		panic(err)
	}

	v := reflect.ValueOf(s)
	length := uint64(v.Len())
	e.Uint64(length)
	for i := 0; i < int(length); i++ {
		e.Auto(v.Index(i).Interface())
	}
}
func (e *binEncoder) Map(m interface{}) {
	kind := reflect.TypeOf(m).Kind()
	if kind != reflect.Map {
		err := xerrors.Errorf("not a map: %s", kind)
		panic(err)
	}

	v := reflect.ValueOf(m)
	length := uint64(v.Len())
	e.Uint64(length)
	iter := v.MapRange()
	for iter.Next() {
		e.Auto(iter.Key().Interface())
		e.Auto(iter.Value().Interface())
	}
}
func (e *binEncoder) Struct(s interface{}) {
	t := reflect.TypeOf(s)
	kind := t.Kind()
	if kind == reflect.Ptr {
		// Dereference pointer.
		v := reflect.ValueOf(s).Elem().Interface()
		e.Struct(v)
		return
	}
	if kind != reflect.Struct {
		err := xerrors.Errorf("not a struct: %s", kind)
		panic(err)
	}

	v := reflect.ValueOf(s)
	fieldIDs := []uint8{}
	fields := map[uint8]reflect.Value{}
	for i := 0; i < t.NumField(); i++ {
		fieldID := parseXDRTag(t.Field(i))
		if fieldID == 0 {
			// Not specified or ignored field.
			continue
		}

		if _, ok := fields[fieldID]; ok {
			err := xerrors.Errorf("duplicated fieldID: %d", fieldID)
			panic(err)
		}
		if !v.Field(i).CanInterface() {
			// Skip unexported field
			continue
		}

		fieldIDs = append(fieldIDs, fieldID)
		fields[fieldID] = v.Field(i)
	}

	if len(fieldIDs) != len(fields) {
		panic("bug")
	}
	if math.MaxUint8 < len(fieldIDs) {
		err := xerrors.Errorf("too many fields: %d", len(fieldIDs))
		panic(err)
	}

	sort.Slice(fieldIDs, func(i, j int) bool {
		return fieldIDs[i] < fieldIDs[j]
	})
	length := uint8(len(fieldIDs))
	e.Uint8(length)
	for _, fieldID := range fieldIDs {
		e.Uint8(fieldID)
		e.Auto(fields[fieldID].Interface())
	}
}
func (e *binEncoder) RawPacket(nsid NSID, flags PacketFlag, data interface{}) {
	sid := parseXDRStructIDTag(reflect.TypeOf(data))

	buf := &bytes.Buffer{}
	enc := NewXDREncoder(utils.WrapMustWriter(buf))
	enc.Struct(data)
	size := uint64(buf.Len())

	e.Uint64(size)
	e.Uint64(uint64(nsid))
	e.Uint8(uint8(flags))
	e.Uint64(sid)
	e.w.MustWriteAll(buf.Bytes())
}
func (e *binEncoder) Auto(v interface{}) {
	switch vv := v.(type) {
	// Fast paths.
	case uint8:
		e.Uint8(vv)
	case bool:
		e.Bool(vv)
	case uint64:
		e.Uint64(vv)
	case string:
		e.String(vv)
	case []byte:
		e.Bytes(vv)
	case time.Time:
		e.Timestamp(vv)

	default: // Slow path.
		kind := reflect.TypeOf(v).Kind()
		switch kind {
		case reflect.Uint8:
			u := uint8(reflect.ValueOf(v).Uint())
			e.Uint8(u)
		case reflect.Bool:
			b := reflect.ValueOf(v).Bool()
			e.Bool(b)
		case reflect.Uint64:
			u := reflect.ValueOf(v).Uint()
			e.Uint64(u)
		case reflect.String:
			s := reflect.ValueOf(v).String()
			e.String(s)
		case reflect.Slice:
			if reflect.TypeOf(v).Elem().Kind() == reflect.Uint8 {
				vv := reflect.ValueOf(v)
				ptr := &reflect.SliceHeader{
					Data: vv.Pointer(),
					Len:  vv.Len(),
					Cap:  vv.Cap(),
				}
				b := *(*[]byte)(unsafe.Pointer(ptr))
				e.Bytes(b)
			} else {
				e.Slice(v)
			}
		case reflect.Map:
			e.Map(v)
		case reflect.Struct:
			e.Struct(v)
		case reflect.Ptr:
			// Dereference pointer.
			v = reflect.ValueOf(v).Elem().Interface()
			e.Auto(v)
		default:
			err := xerrors.Errorf("unsupported type: %s", kind)
			panic(err)
		}
	}
}

func (d *binDecoder) Uint8() uint8 {
	b := make([]byte, 1)
	d.r.MustReadAll(b)
	return b[0]
}
func (d *binDecoder) Bool() bool {
	if d.Uint8() == 0 {
		return false
	}
	return true
}
func (d *binDecoder) Uint64() uint64 {
	b := make([]byte, 8)
	d.r.MustReadAll(b)
	return binary.BigEndian.Uint64(b)
}
func (d *binDecoder) String() string {
	return string(d.Bytes())
}
func (d *binDecoder) Bytes() []byte {
	length := d.Uint64()
	b := make([]byte, length)
	d.r.MustReadAll(b)
	return b
}
func (d *binDecoder) Timestamp() time.Time {
	sec := int64(d.Uint64())
	nsec := int64(d.Uint64())
	return time.Unix(sec, nsec).UTC()
}
func (d *binDecoder) Slice(emptySlice interface{}) interface{} {
	t := reflect.TypeOf(emptySlice)
	return d.slice(t)
}
func (d *binDecoder) slice(t reflect.Type) interface{} {
	kind := t.Kind()
	if kind != reflect.Slice {
		err := xerrors.Errorf("not a slice: %s", kind)
		panic(err)
	}

	elemType := t.Elem()
	length := d.Uint64()
	slice := reflect.MakeSlice(t, int(length), int(length))
	for i := 0; i < int(length); i++ {
		value := d.auto(elemType)
		slice.Index(i).Set(value)
	}
	return slice.Interface()
}
func (d *binDecoder) Map(emptyMapping interface{}) interface{} {
	t := reflect.TypeOf(emptyMapping)
	return d.map_(t)
}
func (d *binDecoder) map_(t reflect.Type) interface{} {
	kind := t.Kind()
	if kind != reflect.Map {
		err := xerrors.Errorf("not a map: %s", kind)
		panic(err)
	}

	keyType := t.Key()
	valueType := t.Elem()
	length := d.Uint64()
	mapping := reflect.MakeMapWithSize(t, int(length))
	for i := 0; i < int(length); i++ {
		key := d.auto(keyType)
		value := d.auto(valueType)
		mapping.SetMapIndex(key, value)
	}
	return mapping.Interface()
}
func (d *binDecoder) Struct(emptyStruct interface{}) interface{} {
	t := reflect.TypeOf(emptyStruct)
	return d.struct_(t)
}
func (d *binDecoder) struct_(t reflect.Type) interface{} {
	var isPointer bool
	if t.Kind() == reflect.Ptr {
		// Dereference pointer.
		t = t.Elem()
		isPointer = true
	}

	// Key: FieldID
	// Value: Field Index of the struct.
	fieldID2Index := map[uint8]int{}
	for i := 0; i < t.NumField(); i++ {
		tag := parseXDRTag(t.Field(i))
		if tag == 0 {
			// Not specified or ignored field.
			continue
		}

		if _, ok := fieldID2Index[tag]; ok {
			err := xerrors.Errorf("duplicate FieldID: %d", tag)
			panic(err)
		}
		fieldID2Index[tag] = i
	}
	if math.MaxUint8 < len(fieldID2Index) {
		err := xerrors.Errorf("too many fields: %d", len(fieldID2Index))
		panic(err)
	}

	// reflect.New() returns pointer to the struct.
	vp := reflect.New(t)
	v := vp.Elem()

	length := d.Uint8()
	if length != uint8(len(fieldID2Index)) {
		err := xerrors.Errorf("mismatch number of fields: local=%d, remote=%d", length, len(fieldID2Index))
		panic(err)
	}

	for i := 0; i < int(length); i++ {
		fieldID := d.Uint8()
		if fieldID == 0 {
			err := xerrors.Errorf("invalid fieldID: %d", fieldID)
			panic(err)
		}

		idx, ok := fieldID2Index[fieldID]
		if !ok {
			err := xerrors.Errorf("not found FieldID: %d", fieldID)
			panic(err)
		}
		value := d.auto(t.Field(idx).Type)
		if !v.Field(idx).CanSet() {
			err := xerrors.Errorf("failed to set the value: FieldID=%d", fieldID)
			panic(err)
		}

		if t.Field(idx).Type != value.Type() {
			// t.Field(idx) and value are same kind.  But those are different types.
			// Must cast to specified type before setting it.
			value = value.Convert(t.Field(idx).Type)
		}
		v.Field(idx).Set(value)
	}

	if isPointer {
		return vp.Interface()
	}
	return v.Interface()
}
func (d *binDecoder) RawPacket() *rawPacket {
	p := &rawPacket{
		size:  d.Uint64(),
		nsid:  NSID(d.Uint64()),
		flags: PacketFlag(d.Uint8()),
		sid:   StructID(d.Uint64()),
		data:  nil,
	}
	if p.size > MaxAllowedPacketSize {
		err := xerrors.Errorf("packet too large: limit=%d receivedPacket=%d", MaxAllowedPacketSize, p.size)
		panic(err)
	}
	if p.flags&3 == 3 {
		err := xerrors.Errorf("create/close session flags are exclusive. but both flags are specified")
		panic(err)
	}
	if p.sid > MaxStructID {
		err := xerrors.Errorf("invalid struct id: %d", p.sid)
		panic(err)
	}

	p.data = make([]byte, p.size)
	d.r.MustReadAll(p.data)
	return p
}
func (d *binDecoder) auto(t reflect.Type) reflect.Value {
	var v interface{}

	kind := t.Kind()
	switch kind {
	case reflect.Uint8:
		v = d.Uint8()
	case reflect.Bool:
		v = d.Bool()
	case reflect.Uint64:
		v = d.Uint64()
	case reflect.String:
		v = d.String()
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			v = d.Bytes()
		} else {
			v = d.slice(t)
		}
	case reflect.Map:
		v = d.map_(t)
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			v = d.Timestamp()
		} else {
			v = d.struct_(t)
		}
	case reflect.Ptr:
		switch t.Elem().Kind() {
		case reflect.Struct:
			v = d.struct_(t)
		default:
			err := xerrors.Errorf("unsupported type: pointer to %s", t.Elem().Kind())
			panic(err)
		}
	default:
		err := xerrors.Errorf("unsupported type: %s", kind)
		panic(err)
	}

	return reflect.ValueOf(v)
}

// parseXDRTag parses "xdr" tag and returns a FieldID.
// If it is ignored field, returns zero.
func parseXDRTag(field reflect.StructField) uint8 {
	tag := field.Tag.Get(XDRTag)
	if tag == "" || tag == "ignore" {
		return 0
	}
	n, err := strconv.ParseUint(tag, 10, 8)
	if err != nil {
		err = xerrors.Errorf("parse FieldID: %w", err)
		panic(err)
	}
	if n == 0 {
		err = xerrors.Errorf("parse FieldID: out of range: %d", n)
		panic(err)
	}
	return uint8(n)
}

// parseXDRStructIDTag parses "xdrid" tag and return a StructID.
// It will panics in the following situations:
// * If p is not struct type.
// * If XXX_XDR_ID field is not found.
// * If xdrid tag on XXX_XDR_ID field is not found.
func parseXDRStructIDTag(p reflect.Type) uint64 {
	if p.Kind() == reflect.Ptr {
		// Dereference pointer.
		p = p.Elem()
	}

	field, ok := p.FieldByName(XDRStructIDField)
	if !ok {
		err := xerrors.Errorf("not found %s field", XDRStructIDField)
		panic(err)
	}

	tag, ok := field.Tag.Lookup(XDRStructIDTag)
	if !ok {
		err := xerrors.Errorf("not found %s tag in %s", XDRStructIDTag, XDRStructIDField)
		panic(err)
	}
	n, err := strconv.ParseUint(tag, 10, 64)
	if err != nil {
		err := xerrors.Errorf("not unsigned integer: %w", err)
		panic(err)
	}
	if n == 0 {
		err := xerrors.Errorf("parse XDRStructID: out of range: %d", n)
		panic(err)
	}
	return n
}
