package errors_test

import (
	"testing"

	"github.com/rislah/fakes/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestErrorNew(t *testing.T) {
	e := errors.New("test")
	assert.Equal(t, "test", e.Error())
	_, ok := e.(errors.Error)
	assert.True(t, ok)
}

func TestErrorWrap(t *testing.T) {
	e1 := errors.New("test_1", errors.Fields{"key_1": "val_1"})
	e2 := errors.Wrap(e1, "wrapped error", errors.Fields{"key_2": "val_2"})

	v, ok := e2.Fields()["key_1"]
	assert.True(t, ok)
	assert.Equal(t, "val_1", v)

	v, ok = e2.Fields()["key_2"]
	assert.True(t, ok)
	assert.Equal(t, "val_2", v)

	unwrapped := errors.Unwrap(e2)
	_, ok = unwrapped.(errors.Error)
	assert.False(t, ok)
}
