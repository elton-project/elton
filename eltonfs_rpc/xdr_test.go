package eltonfs_rpc

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"testing"
	"time"
)

func newEnc() (*bytes.Buffer, XDREncoder) {
	buf := &bytes.Buffer{}
	return buf, NewXDREncoder(utils.WrapMustWriter(buf))
}
func newDec(data ...interface{}) XDRDecoder {
	buf, enc := newEnc()
	for i := range data {
		enc.Auto(data[i])
	}
	return NewXDRDecoder(utils.WrapMustReader(buf))
}
func TestBinEncoder_Uint8(t *testing.T) {
	buf, enc := newEnc()
	enc.Uint8(1)
	enc.Uint8(2)
	enc.Uint8(3)

	assert.Equal(t, []byte{1, 2, 3}, buf.Bytes())
}
func TestBinEncoder_Bool(t *testing.T) {
	buf, enc := newEnc()
	enc.Bool(true)
	enc.Bool(false)

	assert.Equal(t, []byte{1, 0}, buf.Bytes())
}
func TestBinEncoder_Uint64(t *testing.T) {
	buf, enc := newEnc()
	enc.Uint64(0x0123456789abcdef)

	assert.Equal(t, []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}, buf.Bytes())
}
func TestBinEncoder_String(t *testing.T) {
	buf, enc := newEnc()
	enc.String("test")

	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 4, 't', 'e', 's', 't'}, buf.Bytes())
}
func TestBinEncoder_Bytes(t *testing.T) {
	buf, enc := newEnc()
	enc.Bytes([]byte("test"))

	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 4, 't', 'e', 's', 't'}, buf.Bytes())
}
func TestBinEncoder_Timestamp(t *testing.T) {
	timestamp := time.Unix(0x123456, 0x987654)
	buf, enc := newEnc()
	enc.Timestamp(timestamp)

	assert.Equal(t, []byte{
		// UNIX time (sec)
		0, 0, 0, 0, 0, 0x12, 0x34, 0x56,
		// Unix time (nano sec)
		0, 0, 0, 0, 0, 0x98, 0x76, 0x54,
	}, buf.Bytes())
}
func TestBinEncoder_Slice(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		var slice []string
		buf, enc := newEnc()
		enc.Slice(slice)

		assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0}, buf.Bytes())
	})
	t.Run("empty slice", func(t *testing.T) {
		buf, enc := newEnc()
		enc.Slice([]string{})

		assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0}, buf.Bytes())
	})
	t.Run("[]byte", func(t *testing.T) {
		buf, enc := newEnc()
		enc.Slice([]byte("test"))

		assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 4, 't', 'e', 's', 't'}, buf.Bytes())
	})
	t.Run("[]string", func(t *testing.T) {
		buf, enc := newEnc()
		enc.Slice([]string{
			"test",
			"hello world",
		})

		assert.Equal(t,
			[]byte{
				0, 0, 0, 0, 0, 0, 0, 2,
				0, 0, 0, 0, 0, 0, 0, 4,
				't', 'e', 's', 't',
				0, 0, 0, 0, 0, 0, 0, 11,
				'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd',
			}, buf.Bytes())
	})
}
func TestBinEncoder_Map(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		var m map[int]string
		buf, enc := newEnc()
		enc.Map(m)

		assert.Nil(t, m)
		assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0}, buf.Bytes())
	})
	t.Run("empty map", func(t *testing.T) {
		m := map[int]string{}
		buf, enc := newEnc()
		enc.Map(m)

		assert.NotNil(t, m)
		assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0}, buf.Bytes())
	})
	t.Run("map[uint64]string", func(t *testing.T) {
		m := map[uint64]string{
			1: "one",
			2: "two",
		}
		buf, enc := newEnc()
		enc.Map(m)

		pattern1 := []byte{
			0, 0, 0, 0, 0, 0, 0, 2, // length
			0, 0, 0, 0, 0, 0, 0, 1, // key1
			0, 0, 0, 0, 0, 0, 0, 3,
			'o', 'n', 'e', // value1
			0, 0, 0, 0, 0, 0, 0, 2, // key2
			0, 0, 0, 0, 0, 0, 0, 3,
			't', 'w', 'o', // value2
		}
		pattern2 := []byte{
			0, 0, 0, 0, 0, 0, 0, 2,
			0, 0, 0, 0, 0, 0, 0, 2, // key2
			0, 0, 0, 0, 0, 0, 0, 3,
			't', 'w', 'o', // value2
			0, 0, 0, 0, 0, 0, 0, 1, // key1
			0, 0, 0, 0, 0, 0, 0, 3,
			'o', 'n', 'e', // value1
		}
		assert.Contains(t, [][]byte{
			pattern1,
			pattern2,
		}, buf.Bytes())
	})
}
func TestBinEncoder_Struct(t *testing.T) {
	t.Run("encode empty struct", func(t *testing.T) {
		s := struct{}{}
		buf, enc := newEnc()
		enc.Struct(s)

		assert.Equal(t, []byte{0}, buf.Bytes())
	})
	t.Run("no-xdr fields", func(t *testing.T) {
		s := struct {
			A uint64
			B string
		}{
			A: 1,
			B: "2",
		}
		buf, enc := newEnc()
		enc.Struct(s)

		assert.Equal(t, []byte{0}, buf.Bytes())
	})
	t.Run("invalid id", func(t *testing.T) {
		s1 := struct {
			XXX_XDR_ID struct{} `xdrid:"10"`
			A          uint64   `xdr:"0"`
		}{A: 1}
		s2 := struct {
			XXX_XDR_ID struct{} `xdrid:"11"`
			A          uint64   `xdr:"-1"`
		}{A: 1}

		_, enc := newEnc()
		assert.Panics(t, func() {
			enc.Struct(s1)
		})
		assert.Panics(t, func() {
			enc.Struct(s2)
		})
	})
	t.Run("private xdr fields", func(t *testing.T) {
		s := struct {
			a uint64
			b string `xdr:"1"`
		}{
			a: 1,
			b: "2",
		}
		buf, enc := newEnc()
		enc.Struct(s)

		assert.Equal(t, []byte{0}, buf.Bytes())
	})
	t.Run("multiple xdr fields", func(t *testing.T) {
		s := struct {
			A uint8  `xdr:"3"`
			B uint64 `xdr:"2"`
			C string `xdr:"1"`
		}{
			A: 3,
			B: 2,
			C: "1",
		}

		buf, enc := newEnc()
		enc.Struct(s)

		assert.Equal(t, []byte{
			3, // number of fields
			1, // field id
			0, 0, 0, 0, 0, 0, 0, 1,
			'1',                    // field value (string)
			2,                      // field id
			0, 0, 0, 0, 0, 0, 0, 2, // field value (uint64)
			3, // field id
			3, // field value (uint8)
		}, buf.Bytes())
	})
	t.Run("pointer to the struct", func(t *testing.T) {
		s := &struct {
			A string `xdr:"10"`
		}{
			A: "hoge",
		}
		buf, enc := newEnc()
		enc.Struct(s)

		assert.Equal(t, []byte{
			1,  // number of fields
			10, // field id
			0, 0, 0, 0, 0, 0, 0, 4,
			'h', 'o', 'g', 'e', // field value  (string)
		}, buf.Bytes())
	})
}
func TestBinEncoder_Auto(t *testing.T) {
	t.Run("basic types", func(t *testing.T) {
		buf, enc := newEnc()
		enc.Auto(uint8(100))
		enc.Auto(true)
		enc.Auto(false)
		enc.Auto(uint64(0x0123456789abcdef))
		enc.Auto("hello")
		enc.Auto([]byte("world"))

		assert.Equal(t, []byte{
			100,                                            // uint8
			1,                                              // true
			0,                                              // false
			0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, // uint64
			0, 0, 0, 0, 0, 0, 0, 5,
			'h', 'e', 'l', 'l', 'o', // string
			0, 0, 0, 0, 0, 0, 0, 5,
			'w', 'o', 'r', 'l', 'd', // []byte
		}, buf.Bytes())
	})
	t.Run("slice", func(t *testing.T) {
		buf, enc := newEnc()
		enc.Auto([]string{"foo"})

		assert.Equal(t, []byte{
			0, 0, 0, 0, 0, 0, 0, 1,
			0, 0, 0, 0, 0, 0, 0, 3,
			'f', 'o', 'o',
		}, buf.Bytes())
	})
	t.Run("map", func(t *testing.T) {
		buf, enc := newEnc()
		enc.Auto(map[uint8]string{
			1: "one",
		})

		assert.Equal(t, []byte{
			0, 0, 0, 0, 0, 0, 0, 1, // length of map
			1, // key (uint8)
			0, 0, 0, 0, 0, 0, 0, 3,
			'o', 'n', 'e', // value (string)
		}, buf.Bytes())
	})
	t.Run("struct", func(t *testing.T) {
		buf, enc := newEnc()
		enc.Auto(struct {
			A uint8 `xdr:"10"`
		}{
			A: 9,
		})

		assert.Equal(t, []byte{
			1,  // number of fields
			10, // field ID
			9,  // field value
		}, buf.Bytes())
	})
	t.Run("pointer of struct", func(t *testing.T) {
		buf, enc := newEnc()
		enc.Auto(&struct {
			A uint8 `xdr:"10"`
		}{
			A: 9,
		})

		assert.Equal(t, []byte{
			1,  // number of fields
			10, // field ID
			9,  // field value
		}, buf.Bytes())
	})
	t.Run("type-alias", func(t *testing.T) {
		buf, enc := newEnc()
		type Alias string
		enc.Auto(Alias("alias"))

		assert.Equal(t, []byte{
			0, 0, 0, 0, 0, 0, 0, 3, // string length
			'a', 'l', 'i', 'a', 's', // string body
		}, buf.Bytes())
	})
}
func TestBinDecoder_Uint8(t *testing.T) {
	n := uint8(10)
	dec := newDec(n)
	assert.Equal(t, n, dec.Uint8())
}
func TestBinDecoder_Bool(t *testing.T) {
	dec := newDec(true, false)
	assert.Equal(t, true, dec.Bool())
	assert.Equal(t, false, dec.Bool())
}
func TestBinDecoder_Uint64(t *testing.T) {
	n := uint64(0x0123456789abcdef)
	dec := newDec(n)
	assert.Equal(t, n, dec.Uint64())
}
func TestBinDecoder_String(t *testing.T) {
	s1 := "hello"
	s2 := "world"
	dec := newDec(s1, s2)
	assert.Equal(t, s1, dec.String())
	assert.Equal(t, s2, dec.String())
}
func TestBinDecoder_Bytes(t *testing.T) {
	b1 := []byte("hello")
	b2 := []byte("world")
	dec := newDec(b1, b2)
	assert.Equal(t, b1, dec.Bytes())
	assert.Equal(t, b2, dec.Bytes())
}
func TestBinDecoder_Timestamp(t *testing.T) {
	t1 := time.Unix(0x12345678, 0x987654).UTC()
	t2 := time.Unix(0x1234567800000000, 0x12345678).UTC()
	dec := newDec(t1, t2)
	assert.Equal(t, t1, dec.Timestamp())
	assert.Equal(t, t2, dec.Timestamp())
}
func TestBinDecoder_Slice(t *testing.T) {
	s := []string{"foo", "bar"}
	dec := newDec(s)
	assert.Equal(t, s, dec.Slice([]string{}))
}
func TestBinDecoder_Map(t *testing.T) {
	m := map[string]*struct {
		A uint64 `xdr:"2"`
	}{
		"one": {A: 1},
		"two": {A: 2},
	}
	dec := newDec(m)
	assert.Equal(t, m, dec.Map(map[string]*struct {
		A uint64 `xdr:"2"`
	}{}))
}
func TestBinDecoder_Struct(t *testing.T) {
	t.Run("empty struct", func(t *testing.T) {
		s := struct{}{}
		dec := newDec(s)
		assert.Equal(t, s, dec.Struct(struct{}{}))
	})
	t.Run("pointer to the struct", func(t *testing.T) {
		s := &struct {
			A string `xdr:"2"`
		}{"hello world"}
		dec := newDec(s)
		assert.Equal(t, s, dec.Struct(&struct {
			A string `xdr:"2"`
		}{}))
	})
	t.Run("private fields", func(t *testing.T) {
		s := &struct {
			A string `xdr:"2"`
			B string `xdr:"1"`
		}{"foo", "bar"}
		dec := newDec(s)
		assert.Panics(t, func() {
			dec.Struct(struct {
				A string `xdr:"2"`
				b string `xdr:"1"`
			}{})
		})
	})
	t.Run("missing xdr fields", func(t *testing.T) {
		s := &struct {
			A string `xdr:"1"`
			B string `xdr:"2"`
		}{"one", "two"}
		dec := newDec(s)
		assert.Panics(t, func() {
			dec.Struct(struct {
				B string `xdr:"2"`
			}{})
		})
	})
	t.Run("different xdr field names", func(t *testing.T) {
		s := &struct {
			A string `xdr:"1"`
			B string `xdr:"2"`
		}{"one", "two"}
		dec := newDec(s)
		assert.Equal(t, struct {
			X string `xdr:"1"`
			Y string `xdr:"2"`
		}{"one", "two"}, dec.Struct(struct {
			X string `xdr:"1"`
			Y string `xdr:"2"`
		}{}))
	})
}
