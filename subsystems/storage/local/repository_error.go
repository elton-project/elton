package localStorage

import (
	"fmt"
	werror "github.com/sonatard/werror/xerrors"
	"golang.org/x/xerrors"
)

type ObjectTooLargeError struct {
	werror.WrapError
	received uint64
	limit    uint64
}

func NewObjectTooLargeError(received, limit uint64) *ObjectTooLargeError {
	err := &ObjectTooLargeError{
		received: received,
		limit:    limit,
	}
	err.WrapError = werror.Wrap(err, nil, 2)
	return err
}
func (e ObjectTooLargeError) Wrap(next error) error {
	e.WrapError = werror.Wrap(&e, next, 2)
	return &e
}
func (e *ObjectTooLargeError) Error() string {
	return fmt.Sprintf("object too large: received=%d, limit=%d", e.received, e.limit)
}
func (e *ObjectTooLargeError) Is(err error) bool {
	var other *ObjectTooLargeError
	return xerrors.As(err, &other)
}

type ObjectNotFoundError struct {
	werror.WrapError
	key Key
}

func NewObjectNotFoundError(key Key) *ObjectNotFoundError {
	err := &ObjectNotFoundError{
		key: key,
	}
	err.WrapError = werror.Wrap(err, nil, 2)
	return err
}
func (e ObjectNotFoundError) Wrap(next error) error {
	e.WrapError = werror.Wrap(&e, next, 2)
	return &e
}
func (e *ObjectNotFoundError) Error() string {
	return fmt.Sprintf("object not found: key=%s", e.key)
}
func (e *ObjectNotFoundError) Is(err error) bool {
	var other *ObjectNotFoundError
	return xerrors.As(err, &other)
}

type MetadataTooLargeError struct {
	werror.WrapError
}

func NewMetadataTooLargeError() *MetadataTooLargeError {
	err := &MetadataTooLargeError{}
	err.WrapError = werror.Wrap(err, nil, 2)
	return err
}
func (e MetadataTooLargeError) Wrap(next error) error {
	e.WrapError = werror.Wrap(&e, next, 2)
	return &e
}
func (e *MetadataTooLargeError) Error() string {
	return fmt.Sprintf("metadata too large")
}
func (e *MetadataTooLargeError) Is(err error) bool {
	var other *MetadataTooLargeError
	return xerrors.As(err, &other)
}

type InvalidObject struct {
	werror.WrapError
	cause string
}

func NewInvalidObject(cause string) *InvalidObject {
	err := &InvalidObject{
		cause: cause,
	}
	err.WrapError = werror.Wrap(err, nil, 2)
	return err
}
func (e InvalidObject) Wrap(next error) error {
	e.WrapError = werror.Wrap(&e, next, 2)
	return &e
}
func (e *InvalidObject) Error() string {
	return fmt.Sprintf("invalid object: %s", e.cause)
}
func (e *InvalidObject) Is(err error) bool {
	var other *InvalidObject
	return xerrors.As(err, &other)
}
