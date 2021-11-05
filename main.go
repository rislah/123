package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	"github.com/graph-gophers/graphql-go"
	"github.com/jmoiron/sqlx"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rislah/fakes/gql"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/circuitbreaker"
	"github.com/rislah/fakes/internal/geoip"
	"github.com/rislah/fakes/internal/jwt"
	"github.com/rislah/fakes/internal/postgres"
	"github.com/rislah/fakes/loaders"

	"github.com/rislah/fakes/schema"

	"github.com/rislah/fakes/internal/redis"
	"github.com/rislah/fakes/resolvers"

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
	if err := envconfig.Process("fakes", &conf); err != nil {
		log.Fatal(err)
	}

	opts := postgres.Options{
		ConnectionString: fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable binary_parameters=yes", conf.PgHost, conf.PgPort, conf.PgUser, conf.PgPass, conf.PgDB),
		MaxIdleConns:     100,
		MaxOpenConns:     100,
	}

	client, err := postgres.NewClient(opts)
	if err != nil {
		log.Fatal("init postgres client", err)
	}

	log := logger.New(conf.Environment)
	// geoIPDB := initGeoIPDB("./GeoLite2-Country.mmdb")
	jwtWrapper := jwt.NewHS256Wrapper(app.JWTSecret)
	userDB := initUserDB(conf, client, log)
	roleDB := initRoleDB(conf, client, log)
	authenticator := app.NewAuthenticator(userDB, roleDB, jwtWrapper)
	userBackend := app.NewUserBackend(userDB, jwtWrapper)
	roleBackend := app.NewRoleBackend(roleDB)

	backend := &app.Backend{
		Authenticator: authenticator,
		User:          userBackend,
		Role:          roleBackend,
	}

	rootResolver := resolvers.NewRootResolver(backend)
	schemaStr, err := schema.String()
	if err != nil {
		log.Fatal("schema", err)
	}
	schema := graphql.MustParseSchema(schemaStr, rootResolver)
	_ = schema

	dl := loaders.New(backend)

	m := mux.NewRouter()
	m.Handle("/metrics", promhttp.Handler())

	runtime.SetMutexProfileFraction(10)
	runtime.SetBlockProfileRate(int(time.Second.Nanoseconds()))
	go func() {
		http.ListenAndServe(":6060", nil)
	}()

	ms := m.NewRoute().Subrouter()
	ms.Use(dl.AttachMiddleware)
	ms.Handle("/query", &gql.Handler{Schema: schema}).Methods("POST")

	srv := initHTTPServer(conf.ListenAddr, m)
	srv.ListenAndServe()
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

func initUserDB(conf config, client *sqlx.DB, log *logger.Logger) app.UserDB {
	switch conf.Environment {
	case "local":
		return local.NewUserDB()
	case "development":
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

func initRoleDB(conf config, client *sqlx.DB, log *logger.Logger) app.RoleDB {
	roleDBCircuit, err := circuitbreaker.New("postgres_roledb", circuitbreaker.Config{})
	if err != nil {
		log.Fatal("roledb circuit", err)
	}

	db := postgres.NewRoleDB(client, roleDBCircuit)

	return db
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
