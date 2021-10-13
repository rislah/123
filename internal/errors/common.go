package errors

import (
	"fmt"
	"github.com/cep21/circuit"
	"github.com/go-redis/redis/v8"
	"strings"
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

type WrappedError struct {
	StatusCode int
	CodeValue  string
	Message    string
	ShouldLog  bool
}

func (w *WrappedError) Error() string {
	return w.Message
}

type ErrorResponse struct {
	Status     int    `json:"status"`
	Message    string `json:"message"`
	StatusText string `json:"error"`
	ErrorCode  string `json:"error_code"`
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
