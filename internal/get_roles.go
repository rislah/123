package app

import "context"

func (r *roleImpl) GetRoles(ctx context.Context, args RolesQueryArgs) ([]Role, error) {
	var (
		ids             []int    = args.IDs
		names           []string = args.Names
		userIDs         []string = args.UserIDs
		userRoleUserIDs []string = args.UserRoleUserIDs
	)

	if ids != nil {
		return r.roleDB.GetRolesByIDs(ctx, ids)
	}

	if names != nil {
		return r.roleDB.GetRolesByNames(ctx, names)
	}

	if userIDs != nil {
		return r.roleDB.GetRolesByUserIDs(ctx, userIDs)
	}

	if userRoleUserIDs != nil {
		return r.roleDB.GetUserRolesByUserIDs(ctx, userRoleUserIDs)
	}

	return r.roleDB.GetRoles(ctx)
}
