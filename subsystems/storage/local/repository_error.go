package localStorage

import (
	"fmt"
	werror "github.com/sonatard/werror/xerrors"
	"golang.org/x/xerrors"
)

type ObjectTooLarge struct {
	werror.WrapError
	received uint64
	limit    uint64
}

func NewObjectTooLarge(received, limit uint64) *ObjectTooLarge {
	err := &ObjectTooLarge{
		received: received,
		limit:    limit,
	}
	err.WrapError = werror.Wrap(err, nil, 2)
	return err
}
func (e ObjectTooLarge) Wrap(next error) error {
	e.WrapError = werror.Wrap(&e, next, 2)
	return &e
}
func (e *ObjectTooLarge) Error() string {
	return fmt.Sprintf("object too large: received=%d, limit=%d", e.received, e.limit)
}
func (e *ObjectTooLarge) Is(err error) bool {
	var other ObjectTooLarge
	return xerrors.As(err, &other)
}
