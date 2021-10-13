package main

import (
	"github.com/cep21/circuit/v3"
	"github.com/cep21/circuit/v3/closers/hystrix"
	"github.com/rislah/fakes/api"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/logger"
	"github.com/rislah/fakes/internal/redis"
	promreporter "github.com/uber-go/tally/prometheus"
	"go.uber.org/zap"
	"net/http"
	"time"
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
		//case "production":

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

	hystrixFactory := hystrix.Factory{
		CreateConfigureOpener: []func(circuitName string) hystrix.ConfigureOpener{openerFactory},
		CreateConfigureCloser: []func(circuitName string) hystrix.ConfigureCloser{closerFactory},
	}

	manager := circuit.Manager{
		DefaultCircuitProperties: []circuit.CommandPropertiesConstructor{
			hystrixFactory.Configure,
		},
	}

	rd, err := redis.NewClient("localhost:6379", &manager, log)
	if err != nil {
		log.Fatal("starting redis", zap.Error(err))
	}

	config := app.ServiceConfig{UserDB: db}
	r := promreporter.NewReporter(promreporter.Options{})
	service := app.NewUserService(config)
	srv := api.NewServer(service, rd, log, r)

	srv.Handle("/metrics", r.HTTPHandler()).Methods("GET")

	log.Info("listening")
	http.ListenAndServe(":8888", srv)
}
