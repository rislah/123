package app

import (
	"context"

	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/jwt"
)

const JWTSecret = "secret"

type UserBackend interface {
	CreateUser(ctx context.Context, user User) error
	GetUsers(ctx context.Context) ([]User, error)
}

type UserDB interface {
	CreateUser(ctx context.Context, user User) error
	GetUsers(ctx context.Context) ([]User, error)
	GetUserByUsername(ctx context.Context, username string) (User, error)
}

type User struct {
	Username string
	Password string
	Role     string
}

type userImpl struct {
	userDB     UserDB
	jwtWrapper jwt.Wrapper
}

func NewUserBackend(db UserDB, jwtWrapper jwt.Wrapper) UserBackend {
	if db == nil {
		panic("database is required")
	}

	return &userImpl{
		userDB:     db,
		jwtWrapper: jwtWrapper,
	}
}

var (
	ErrUserNotFound = &errors.WrappedError{
		Code: errors.ErrNotFound,
		Msg:  "User not found",
	}
)
