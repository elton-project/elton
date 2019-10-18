package eltonfs_rpc

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"golang.org/x/xerrors"
	"net"
	"testing"
	"time"
)

func TestClientS_Setup(t *testing.T) {
	t.Run("no_error", func(t *testing.T) {
		// Prepare server response (Setup2).
		conn := &fakeConn{}
		enc := NewXDREncoder(utils.WrapMustWriter(&conn.ReadBuf))
		enc.Struct(&Setup2{
			Error:      0,
			Reason:     "",
			ServerName: "eltonfs",
		})
		// Prepare expected setup request (Setup1)
		req := &bytes.Buffer{}
		NewXDREncoder(utils.WrapMustWriter(req)).Struct(&Setup1{
			ClientName: "eltonfs-helper",
		})

		// Run the setup.
		s := NewClientSession(conn)
		err := s.Setup()
		assert.NoError(t, err)

		// Check setup request (Setup1).
		assert.NoError(t, err)
		assert.Equal(t,
			req.Bytes(),
			conn.WriteBuf.Bytes(),
		)
	})
}

type fakeConn struct {
	ReadBuf     bytes.Buffer
	WriteBuf    bytes.Buffer
	writeClosed bool
}

func (f *fakeConn) Read(b []byte) (n int, err error) {
	return f.ReadBuf.Read(b)
}

func (f *fakeConn) Write(b []byte) (n int, err error) {
	if f.writeClosed {
		return 0, xerrors.New("write on closed conn")
	}
	return f.WriteBuf.Write(b)
}

func (f *fakeConn) Close() error {
	f.writeClosed = true
	return nil
}

func (f *fakeConn) LocalAddr() net.Addr {
	panic("implement me")
}

func (f *fakeConn) RemoteAddr() net.Addr {
	panic("implement me")
}

func (f *fakeConn) SetDeadline(t time.Time) error {
	panic("implement me")
}

func (f *fakeConn) SetReadDeadline(t time.Time) error {
	panic("implement me")
}

func (f *fakeConn) SetWriteDeadline(t time.Time) error {
	panic("implement me")
}
