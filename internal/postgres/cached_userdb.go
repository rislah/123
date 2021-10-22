package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/cep21/circuit/v3"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/metrics"
	"github.com/rislah/fakes/internal/redis"
)

type cacheKey string

func (c cacheKey) String() string {
	return string(c)
}

const (
	UsersKey cacheKey = "users"
)

type postgresCachedUserDB struct {
	userDB  *postgresUserDB
	redis   redis.Client
	metrics metrics.Metrics
}

var _ app.UserDB = &postgresCachedUserDB{}

func NewCachedUserDB(pg *sql.DB, rd redis.Client, cc *circuit.Circuit, metrics metrics.Metrics) (*postgresCachedUserDB, error) {
	pgUserDB := &postgresUserDB{pg: pg, circuit: cc}
	cdb := &postgresCachedUserDB{
		userDB:  pgUserDB,
		redis:   rd,
		metrics: metrics,
	}
	return cdb, nil
}

func (cdb *postgresCachedUserDB) CreateUser(ctx context.Context, user app.User) error {
	if err := cdb.userDB.CreateUser(ctx, user); err != nil {
		return err
	}
	if err := cdb.redis.Del(UsersKey.String()); err != nil {
		return err
	}
	return nil
}

func (cdb *postgresCachedUserDB) GetUsers(ctx context.Context) ([]app.User, error) {
	resp, err := cdb.redis.SMembers(ctx, UsersKey.String())
	if err != nil {
		return nil, err
	}

	users := []app.User{}
	if len(resp) > 0 {
		for _, r := range resp {
			user := app.User{}
			err := json.Unmarshal([]byte(r), &user)
			if err != nil {
				return nil, err
			}

			users = append(users, user)
		}

		cdb.metrics.Inc("cachedb_hit")

		return users, nil
	}

	users, err = cdb.userDB.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	marshalledUsers := []interface{}{}
	if len(users) > 0 {
		for _, usr := range users {
			b, err := json.Marshal(usr.Sanitize())
			if err != nil {
				return nil, err
			}

			marshalledUsers = append(marshalledUsers, b)

		}
	}

	if err := cdb.redis.SAdd(ctx, UsersKey.String(), marshalledUsers...); err != nil {
		return nil, err
	}

	cdb.metrics.Inc("cachedb_miss")

	return users, nil
}

func (cdb *postgresCachedUserDB) GetUserByUsername(ctx context.Context, username string) (app.User, error) {
	return cdb.userDB.GetUserByUsername(ctx, username)
}
