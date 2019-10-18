package eltonfs_rpc

import (
	"bytes"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"golang.org/x/xerrors"
	"net"
)

type PacketFlag uint8

const (
	_ PacketFlag = 1 << iota // ignore first value (0) by assigning to blank identifier
	CreateSessionFlag
	CloseSessionFlag
	ErrorSessionFlag
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
		W:    w,
		Enc:  NewXDREncoder(w),
		Dec:  NewXDRDecoder(r),
	}
}

type clientS struct {
	Conn net.Conn
	W    utils.MustWriter
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
func (s *clientS) sendPacket(nsid uint64, flags PacketFlag, data interface{}) error {
	buf := &bytes.Buffer{}
	enc := NewXDREncoder(utils.WrapMustWriter(buf))
	enc.Struct(data)
	size := uint64(buf.Len())
	sid := uint64(0) // StructID

	err := HandlePanic(func() error {
		s.Enc.Uint64(size)
		s.Enc.Uint64(nsid)
		s.Enc.Uint8(uint8(flags))
		s.Enc.Uint64(sid)
		s.W.MustWriteAll(buf.Bytes())
		return nil
	})
	if err != nil {
		return xerrors.Errorf("sendPacket: %w", err)
	}
	return nil
}

type clientNS struct {
	S    *clientS
	NSID uint64

	established bool
	closedC2S   bool
	closedS2C   bool
}

func (ns *clientNS) Send(v interface{}) error {
	if ns.closedS2C {
		return xerrors.Errorf("the nested session (NSID=%d) is already closed", ns.NSID)
	}

	var flags PacketFlag
	if !ns.established {
		flags |= CreateSessionFlag
	}
	return ns.S.sendPacket(ns.NSID, flags, v)
}
func (ns *clientNS) SendErr(err *SessionError) error {
	if ns.closedS2C {
		return xerrors.Errorf("the nested session (NSID=%d) is already closed", ns.NSID)
	}

	var flags PacketFlag
	if !ns.established {
		flags |= CreateSessionFlag
	}
	flags |= ErrorSessionFlag
	return ns.S.sendPacket(ns.NSID, flags, err)
}
func (ns *clientNS) Recv(empty interface{}) (interface{}, error) { panic("todo") }
func (ns *clientNS) CloseWithError(err SessionError) error {
	if !ns.established {
		return xerrors.Errorf("the nested session (NSID=%d) is not established", ns.NSID)
	}
	if ns.closedC2S {
		return xerrors.Errorf("the nested session (NSID=%d) is already closed", ns.NSID)
	}

	return ns.S.sendPacket(ns.NSID, CloseSessionFlag|ErrorSessionFlag, err)
}
func (ns *clientNS) Close() error {
	if !ns.established {
		return xerrors.Errorf("the nested session (NSID=%d) is not established", ns.NSID)
	}
	if ns.closedC2S {
		return xerrors.Errorf("the nested session (NSID=%d) is already closed", ns.NSID)
	}
	return ns.S.sendPacket(ns.NSID, CloseSessionFlag, nil)
}
func (ns *clientNS) IsSendable() bool {
	return ns.closedC2S
}
func (ns *clientNS) IsReceivable() bool {
	return ns.established && !ns.closedS2C
}
