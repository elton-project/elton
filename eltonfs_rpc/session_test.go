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
	t.Run("invalid response", func(t *testing.T) {
		// Prepare invalid server response (Setup2).
		conn := &fakeConn{}
		conn.ReadBuf.Write([]byte("invalid response"))

		// Run the setup.
		s := NewClientSession(conn)
		err := s.Setup()
		assert.Error(t, err)
	})
}
func TestClientS_New(t *testing.T) {
	t.Run("send and recv", func(t *testing.T) {
		conn := &fakeConn{}

		s := NewClientSession(conn)
		s.(*clientS).setupOK = true

		ns, err := s.New()
		assert.NoError(t, err)
		assert.NotNil(t, ns)
		assert.Equal(t, []byte(nil), conn.WriteBuf.Bytes())

		err = ns.Send(&Ping{})
		assert.NoError(t, err)
		expected := func() []byte {
			// Prepare expected packet
			buf := &bytes.Buffer{}
			enc := NewXDREncoder(utils.WrapMustWriter(buf))
			enc.Uint64(1)                       // Number of data size in bytes.
			enc.Uint64(1<<63 | 1)               // SessionID
			enc.Uint8(uint8(CreateSessionFlag)) // Flags
			enc.Uint64(3)                       // StructID (ping)
			enc.Uint8(0)                        // Number of fields in struct.
			return buf.Bytes()
		}()
		assert.Equal(t, expected, conn.WriteBuf.Bytes())

		go func() {
			enc := NewXDREncoder(utils.WrapMustWriter(&conn.ReadBuf))
			enc.Uint64(1)         // Number of data size in bytes.
			enc.Uint64(1<<63 | 1) // SessionID
			enc.Uint8(0)          // Flags
			enc.Uint64(3)         // StructID ()
			enc.Uint8(0)          // Number of fields in struct.
		}()
		ping, err := ns.Recv(&Ping{})
		assert.NoError(t, err)
		assert.Equal(t, &Ping{}, ping)
	})
	t.Run("try to create nested session before setup", func(t *testing.T) {
		conn := &fakeConn{}
		s := NewClientSession(conn)
		ns, err := s.New()
		assert.EqualError(t, err, "setup is not complete")
		assert.Nil(t, ns)
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
