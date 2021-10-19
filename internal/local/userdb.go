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
		if value.Firstname == user.Firstname && value.Lastname == user.Lastname {
			found = true
			break
		}
	}

	if found {
		return errors.New("unique constraint error")
	}

	ld.users = append(ld.users, user)
	return nil
}

func (ld *localDB) GetUsers(ctx context.Context) ([]app.User, error) {
	return ld.users, nil
}

func (ld *localDB) flushAll() error {
	ld.users = ld.users[:0]
	return nil
}
