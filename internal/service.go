package app

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	errorsx "github.com/rislah/fakes/internal/errors"
)

var (
	ErrUsersNotFound = &errorsx.WrappedError{
		StatusCode: http.StatusNotFound,
		CodeValue:  "users_not_found",
		Message:    "Users not found",
	}
)

type ServiceConfig struct {
	UserDB UserDB
}

type Service struct {
	userDB UserDB
}

func NewUserService(config ServiceConfig) Service {
	if config.UserDB == nil {
		panic("database is required")
	}

	s := Service{
		userDB: config.UserDB,
	}

	return s
}

func (s *Service) CreateUser(ctx context.Context, user User) error {
	return s.userDB.CreateUser(ctx, user)
}

func (s *Service) GetUsers(ctx context.Context) ([]User, error) {
	users, err := s.userDB.GetUsers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getUsers")
	}

	if len(users) == 0 {
		return nil, ErrUsersNotFound
	}

	return users, nil
}
