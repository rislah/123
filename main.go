package main

import (
	"net/http"
	"time"

	"github.com/rislah/fakes/api"
	"github.com/rislah/fakes/internal/metrics"
	"github.com/rislah/fakes/internal/redis"
	"github.com/segmentio/stats/v4"
	prom "github.com/segmentio/stats/v4/prometheus"

	"github.com/cep21/circuit/metrics/rolling"
	"github.com/cep21/circuit/v3"
	"github.com/cep21/circuit/v3/closers/hystrix"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/logger"
)

const (
	environment = "development"
)

func main() {
	log := logger.New(environment)

	var db app.UserDB

	switch environment {
	case "development":
		db = local.NewUserDB()
	}

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
	sf := rolling.StatFactory{}

	hystrixFactory := hystrix.Factory{
		CreateConfigureOpener: []func(circuitName string) hystrix.ConfigureOpener{openerFactory},
		CreateConfigureCloser: []func(circuitName string) hystrix.ConfigureCloser{closerFactory},
	}

	statsEngine := stats.NewEngine("mysvc", stats.DefaultEngine.Handler)
	statsEngine.Register(prom.DefaultHandler)

	mtr := metrics.New(statsEngine)
	scf := metrics.CommandFactory{Metrics: mtr}

	manager := circuit.Manager{
		DefaultCircuitProperties: []circuit.CommandPropertiesConstructor{
			scf.CommandProperties,
			sf.CreateConfig,
			hystrixFactory.Configure,
		},
	}

	rd, err := redis.NewClient("localhost:6379", &manager, &mtr, log)
	if err != nil {
		log.Fatal("starting redis", err, nil)
	}

	config := app.ServiceConfig{UserDB: db}
	service := app.NewUserService(config)
	srv := api.NewServer(service, rd, mtr, log)

	httpSrv := &http.Server{
		Addr:         ":8888",
		Handler:      srv,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Info("listening", nil)
	httpSrv.ListenAndServe()
}
