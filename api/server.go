package api

import (
	"context"
	"net/http"
	"time"

	"github.com/rislah/fakes/internal/app/user"
	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/jwt"
	"github.com/rislah/fakes/internal/ratelimiter"

	"github.com/rislah/fakes/internal/geoip"
	"github.com/rislah/fakes/internal/metrics"
	"github.com/segmentio/stats/v4/prometheus"

	"github.com/gorilla/mux"
	"github.com/rislah/fakes/internal/logger"
	"github.com/rislah/fakes/internal/redis"
)

type Mux struct {
	*mux.Router
	userService             user.User
	userRegisterRatelimiter *ratelimiter.Ratelimiter
	userLoginRatelimiter    *ratelimiter.Ratelimiter
	jwtWrapper              jwt.Wrapper
	logger                  *logger.Logger
}

func NewMux(service user.User, jwtWrapper jwt.Wrapper, gip geoip.GeoIP, client redis.Client, mtr metrics.Metrics, logger *logger.Logger) *Mux {
	router := mux.NewRouter()
	router.Handle("/metrics", prometheus.DefaultHandler).Methods("GET")

	userRegisterRatelimiter := ratelimiter.NewRateLimiter(&ratelimiter.Options{
		Name:           "user_register",
		Datastore:      ratelimiter.NewRedisDatastore(client),
		LimitPerMinute: 10,
		WindowInterval: 1 * time.Minute,
		BucketInterval: 5 * time.Second,
		WriteHeaders:   true,
		DevMode:        true,
	})

	userLoginRatelimiter := ratelimiter.NewRateLimiter(&ratelimiter.Options{
		Name:           "user_login",
		Datastore:      ratelimiter.NewRedisDatastore(client),
		LimitPerMinute: 10,
		WindowInterval: 1 * time.Minute,
		BucketInterval: 5 * time.Second,
		WriteHeaders:   true,
		DevMode:        false,
	})

	s := &Mux{
		Router:                  router,
		userService:             service,
		userRegisterRatelimiter: userRegisterRatelimiter,
		userLoginRatelimiter:    userLoginRatelimiter,
		jwtWrapper:              jwtWrapper,
		logger:                  logger,
	}

	subRouter := router.NewRoute().Subrouter()
	// subRouter.Use(requestsLoggerMiddleware(logger, gip))
	subRouter.Use(metricsMiddleware(mtr))
	subRouter.Use(contextMiddleWare)
	subRouter.Handle("/users", s.wrap(s.GetUsers)).Methods("GET")
	subRouter.Handle("/login", s.wrap(s.Login)).Methods("POST")
	subRouter.Handle("/register", s.wrap(s.CreateUser)).Methods("POST")
	subRouter.Handle("/test", s.withAuthentication(s.test, "asd", "ASD", "ASDASD")).Methods("GET")

	return s
}

type apiFunc func(ctx context.Context, response *Response, request *http.Request) error

func (s *Mux) test(ctx context.Context, response *Response, request *http.Request) error {
	response.WriteJSON(map[string]string{"asdasd": "asdasd"})
	return nil
}

func (s *Mux) withAuthentication(handler apiFunc, roles ...string) http.Handler {
	return AuthenticationMiddleware(s.wrap(handler), s.jwtWrapper, roles...)
}

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
	})
}
