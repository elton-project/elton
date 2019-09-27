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
	return e.Msg
}
func (e InputError) Wrap(next error) error {
	e.WrapError = werror.Wrap(&e, next, 2)
	return &e
}
func (e *InputError) Is(err error) bool {
	var ie *InputError
	return errors.As(err, &ie)
}

// Errors caused by incorrect data or incorrect operations.
var (
	ErrDupVolumeID          = &InputError{Msg: "duplicate volume id"}
	ErrDupVolumeName        = &InputError{Msg: "duplicate volume name"}
	ErrNotFoundVolume       = &InputError{Msg: "not found volume"}
	ErrNotFoundCommit       = &InputError{Msg: "not found commit"}
	ErrNotFoundTree         = &InputError{Msg: "not found tree"}
	ErrNotFoundProp         = &InputError{Msg: "not found property"}
	ErrAlreadyExists        = &InputError{Msg: "already exists"}
	ErrNotAllowedReplace    = &InputError{Msg: "replacement not allowed"}
	ErrCrossVolumeCommit    = &InputError{Msg: "cross-volume commit"}
	ErrMismatchParentCommit = &InputError{Msg: "mismatch parent commit id"}
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
	return errors.As(err, &ie)
}

// Internal errors in database.
var (
	IErrInitialize = &InternalError{Msg: "initialize db"}
	IErrDatabase   = &InternalError{Msg: "database error"}
	IErrOpen       = &InternalError{Msg: "open database"}
	IErrClose      = &InternalError{Msg: "close database"}
	IErrDelete     = &InternalError{Msg: "delete record"}
)
