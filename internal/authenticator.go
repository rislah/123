package app

import (
	"context"
	"database/sql"
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
	if err := creds.Password.ValidateLength(); err != nil {
		return User{}, err
	}

	usr, err := a.userDB.GetUserByUsername(ctx, string(creds.Username))
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}

	if err := credentials.ValidatePassword(usr.Password, creds.Password); err != nil {
		return User{}, nil
	}

	return usr, nil
}

func (a authenticatorImpl) GenerateJWT(usr User) (string, error) {
	usrClaims := jwt.NewUserClaims(usr.Username, usr.Role)
	tokenStr, err := a.jwtWrapper.Encode(usrClaims)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}
