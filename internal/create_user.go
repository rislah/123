package app

import (
	"context"
)

func (u *userImpl) CreateUser(ctx context.Context, user User) error {
	err := u.userDB.CreateUser(ctx, user)
	if err != nil {
		return err
	}

	return nil
}
