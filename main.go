package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/circuitbreaker"
	"github.com/rislah/fakes/internal/geoip"
	"github.com/rislah/fakes/internal/jwt"
	"github.com/rislah/fakes/internal/postgres"

	"github.com/rislah/fakes/api"
	"github.com/rislah/fakes/internal/redis"

	"github.com/cep21/circuit/v3"
	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/logger"
)

type config struct {
	ListenAddr  string `default:":8080"`
	Environment string `default:"development"`
	PgHost      string `default:"127.0.0.1"`
	PgPort      string `default:"5432"`
	PgUser      string `default:"user"`
	PgPass      string `default:"parool"`
	PgDB        string `default:"user"`
	RedisHost   string `default:"localhost"`
	RedisPort   string `default:"6379"`
}

func main() {
	var conf config
	err := envconfig.Process("fakes", &conf)
	if err != nil {
		log.Fatal(err)
	}

	log := logger.New(conf.Environment)
	geoIPDB := initGeoIPDB("./GeoLite2-Country.mmdb")
	jwtWrapper := jwt.NewHS256Wrapper(app.JWTSecret)
	userDB := initUserDB(conf, log)
	authenticator := app.NewAuthenticator(userDB, jwtWrapper)
	userBackend := app.NewUserBackend(userDB, jwtWrapper)
	ratelimiterRedisCB, err := circuitbreaker.New("redis_ratelimiter", circuitbreaker.Config{})
	if err != nil {
		log.Fatal("error creating rate limiter cb", err)
	}
	ratelimiterRedis := initRedis(conf, ratelimiterRedisCB, log)
	mux := api.NewMux(userBackend, authenticator, jwtWrapper, geoIPDB, ratelimiterRedis, log)
	httpSrv := initHTTPServer(conf.ListenAddr, mux)

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGUSR2)
	log.Info("Listening on addr " + httpSrv.Addr)
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

func initUserDB(conf config, log *logger.Logger) app.UserDB {
	switch conf.Environment {
	case "local":
		return local.NewUserDB()
	case "development":
		opts := postgres.Options{
			ConnectionString: fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", conf.PgHost, conf.PgPort, conf.PgUser, conf.PgPass, conf.PgDB),
			MaxIdleConns:     100,
			MaxOpenConns:     100,
		}

		client, err := postgres.NewClient(opts)
		if err != nil {
			log.Fatal("init postgres client", err)
		}

		userDBCircuit, err := circuitbreaker.New("postgres_userdb", circuitbreaker.Config{})
		if err != nil {
			log.Fatal("error creating userdb circuit", err)
		}

		redisCircuit, err := circuitbreaker.New("redis_cache_userdb", circuitbreaker.Config{})
		if err != nil {
			log.Fatal("error creating redis cache circuit", err)
		}

		rd := initRedis(conf, redisCircuit, log)
		db, err := postgres.NewCachedUserDB(client, rd, userDBCircuit)
		if err != nil {
			log.Fatal("init cached userdb", err)
		}

		return db
	default:
		panic("unknown environment")
	}

}

// func initMetrics(cm *circuit.Manager) metrics.Metrics {
// 	statsEngine := stats.NewEngine("app", stats.DefaultEngine.Handler)
// 	statsEngine.Register(prom.DefaultHandler)

// 	mtr := metrics.New(statsEngine)
// 	scf := metrics.NewCommandFactory(mtr)
// 	sf := rolling.StatFactory{}

// 	cm.DefaultCircuitProperties = append(cm.DefaultCircuitProperties, scf.CommandProperties, sf.CreateConfig)

// 	return mtr
// }

func initRedis(conf config, cb *circuit.Circuit, log *logger.Logger) redis.Client {
	redis, err := redis.NewClient(fmt.Sprintf("%s:%s", conf.RedisHost, conf.RedisPort), cb, log)
	if err != nil {
		log.Fatal("init redis", err)
	}
	return redis
}

func initGeoIPDB(filePath string) geoip.GeoIP {
	geoIPDB, err := geoip.New(filePath)
	if err != nil {
		log.Fatal("opening geoip database", err)
	}
	return geoIPDB
}
