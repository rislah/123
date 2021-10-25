package app

import "context"

func (u *userImpl) GetUserByUsername(ctx context.Context, username string) (User, error) {
	user, err := u.userDB.GetUserByUsername(ctx, username)
	if err != nil {
		return User{}, err
	}

	if user.IsEmpty() {
		return User{}, ErrUserNotFound
	}

	return user, nil
}
