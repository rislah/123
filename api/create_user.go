package api

import (
	"context"
	"encoding/json"
	"net"
	"net/http"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/credentials"
	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/ratelimiter"
)

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	Username string `json:"username"`
}

func (s *Mux) CreateUser(ctx context.Context, response *Response, req *http.Request) error {
	ip := ctx.Value(RemoteIPContextKey).(net.IP)
	if ip != nil {
		field := ratelimiter.Field{
			Scope:      "ip",
			Identifier: ip.String(),
		}

		throttled, err := s.userRegisterRatelimiter.ShouldThrottle(ctx, response, field)
		if err != nil {
			s.logger.LogRequestError(errors.Wrap(err, "userRegisterRateLimiter"), req)
		}

		if throttled {
			response.WriteHeader(int(errors.ErrRateLimited))
			return response.WriteJSON(errors.NewErrorResponse("You have been rate limited", int(errors.ErrRateLimited)))
		}
	}

	var createUserReq CreateUserRequest
	if err := json.NewDecoder(req.Body).Decode(&createUserReq); err != nil {
		return err
	}

	creds := credentials.New(createUserReq.Username, createUserReq.Password)
	if err := creds.Valid(); err != nil {
		return errors.IsWrappedErrorWriteErrorResponse(ctx, response, err)
	}

	bcryptHash, err := creds.Password.GenerateBCrypt()
	if err != nil {
		return errors.Wrap(err, "createuser: generating bcrypt hash")
	}

	if err := s.userBackend.CreateUser(ctx, app.User{
		Username: createUserReq.Username,
		Password: bcryptHash,
	}); err != nil {
		return errors.IsWrappedErrorWriteErrorResponse(ctx, response, err)
	}

	response.WriteHeader(http.StatusCreated)

	return response.WriteJSON(CreateUserResponse{
		Username: createUserReq.Username,
	})
}
