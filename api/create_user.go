package api

import (
	"context"
	"encoding/json"
	"net"
	"net/http"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/password"
	"github.com/rislah/fakes/internal/ratelimiter"
)

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

	var user app.User
	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		return err
	}

	pass := password.Password(user.Password)
	passHash, err := pass.GenerateBCrypt()
	if err != nil {
		return err
	}

	_, err = pass.ValidatePassword(user.Username)
	if err != nil {
		if e, ok := errors.IsWrappedError(ctx, err); ok {
			response.WriteHeader(int(e.Code))
			return response.WriteJSON(errors.NewErrorResponse(e.Msg, int(e.Code)))
		}
	}

	user.Password = passHash

	if err := s.userService.CreateUser(ctx, user); err != nil {
		return err
	}

	return nil
}
