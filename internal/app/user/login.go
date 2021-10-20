package user

import (
	"context"
	"database/sql"

	"github.com/rislah/fakes/internal/jwt"
	"github.com/rislah/fakes/internal/password"
)

func (u userImpl) Login(ctx context.Context, username, clearTextPassword string) (string, error) {
	user, err := u.userDB.GetUserByUsername(ctx, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrUserNotFound
		}
		return "", err
	}

	if err := u.comparePasswords(user.Password, clearTextPassword); err != nil {
		return "", err
	}

	token, err := u.jwt.Encode(jwt.NewUserClaims(user.Username, "asd"))
	if err != nil {
		return "", err
	}

	return token, nil
}

func (u userImpl) comparePasswords(hashedPassword, clearTextPassword string) error {
	pass := password.NewPassword(clearTextPassword)
	compare, err := pass.CompareBCrypt(hashedPassword)
	if err != nil {
		return err
	}

	if !compare {
		return ErrLoginBadCredentials
	}

	return nil
}
