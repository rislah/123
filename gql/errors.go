package gql

import (
	"context"

	"github.com/cep21/circuit/v3"
	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/logger"
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
	case *errors.WrappedError:
		return err.Error()
	default:
		if typeError == context.DeadlineExceeded {
			return TypeServiceTimeout
		}

		logger.SharedGlobalLogger.Error("Internal Error", err)

		return TypeInternal
	}
}
