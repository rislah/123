package user

import (
	"context"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/jwt"
)

const JWTSecret = "secret"

type User interface {
	CreateUser(ctx context.Context, user app.User) error
	GetUsers(ctx context.Context) ([]app.User, error)
	Login(ctx context.Context, username, password string) (string, error)
}

type userImpl struct {
	userDB app.UserDB
	jwt    jwt.Wrapper
}

func NewUser(db app.UserDB, jwt jwt.Wrapper) User {
	if db == nil {
		panic("database is required")
	}

	return &userImpl{
		userDB: db,
		jwt:    jwt,
	}
}

var (
	ErrUserNotFound = &errors.WrappedError{
		Code: errors.ErrNotFound,
		Msg:  "User not found",
	}
	ErrLoginBadCredentials = &errors.WrappedError{
		Code: errors.ErrBadRequest,
		Msg:  "Username or password is incorrect",
	}
)
