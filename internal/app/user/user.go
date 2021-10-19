package user

import (
	"context"

	app "github.com/rislah/fakes/internal"
)

type User interface {
	CreateUser(ctx context.Context, user app.User) error
	GetUsers(ctx context.Context) ([]app.User, error)
}

type userImpl struct {
	userDB app.UserDB
}

func NewUser(db app.UserDB) User {
	if db == nil {
		panic("database is required")
	}

	return &userImpl{userDB: db}
}
