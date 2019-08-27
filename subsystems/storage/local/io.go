package localStorage

import (
	"github.com/yuuki0xff/goapptrace/tracer/util"
	"github.com/yuuki0xff/pathlib"
	"io"
	"os"
)

type mustWriter struct {
	io.Writer
}

func (w mustWriter) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	if err != nil {
		panic(err)
	}
	return
}
func WithMustWriter(w io.Writer, fn func(w io.Writer) error) error {
	var panicErr, fnErr error
	panicErr = util.PanicHandler(func() {
		must := mustWriter{w}
		fnErr = fn(&must)
	})
	if panicErr != nil {
		return panicErr
	}
	return fnErr
}

type MustReadSeeker interface {
	io.ReadSeeker
	IgnoreEOF() MustReadSeeker
}

type mustReadSeeker struct {
	io.ReadSeeker
	ignoreEOF bool
}

func (rs *mustReadSeeker) Read(p []byte) (n int, err error) {
	n, err = rs.ReadSeeker.Read(p)
	rs.checkError(err)
	return
}
func (rs *mustReadSeeker) Seek(offset int64, whence int) (pos int64, err error) {
	pos, err = rs.ReadSeeker.Seek(offset, whence)
	rs.checkError(err)
	return
}
func (rs *mustReadSeeker) IgnoreEOF() MustReadSeeker {
	return &mustReadSeeker{
		ReadSeeker: rs.ReadSeeker,
		ignoreEOF:  true,
	}
}
func (rs *mustReadSeeker) checkError(err error) {
	if err != nil {
		if rs.ignoreEOF && err == io.EOF {
			// Ignore this error.
			return
		}
		panic(err)
	}
}
func WithMustReadSeeker(rs io.ReadSeeker, fn func(rs MustReadSeeker) error) error {
	var panicErr, fnErr error
	panicErr = util.PanicHandler(func() {
		must := mustReadSeeker{
			ReadSeeker: rs,
		}
		fnErr = fn(&must)
	})
	if panicErr != nil {
		return panicErr
	}
	return fnErr
}

func AtomicWrite(target, tmp pathlib.Path, fn func(writer io.Writer) error) error {
	f, err := tmp.OpenRW(os.O_CREATE|os.O_EXCL, 0777)
	if err != nil {
		return err
	}
	defer tmp.Unlink()
	defer f.Close()

	if err := fn(f); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := tmp.Rename(target); err != nil {
		return err
	}
	return nil
}
