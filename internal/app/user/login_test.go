package user_test

import (
	"context"
	"testing"
	"time"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/app/user"
	"github.com/rislah/fakes/internal/jwt"
	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/password"
	"github.com/stretchr/testify/assert"
)

type loginTestCase struct {
	svc user.User
	db  app.UserDB

	user app.User
}

func TestLogin(t *testing.T) {
	tests := []struct {
		scenario string
		user     app.User
		test     func(ctx context.Context, testCase loginTestCase)
	}{
		{
			scenario: "should login with correct credentials",
			user: app.User{
				Username: "test_username",
				Password: "test_password",
			},
			test: func(ctx context.Context, testcase loginTestCase) {
				user := testcase.user
				hash, err := password.NewPassword(testcase.user.Password).GenerateBCrypt()
				assert.NoError(t, err)

				user.Password = hash
				err = testcase.db.CreateUser(ctx, user)
				assert.NoError(t, err)

				jwt, err := testcase.svc.Login(ctx, testcase.user.Username, testcase.user.Password)
				assert.NoError(t, err)
				assert.NotEmpty(t, jwt)
			},
		},
		{
			scenario: "should return false if user not found",
			user: app.User{
				Username: "test_username",
				Password: "test_password",
			},
			test: func(ctx context.Context, testcase loginTestCase) {
				_, err := testcase.svc.Login(ctx, testcase.user.Username, testcase.user.Password)
				assert.Error(t, err)
				assert.Equal(t, user.ErrUserNotFound, err)
			},
		},
		{
			scenario: "should return false with incorrect credentials",
			user: app.User{
				Username: "test_username",
				Password: "test_password",
			},
			test: func(ctx context.Context, testcase loginTestCase) {
				tcUser := testcase.user
				hash, err := password.NewPassword("").GenerateBCrypt()
				assert.NoError(t, err)

				tcUser.Password = hash
				err = testcase.db.CreateUser(ctx, tcUser)
				assert.NoError(t, err)

				_, err = testcase.svc.Login(ctx, testcase.user.Username, testcase.user.Password)
				assert.Error(t, err)
				assert.Equal(t, user.ErrLoginBadCredentials, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			db, teardown, err := local.MakeUserDB()
			assert.NoError(t, err)
			defer teardown()

			jwtWrapper := jwt.NewHS256Wrapper("secret")
			svc := user.NewUser(db, jwtWrapper)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			test.test(ctx, loginTestCase{
				svc:  svc,
				db:   db,
				user: test.user,
			})
		})
	}
}
