package utils

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestMustW_MustWrite(t *testing.T) {
	buf := &bytes.Buffer{}
	w := WrapMustWriter(buf)

	assert.NotPanics(t, func() {
		w.MustWrite([]byte("foo "))
		w.MustWrite([]byte("bar"))
	})
	assert.Equal(t, "foo bar", buf.String())
}
func TestMustR_MustRead(t *testing.T) {
	buf := bytes.NewBufferString("foo bar")
	r := WrapMustReader(buf)

	assert.NotPanics(t, func() {
		b := [4]byte{}
		n := r.MustRead(b[:])
		assert.Equal(t, 4, n)
		assert.Equal(t, []byte("foo "), b[:n])
	})
	assert.NotPanics(t, func() {
		b := [10]byte{}
		n := r.MustRead(b[:])
		assert.Equal(t, 3, n)
		assert.Equal(t, []byte("bar"), b[:n])
	})
	assert.PanicsWithValue(t, io.EOF, func() {
		b := [10]byte{}
		r.MustRead(b[:])
	})
}
func TestMustR_MustReadAll(t *testing.T) {
	t.Run("partial_read", func(t *testing.T) {
		buf := bytes.NewBufferString("foo bar")
		r := WrapMustReader(&partialReader{R: buf})

		assert.NotPanics(t, func() {
			b := [4]byte{}
			r.MustReadAll(b[:])
			assert.Equal(t, []byte("foo "), b[:])
		})
	})
	t.Run("reached_to_EOF_and_buffer_is_full", func(t *testing.T) {
		buf := bytes.NewBufferString("foo")
		r := WrapMustReader(&partialReader{R: buf})

		assert.NotPanics(t, func() {
			b := [3]byte{}
			r.MustReadAll(b[:])
			assert.Equal(t, []byte("foo"), b[:])
		})
	})
	t.Run("reached_to_EOF_when_reading", func(t *testing.T) {
		buf := bytes.NewBufferString("foo")
		r := WrapMustReader(&partialReader{R: buf})

		assert.PanicsWithValue(t, io.EOF, func() {
			b := [10]byte{}
			r.MustReadAll(b[:])
		})
	})
}

type partialReader struct {
	R io.Reader
}

func (r *partialReader) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return r.R.Read(b)
	}
	return r.R.Read(b[:1])
}
