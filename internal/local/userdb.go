package local

import (
	"context"
	"database/sql"
	"fmt"

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
		fmt.Println("????????????????????????????????????????")
		return errors.New("unique constraint error")
	}

	ld.users = append(ld.users, user)
	return nil
}

func (ld *localDB) GetUserByUsername(ctx context.Context, username string) (app.User, error) {
	for _, value := range ld.users {
		if value.Username == username {
			return value, nil
		}
	}

	return app.User{}, sql.ErrNoRows
}

func (ld *localDB) GetUsers(ctx context.Context) ([]app.User, error) {
	return ld.users, nil
}

func (ld *localDB) flushAll() error {
	ld.users = ld.users[:0]
	return nil
}
