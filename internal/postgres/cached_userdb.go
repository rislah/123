package postgres

import (
	"context"
	"database/sql"

	"github.com/cep21/circuit/v3"
	app "github.com/rislah/fakes/internal"
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
	db    *postgresUserDB
	redis redis.Client
}

// var _ app.UserDB = &postgresCachedUserDB{}

func NewCachedUserDB(pg *sql.DB, rd redis.Client, cc *circuit.Circuit) (*postgresCachedUserDB, error) {
	pgUserDB := &postgresUserDB{pg: pg, circuit: cc}
	cdb := &postgresCachedUserDB{
		db:    pgUserDB,
		redis: rd,
	}
	return cdb, nil
}

func (cdb *postgresCachedUserDB) CreateUser(ctx context.Context, user app.User) error {
	if err := cdb.CreateUser(ctx, user); err != nil {
		return err
	}
	if err := cdb.redis.Del(UsersKey.String()); err != nil {
		return err
	}
	return nil
}

func (cdb *postgresCachedUserDB) GetUsers(ctx context.Context) ([]app.User, error) {
	// var users []app.User
	// resp, err :+ cdb.redis.Get("users")

	// json.Unmarshal([]byte(resp), u)
	return cdb.GetUsers(ctx)
}
