package local

import (
	"context"

	"github.com/pkg/errors"
	app "github.com/rislah/fakes/internal"
)

type localDB struct {
	users []app.User
}

func NewUserDB() *localDB {
	return &localDB{}
}

func MakeUserDB() (app.UserDB, func() error, error) {
	db := NewUserDB()
	return NewUserDB(), db.flushAll, nil
}

var _ app.UserDB = &localDB{}

func (ld *localDB) CreateUser(ctx context.Context, user app.User) error {
	var found bool
	for _, value := range ld.users {
		if value.Username == user.Username {
			found = true
			break
		}
	}

	if found {
		return errors.New("unique constraint error")
	}

	usr := user
	if usr.Role == "" {
		usr.Role = "guest"
	}

	ld.users = append(ld.users, usr)
	return nil
}

func (ld *localDB) GetUserByUsername(ctx context.Context, username string) (app.User, error) {
	for _, value := range ld.users {
		if value.Username == username {
			return value, nil
		}
	}

	return app.User{}, nil
}

func (ld *localDB) GetUsers(ctx context.Context) ([]app.User, error) {
	return ld.users, nil
}

func (ld *localDB) flushAll() error {
	ld.users = ld.users[:0]
	return nil
}
