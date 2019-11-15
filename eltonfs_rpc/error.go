package eltonfs_rpc

import (
	"errors"
	"fmt"
)

type ErrorID uint64

const (
	_ ErrorID = iota // ignore first value (0) by assigning to blank identifier
	Cancelled
	InvalidStruct
	UnsupportedStruct
	Unauthorized
	Aborted
	Unimplemented
	Internal
	Unavailable
)

var errorMsgs = map[ErrorID]string{
	Cancelled:         "cancelled",
	InvalidStruct:     "invalid struct",
	UnsupportedStruct: "unsupported struct",
	Unauthorized:      "unauthorized",
	Aborted:           "aborted",
	Unimplemented:     "unimplemented",
	Internal:          "internal error",
	Unavailable:       "unavailable",
}

func (id ErrorID) String() string {
	msg := errorMsgs[id]
	if msg == "" {
		return "unknown"
	}
	return msg
}

const SessionErrorStructID = 4

// SessionError represents an error in the session or nested sessions.
type SessionError struct {
	XXX_XDR_ID struct{} `xdrid:"4"`
	ErrorID    ErrorID  `xdr:"1"`
	Reason     string   `xdr:"2"`
}

func (e *SessionError) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorID, e.Reason)
}
func (e *SessionError) Is(err error) bool {
	var e2 *SessionError
	if !errors.As(err, &e2) {
		return false
	}
	if e2.ErrorID == 0 {
		// Only check the error type.
		return true
	}
	return e.ErrorID == e2.ErrorID
}
