package tests

import (
	"context"
	"testing"
	"time"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/credentials"
	"github.com/rislah/fakes/internal/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type authenticatorTestCase struct {
	creds      credentials.Credentials
	auth       app.Authenticator
	db         app.UserDB
	jwtWrapper jwt.Wrapper
}

func TestAuthenticator(t *testing.T, makeUserDB MakeUserDB) {
	tests := []struct {
		scenario string
		creds    credentials.Credentials
		test     func(ctx context.Context, testCase authenticatorTestCase)
	}{
		{
			scenario: "user doesnt exist",
			creds: credentials.Credentials{
				Username: "test_username",
				Password: "p@r00l!2$",
			},
			test: func(ctx context.Context, testCase authenticatorTestCase) {
				_, err := testCase.auth.AuthenticatePassword(ctx, testCase.creds)
				assert.Error(t, err)
				assert.Equal(t, app.ErrUserNotFound, err)
			},
		},
		{
			scenario: "fail length",
			creds: credentials.Credentials{
				Username: "test_username",
				Password: "a",
			},
			test: func(ctx context.Context, testCase authenticatorTestCase) {
				_, err := testCase.auth.AuthenticatePassword(ctx, testCase.creds)
				assert.Error(t, err)
				assert.Equal(t, credentials.ErrPasswordLength, err)
			},
		},
		{
			scenario: "creates valid jwt",
			creds: credentials.Credentials{
				Username: "test_username",
				Password: "p@r00l!2$",
			},
			test: func(ctx context.Context, testCase authenticatorTestCase) {
				user := app.User{
					Username: testCase.creds.Username.String(),
					Password: testCase.creds.Password.String(),
					Role:     app.GuestRole,
				}

				tokenStr, err := testCase.auth.GenerateJWT(user)
				assert.NoError(t, err)
				assert.NotEmpty(t, tokenStr)

				token, err := testCase.jwtWrapper.Decode(tokenStr, &jwt.UserClaims{})
				assert.NoError(t, err)

				tokenUsrClaims, ok := token.Claims.(*jwt.UserClaims)
				assert.True(t, ok)
				assert.Equal(t, testCase.creds.Username.String(), tokenUsrClaims.Username)
				assert.Equal(t, app.GuestRole.String(), tokenUsrClaims.Role)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			db, teardown, err := makeUserDB()
			assert.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			jwtWrapper := jwt.NewHS256Wrapper("secret")
			a := app.NewAuthenticator(db, jwtWrapper)
			test.test(ctx, authenticatorTestCase{
				creds:      test.creds,
				auth:       a,
				db:         db,
				jwtWrapper: jwtWrapper,
			})

			defer func() {
				err := teardown()
				require.NoError(t, err)
				cancel()
			}()
		})
	}
}
