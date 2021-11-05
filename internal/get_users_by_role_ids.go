package app

import (
	"context"
)

func (u *userImpl) GetUsersByRoleIDs(ctx context.Context, roleIDs []int) (map[int][]User, error) {
	userRoles, err := u.userDB.GetUsersByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}

	m := map[int][]User{}
	for _, userRole := range userRoles {
		m[userRole.Role.ID] = append(m[userRole.Role.ID], userRole.User)
	}

	return m, nil
}
