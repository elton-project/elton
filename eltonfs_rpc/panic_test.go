package eltonfs_rpc

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandlePanic(t *testing.T) {
	targetError := errors.New("ok")

	t.Run("should_return_original_error", func(t *testing.T) {
		err := HandlePanic(func() error {
			return targetError
		})
		assert.Equal(t, targetError, err)
	})
	t.Run("should_handle_panic", func(t *testing.T) {
		err := HandlePanic(func() error {
			panic(targetError)
		})
		assert.Equal(t, targetError, err)
	})
	t.Run("should_handle_non_error_object", func(t *testing.T) {
		err := HandlePanic(func() error {
			panic("string object")
		})
		assert.EqualError(t, err, "string object")
	})
}
