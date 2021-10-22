package app_test

import (
	"context"
	"github.com/rislah/fakes/internal/credentials"
	"testing"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/jwt"
	"github.com/rislah/fakes/internal/local"
	"github.com/stretchr/testify/assert"
)

func TestUserImpl_CreateUser(t *testing.T) {
	tests := []struct {
		name  string
		creds credentials.Credentials
		test  func(ctx context.Context, t *testing.T, creds credentials.Credentials, userBackend app.UserBackend, db app.UserDB)
	}{
		{
			name:  "below username min length",
			creds: credentials.New("aaa", "parool123!"),
			test: func(ctx context.Context, t *testing.T, creds credentials.Credentials, userBackend app.UserBackend, db app.UserDB) {
				err := userBackend.CreateUser(ctx, creds)
				assert.Error(t, err)
				assert.Equal(t, credentials.ErrUsernameLength, err)
			},
		},
		{
			name:  "below password min length",
			creds: credentials.New("kastuaja123", "parool"),
			test: func(ctx context.Context, t *testing.T, creds credentials.Credentials, userBackend app.UserBackend, db app.UserDB) {
				err := userBackend.CreateUser(ctx, creds)
				assert.Error(t, err)
				assert.Equal(t, credentials.ErrPasswordLength, err)
			},
		},
		{
			name:  "below password complexity",
			creds: credentials.New("kasutaja", "parool123"),
			test: func(ctx context.Context, t *testing.T, creds credentials.Credentials, userBackend app.UserBackend, db app.UserDB) {
				err := userBackend.CreateUser(ctx, creds)
				assert.Error(t, err)
				assert.Equal(t, credentials.ErrPasswordNotComplexEnough, err)
			},
		},
		{
			name:  "user already exists",
			creds: credentials.New("kasutaja", "parool123!"),
			test: func(ctx context.Context, t *testing.T, creds credentials.Credentials, userBackend app.UserBackend, db app.UserDB) {
				hashedPassword, err := creds.Password.GenerateBCrypt()
				assert.NoError(t, err)

				err = db.CreateUser(ctx, app.User{
					Username: creds.Username.String(),
					Password: hashedPassword,
				})
				assert.NoError(t, err)

				err = userBackend.CreateUser(ctx, creds)
				assert.Error(t, err)
				assert.Equal(t, app.ErrUserAlreadyExists, err)
			},
		},
		{
			name:  "should create user",
			creds: credentials.New("kasutaja", "parool123!"),
			test: func(ctx context.Context, t *testing.T, creds credentials.Credentials, userBackend app.UserBackend, db app.UserDB) {
				err := userBackend.CreateUser(ctx, creds)
				assert.NoError(t, err)

				usr, err := db.GetUserByUsername(ctx, creds.Username.String())
				assert.NoError(t, err)
				assert.Equal(t, creds.Username.String(), usr.Username)
				assert.NotEmpty(t, usr.Password)
				assert.NotEmpty(t, usr.Role)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userDB, teardown, err := local.MakeUserDB()
			defer teardown()
			assert.NoError(t, err)

			jwtWrapper := jwt.NewHS256Wrapper("wrap")
			userBackend := app.NewUserBackend(userDB, jwtWrapper)
			test.test(context.Background(), t, test.creds, userBackend, userDB)
		})
	}
}
