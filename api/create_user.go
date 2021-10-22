package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rislah/fakes/internal/credentials"
	"github.com/rislah/fakes/internal/errors"
)

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	Username string `json:"username"`
}

func (s *Mux) CreateUser(ctx context.Context, response *Response, req *http.Request) error {
	var createUserReq CreateUserRequest
	if err := json.NewDecoder(req.Body).Decode(&createUserReq); err != nil {
		return err
	}

	creds := credentials.New(createUserReq.Username, createUserReq.Password)
	if err := s.userBackend.CreateUser(ctx, creds); err != nil {
		return errors.IsWrappedErrorWriteErrorResponse(ctx, response, err)
	}

	response.WriteHeader(http.StatusCreated)
	return response.WriteJSON(CreateUserResponse{
		Username: createUserReq.Username,
	})
}
