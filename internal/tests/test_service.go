package tests

import (
	"context"
	"testing"
	"time"

	app "github.com/rislah/fakes/internal"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	service app.Service
	db      app.UserDB
}

func TestService(t *testing.T, makeUserDB MakeUserDB) {
	t.Parallel()

	tests := []struct {
		name string
		test func(ctx context.Context, t *testing.T, testCase testCase)
	}{
		{
			name: "should return error if users is empty",
			test: func(ctx context.Context, t *testing.T, testCase testCase) {
				users, err := testCase.service.GetUsers(ctx)
				if assert.Error(t, err) {
					assert.Equal(t, app.ErrUsersNotFound, err)
				}
				assert.Nil(t, users)
			},
		},
	}

	for _, tc := range tests {
		test := tc

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			db, err := makeUserDB()
			if err != nil {
				t.Fatal(err)
			}

			config := app.ServiceConfig{
				UserDB: db,
			}

			service := app.NewUserService(config)

			test.test(ctx, t, testCase{
				service: service,
				db:      db,
			})
		})
	}
}
