package eltonfs_rpc

import (
	"bytes"
	"context"
	"fmt"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"golang.org/x/xerrors"
	"net"
	"reflect"
	"sync"
)

type PacketFlag uint8

const (
	_ PacketFlag = 1 << iota // ignore first value (0) by assigning to blank identifier
	CreateSessionFlag
	CloseSessionFlag
	ErrorSessionFlag
)

const (
	SendQueueSize = 64
	RecvQueueSize = 16
)

const (
	// Flag value for the NS started from client side.
	nsidClientFlag = 1 << 31
	// Maximum value of NSID.
	nsidMaxValue = (1 << 32) - 1
)

type ClientSession interface {
	// Setup initialize the connection.
	Setup() error
	// New creates new nested session.
	New() (ClientNS, error)
	Serve(ctx context.Context) error
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
		R:    r,
		W:    w,
		Enc:  NewXDREncoder(w),
		Dec:  NewXDRDecoder(r),
	}
}

type clientS struct {
	Conn net.Conn
	R    utils.MustReader
	W    utils.MustWriter
	Enc  XDREncoder
	Dec  XDRDecoder

	setupOK bool
	// closed is closed when clientS.Close() called.  The purpose is notify workers of server termination event.
	//
	// Usage:
	//	select {
	//	case <-other:
	//		// Do something.
	//	case <-s.closed:
	//		return
	//	}
	closed chan struct{}
	// Nested Sessions
	// Key: nested session IDs
	// Value: *clientNS
	nss      map[uint64]*clientNS
	lastNSID uint64
	nssLock  sync.RWMutex

	// Queue for packets waiting to be sent.
	// Elements are serialized packets.
	sendQ chan []byte
	// Queue for packets waiting to be received.
	// A lock must be acquired before access to the recvQ.
	recvQ     map[uint64]chan *rawPacket
	recvQLock sync.RWMutex
}

type rawPacket struct {
	size  uint64
	nsid  uint64
	flags PacketFlag
	sid   uint64
	data  []byte
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

		// Start workers.
		s.closed = make(chan struct{})
		s.sendQ = make(chan []byte, SendQueueSize)
		s.recvQ = map[uint64]chan *rawPacket{}
		s.nss = map[uint64]*clientNS{}
		go s.recvWorker()
		go s.sendWorker()
		return nil
	})
}
func (s *clientS) New() (ClientNS, error) {
	if !s.setupOK {
		return nil, xerrors.New("setup is not complete")
	}

	s.nssLock.Lock()
	defer s.nssLock.Unlock()

	// Get next NSID.
	nextNSID := s.lastNSID
	for {
		nextNSID = ((nextNSID + 1) | 1<<31) & (1<<32 - 1)
		if _, ok := s.nss[nextNSID]; ok {
			continue
		}
		break
	}
	s.lastNSID = nextNSID

	// Create new clientNS instance.
	ns := &clientNS{
		S:        s,
		NSID:     nextNSID,
		sendable: true,
	}
	s.nss[nextNSID] = ns

	// Create channel.
	s.recvQLock.Lock()
	defer s.recvQLock.Unlock()
	s.recvQ[nextNSID] = make(chan *rawPacket, RecvQueueSize)

	return ns, nil
}
func (s *clientS) Serve(ctx context.Context) error {
	select {
	case <-ctx.Done():
	case <-s.closed:
	}
	return s.Close()
}
func (s *clientS) Close() error {
	if s.closed != nil {
		close(s.closed)
	}
	// TODO: wait for workers.
	return nil
}
func (s *clientS) sendPacket(nsid uint64, flags PacketFlag, data interface{}) error {
	err := HandlePanic(func() error {
		sid := parseXDRStructIDTag(reflect.TypeOf(data))

		buf := &bytes.Buffer{}
		enc := NewXDREncoder(utils.WrapMustWriter(buf))
		enc.Struct(data)
		size := uint64(buf.Len())

		buf2 := &bytes.Buffer{}
		enc = NewXDREncoder(utils.WrapMustWriter(buf2))
		enc.Uint64(size)
		enc.Uint64(nsid)
		enc.Uint8(uint8(flags))
		enc.Uint64(sid)
		buf2.Write(buf.Bytes())

		s.sendQ <- buf2.Bytes()
		return nil
	})
	if err != nil {
		return xerrors.Errorf("sendPacket: %w", err)
	}
	return nil
}
func (s *clientS) recvPacket(nsid uint64, empty interface{}) (data interface{}, flags PacketFlag, err error) {
	s.recvQLock.RLock()
	defer s.recvQLock.RUnlock()

	ch := s.recvQ[nsid]
	if ch == nil {
		err := xerrors.Errorf("not found channel: nsid=%d", nsid)
		panic(err)
	}

	// Receive a packet.
	p := <-ch

	// Decode it.
	buf := utils.WrapMustReader(bytes.NewBuffer(p.data))
	dec := NewXDRDecoder(buf)
	data = dec.Struct(empty)
	return data, p.flags, nil
}
func (s *clientS) recvWorker() {
	go func() {
		<-s.closed
		s.Conn.Close()
	}()

	err := HandlePanic(func() error {
		for {
			p := &rawPacket{
				size:  s.Dec.Uint64(),
				nsid:  s.Dec.Uint64(),
				flags: PacketFlag(s.Dec.Uint8()),
				sid:   s.Dec.Uint64(),
				data:  nil,
			}
			p.data = make([]byte, p.size)
			s.R.MustReadAll(p.data)

			if p.nsid > nsidMaxValue {
				err := xerrors.Errorf("NSID is out-of-range")
				panic(err)
			}

			s.recvQLock.RLock()
			ch := s.recvQ[p.nsid]
			if ch == nil {
				s.recvQLock.RUnlock()
				if p.flags|CreateSessionFlag == 0 {
					err := xerrors.Errorf("not found channel: nsid=%d", p.nsid)
					panic(err)
				}
				// New nested session was created by server.
				ns := &clientNS{
					S:           s,
					NSID:        p.nsid,
					established: true,
					sendable:    true,
					receivable:  true,
				}
				s.nssLock.Lock()
				s.nss[ns.NSID] = ns
				s.nssLock.Unlock()
				s.recvQLock.Lock()
				s.recvQ[ns.NSID] = make(chan *rawPacket, RecvQueueSize)
				s.recvQLock.Unlock()
			}
			select {
			case ch <- p:
			case <-s.closed:
				return nil
			}
			s.recvQLock.RUnlock()
		}
	})
	// TODO
	_ = err
}
func (s *clientS) sendWorker() {
	err := HandlePanic(func() error {
		for {
			select {
			case b := <-s.sendQ:
				s.W.MustWriteAll(b)
			case <-s.closed:
				return nil
			}
		}
		return nil
	})
	// TODO
	_ = err
}

