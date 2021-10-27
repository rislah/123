package api

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/jwt"
	"github.com/rislah/fakes/internal/ratelimiter"

	"github.com/rislah/fakes/internal/geoip"

	"github.com/gorilla/mux"
	"github.com/rislah/fakes/internal/logger"
	"github.com/rislah/fakes/internal/redis"
)

type Mux struct {
	*mux.Router
	userBackend             app.UserBackend
	authenticator           app.Authenticator
	userRegisterRatelimiter *ratelimiter.Ratelimiter
	userLoginRatelimiter    *ratelimiter.Ratelimiter
	globalRatelimiter       *ratelimiter.Ratelimiter
	jwtWrapper              jwt.Wrapper
	logger                  *logger.Logger
}

func NewMux(userBackend app.UserBackend, authenticator app.Authenticator, jwtWrapper jwt.Wrapper, gip geoip.GeoIP, client redis.Client, logger *logger.Logger) *Mux {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

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
		DevMode:        true,
	})

	globalRateLimiter := ratelimiter.NewRateLimiter(&ratelimiter.Options{
		Name:           "global",
		Datastore:      ratelimiter.NewRedisDatastore(client),
		LimitPerMinute: 50000,
		WindowInterval: 1 * time.Minute,
		BucketInterval: 5 * time.Second,
		WriteHeaders:   true,
		DevMode:        true,
	})

	s := &Mux{
		Router:                  router,
		userBackend:             userBackend,
		authenticator:           authenticator,
		userRegisterRatelimiter: userRegisterRatelimiter,
		userLoginRatelimiter:    userLoginRatelimiter,
		globalRatelimiter:       globalRateLimiter,
		jwtWrapper:              jwtWrapper,
		logger:                  logger,
	}

	subRouter := router.NewRoute().Subrouter()
	//subRouter.Use(requestsLoggerMiddleware(logger, gip))
	subRouter.Use(metricsMiddleware)
	subRouter.Use(contextMiddleWare)
	subRouter.Use(s.ratelimiterMiddleware)

	routeModule := NewRouteModule(jwtWrapper)
	routeModule.Get("/testauth", s.test).Permissions("viewTest")
	// routeModule.Get("/users", s.GetUsers)
	routeModule.Post("/register", s.CreateUser)
	routeModule.Post("/login", s.Login)
	routeModule.InjectRoutes(subRouter)

	return s
}

func (s *Mux) test(ctx context.Context, response *Response, request *http.Request) error {
	response.WriteJSON(map[string]string{"asdasd": "asdasd"})
	return nil
}

func (s *Mux) wrap(handler ApiFunc) http.Handler {
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
