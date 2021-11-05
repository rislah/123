package queries

import (
	"context"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/loaders"
)

type UserResolver struct {
	user    app.User
	backend *app.Backend
}

func NewUserResolver(usr app.User, data *app.Backend) *UserResolver {
	usr = usr.Sanitize()
	return &UserResolver{
		user:    usr,
		backend: data,
	}
}

func NewUsersByRoleIDResolver(ctx context.Context, data *app.Backend, roleID int) (*[]*UserResolver, error) {
	users, err := loaders.LoadUsersByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	usersResolver := make([]*UserResolver, 0, len(users))
	for _, user := range users {
		usersResolver = append(usersResolver, NewUserResolver(user, data))
	}

	return &usersResolver, nil
}

func NewUserListResolver(ctx context.Context, data *app.Backend) (*[]*UserResolver, error) {
	users, err := data.User.GetUsers(ctx, app.UsersQueryArgs{})
	if err != nil {
		return nil, err
	}

	usersResolver := make([]*UserResolver, 0, len(users))
	for _, user := range users {
		usersResolver = append(usersResolver, NewUserResolver(user, data))
	}

	return &usersResolver, nil
}

func (r *QueryResolver) Users(ctx context.Context) (*[]*UserResolver, error) {
	return NewUserListResolver(ctx, r.Backend)
}

func (u *UserResolver) ID() string {
	return u.user.UserID
}

func (u *UserResolver) Username() string {
	return u.user.Username
}

func (u *UserResolver) Role(ctx context.Context) (*RoleResolver, error) {
	return NewRoleResolverByUserID(ctx, u.backend, u.user.UserID)
}
