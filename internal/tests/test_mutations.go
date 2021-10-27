package tests

import (
	"context"
	"testing"
	"time"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/credentials"
	"github.com/rislah/fakes/internal/jwt"
	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/loaders"
	"github.com/rislah/fakes/resolvers/mutations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type registerTestCase struct {
	data      *app.Data
	mutations *mutations.MutationResolver
	args      mutations.UserRegisterArgs
}

func TestRegister(t *testing.T, makeUserDB MakeUserDB, makeRoleDB MakeRoleDB) {
	tests := []struct {
		scenario string
		args     mutations.UserRegisterArgs
		test     func(ctx context.Context, testCase registerTestCase)
	}{
		{
			scenario: "register user",
			args: mutations.UserRegisterArgs{
				Input: mutations.UserRegisterInput{
					Username: "kasutaja",
					Password: "p@r00l!@#",
				},
			},
			test: func(ctx context.Context, testCase registerTestCase) {
				payload, err := testCase.mutations.Register(ctx, testCase.args)
				assert.NoError(t, err)
				assert.NotNil(t, payload)
				assert.Equal(t, testCase.args.Input.Username, payload.Username())
			},
		},
		{
			scenario: "username empty",
			args: mutations.UserRegisterArgs{
				Input: mutations.UserRegisterInput{
					Username: "",
					Password: "p@r00l!@#",
				},
			},
			test: func(ctx context.Context, testCase registerTestCase) {
				payload, err := testCase.mutations.Register(ctx, testCase.args)
				assert.ErrorIs(t, err, credentials.ErrUsernameMissing)
				assert.Nil(t, payload)
			},
		},
		{
			scenario: "username min length",
			args: mutations.UserRegisterArgs{
				Input: mutations.UserRegisterInput{
					Username: "uas",
					Password: "p@r00l!@#",
				},
			},
			test: func(ctx context.Context, testCase registerTestCase) {
				payload, err := testCase.mutations.Register(ctx, testCase.args)
				assert.ErrorIs(t, err, credentials.ErrUsernameLength)
				assert.Nil(t, payload)
			},
		},
		{
			scenario: "password min length",
			args: mutations.UserRegisterArgs{
				Input: mutations.UserRegisterInput{
					Username: "kasutaja",
					Password: "asd",
				},
			},
			test: func(ctx context.Context, testCase registerTestCase) {
				payload, err := testCase.mutations.Register(ctx, testCase.args)
				assert.ErrorIs(t, err, credentials.ErrPasswordLength)
				assert.Nil(t, payload)
			},
		},
		{
			scenario: "password empty",
			args: mutations.UserRegisterArgs{
				Input: mutations.UserRegisterInput{
					Username: "kastuaja",
					Password: "",
				},
			},
			test: func(ctx context.Context, testCase registerTestCase) {
				payload, err := testCase.mutations.Register(ctx, testCase.args)
				assert.ErrorIs(t, err, credentials.ErrPasswordMissing)
				assert.Nil(t, payload)
			},
		},
		{
			scenario: "password complexity",
			args: mutations.UserRegisterArgs{
				Input: mutations.UserRegisterInput{
					Username: "kastuaja",
					Password: "testtest",
				},
			},
			test: func(ctx context.Context, testCase registerTestCase) {
				payload, err := testCase.mutations.Register(ctx, testCase.args)
				assert.ErrorIs(t, err, credentials.ErrPasswordNotComplexEnough)
				assert.Nil(t, payload)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			roleDB, roleDBTeardown, err := makeRoleDB()
			require.NoError(t, err)

			userDB, userDBTeardown, err := makeUserDB()
			require.NoError(t, err)

			data := &app.Data{
				RoleDB: roleDB,
				UserDB: userDB,
				User:   app.NewUserBackend(userDB, jwt.NewHS256Wrapper("secret")),
			}

			if u, ok := userDB.(local.LocalUserDB); ok {
				u.SetRoleDB(roleDB)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			lds := loaders.New(data)
			ctxWithLoaders := lds.Attach(ctx)

			defer func() {
				cancel()
				require.NoError(t, roleDBTeardown())
				require.NoError(t, userDBTeardown())
			}()

			test.test(ctxWithLoaders, registerTestCase{
				mutations: &mutations.MutationResolver{Data: data},
				args:      test.args,
				data:      data,
			})
		})
	}
}

type loginTestCase struct {
	data             *app.Data
	mutationResolver *mutations.MutationResolver
	jwtWrapper       jwt.Wrapper

	args mutations.UserLoginArgs
}

func TestLogin(t *testing.T, makeRoleDB MakeRoleDB, makeUserDB MakeUserDB) {
	tests := []struct {
		name string
		args mutations.UserLoginArgs
		test func(ctx context.Context, apiTestCase loginTestCase)
	}{
		{
			name: "should return error if password doesnt meet length",
			args: mutations.UserLoginArgs{
				Input: mutations.UserLoginInput{
					Username: "testasdasdasd",
					Password: "test",
				},
			},
			test: func(ctx context.Context, apiTestCase loginTestCase) {
				resp, err := apiTestCase.mutationResolver.Login(ctx, apiTestCase.args)
				assert.Nil(t, resp)
				assert.ErrorIs(t, err, credentials.ErrPasswordLength)
			},
		},
		{
			name: "should return error if user doesnt meet length",
			args: mutations.UserLoginArgs{
				Input: mutations.UserLoginInput{
					Username: "tst",
					Password: "testtesttest",
				},
			},
			test: func(ctx context.Context, apiTestCase loginTestCase) {
				resp, err := apiTestCase.mutationResolver.Login(ctx, apiTestCase.args)
				assert.Nil(t, resp)
				assert.ErrorIs(t, err, credentials.ErrUsernameLength)
			},
		},
		{
			name: "should return error if user empty",
			args: mutations.UserLoginArgs{
				Input: mutations.UserLoginInput{
					Username: "",
					Password: "testtesttest",
				},
			},
			test: func(ctx context.Context, apiTestCase loginTestCase) {
				resp, err := apiTestCase.mutationResolver.Login(ctx, apiTestCase.args)
				assert.Nil(t, resp)
				assert.ErrorIs(t, err, credentials.ErrUsernameMissing)
			},
		},
		{
			name: "should return error if login empty",
			args: mutations.UserLoginArgs{
				Input: mutations.UserLoginInput{
					Username: "teststest",
					Password: "",
				},
			},
			test: func(ctx context.Context, apiTestCase loginTestCase) {
				resp, err := apiTestCase.mutationResolver.Login(ctx, apiTestCase.args)
				assert.Nil(t, resp)
				assert.ErrorIs(t, err, credentials.ErrPasswordMissing)
			},
		},
		{
			name: "should return error if user not found",
			args: mutations.UserLoginArgs{
				Input: mutations.UserLoginInput{
					Username: "testtest",
					Password: "testtest",
				},
			},
			test: func(ctx context.Context, apiTestCase loginTestCase) {
				resp, err := apiTestCase.mutationResolver.Login(ctx, apiTestCase.args)
				assert.Nil(t, resp)
				assert.ErrorIs(t, err, app.ErrUserNotFound)
			},
		},
		{
			name: "should return valid jwt on login",
			args: mutations.UserLoginArgs{
				Input: mutations.UserLoginInput{
					Username: "testtest",
					Password: "p@r00l!@#",
				},
			},
			test: func(ctx context.Context, apiTestCase loginTestCase) {
				creds := credentials.New(apiTestCase.args.Input.Username, apiTestCase.args.Input.Password)
				err := apiTestCase.data.User.CreateUser(ctx, creds)
				assert.NoError(t, err)

				loginPayload, err := apiTestCase.mutationResolver.Login(ctx, apiTestCase.args)
				assert.NoError(t, err)
				assert.NotEmpty(t, loginPayload)

				jwtToken := loginPayload.Token()
				token, err := apiTestCase.jwtWrapper.Decode(jwtToken, &jwt.UserClaims{})
				assert.NoError(t, err)

				userToken, ok := token.Claims.(*jwt.UserClaims)
				assert.True(t, ok)
				assert.Equal(t, apiTestCase.args.Input.Username, userToken.Username)
				assert.Equal(t, "guest", userToken.Role)
			},
		},
		{
			name: "should return error on incorrect password",
			args: mutations.UserLoginArgs{
				Input: mutations.UserLoginInput{
					Username: "testtest",
					Password: "p@r00l!@#",
				},
			},
			test: func(ctx context.Context, apiTestCase loginTestCase) {
				creds := credentials.New(apiTestCase.args.Input.Username, apiTestCase.args.Input.Password+"2")
				err := apiTestCase.data.User.CreateUser(ctx, creds)
				assert.NoError(t, err)

				_, err = apiTestCase.mutationResolver.Login(ctx, apiTestCase.args)
				assert.Error(t, err)
				assert.ErrorIs(t, err, credentials.ErrPasswordMismatch)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			roleDB, roleDBTeardown, err := makeRoleDB()
			require.NoError(t, err)

			userDB, userDBTeardown, err := makeUserDB()
			require.NoError(t, err)

			jwtWrapper := jwt.NewHS256Wrapper("secret")
			authenticator := app.NewAuthenticator(userDB, roleDB, jwtWrapper)

			data := &app.Data{
				User:          app.NewUserBackend(userDB, jwtWrapper),
				RoleDB:        roleDB,
				UserDB:        userDB,
				Authenticator: authenticator,
			}

			if u, ok := userDB.(local.LocalUserDB); ok {
				u.SetRoleDB(roleDB)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			lds := loaders.New(data)
			ctxWithLoaders := lds.Attach(ctx)

			defer func() {
				cancel()
				require.NoError(t, roleDBTeardown())
				require.NoError(t, userDBTeardown())
			}()

			test.test(ctxWithLoaders, loginTestCase{
				data:             data,
				mutationResolver: &mutations.MutationResolver{Data: data},
				args:             test.args,
				jwtWrapper:       jwtWrapper,
			})

		})
	}
}
