package api

import (
	"context"
	"net/http"
	"time"

	"github.com/rislah/fakes/internal/app/user"
	"github.com/rislah/fakes/internal/errors"

	"github.com/rislah/fakes/internal/geoip"
	"github.com/rislah/fakes/internal/metrics"
	"github.com/segmentio/stats/v4/prometheus"

	"github.com/gorilla/mux"
	"github.com/rislah/fakes/internal/logger"
	"github.com/rislah/fakes/internal/redis"
	"github.com/rislah/fakes/internal/throttler"
)

type Mux struct {
	*mux.Router
	userService      user.User
	requestThrottler throttler.Throttler
	logger           *logger.Logger
}

func NewMux(service user.User, gip geoip.GeoIP, client redis.Client, mtr metrics.Metrics, logger *logger.Logger) *Mux {
	router := mux.NewRouter()
	router.Handle("/metrics", prometheus.DefaultHandler).Methods("GET")

	s := &Mux{
		Router:      router,
		userService: service,
		requestThrottler: throttler.New(&throttler.Config{
			Client:             client,
			KeyType:            "request",
			AttemptLimit:       400,
			AttemptWindow:      10 * time.Minute,
			BaseTimeout:        1 * time.Minute,
			MaxTimeout:         24 * time.Hour,
			TimeoutScaleFactor: 30.0,
		}),
		logger: logger,
	}

	subRouter := router.NewRoute().Subrouter()
	subRouter.Use(requestsLoggerMiddleware(logger, gip))
	subRouter.Use(metricsMiddleware(mtr))
	subRouter.Use(contextMiddleWare)
	subRouter.Handle("/users", s.wrap(s.GetUsers)).Methods("GET")

	return s
}

type apiFunc func(ctx context.Context, response *Response, request *http.Request) (error)

func (s *Mux) wrap(handler apiFunc) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		resp := &Response{ResponseWriter: rw}
		err := handler(r.Context(), resp, r)
		if err != nil {
			if !resp.WasWritten() {
				resp.WriteHeader(http.StatusInternalServerError)
			}

			switch resp.Status() {
			case http.StatusInternalServerError:
				s.logger.LogRequestError(err, r)
				resp.WriteJSON(errors.NewErrorResponse("Internal server error has occured", http.StatusInternalServerError))
			}

			return
		}

		if !resp.WasWritten() {
			resp.WriteHeader(http.StatusOK)
		}

	})
}
