package postgres

import (
	"context"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/redis"
)

type postgresCachedUserDB struct {
	db    *postgresUserDB
	redis redis.Client
}

// var _ app.UserDB = &postgresCachedUserDB{}

func NewCachedUserDB(rd redis.Client, opts Options) (*postgresCachedUserDB, error) {
	db, err := NewUserDB(opts)
	if err != nil {
		return nil, err
	}

	cdb := &postgresCachedUserDB{
		db:    db,
		redis: rd,
	}

	return cdb, nil
}

func (cdb *postgresCachedUserDB) CreateUser(ctx context.Context, user app.User) error {
	return cdb.CreateUser(ctx, user)
}

func (cdb *postgresCachedUserDB) GetUsers(ctx context.Context) ([]app.User, error) {
	// var users []app.User
	// resp, err :+ cdb.redis.Get("users")

	// json.Unmarshal([]byte(resp), u)
	return cdb.GetUsers(ctx)
}
