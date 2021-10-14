package api

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/rislah/fakes/internal/encoder"
	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/logger"
	"github.com/rislah/fakes/internal/throttler"
)

func (s *Server) GetUsers(w http.ResponseWriter, r *http.Request) {
	if shouldThrottle(r.Context(), s.requestThrottler, s.logger) {
		encoder.ServeJSON(w, errors.ErrorResponse{
			Status:     http.StatusTooManyRequests,
			Message:    "You have requested this endpoint too many times. Please try again later.",
			StatusText: http.StatusText(http.StatusTooManyRequests),
			ErrorCode:  "too_many_attempts",
		}, http.StatusTooManyRequests)
		return
	}

	users, err := s.userService.GetUsers(r.Context())
	if err != nil {
		if err := encoder.ServeError(w, r, err); err != nil {
			s.logger.Error("error sending response", err, nil)
			return
		}
		return
	}
	encoder.ServeJSON(w, users, 200)
}

func shouldThrottle(ctx context.Context, th throttler.Throttler, l *logger.Logger) bool {
	ip := ctx.Value("remote_ip").(net.IP)

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
