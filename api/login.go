package api

import (
	"context"
	"encoding/json"
	"net"
	"net/http"

	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/password"
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
	throttled := s.isLoginThrottled(ctx, response, req)
	if throttled {
		response.WriteHeader(http.StatusTooManyRequests)
		return response.WriteJSON(errors.NewErrorResponse("You are being rate limited", http.StatusTooManyRequests))
	}

	var loginReq LoginRequest
	if err := json.NewDecoder(req.Body).Decode(&loginReq); err != nil {
		return err
	}

	p := password.NewPassword(loginReq.Password)
	if err := p.ValidateLength(); err != nil {
		if e, ok := errors.IsWrappedError(ctx, err); ok {
			response.WriteHeader(int(e.Code))
			return response.WriteJSON(errors.NewErrorResponse(e.Msg, int(e.Code)))
		}
	}

	token, err := s.userService.Login(ctx, loginReq.Username, loginReq.Password)
	if err != nil {
		if e, ok := errors.IsWrappedError(ctx, err); ok {
			response.WriteHeader(int(e.Code))
			return response.WriteJSON(errors.NewErrorResponse(e.Msg, int(e.Code)))
		}
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
