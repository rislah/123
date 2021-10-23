package app

import (
	"context"

	"github.com/rislah/fakes/internal/credentials"
	"github.com/rislah/fakes/internal/jwt"
)

type Authenticator interface {
	AuthenticatePassword(context.Context, credentials.Credentials) (User, error)
	GenerateJWT(User) (string, error)
}

type authenticatorImpl struct {
	userDB     UserDB
	jwtWrapper jwt.Wrapper
}

func NewAuthenticator(userdb UserDB, jwtWrapper jwt.Wrapper) authenticatorImpl {
	return authenticatorImpl{
		userdb,
		jwtWrapper,
	}
}

func (a authenticatorImpl) AuthenticatePassword(ctx context.Context, creds credentials.Credentials) (User, error) {
	if err := creds.Valid(); err != nil {
		return User{}, err
	}

	usr, err := a.userDB.GetUserByUsername(ctx, string(creds.Username))
	if err != nil {
		return User{}, err
	}

	if usr.IsEmpty() {
		return User{}, ErrUserNotFound
	}

	if err := credentials.ComparePassword(usr.Password, creds.Password); err != nil {
		return User{}, err
	}

	return usr, nil
}

func (a authenticatorImpl) GenerateJWT(usr User) (string, error) {
	usrClaims := jwt.NewUserClaims(usr.Username, usr.Role.String())
	tokenStr, err := a.jwtWrapper.Encode(usrClaims)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}
