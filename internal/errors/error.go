package errors

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/cep21/circuit"
	"github.com/go-redis/redis/v8"
)

type Error interface {
	Error() string
	String() string
	Fields() Fields
}

type Fields map[string]interface{}

type errorImpl struct {
	error
	wrapMessages []string
	fields       Fields
}

func New(val interface{}, fields ...Fields) Error {
	return NewWithSkip(val, 1, fields...)
}

func Wrap(err error, message string, fields ...Fields) Error {
	if err == nil {
		return nil
	}

	e := NewWithSkip(err, 1, fields...)
	wrappedErr := e.(*errorImpl)
	wrappedErr.wrapMessages = append([]string{message}, wrappedErr.wrapMessages...)
	return wrappedErr
}

func Unwrap(err error) error {
	return Cause(err)
}

func NewWithSkip(val interface{}, skip int, fields ...Fields) Error {
	var err *errorImpl
	switch e := val.(type) {
	case nil:
		return nil
	case *errorImpl:
		err = cloneErrorImpl(e)
	case error:
		err = &errorImpl{error: e}
	default:
		err = &errorImpl{error: fmt.Errorf("%v", e)}
	}

	if err.fields == nil {
		err.fields = make(map[string]interface{})
	}

	if _, ok := err.fields["stack"]; !ok {
		err.fields["stack"] = BuildStack(3)
	}

	for _, f := range fields {
		for k, v := range f {
			err.fields[k] = v
		}
	}

	return err
}

func (e *errorImpl) Error() string {
	tokens := append(e.wrapMessages, e.error.Error())
	return strings.Join(tokens, ":")
}

func (e *errorImpl) String() string {
	return e.Error()
}

func (e *errorImpl) Fields() Fields {
	return e.fields
}

func (e *errorImpl) Cause() error {
	return e.error
}

type causer interface {
	Cause() error
}

func Cause(err error) error {
	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}

		err = cause.Cause()
	}

	return err
}

func cloneErrorImpl(err *errorImpl) *errorImpl {
	clonedWrapMessages := make([]string, len(err.wrapMessages))
	copy(clonedWrapMessages, err.wrapMessages)

	clonedFields := make(Fields, len(err.fields))
	for key, val := range err.fields {
		clonedFields[key] = val
	}

	return &errorImpl{
		fields:       clonedFields,
		wrapMessages: clonedWrapMessages,
		error:        err.error,
	}
}

type ErrorCode int

const (
	ErrBadRequest  ErrorCode = 404
	ErrRateLimited ErrorCode = 429
)

type WrappedError struct {
	Code ErrorCode
	Msg  string
}

func (w WrappedError) Error() string {
	return w.Msg
}

type ErrorResponse struct {
	Status     int    `json:"status"`
	Message    string `json:"message"`
	StatusText string `json:"error"`
}

func NewErrorResponse(msg string, code int) ErrorResponse {
	return ErrorResponse{
		Status:     code,
		StatusText: http.StatusText(code),
		Message:    msg,
	}
}

func IsWrappedRedisNilError(err error) bool {
	if err == nil {
		return false
	}

	simpleErr, ok := err.(*circuit.SimpleBadRequest)
	if ok {
		err = simpleErr.Err
	}

	return err == redis.Nil

}

func IsWrappedError(ctx context.Context, err error) (WrappedError, bool) {
	if err == nil {
		return WrappedError{}, false
	}

	select {
	case <-ctx.Done():
		if ctx.Err() == context.Canceled {
			return WrappedError{}, false
		}
	default:
	}

	if e, ok := Unwrap(err).(WrappedError); ok {
		return e, true
	}

	return WrappedError{}, false
}
