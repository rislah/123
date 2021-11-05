package app

import "context"

func (u *userImpl) GetUserByUsername(ctx context.Context, username string) (User, error) {
	if username == "" {
		return User{}, nil
	}

	return u.userDB.GetUserByUsername(ctx, username)
}
