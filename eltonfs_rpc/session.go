package eltonfs_rpc

import (
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"golang.org/x/xerrors"
	"net"
)

type ClientSession interface {
	// Setup initialize the connection.
	Setup() error
	// New creates new nested session.
	New() (ClientNS, error)
	// Close closes this session.
	// If nested connections is not closed, it will return an error.
	Close() error
}

// ClientNS represents a nested session.
type ClientNS interface {
	// Send sends a struct to the opponent.
	Send(v interface{}) error
	// SendErr sends an error to the opponent.
	SendErr(err *SessionError) error
	// Recv receives a struct of specified type from the opponent.
	Recv(empty interface{}) (interface{}, error)
	// Close notifies error to opponent and close a nested session.
	CloseWithError(err SessionError) error
	// Close closes this nested session.
	Close() error
	// IsSendable returns a boolean value whether data can be sent to the opponent.
	IsSendable() bool
	// IsReceivable returns a boolean value whether data can be received from the opponent.
	IsReceivable() bool
}

func NewClientSession(conn net.Conn) ClientSession {
	w := utils.WrapMustWriter(conn)
	r := utils.WrapMustReader(conn)
	return &clientS{
		Conn: conn,
		Enc:  NewXDREncoder(w),
		Dec:  NewXDRDecoder(r),
	}
}

type clientS struct {
	Conn net.Conn
	Enc  XDREncoder
	Dec  XDRDecoder

	setupOK bool
	// Nested Sessions
	// Key: nested session IDs
	// Value: *clientNS
	nss      map[uint64]*clientNS
	lastNSID uint64
}

func (s *clientS) Setup() error {
	return HandlePanic(func() error {
		s.Enc.Struct(&Setup1{
			ClientName: "eltonfs-helper",
		})
		s2 := s.Dec.Struct(&Setup2{}).(*Setup2)
		if s2.Error != 0 {
			return xerrors.Errorf(
				"%d - %s: %s",
				s2.Error,
				ErrorID(s2.Error).String(),
				s2.Reason,
			)
		}
		s.setupOK = true
		return nil
	})
}
func (s *clientS) New() (ClientNS, error) {
	if !s.setupOK {
		return nil, xerrors.New("setup is not complete")
	}
	// Initialize　s.nss
	if s.nss == nil {
		s.nss = map[uint64]*clientNS{}
	}

	// Get next NSID.
	var nextNSID uint64
	for {
		nextNSID = (s.lastNSID + 1) | 1<<63
		if _, ok := s.nss[nextNSID]; ok {
			continue
		}
		break
	}
	s.lastNSID = nextNSID

	// Create new clientNS instance.
	ns := &clientNS{
		S:    s,
		NSID: nextNSID,
	}
	s.nss[nextNSID] = ns
	return ns, nil
}
func (s *clientS) Close() error { panic("todo") }

type clientNS struct {
	S    *clientS
	NSID uint64
}

func (ns *clientNS) Send(v interface{}) error                    { panic("todo") }
func (ns *clientNS) SendErr(err *SessionError) error             { panic("todo") }
func (ns *clientNS) Recv(empty interface{}) (interface{}, error) { panic("todo") }
func (ns *clientNS) CloseWithError(err SessionError) error       { panic("todo") }
func (ns *clientNS) Close() error                                { panic("todo") }
func (ns *clientNS) IsSendable() bool                            { panic("todo") }
func (ns *clientNS) IsReceivable() bool                          { panic("todo") }
