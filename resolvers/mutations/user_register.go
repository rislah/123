package mutations

import (
	"context"

	"github.com/rislah/fakes/internal/credentials"
)

type UserRegisterPayloadResolver struct {
	username string
}

func (u *UserRegisterPayloadResolver) Username() string {
	return u.username
}

func NewUserRegisterPayloadResolver(username string) *UserRegisterPayloadResolver {
	return &UserRegisterPayloadResolver{username: username}
}

type UserRegisterInput struct {
	Username string
	Password string
}

type UserRegisterArgs struct {
	Input UserRegisterInput
}

func (m *MutationResolver) Register(ctx context.Context, args UserRegisterArgs) (*UserRegisterPayloadResolver, error) {
	creds := credentials.New(args.Input.Username, args.Input.Password)
	err := m.Backend.User.CreateUser(ctx, creds)
	if err != nil {
		return nil, err
	}

	return NewUserRegisterPayloadResolver(args.Input.Username), nil
}
