package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/circuitbreaker"
	"github.com/rislah/fakes/internal/geoip"
	"github.com/rislah/fakes/internal/jwt"
	"github.com/rislah/fakes/internal/postgres"

	"github.com/rislah/fakes/api"
	"github.com/rislah/fakes/internal/metrics"
	"github.com/rislah/fakes/internal/redis"
	"github.com/segmentio/stats/v4"
	prom "github.com/segmentio/stats/v4/prometheus"

	"github.com/cep21/circuit/metrics/rolling"
	"github.com/cep21/circuit/v3"
	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/logger"
)

var (
	environment = "development"
	pgHost      = "127.0.0.1"
	pgPort      = "5432"
	pgUser      = "user"
	pgPass      = "parool"
	pgDB        = "user"
)

func main() {
	log := logger.New(environment)
	circuitManager := circuitbreaker.NewDefault()
	metrics := initMetrics(circuitManager)

	redis, err := redis.NewClient("localhost:6379", circuitManager, log)
	if err != nil {
		log.Fatal("starting redis", err)
	}

	geoIPDB, err := geoip.New("./GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatal("opening geoip database", err)
	}

	jwtWrapper := jwt.NewHS256Wrapper(app.JWTSecret)
	userDB, err := initUserDB(redis, circuitManager, metrics)
	authenticator := app.NewAuthenticator(userDB, jwtWrapper)
	userBackend := app.NewUserBackend(userDB, jwtWrapper)
	mux := api.NewMux(userBackend, authenticator, jwtWrapper, geoIPDB, redis, metrics, log)
	httpSrv := initHTTPServer(":8080", mux)

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGUSR2)

	log.Info(httpSrv.Addr)
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil {
			log.Fatal("starting server", err)
		}
	}()

	<-stopCh
	log.Info("stopping server")
}

func initHTTPServer(addr string, handler http.Handler) *http.Server {
	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return httpSrv
}

func initUserDB(rd redis.Client, cm *circuit.Manager, metrics metrics.Metrics) (app.UserDB, error) {
	switch environment {
	case "local":
		return local.NewUserDB(), nil
	case "development":
		opts := postgres.Options{
			ConnectionString: fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", pgHost, pgPort, pgUser, pgPass, pgDB),
			MaxIdleConns:     100,
			MaxOpenConns:     100,
		}

		client, err := postgres.NewClient(opts)
		if err != nil {
			return nil, err
		}

		cc := cm.MustCreateCircuit(
			"postgres_userdb",
			circuit.Config{
				Execution: circuit.ExecutionConfig{
					Timeout: 300 * time.Millisecond,
				},
			},
		)

		// return postgres.NewUserDB(client, cc)
		return postgres.NewCachedUserDB(client, rd, cc, metrics)

	default:
		panic("unknown environment")
	}

}

func initMetrics(cm *circuit.Manager) metrics.Metrics {
	statsEngine := stats.NewEngine("app", stats.DefaultEngine.Handler)
	statsEngine.Register(prom.DefaultHandler)

	mtr := metrics.New(statsEngine)
	scf := metrics.CommandFactory{Metrics: mtr}
	sf := rolling.StatFactory{}

	cm.DefaultCircuitProperties = append(cm.DefaultCircuitProperties, scf.CommandProperties, sf.CreateConfig)

	return mtr
}
