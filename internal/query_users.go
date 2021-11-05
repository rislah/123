package app

import "context"

func (u *userImpl) GetUsers(ctx context.Context, args UsersQueryArgs) ([]User, error) {
	if args.UserIDs != nil {
		return u.userDB.GetUsersByIDs(ctx, args.UserIDs)
	}

	if args.RoleID != 0 {
		return u.userDB.GetUsersByRoleID(ctx, args.RoleID)
	}

	return u.userDB.GetUsers(ctx)
}
