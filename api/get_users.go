package api

import (
	"context"
	"net/http"

	"github.com/rislah/fakes/internal/errors"
)

func (s *Mux) GetUsers(ctx context.Context, response *Response, request *http.Request) error {
	users, err := s.userBackend.GetUsers(ctx)
	if err != nil {
		if e, ok := errors.IsWrappedError(ctx, err); ok {
			response.WriteHeader(int(e.Code))
			return response.WriteJSON(errors.NewErrorResponse(e.Msg, int(e.Code)))
		}
		return err
	}

	return response.WriteJSON(users)
}
