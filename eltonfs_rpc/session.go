package eltonfs_rpc

import (
	"bytes"
	"context"
	"fmt"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"golang.org/x/xerrors"
	"log"
	"net"
	"sync"
	"time"
)

type StructID uint64
type PacketFlag uint8

const (
	CreateSessionFlag PacketFlag = 1 << iota
	CloseSessionFlag
	ErrorSessionFlag
)

const (
	MinNSID       = 1
	MaxNSID       = (1 << 32) - 1
	MinServerNSID = 1
	MaxServerNSID = (1 << 31) - 1
	MinClientNSID = 1 << 31
	MaxClientNSID = (1 << 32) - 1
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
		Conn:   conn,
		R:      r,
		W:      w,
		Enc:    NewXDREncoder(w),
		Dec:    NewXDRDecoder(r),
		Handle: defaultHandler,
	}
}

type NSID uint64
type clientS struct {
	Conn   net.Conn
	R      utils.MustReader
	W      utils.MustWriter
	Enc    XDREncoder
	Dec    XDRDecoder
	Handle RpcHandler

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
	closed  <-chan struct{}
	doClose func()
	// Nested Sessions
	// Key: nested session IDs
	// Value: *clientNS
	nss      map[NSID]*clientNS
	lastNSID NSID
	nssLock  sync.RWMutex

	// Queue for packets waiting to be sent.
	// Elements are serialized packets.
	sendQ chan []byte
	// Queue for packets waiting to be received.
	// A lock must be acquired before access to the recvQ.  And should release the lock before receive data from channel
	// to prevent dead-lock.
	recvQ     map[NSID]chan *rawPacket
	recvQLock sync.RWMutex
}

type rawPacket struct {
	size  uint64
	nsid  NSID
	flags PacketFlag
	sid   StructID
	data  []byte
}

