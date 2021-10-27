package queries

import (
	"context"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/loaders"
)

type UserResolver struct {
	user app.User
	data *app.Data
}

func NewUserResolver(usr app.User, data *app.Data) *UserResolver {
	usr = usr.Sanitize()
	return &UserResolver{
		user: usr,
		data: data,
	}
}

func NewUsersByRoleIDResolver(ctx context.Context, data *app.Data, roleID int) (*[]*UserResolver, error) {
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

func NewUserListResolver(ctx context.Context, data *app.Data) (*[]*UserResolver, error) {
	users, err := data.UserDB.GetUsers(ctx)
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
	return NewUserListResolver(ctx, r.Data)
}

func (u *UserResolver) ID() string {
	return u.user.UserID
}

func (u *UserResolver) Username() string {
	return u.user.Username
}

func (u *UserResolver) Role(ctx context.Context) (*RoleResolver, error) {
	return NewRoleResolverByUserID(ctx, u.data, u.user.UserID)
}
