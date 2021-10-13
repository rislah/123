package api

import (
	"context"
	"fmt"
	"github.com/cep21/circuit/v3"
	"github.com/rislah/fakes/internal/encoder"
	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/metrics"
	"github.com/uber-go/tally"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/logger"
	"github.com/rislah/fakes/internal/redis"
	"github.com/rislah/fakes/internal/throttler"
)

type Server struct {
	*mux.Router
	userService      *app.Service
	requestThrottler throttler.Throttler
	logger           *logger.Logger
}

func NewServer(service *app.Service, client redis.Client, logger *logger.Logger, reporter tally.CachedStatsReporter) *Server {
	router := mux.NewRouter()

	s := &Server{
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

	m := metrics.New(reporter)
	//m.NewCounter("http_status", metrics.Tags{Key: "status", Value: "5xx"})
	//m.NewCounter("http_status", metrics.Tags{Key: "status", Value: "4xx"})
	//m.NewCounter("http_status", metrics.Tags{Key: "status", Value: "3xx"})
	//m.NewCounter("http_status", metrics.Tags{Key: "status", Value: "2xx"})

	router.Use(contextMiddleWare)
	router.Use(requestThrottler(s.requestThrottler, s.logger))
	router.Use(metricsMiddleware(m))

	router.HandleFunc("/users", s.GetUsers).Methods("GET")
	// router.Handle("/users", s.CreateUser).Methods("POST")

	return s
}

type apiFunc func(ctx context.Context, req *http.Request) (interface{}, error)

func contextMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipPort := strings.Split(r.RemoteAddr, ":")
		ip := net.ParseIP(ipPort[0])
		ctx := context.WithValue(r.Context(), "remote_ip", ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requestThrottler(th throttler.Throttler, l *logger.Logger) func(handler http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.RequestURI == "/metrics" {
				h.ServeHTTP(w, r)
			}

			ctx := r.Context()
			ip := ctx.Value("remote_ip").(net.IP)

			keys := []throttler.ID{
				{
					Key:  ip.String(),
					Type: "ip",
				},
			}

			shouldThrottle, err := th.Incr(ctx, keys)
			if err != nil {
				unwrappedErr := errors.Unwrap(err)
				if e, ok := unwrappedErr.(circuit.Error); ok {
					fmt.Println(e)
					encoder.ServeJSON(w, encoder.ErrInternalServerError, http.StatusInternalServerError)
					return
				}
				l.LogError(err, "cant update throttler")
				shouldThrottle = false
			}

			if shouldThrottle {
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
