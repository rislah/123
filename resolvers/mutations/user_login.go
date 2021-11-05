package mutations

import (
	"context"

	"github.com/rislah/fakes/internal/credentials"
)

type UserLoginPayloadResolver struct {
	token string
}

func (u *UserLoginPayloadResolver) Token() string {
	return u.token
}

func NewUserLoginPayloadResolver(token string) *UserLoginPayloadResolver {
	return &UserLoginPayloadResolver{token: token}
}

type UserLoginInput struct {
	Username string
	Password string
}

type UserLoginArgs struct {
	Input UserLoginInput
}

func (m *MutationResolver) Login(ctx context.Context, args UserLoginArgs) (*UserLoginPayloadResolver, error) {
	creds := credentials.New(args.Input.Username, args.Input.Password)
	user, err := m.Backend.Authenticator.AuthenticatePassword(ctx, creds)
	if err != nil {
		return nil, err
	}

	token, err := m.Backend.Authenticator.GenerateJWT(ctx, user)
	if err != nil {
		return nil, err
	}

	return NewUserLoginPayloadResolver(token), nil
}
