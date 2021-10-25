package app

import "context"

func (u *userImpl) GetUserRolesByUserIDs(ctx context.Context, userIDs []string) ([]Role, error) {
	roles, err := u.userDB.GetUserRolesByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	return roles, nil
}
