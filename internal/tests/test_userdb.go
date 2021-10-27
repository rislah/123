package tests

import (
	"context"
	"testing"
	"time"

	app "github.com/rislah/fakes/internal"

	"github.com/stretchr/testify/assert"
)

type MakeUserDB func() (app.UserDB, func() error, error)

func TestUserDB(t *testing.T, makeUserDB MakeUserDB) {
	tests := []struct {
		name       string
		users      []app.User
		cachedTest bool
		test       func(ctx context.Context, t *testing.T, db app.UserDB, users ...app.User)
	}{
		// {
		// 	name: "create a user and read it back",
		// 	users: []app.User{
		// 		{
		// 			Username: "user",
		// 			Password: "pw",
		// 			Role:     "guest",
		// 		},
		// 	},
		// 	test: func(ctx context.Context, t *testing.T, db app.UserDB, users ...app.User) {
		// 		err := db.CreateUser(ctx, users[0])
		// 		assert.NoError(t, err)

		// 		usersCreated, err := db.GetUsers(ctx)
		// 		assert.NoError(t, err)

		// 		assert.Equal(t, len(usersCreated), 1)
		// 		assert.Equal(t, usersCreated[0].Username, users[0].Username)
		// 		assert.Equal(t, usersCreated[0].Password, users[0].Password)
		// 	},
		// },
		// {
		// 	name: "get all users",
		// 	users: []app.User{
		// 		{
		// 			Username: "user1",
		// 			Password: "pw",
		// 			Role:     "guest",
		// 		},
		// 		{
		// 			Username: "user2",
		// 			Password: "pw",
		// 			Role:     "guest",
		// 		},
		// 	},
		// 	test: func(ctx context.Context, t *testing.T, db app.UserDB, users ...app.User) {
		// 		for _, usr := range users {
		// 			err := db.CreateUser(ctx, usr)
		// 			assert.NoError(t, err)
		// 		}

		// 		res, err := db.GetUsers(ctx)
		// 		assert.NoError(t, err)
		// 		assert.Equal(t, len(res), 2)
		// 	},
		// },
		// {
		// 	name:  "no users",
		// 	users: []app.User{},
		// 	test: func(ctx context.Context, t *testing.T, db app.UserDB, users ...app.User) {
		// 		res, err := db.GetUsers(ctx)
		// 		assert.NoError(t, err)
		// 		assert.Empty(t, res)
		// 	},
		// },
		// {
		// 	name: "getbyusername when exists",
		// 	users: []app.User{
		// 		{
		// 			Username: "user1",
		// 			Password: "pw",
		// 			Role:     "guest",
		// 		},
		// 	},
		// 	test: func(ctx context.Context, t *testing.T, db app.UserDB, users ...app.User) {
		// 		err := db.CreateUser(ctx, users[0])
		// 		assert.NoError(t, err)

		// 		res, err := db.GetUserByUsername(ctx, users[0].Username)
		// 		assert.NoError(t, err)
		// 		assert.NotEmpty(t, res)
		// 		assert.Equal(t, users[0].Username, res.Username)
		// 		assert.Equal(t, users[0].Role, res.Role)
		// 		assert.Equal(t, users[0].Password, res.Password)

		// 	},
		// },
		// {
		// 	name: "getbyusername when doesnt exist",
		// 	users: []app.User{
		// 		{
		// 			Username: "user1",
		// 			Password: "pw",
		// 			Role:     "guest",
		// 		},
		// 	},
		// 	test: func(ctx context.Context, t *testing.T, db app.UserDB, users ...app.User) {
		// 		res, err := db.GetUserByUsername(ctx, users[0].Username)
		// 		assert.NoError(t, err)
		// 		assert.Empty(t, res)
		// 	},
		// },
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			db, teardown, err := makeUserDB()
			if err != nil {
				t.Fatal(err)
			}

			defer func() {
				err := teardown()
				assert.NoError(t, err)
			}()

			test.test(ctx, t, db, test.users...)
		})
	}
}
