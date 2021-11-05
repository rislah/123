package postgres

import (
	"context"
	"encoding/json"

	"github.com/cep21/circuit/v3"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/redis"
)

var (
	cacheHit = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_hit_total",
	}, []string{"method"})

	cacheMiss = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_miss_total",
	}, []string{"method"})
)

func init() {
	prometheus.Register(cacheHit)
	prometheus.Register(cacheMiss)
}

type cacheKey string

func (c cacheKey) String() string {
	return string(c)
}

const (
	UsersKey cacheKey = "users"
)

type postgresCachedUserDB struct {
	userDB *postgresUserDB
	redis  redis.Client
}

// var _ app.UserDB = &postgresCachedUserDB{}

func NewCachedUserDB(pg *sqlx.DB, rd redis.Client, cc *circuit.Circuit) (*postgresCachedUserDB, error) {
	pgUserDB := &postgresUserDB{pg: pg, circuit: cc}
	cdb := &postgresCachedUserDB{
		userDB: pgUserDB,
		redis:  rd,
	}
	return cdb, nil
}

func (cdb *postgresCachedUserDB) CreateUser(ctx context.Context, user app.User) error {
	if err := cdb.userDB.CreateUser(ctx, user); err != nil {
		return errors.New(err)
	}
	if err := cdb.redis.Del(UsersKey.String()); err != nil {
		return errors.New(err)
	}
	return nil
}

func (cdb *postgresCachedUserDB) GetUsers(ctx context.Context) ([]app.User, error) {
	resp, err := cdb.redis.SMembers(ctx, UsersKey.String())
	if err != nil {
		return nil, errors.New(err)
	}

	if len(resp) > 0 {
		users := []app.User{}
		for _, r := range resp {
			user := app.User{}
			err := json.Unmarshal([]byte(r), &user)
			if err != nil {
				return nil, errors.New(err)
			}
			users = append(users, user)
		}

		cacheHit.WithLabelValues("getUsers").Inc()

		return users, nil
	}

	users, err := cdb.userDB.GetUsers(ctx)
	if err != nil {
		return nil, errors.New(err)
	}

	if len(users) > 0 {
		marshalledUsers := []interface{}{}
		if err != nil {
			return nil, err
		}

		for _, usr := range users {
			b, err := json.Marshal(usr.Sanitize())
			if err != nil {
				return nil, errors.New(err)
			}
			marshalledUsers = append(marshalledUsers, b)
		}

		if err := cdb.redis.SAdd(ctx, UsersKey.String(), marshalledUsers...); err != nil {
			return nil, errors.New(err)
		}
	}

	cacheMiss.WithLabelValues("getUsers").Inc()
	return users, nil
}

func (cdb *postgresCachedUserDB) GetUserByUsername(ctx context.Context, username string) (app.User, error) {
	return cdb.userDB.GetUserByUsername(ctx, username)
}

func (cdb *postgresCachedUserDB) GetUsersByIDs(ctx context.Context, userIDs []string) ([]app.User, error) {
	return cdb.userDB.GetUsersByIDs(ctx, userIDs)
}

func (cdb *postgresCachedUserDB) GetUserRoleByUserID(ctx context.Context, userID string) (app.Role, error) {
	return cdb.userDB.GetUserRoleByUserID(ctx, userID)
}

func (cdb *postgresCachedUserDB) GetUserRolesByUserIDs(ctx context.Context, userIDs []string) ([]*app.Role, error) {
	return cdb.userDB.GetUserRolesByUserIDs(ctx, userIDs)
}

func (cdb *postgresCachedUserDB) GetUsersByRoleID(ctx context.Context, roleID int) ([]app.User, error) {
	return cdb.userDB.GetUsersByRoleID(ctx, roleID)
}

func (cdb *postgresCachedUserDB) GetUsersByRoleIDs(ctx context.Context, roleIDs []int) ([]app.UserRole, error) {
	return cdb.userDB.GetUsersByRoleIDs(ctx, roleIDs)
}
