package eltonfs_rpc

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"io"
	"io/ioutil"
	"net"
	"testing"
	"time"
)

// newClientSession returns a session that it has been set up.
func newClientSession() (server *fakeConn, s ClientSession) {
	var client *fakeConn
	client, server = newFakeConn()

	// Prepare server response (Setup2).
	enc := NewXDREncoder(utils.WrapMustWriter(server))
	enc.Struct(&Setup2{
		Error:      0,
		Reason:     "",
		ServerName: "eltonfs",
	})

	s = NewClientSession(client)
	err := s.Setup()
	if err != nil {
		panic(err)
	}
	return
}

func TestClientS_Setup(t *testing.T) {
	t.Run("no_error", func(t *testing.T) {
		// Prepare server response (Setup2).
		client, server := newFakeConn()
		enc := NewXDREncoder(utils.WrapMustWriter(server))
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
		s := NewClientSession(client)
		defer s.Close()
		err := s.Setup()
		assert.NoError(t, err)

		// Check setup request (Setup1).
		assert.NoError(t, err)
		assert.Equal(t,
			req.Bytes(),
			server.MustReadAll(),
		)
	})
	t.Run("invalid response", func(t *testing.T) {
		// Prepare invalid server response (Setup2).
		client, server := newFakeConn()
		utils.WrapMustWriter(server).MustWriteAll([]byte("invalid response"))

		s := NewClientSession(client)
		defer client.Close()
		// Run the setup.
		err := s.Setup()
		assert.Error(t, err)
	})
}
func TestClientS_New(t *testing.T) {
	t.Run("send and recv", func(t *testing.T) {
		server, s := newClientSession()
		defer s.Close()

		ns, err := s.New()
		assert.NoError(t, err)
		assert.NotNil(t, ns)
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
		assert.Equal(t, expected, server.MustReadAll())

		go func() {
			enc := NewXDREncoder(utils.WrapMustWriter(server))
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

func newFakeConn() (*fakeConn, *fakeConn) {
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	return &fakeConn{
			r: r1,
			w: w2,
		}, &fakeConn{
			r: r2,
			w: w1,
		}
}

type fakeConn struct {
	r *io.PipeReader
	w *io.PipeWriter
}

func (f *fakeConn) Read(b []byte) (n int, err error) {
	return f.r.Read(b)
}

func (f *fakeConn) Write(b []byte) (n int, err error) {
	return f.w.Write(b)
}

func (f *fakeConn) Close() error {
	return f.w.Close()
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
func (f *fakeConn) MustReadAll() []byte {
	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	return b
}
