package api

import (
	"context"
	"net/http"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/errors"
)

type GetUsersResponse struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Role     app.Role `json:"role"`
}

func (s *Mux) GetUsers(ctx context.Context, response *Response, request *http.Request) error {
	users, err := s.userBackend.GetUsers(ctx)
	if err != nil {
		if e, ok := errors.IsWrappedError(ctx, err); ok {
			response.WriteHeader(int(e.Code))
			return response.WriteJSON(errors.NewErrorResponse(e.Msg, int(e.Code)))
		}
		return err
	}

	var res []GetUsersResponse
	for _, usr := range users {
		res = append(res, GetUsersResponse{
			UserID:   usr.UserID,
			Username: usr.Username,
			Role:     usr.Role,
		})
	}

	return response.WriteJSON(res)
}
