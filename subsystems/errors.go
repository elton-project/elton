package subsystems

import (
	"fmt"
	werror "github.com/sonatard/werror/xerrors"
	"golang.org/x/xerrors"
)

type RestartRequest struct {
	werror.WrapError
	Reason string
}

func NewRestartRequest(reason string) *RestartRequest {
	err := &RestartRequest{
		Reason: reason,
	}
	err.WrapError = werror.Wrap(err, nil, 2)
	return err
}
func (e RestartRequest) Wrap(next error) error {
	e.WrapError = werror.Wrap(&e, next, 2)
	return &e
}
func (e *RestartRequest) Error() string {
	return fmt.Sprintf("RestartRequest: %s", e.Reason)
}
func (e *RestartRequest) Is(err error) bool {
	var other *RestartRequest
	return xerrors.As(err, &other)
}
