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
	return &UserResolver{
		user: usr.Sanitize(),
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
		usersResolver = append(usersResolver, NewUserResolver(*user, data))
	}

	return &usersResolver, nil
}

func (r *QueryResolver) Users(ctx context.Context) (*[]*UserResolver, error) {
	users, err := r.Data.UserDB.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		loaders.PrimeUsers(ctx, &user)
	}

	usersResolver := make([]*UserResolver, 0, len(users))
	for _, usr := range users {
		userResolver := NewUserResolver(usr, r.Data)
		usersResolver = append(usersResolver, userResolver)
	}

	return &usersResolver, nil
}

func (u *UserResolver) ID() string {
	return u.user.UserID
}

func (u *UserResolver) Username() string {
	return u.user.Username
}

func (u *UserResolver) Role(ctx context.Context) *RoleResolver {
	return NewRoleResolverByUserID(ctx, u.data, u.user.UserID)
}
