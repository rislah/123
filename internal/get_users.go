package app

import (
	"context"
)

func (u userImpl) GetUsers(ctx context.Context) ([]User, error) {
	users, err := u.userDB.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, ErrUsersNotFound
	}

	return users, nil
}
