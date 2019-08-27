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

type mustReadSeeker struct {
	io.ReadSeeker
}

func (rs mustReadSeeker) Read(p []byte) (n int, err error) {
	n, err = rs.ReadSeeker.Read(p)
	rs.ReadSeeker.Seek()
	if err != nil {
		panic(err)
	}
	return
}
func (rs mustReadSeeker) Seek(offset int64, whence int) (pos int64, err error) {
	pos, err = rs.ReadSeeker.Seek(offset, whence)
	if err != nil {
		panic(err)
	}
	return

}
func WithMustReadSeeker(rs io.ReadSeeker, fn func(rs io.ReadSeeker) error) error {
	var panicErr, fnErr error
	panicErr = util.PanicHandler(func() {
		must := mustReadSeeker{rs}
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