type clientNS struct {
	S    *clientS
	NSID uint64

	established bool
	sendable    bool
	receivable  bool
}

func (ns *clientNS) Send(v interface{}) error {
	if ns.receivable {
		return xerrors.Errorf("the nested session (NSID=%d) is already closed", ns.NSID)
	}

	var flags PacketFlag
	if !ns.established {
		flags |= CreateSessionFlag
	}

	err := ns.S.sendPacket(ns.NSID, flags, v)
	if err != nil {
		return xerrors.Errorf("ClientSession.Send: %w", err)
	}
	ns.established = true
	return nil
}
func (ns *clientNS) SendErr(se *SessionError) error {
	if ns.receivable {
		return xerrors.Errorf("the nested session (NSID=%d) is already closed", ns.NSID)
	}

	var flags PacketFlag
	if !ns.established {
		flags |= CreateSessionFlag
	}
	flags |= ErrorSessionFlag

	err := ns.S.sendPacket(ns.NSID, flags, se)
	if err != nil {
		return xerrors.Errorf("ClientSession.SendErr: %w", err)
	}
	ns.established = true
	return nil
}
func (ns *clientNS) Recv(empty interface{}) (interface{}, error) {
	if !ns.established {
		return nil, xerrors.Errorf("the nested session (NSID=%d) is not established", ns.NSID)
	}
	if ns.receivable {
		return nil, xerrors.Errorf("the nested session (NSID=%d) is already closed", ns.NSID)
	}

	data, flags, err := ns.S.recvPacket(ns.NSID, empty)
	if err != nil {
		return nil, xerrors.Errorf("recv: %w", err)
	}
	if flags&CloseSessionFlag != 0 {
		ns.receivable = true
	}
	if flags&ErrorSessionFlag != 0 {
		return nil, xerrors.New(fmt.Sprint(data))
	}
	return data, err
}
func (ns *clientNS) CloseWithError(se SessionError) error {
	if !ns.established {
		return xerrors.Errorf("the nested session (NSID=%d) is not established", ns.NSID)
	}
	if ns.sendable {
		return xerrors.Errorf("the nested session (NSID=%d) is already closed", ns.NSID)
	}

	err := ns.S.sendPacket(ns.NSID, CloseSessionFlag|ErrorSessionFlag, se)
	ns.sendable = true
	return err
}
func (ns *clientNS) Close() error {
	if !ns.established {
		return xerrors.Errorf("the nested session (NSID=%d) is not established", ns.NSID)
	}
	if ns.sendable {
		return xerrors.Errorf("the nested session (NSID=%d) is already closed", ns.NSID)
	}

	err := ns.S.sendPacket(ns.NSID, CloseSessionFlag, &Ping{})
	ns.sendable = true
	return err
}
func (ns *clientNS) IsSendable() bool {
	return ns.sendable
}
func (ns *clientNS) IsReceivable() bool {
	return ns.receivable
}