func (s *clientS) Setup() error {
	return HandlePanic(func() error {
		s.Enc.RawPacket(0, 0, &Setup1{
			ClientName: "eltonfs-helper",
		})

		raw, data := s.recvPacketDirect(&Setup2{})
		if raw.flags != 0 {
			return xerrors.Errorf("received an invalid packet: flags should 0, but %d", raw.flags)
		}
		s2 := data.(*Setup2)
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
		ctx, cancel := context.WithCancel(context.Background())
		s.closed = ctx.Done()
		s.doClose = cancel
		s.sendQ = make(chan []byte, SendQueueSize)
		s.recvQ = map[NSID]chan *rawPacket{}
		s.nss = map[NSID]*clientNS{}
		go s.recvWorker()
		go s.sendWorker()
		go s.pinger()
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
	ns := newClientNS(s, nextNSID, true)
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
	if s.doClose != nil {
		s.doClose()
	}
	// TODO: wait for workers.
	return nil
}
func (s *clientS) validateSendPacket(nsid NSID, flags PacketFlag) error {
	if !(MinNSID <= nsid && nsid <= MaxNSID) {
		return xerrors.Errorf("invalid nsid: nsid=%d, flags=%d", nsid, flags)
	}
	if flags&CreateSessionFlag != 0 && !(MinClientNSID <= nsid && nsid <= MaxClientNSID) {
		return xerrors.Errorf("invalid nsid: nsid=%d, flags=%d", nsid, flags)
	}
	return nil
}
func (s *clientS) validateRecvedPacket(nsid NSID, flags PacketFlag) error {
	if !(MinNSID <= nsid && nsid <= MaxNSID) {
		return xerrors.Errorf("invalid nsid: nsid=%d, flags=%d", nsid, flags)
	}
	if flags&CreateSessionFlag != 0 && !(MinServerNSID <= nsid && nsid <= MaxServerNSID) {
		return xerrors.Errorf("invalid nsid: nsid=%d, flags=%d", nsid, flags)
	}
	return nil
}
func (s *clientS) sendPacket(nsid NSID, flags PacketFlag, data interface{}) error {
	err := HandlePanic(func() error {
		if err := s.validateSendPacket(nsid, flags); err != nil {
			panic(err)
		}

		buf := &bytes.Buffer{}
		enc := NewXDREncoder(utils.WrapMustWriter(buf))
		enc.RawPacket(nsid, flags, data)
		s.sendQ <- buf.Bytes()
		return nil
	})
	if err != nil {
		return xerrors.Errorf("sendPacket: %w", err)
	}
	return nil
}
func (s *clientS) recvRawPacket(nsid NSID) *rawPacket {
	s.recvQLock.RLock()
	ch := s.recvQ[nsid]
	s.recvQLock.RUnlock()

	if ch == nil {
		err := xerrors.Errorf("not found channel: nsid=%d", nsid)
		panic(err)
	}

	// Receive a packet.
	return <-ch
}
func (s *clientS) recvPacket(nsid NSID, empty interface{}) (data interface{}, flags PacketFlag, err error) {
	p := s.recvRawPacket(nsid)

	// Decode it.
	buf := utils.WrapMustReader(bytes.NewBuffer(p.data))
	dec := NewXDRDecoder(buf)
	data = dec.Struct(empty)
	return data, p.flags, nil
}
func (s *clientS) recvPacketDirect(empty interface{}) (p *rawPacket, data interface{}) {
	p = s.Dec.RawPacket()

	// Decode
	buf := utils.WrapMustReader(bytes.NewBuffer(p.data))
	dec := NewXDRDecoder(buf)
	data = dec.Struct(empty)

	return p, data
}
func (s *clientS) recvWorker() {
	go func() {
		<-s.closed
		s.Conn.Close()
	}()

	err := HandlePanic(func() error {
		for {
			p := s.Dec.RawPacket()
			if err := s.validateRecvedPacket(p.nsid, p.flags); err != nil {
				panic(err)
			}

			s.recvQLock.RLock()
			ch := s.recvQ[p.nsid]
			s.recvQLock.RUnlock()

			if ch == nil {
				if p.flags|CreateSessionFlag == 0 {
					err := xerrors.Errorf("not found channel: nsid=%d", p.nsid)
					panic(err)
				}

				// New nested session was created by server.
				ns := newClientNS(s, p.nsid, false)
				s.nssLock.Lock()
				s.nss[ns.nsid] = ns
				s.nssLock.Unlock()
				s.recvQLock.Lock()
				ch = make(chan *rawPacket, RecvQueueSize)
				s.recvQ[ns.nsid] = ch
				s.recvQLock.Unlock()

				// Start handler.
				go s.Handle(ns, p.sid, p.flags)
			}

			select {
			case ch <- p:
			case <-s.closed:
				return nil
			}
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
func (s *clientS) pinger() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// TODO: detect connection error and close session.
	for {
		select {
		case <-ticker.C:
			ns, err := s.New()
			if err != nil {
				panic(xerrors.Errorf("pinger: creating new ns: %w", err))
			}
			if err := ns.Send(&Ping{}); err != nil {
				panic(xerrors.Errorf("pinger: sending ping: %w", err))
			}
			if _, err := ns.Recv(&Ping{}); err != nil {
				panic(xerrors.Errorf("pinger: receiving ping: %w", err))
			}
			if err := ns.Close(); err != nil {
				panic(xerrors.Errorf("pinger: closing ns: %w", err))
			}
			log.Println("ping OK")
		case <-s.closed:
			return
		}
	}
}
func (s *clientS) unregisterNS(nsid NSID) {
	s.nssLock.Lock()
	delete(s.nss, nsid)
	s.nssLock.Unlock()
}

// newClientNS initializes clientNS struct.
// If nested session created by opponent, isClient must be false.  Otherwise isClient must be true.
func newClientNS(s *clientS, nsid NSID, isClient bool) *clientNS {
	ns := &clientNS{
		S:    s,
		nsid: nsid,
	}
	if isClient {
		ns.sendable = true
	} else {
		ns.established = true
		ns.sendable = true
		ns.receivable = true
	}
	return ns
}

type clientNS struct {
	S *clientS

	nsid        NSID
	established bool
	sendable    bool
	receivable  bool
}

func (ns *clientNS) Send(v interface{}) error       { return ns.sendWithFlags(v, 0) }
func (ns *clientNS) SendErr(se *SessionError) error { return ns.sendWithFlags(se, ErrorSessionFlag) }
func (ns *clientNS) sendWithFlags(v interface{}, flags PacketFlag) error {
	if !ns.sendable {
		return xerrors.Errorf("the nested session (NSID=%d) is already closed", ns.nsid)
	}
	if !ns.established {
		flags |= CreateSessionFlag
	}

	err := ns.S.sendPacket(ns.nsid, flags, v)
	if err != nil {
		return xerrors.Errorf("ClientSession.Send: %w", err)
	}
	if !ns.established {
		ns.established = true
		ns.receivable = true
	}
	return nil
}
func (ns *clientNS) Recv(empty interface{}) (interface{}, error) {
	if !ns.receivable {
		return nil, xerrors.Errorf("the nested session (NSID=%d) is not receivable", ns.nsid)
	}

	data, flags, err := ns.S.recvPacket(ns.nsid, empty)
	if err != nil {
		return nil, xerrors.Errorf("recv: %w", err)
	}
	if flags&CloseSessionFlag != 0 {
		ns.receivable = false
	}
	if flags&ErrorSessionFlag != 0 {
		return nil, xerrors.New(fmt.Sprint(data))
	}
	return data, err
}
func (ns *clientNS) CloseWithError(se SessionError) error { return ns.closeWith(se, ErrorSessionFlag) }
func (ns *clientNS) Close() error                         { return ns.closeWith(&Ping{}, 0) }
func (ns *clientNS) closeWith(v interface{}, flags PacketFlag) error {
	var err error
	if !ns.established {
		return xerrors.Errorf("the nested session (NSID=%d) is not established", ns.nsid)
	}
	if !ns.sendable {
		return xerrors.Errorf("the nested session (NSID=%d) is already closed", ns.nsid)
	}
	ns.sendable = false

	if !ns.receivable {
		// Closed from both direction.
		//
		// 1. Closed from kmod.  Send a packet with close flags.
		// 2. Closed from UMH.  Send a packet with close flags and release memory.
		//       ↑↑↑  NOW
		err = ns.S.sendPacket(ns.nsid, CloseSessionFlag|flags, v)
		ns.S.unregisterNS(ns.nsid)
		// 3. Kmod receives a packet with close flags.  Should release memory.
	} else {
		// Closed from UMH.  Wait for close ns from kmod.
		//
		// 1. Closed from UMH.  Send a packet with close flags.
		//       ↑↑↑  NOW
		// 2. Closed from kmod.  Send a packet with close flags and release memory.
		// 3. Receive a packet with close flags from kmod.  Should release memory.

		// Send a packet with close flags.
		err = ns.S.sendPacket(ns.nsid, CloseSessionFlag|flags, v)

		// Wait for receive a packet with close flags.
		for {
			p := ns.S.recvRawPacket(ns.nsid)
			closed := p.flags&CloseSessionFlag != 0
			isPing := p.sid == PingStructID

			if !closed {
				log.Printf("receive a packet after closed: nsid=%d, struct_id=%d", p.nsid, p.sid)
			}
			if closed && !isPing {
				log.Printf("nested session closed with unexpected struct type: nsid=%d, struct_id=%d", p.nsid, p.sid)
			}

			if closed {
				break
			}
		}

		// Received a closed packet.  Should unregister the ns.
		ns.S.unregisterNS(ns.nsid)
	}

	return err
}
func (ns *clientNS) IsSendable() bool {
	return ns.sendable
}
func (ns *clientNS) IsReceivable() bool {
	return ns.receivable
}
