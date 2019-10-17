package eltonfs_rpc

import (
	"net"
)

type ClientSession interface {
	// Setup initialize the connection.
	Setup() error
	// New creates new nested session.
	New() ClientNS
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

type clientS struct {
	Conn net.Conn
}

func (s *clientS) Setup() error  { panic("todo") }
func (s *clientS) New() ClientNS { panic("todo") }
func (s *clientS) Close() error  { panic("todo") }

type clientNS struct {
	s clientS
}

func (ns *clientNS) Send(v interface{}) error                    { panic("todo") }
func (ns *clientNS) SendErr(err *SessionError) error             { panic("todo") }
func (ns *clientNS) Recv(empty interface{}) (interface{}, error) { panic("todo") }
func (ns *clientNS) CloseWithError(err SessionError) error       { panic("todo") }
func (ns *clientNS) Close() error                                { panic("todo") }
func (ns *clientNS) IsSendable() bool                            { panic("todo") }
func (ns *clientNS) IsReceivable() bool                          { panic("todo") }
