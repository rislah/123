package api

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/logger"
	"github.com/rislah/fakes/internal/throttler"
)

func (s *Mux) GetUsers(ctx context.Context, response *Response, request *http.Request) error {
	if shouldThrottle(ctx, s.requestThrottler, s.logger) {
		response.WriteHeader(int(errors.ErrRateLimited))
		return response.WriteJSON(errors.NewErrorResponse("You have been rate limited", int(errors.ErrRateLimited)))
	}

	users, err := s.userService.GetUsers(ctx)
	if err != nil {
		if e, ok := errors.IsWrappedError(ctx, err); ok {
			response.WriteHeader(int(e.Code))
			return response.WriteJSON(errors.NewErrorResponse(e.Msg, int(e.Code)))
		}
		return err
	}

	return response.WriteJSON(users)
}

func shouldThrottle(ctx context.Context, th throttler.Throttler, l *logger.Logger) bool {
	ip := ctx.Value(contextIPKey).(net.IP)

	keys := []throttler.ID{
		{
			Key:  ip.String(),
			Type: "ip",
		},
	}

	shouldThrottle, err := th.Incr(ctx, keys)
	if err != nil {
		fmt.Println(err)
		return true
	}

	return shouldThrottle
}
