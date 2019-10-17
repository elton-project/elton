package eltonfs_rpc

import (
	"fmt"
	"golang.org/x/xerrors"
)

func HandlePanic(fn func() error) (err error) {
	defer func() {
		if o := recover(); o != nil {
			var ok bool
			err, ok = o.(error)
			if !ok {
				err = xerrors.New(fmt.Sprint(o))
			}
		}
	}()
	err = fn()
	return
}
