package controller_db

import (
	"errors"
	werror "github.com/sonatard/werror/xerrors"
)

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

var ErrDupVolumeID = &InputError{Msg: "duplicate volume id"}
var ErrDupVolumeName = &InputError{Msg: "duplicate volume name"}
var ErrNotFoundVolume = &InputError{Msg: "not found volume"}
var ErrNotFoundCommit = &InputError{Msg: "not found commit"}
var ErrNotFoundTree = &InputError{Msg: "not found tree"}
var ErrNotFoundProp = &InputError{Msg: "not found property"}
var ErrAlreadyExists = &InputError{Msg: "already exists"}
var ErrNotAllowedReplace = &InputError{Msg: "replacement not allowed"}

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

var IErrInitialize = &InternalError{Msg: "initialize db"}
var IErrDatabase = &InternalError{Msg: "database error"}
var IErrOpen = &InternalError{Msg: "open database"}
var IErrClose = &InternalError{Msg: "close database"}
var IErrDelete = &InternalError{Msg: "delete record"}
