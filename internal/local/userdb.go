package local

import (
	"context"
	"github.com/pkg/errors"
	app "github.com/rislah/fakes/internal"
	"sync"
)

type localDB struct {
	users []app.User
	mu    sync.RWMutex
}

func NewUserDB() *localDB {
	return &localDB{}
}

func MakeUserDB() (app.UserDB, error) {
	return NewUserDB(), nil
}

var _ app.UserDB = &localDB{}

func (ld *localDB) CreateUser(ctx context.Context, user app.User) error {
	var found bool

	ld.mu.RLock()

	for _, value := range ld.users {
		if value.Firstname == user.Firstname && value.Lastname == user.Lastname {
			found = true
			break
		}
	}

	ld.mu.RUnlock()

	if found {
		return errors.New("unique constraint error")
	}

	ld.mu.Lock()
	defer ld.mu.Unlock()

	ld.users = append(ld.users, user)

	return nil
}

func (ld *localDB) GetUsers(ctx context.Context) ([]app.User, error) {
	ld.mu.Lock()
	defer ld.mu.Unlock()
	return ld.users, nil
}
