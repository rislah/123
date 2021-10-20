package app_test

import (
	"context"
	"testing"
	"time"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/jwt"
	"github.com/rislah/fakes/internal/local"
	"github.com/stretchr/testify/assert"
)

func TestUserImpl_GetUsers(t *testing.T) {
	tests := []struct {
		name string
		test func(ctx context.Context, t *testing.T, userBackend app.UserBackend)
	}{
		{
			name: "should return error if users is empty",
			test: func(ctx context.Context, t *testing.T, svc app.UserBackend) {
				_, err := svc.GetUsers(ctx)
				assert.Error(t, err)
			},
		},
		{
			name: "should return users",
			test: func(ctx context.Context, t *testing.T, svc app.UserBackend) {
				err := svc.CreateUser(ctx, app.User{
					Username: "asd",
					Password: "asdasd",
				})
				assert.NoError(t, err)
				resp, err := svc.GetUsers(ctx)
				assert.NoError(t, err)
				assert.Len(t, resp, 1)
			},
		},
	}

	for _, tc := range tests {
		test := tc
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			db, teardown, err := local.MakeUserDB()
			defer teardown()
			assert.NoError(t, err)

			jwtWrapper := jwt.NewHS256Wrapper("wrapper")
			svc := app.NewUserBackend(db, jwtWrapper)

			defer func() {
				assert.NoError(t, teardown())
			}()

			test.test(ctx, t, svc)
		})
	}
}
