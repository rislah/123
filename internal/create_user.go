package app

import (
	"context"

	"github.com/rislah/fakes/internal/credentials"
)

func (u *userImpl) CreateUser(ctx context.Context, creds credentials.Credentials) error {
	if err := creds.Valid(); err != nil {
		return err
	}

	if _, err := creds.Password.ValidateStrength(creds.Username.String()); err != nil {
		return err
	}

	usr, err := u.userDB.GetUserByUsername(ctx, creds.Username.String())
	if err != nil {
		return err
	}

	if !usr.IsEmpty() {
		return ErrUserAlreadyExists
	}

	hash, err := creds.Password.GenerateBCrypt()
	if err != nil {
		return err
	}

	err = u.userDB.CreateUser(ctx, User{
		Username: creds.Username.String(),
		Password: hash,
	})

	if err != nil {
		return err
	}

	return nil
}
