package utils

import (
	"golang.org/x/xerrors"
	"io"
)

type MustWriter interface {
	MustWrite([]byte) (n int)
	MustWriteAll([]byte)
}
type MustReader interface {
	MustRead([]byte) (n int)
	MustReadAll([]byte)
}

func WrapMustWriter(w io.Writer) MustWriter {
	return &mustW{Writer: w}
}
func WrapMustReader(r io.Reader) MustReader {
	return &mustR{Reader: r}
}

type mustW struct {
	io.Writer
}

func (w *mustW) MustWrite(b []byte) (n int) {
	n, err := w.Write(b)
	if err != nil {
		err := xerrors.Errorf("must write: %w", err)
		panic(err)
	}
	return n
}
func (w *mustW) MustWriteAll(b []byte) {
	for len(b) > 0 {
		n := w.MustWrite(b)
		b = b[n:]
	}
}

type mustR struct {
	io.Reader
}

func (r *mustR) MustRead(b []byte) (n int) {
	var err error
	n, err = r.Read(b)
	if err != nil {
		if err == io.EOF {
			if n == 0 {
				// Reached to EOF.  We can not read any more.
				panic(err)
			}
			// Reached to EOF.  But we succeed to read a little.
			// Ignore EOF error.
			return
		}
		// Unexpected error
		err := xerrors.Errorf("must read: %w", err)
		panic(err)
	}
	return
}
func (r *mustR) MustReadAll(b []byte) {
	for len(b) > 0 {
		n := r.MustRead(b)
		b = b[n:]
	}
}
