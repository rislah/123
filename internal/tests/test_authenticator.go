package tests

import (
	"context"
	"testing"
	"time"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/credentials"
	"github.com/rislah/fakes/internal/jwt"
	"github.com/rislah/fakes/internal/local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type authenticatorTestCase struct {
	creds      credentials.Credentials
	auth       app.Authenticator
	db         app.UserDB
	jwtWrapper jwt.Wrapper
}

func TestAuthenticator(t *testing.T, makeRoleDB MakeRoleDB, makeUserDB MakeUserDB) {
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
			scenario: "min password length",
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
				}
				err := testCase.db.CreateUser(ctx, user)
				assert.NoError(t, err)

				byUsername, err := testCase.db.GetUserByUsername(ctx, user.Username)
				assert.NoError(t, err)

				user.UserID = byUsername.UserID

				tokenStr, err := testCase.auth.GenerateJWT(ctx, user)
				assert.NoError(t, err)
				assert.NotEmpty(t, tokenStr)

				token, err := testCase.jwtWrapper.Decode(tokenStr, &jwt.UserClaims{})
				assert.NoError(t, err)

				tokenUsrClaims, ok := token.Claims.(*jwt.UserClaims)
				assert.True(t, ok)
				assert.Equal(t, testCase.creds.Username.String(), tokenUsrClaims.Username)
				assert.Equal(t, app.GuestRoleType.String(), tokenUsrClaims.Role)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			userDB, userTeardown, err := makeUserDB()
			require.NoError(t, err)

			roleDB, roleTearDown, err := makeRoleDB()
			require.NoError(t, err)

			if u, ok := userDB.(local.LocalUserDB); ok {
				u.SetRoleDB(roleDB)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			jwtWrapper := jwt.NewHS256Wrapper("secret")
			a := app.NewAuthenticator(userDB, roleDB, jwtWrapper)
			test.test(ctx, authenticatorTestCase{
				creds:      test.creds,
				auth:       a,
				db:         userDB,
				jwtWrapper: jwtWrapper,
			})

			defer func() {
				err := userTeardown()
				require.NoError(t, err)
				err = roleTearDown()
				require.NoError(t, err)
				cancel()
			}()
		})
	}
}
