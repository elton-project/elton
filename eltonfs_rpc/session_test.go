package eltonfs_rpc

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"io"
	"io/ioutil"
	"net"
	"sync"
	"testing"
	"time"
)

// newClientSession returns a session that it has been set up.
func newClientSession() (server *fakeConn, s ClientSession, closer func()) {
	var client *fakeConn
	client, server = newFakeConn()
	closer = func() {
		client.Close()
		server.Close()
	}

	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(2)
	go func() {
		defer wg.Done()
		// Send response (Setup2) to client.
		enc := NewXDREncoder(utils.WrapMustWriter(server))
		enc.Struct(&Setup2{
			Error:      0,
			Reason:     "",
			ServerName: "eltonfs",
		})
	}()
	go func() {
		defer wg.Done()
		// Read request (Setup1) from client.
		dec := NewXDRDecoder(utils.WrapMustReader(server))
		dec.Struct(&Setup1{})
	}()

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
		defer client.Close()
		defer server.Close()

		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(2)
		go func() {
			defer wg.Done()
			// Send data to client.
			enc := NewXDREncoder(utils.WrapMustWriter(server))
			enc.Struct(&Setup2{
				Error:      0,
				Reason:     "",
				ServerName: "eltonfs",
			})
		}()
		go func() {
			defer wg.Done()
			// Prepare expected setup request (Setup1)
			expected := func() []byte {
				req := &bytes.Buffer{}
				NewXDREncoder(utils.WrapMustWriter(req)).Struct(&Setup1{
					ClientName: "eltonfs-helper",
				})
				return req.Bytes()
			}()
			// Read data from client and check setup request (Setup1).
			assert.Equal(t,
				expected,
				server.MustReadAll(),
			)
		}()

		// Run the setup.
		s := NewClientSession(client)
		defer s.Close()
		err := s.Setup()
		assert.NoError(t, err)
	})
	t.Run("invalid response", func(t *testing.T) {
		// Prepare invalid server response (Setup2).
		client, server := newFakeConn()

		var wg sync.WaitGroup
		// Should close connections before call wg.Wait() to prevent deadlock.
		defer wg.Wait()
		defer client.Close()
		defer server.Close()
		wg.Add(2)
		go func() {
			defer wg.Done()
			// Send data to client.
			server.Write([]byte("invalid response"))
		}()
		go func() {
			defer wg.Done()
			// Read data from client.
			dec := NewXDRDecoder(utils.WrapMustReader(server))
			dec.Struct(&Setup1{})
		}()

		s := NewClientSession(client)
		defer s.Close()
		// Run the setup.
		err := s.Setup()
		assert.Error(t, err)
	})
}
func TestClientS_New(t *testing.T) {
	t.Run("send and recv", func(t *testing.T) {
		server, s, closer := newClientSession()
		defer closer()
		defer s.Close()

		// s.New() should not send any packet.
		ns, err := s.New()
		assert.NoError(t, err)
		assert.NotNil(t, ns)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Prepare expected packet
			dec := NewXDRDecoder(utils.WrapMustReader(server))
			assert.Equal(t, dec.Uint64(), uint64(1))               // Number of data size in bytes.
			assert.Equal(t, dec.Uint64(), uint64(1<<63|1))         // SessionID
			assert.Equal(t, dec.Uint8(), uint8(CreateSessionFlag)) // Flags
			assert.Equal(t, dec.Uint64(), uint64(3))               // StructID (ping)
			assert.Equal(t, dec.Uint8(), uint8(0))                 // Number of fields in struct.
		}()
		err = ns.Send(&Ping{})
		assert.NoError(t, err)
		wg.Wait()

		wg.Add(1)
		go func() {
			defer wg.Done()
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
		wg.Wait()
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
