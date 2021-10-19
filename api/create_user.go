package api

import (
	"context"
	"encoding/json"
	"net/http"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/errors"
)

func (s *Mux) CreateUser(ctx context.Context, response *Response, req *http.Request) error {
	var user app.User
	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		return err
	}

	err := s.userService.CreateUser(ctx, user)
	if err != nil {
		if e, ok := errors.IsWrappedError(ctx, err); ok {
			response.WriteHeader(int(e.Code))
			return response.WriteJSON(errors.NewErrorResponse(e.Msg, int(e.Code)))
		}
		return err
	}

	return nil
}
