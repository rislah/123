package app

import (
	"context"

	"github.com/rislah/fakes/internal/credentials"

	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/jwt"
)

const JWTSecret = "secret"

type UserBackend interface {
	CreateUser(ctx context.Context, creds credentials.Credentials) error
	GetUsers(ctx context.Context) ([]User, error)
}

type UserDB interface {
	CreateUser(ctx context.Context, user User) error
	GetUsers(ctx context.Context) ([]User, error)
	GetUserByUsername(ctx context.Context, username string) (User, error)
}

type User struct {
	UserID   string
	Username string
	Password string
	Role     string
}

func (u User) IsEmpty() bool {
	return u.Username == "" || u.Role == "" || u.Password == ""
}

func (u *User) Sanitize() *User {
	u.Password = ""
	return u
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
	ErrUsersNotFound = &errors.WrappedError{
		Code: errors.ErrNotFound,
		Msg:  "Users not found",
	}
	ErrUserAlreadyExists = &errors.WrappedError{
		Code: errors.ErrConflict,
		Msg:  "User already exists",
	}
)
