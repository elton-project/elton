package controller_db

import "fmt"

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
	Err error
}

func wrapInternalError(msg string, err error) error {
	if err == nil {
		return nil
	}

	wrapped := &InternalError{err}
	if msg == "" {
		return wrapped
	}
	return fmt.Errorf("%s: %w", msg, wrapped)
}
func (e *InternalError) Error() string {
	return e.Err.Error()
}
