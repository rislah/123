package encoder

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/rislah/fakes/internal/errors"
)

var (
	ErrInternalServerError = errors.ErrorResponse{
		Status:     http.StatusInternalServerError,
		Message:    "Internal server error has occurred",
		StatusText: http.StatusText(http.StatusInternalServerError),
		ErrorCode:  "internal_server_error",
	}
)

func ServeError(w http.ResponseWriter, r *http.Request, err error) error {
	unwrappedErr := errors.Unwrap(err)
	if e, ok := unwrappedErr.(*errors.WrappedError); ok {
		if e.StatusCode == http.StatusInternalServerError {
			return ServeJSON(w, ErrInternalServerError, e.StatusCode)

		}

		return ServeJSON(w, errors.ErrorResponse{
			Status:     e.StatusCode,
			Message:    e.Message,
			StatusText: http.StatusText(e.StatusCode),
			ErrorCode:  e.CodeValue,
		}, e.StatusCode)
	}

	return ServeJSON(w, ErrInternalServerError, http.StatusInternalServerError)
}

func ServeJSON(w http.ResponseWriter, v interface{}, code int) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}
