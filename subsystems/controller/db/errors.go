package controller_db

import (
	"errors"
	werror "github.com/sonatard/werror/xerrors"
)

// InputError represents an error that the database received incorrect data or incorrect operations.
type InputError struct {
	Msg string
	werror.WrapError
}

func (e *InputError) Error() string {
	return e.Msg + ": " + errors.Unwrap(e).Error()
}
func (e InputError) Wrap(next error) error {
	e.WrapError = werror.Wrap(&e, next, 2)
	return &e
}
func (e *InputError) Is(err error) bool {
	var ie *InputError
	if !errors.As(err, &ie) {
		return false
	}
	if ie.Msg == "" {
		// Only check the error type.
		return true
	}
	return e.Msg == ie.Msg
}

// Errors caused by incorrect data or incorrect operations.
var (
	ErrDupVolumeID         = &InputError{Msg: "duplicate volume id"}
	ErrDupVolumeName       = &InputError{Msg: "duplicate volume name"}
	ErrNotFoundVolume      = &InputError{Msg: "not found volume"}
	ErrNotFoundCommit      = &InputError{Msg: "not found commit"}
	ErrNotFoundTree        = &InputError{Msg: "not found tree"}
	ErrNotFoundProp        = &InputError{Msg: "not found property"}
	ErrNotFoundNode        = &InputError{Msg: "not found node"}
	ErrAlreadyExists       = &InputError{Msg: "already exists"}
	ErrNodeAlreadyExists   = &InputError{Msg: "node already exists"}
	ErrNotAllowedReplace   = &InputError{Msg: "replacement not allowed"}
	ErrCrossVolumeCommit   = &InputError{Msg: "cross-volume commit"}
	ErrInvalidParentCommit = &InputError{Msg: "invalid parent commit"}
	ErrInvalidTree         = &InputError{Msg: "invalid tree"}
	ErrLatestCommitUpdated = &InputError{Msg: "latest commit is updated by other thread"}
)

// InternalError represents an error of database internal error.
type InternalError struct {
	Msg string
	werror.WrapError
}

func (e *InternalError) Error() string {
	return e.Msg
}
func (e InternalError) Wrap(next error) error {
	e.WrapError = werror.Wrap(&e, next, 2)
	return &e
}
func (e *InternalError) Is(err error) bool {
	var ie *InternalError
	if !errors.As(err, &ie) {
		return false
	}
	if ie.Msg == "" {
		// Only check the error type.
		return true
	}
	return e.Msg == ie.Msg
}

// Internal errors in database.
// All internal errors have names that start with "IErr".
var (
	IErrInitialize = &InternalError{Msg: "initialize db"}
	IErrDatabase   = &InternalError{Msg: "database error"}
	IErrOpen       = &InternalError{Msg: "open database"}
	IErrClose      = &InternalError{Msg: "close database"}
	IErrDelete     = &InternalError{Msg: "delete record"}
)
