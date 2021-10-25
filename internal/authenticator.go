package app

import (
	"context"

	"github.com/rislah/fakes/internal/credentials"
	"github.com/rislah/fakes/internal/jwt"
)

type Authenticator interface {
	AuthenticatePassword(ctx context.Context, creds credentials.Credentials) (User, error)
	GenerateJWT(ctx context.Context, user User) (string, error)
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

func (a authenticatorImpl) GenerateJWT(ctx context.Context, usr User) (string, error) {
	role, err := a.userDB.GetUserRoleByUserID(ctx, usr.UserID)
	if err != nil {
		return "", err
	}

	usrClaims := jwt.NewUserClaims(usr.Username, role.Name.String())
	tokenStr, err := a.jwtWrapper.Encode(usrClaims)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}
