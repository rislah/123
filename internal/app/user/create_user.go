package user

import (
	"context"

	app "github.com/rislah/fakes/internal"
)

func (u *userImpl) CreateUser(ctx context.Context, user app.User) error {
	err := u.userDB.CreateUser(ctx, user)
	if err != nil {
		return err
	}

	return nil
}
