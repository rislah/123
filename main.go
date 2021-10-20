package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/geoip"
	"github.com/rislah/fakes/internal/jwt"
	"github.com/rislah/fakes/internal/postgres"

	jwtPkg "github.com/golang-jwt/jwt/v4"
	"github.com/rislah/fakes/api"
	"github.com/rislah/fakes/internal/metrics"
	"github.com/rislah/fakes/internal/redis"
	"github.com/segmentio/stats/v4"
	prom "github.com/segmentio/stats/v4/prometheus"

	"github.com/cep21/circuit/metrics/rolling"
	"github.com/cep21/circuit/v3"
	"github.com/cep21/circuit/v3/closers/hystrix"
	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/logger"
)

const (
	environment = "development"
)

func main() {
	var (
		err error
		db  app.UserDB
	)

	log := logger.New(environment)

	switch environment {
	case "development":
		db = local.NewUserDB()
	case "production":
		opts := postgres.Options{
			ConnectionString: "",
			MaxIdleConns:     100,
			MaxOpenConns:     100,
		}
		db, err = postgres.NewUserDB(opts)
		if err != nil {
			log.Fatal("opening postgres database", err)
		}
	}

	statsEngine := stats.NewEngine("app", stats.DefaultEngine.Handler)
	statsEngine.Register(prom.DefaultHandler)

	mtr := metrics.New(statsEngine)
	scf := metrics.CommandFactory{Metrics: mtr}
	sf := rolling.StatFactory{}

	manager := makeCircuitBreaker()
	manager.DefaultCircuitProperties = append(manager.DefaultCircuitProperties, scf.CommandProperties, sf.CreateConfig)

	rd, err := redis.NewClient("localhost:6379", &manager, log)
	if err != nil {
		log.Fatal("starting redis", err)
	}

	gip, err := geoip.New("./GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatal("opening geoip database", err)
	}

	jw := jwt.Wrapper{
		Algorithm: jwtPkg.SigningMethodHS256,
		Secret:    app.JWTSecret,
	}

	authenticator := app.NewAuthenticator(db, jw)
	userBackend := app.NewUserBackend(db, jw)
	mux := api.NewMux(userBackend, authenticator, jw, gip, rd, mtr, log)
	httpSrv := makeHTTPServer(":8080", mux)

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

func makeCircuitBreaker() circuit.Manager {
	openerFactory := func(circuitName string) hystrix.ConfigureOpener {
		return hystrix.ConfigureOpener{
			ErrorThresholdPercentage: 50,
			RequestVolumeThreshold:   10,
			RollingDuration:          10 * time.Second,
			NumBuckets:               10,
		}
	}

	closerFactory := func(circuitName string) hystrix.ConfigureCloser {
		return hystrix.ConfigureCloser{
			SleepWindow:                  5 * time.Second,
			HalfOpenAttempts:             1,
			RequiredConcurrentSuccessful: 1,
		}
	}

	hystrixFactory := hystrix.Factory{
		CreateConfigureOpener: []func(circuitName string) hystrix.ConfigureOpener{openerFactory},
		CreateConfigureCloser: []func(circuitName string) hystrix.ConfigureCloser{closerFactory},
	}

	manager := circuit.Manager{
		DefaultCircuitProperties: []circuit.CommandPropertiesConstructor{
			hystrixFactory.Configure,
		},
	}

	return manager
}

func makeHTTPServer(addr string, handler http.Handler) *http.Server {
	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return httpSrv
}
