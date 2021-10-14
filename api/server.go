package api

import (
	"time"

	"github.com/rislah/fakes/internal/geoip"
	"github.com/rislah/fakes/internal/metrics"
	"github.com/segmentio/stats/v4/prometheus"

	"github.com/gorilla/mux"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/logger"
	"github.com/rislah/fakes/internal/redis"
	"github.com/rislah/fakes/internal/throttler"
)

type Server struct {
	*mux.Router
	userService      app.Service
	requestThrottler throttler.Throttler
	logger           *logger.Logger
}

func NewServer(service app.Service, client redis.Client, mtr metrics.Metrics, logger *logger.Logger) *Server {
	router := mux.NewRouter()
	router.Handle("/metrics", prometheus.DefaultHandler).Methods("GET")

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

	gip, err := geoip.New("./GeoLite2-Country.mmdb")
	if err != nil {
		logger.Fatal("Couldn't open GeoIP database", err, nil)
	}

	subRouter := router.NewRoute().Subrouter()
	subRouter.Use(requestsLoggerMiddleware(logger, gip))
	subRouter.Use(metricsMiddleware(mtr))
	subRouter.Use(contextMiddleWare())
	subRouter.HandleFunc("/users", s.GetUsers).Methods("GET")

	return s
}
