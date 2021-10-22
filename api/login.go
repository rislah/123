package api

import (
	"context"
	"encoding/json"
	"github.com/rislah/fakes/internal/credentials"
	"net"
	"net/http"

	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/ratelimiter"
)

type LoginResponse struct {
	Token string `json:"token"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Mux) Login(ctx context.Context, response *Response, req *http.Request) error {
	var loginReq LoginRequest
	if err := json.NewDecoder(req.Body).Decode(&loginReq); err != nil {
		return err
	}

	creds := credentials.New(loginReq.Username, loginReq.Password)
	usr, err := s.authenticator.AuthenticatePassword(ctx, creds)
	if err != nil {
		return errors.IsWrappedErrorWriteErrorResponse(ctx, response, err)
	}

	token, err := s.authenticator.GenerateJWT(usr)
	if err != nil {
		return err
	}

	return response.WriteJSON(LoginResponse{
		Token: token,
	})
}

func (s *Mux) isLoginThrottled(ctx context.Context, response *Response, req *http.Request) bool {
	ip := ctx.Value(RemoteIPContextKey).(net.IP)
	field := ratelimiter.Field{
		Scope:      "ip",
		Identifier: ip.String(),
	}

	throttled, err := s.userLoginRatelimiter.ShouldThrottle(ctx, response, field)
	if err != nil {
		s.logger.LogRequestError(errors.Wrap(err, "userRegisterRateLimiter"), req)
		return false
	}

	return throttled
}
