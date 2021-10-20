package app

import (
	"context"

	"github.com/rislah/fakes/internal/errors"
)

func (u userImpl) GetUsers(ctx context.Context) ([]User, error) {
	users, err := u.userDB.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, &errors.WrappedError{
			Code: errors.ErrBadRequest,
			Msg:  "No users found",
		}
	}

	return users, nil
}
