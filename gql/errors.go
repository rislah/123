package gql

import (
	"context"

	"github.com/cep21/circuit/v3"
	"github.com/rislah/fakes/internal/errors"
)

const (
	TypeInternal           = "internal error"
	TypeServiceUnavailable = "service is unavailable"
	TypeServiceTimeout     = "service timed out"
)

func sanitizeResolverError(err error) string {
	switch typeError := errors.Cause(err).(type) {
	case circuit.Error:
		return TypeServiceUnavailable
	default:
		if typeError == context.DeadlineExceeded {
			return TypeServiceTimeout
		}

		return err.Error()
	}
}

type Error struct {
	cause   error
	errType string
}

func (e *Error) Cause() error {
	if e == nil {
		return nil
	}

	return e.cause
}

func (e *Error) String() string {
	if e == nil {
		return ""
	}

	return e.cause.Error()
}

func (e *Error) Type() string {
	if e == nil {
		return ""
	}

	if e.errType == "" {
		return ""
	}

	return e.errType
}
