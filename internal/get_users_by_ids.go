package app

import "context"

func (u *userImpl) GetUsersByIDs(ctx context.Context, userIDs []string) ([]User, error) {
	users, err := u.userDB.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, ErrUsersNotFound
	}

	return users, nil
}
