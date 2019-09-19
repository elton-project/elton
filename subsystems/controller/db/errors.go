package controller_db

import werror "github.com/sonatard/werror/xerrors"

type InputError struct {
	Msg string
}

func (e *InputError) Error() string {
	return e.Msg
}

var ErrDupVolumeID = &InputError{"duplicate volume id"}
var ErrDupVolumeName = &InputError{"duplicate volume name"}
var ErrNotFoundVolume = &InputError{"not found volume"}
var ErrNotFoundCommit = &InputError{"not found commit"}
var ErrNotFoundTree = &InputError{"not found tree"}

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

var IErrInitialize = &InternalError{Msg: "initialize db"}
var IErrDatabase = &InternalError{Msg: "database error"}
var IErrOpen = &InternalError{Msg: "open database"}
var IErrClose = &InternalError{Msg: "close database"}
var IErrDelete = &InternalError{Msg: "delete record"}
