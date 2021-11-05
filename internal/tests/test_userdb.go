package tests

import (
	"context"
	"testing"
	"time"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/local"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MakeUserDB func() (app.UserDB, func() error, error)

func TestUserDB(t *testing.T, makeUserDB MakeUserDB, makeRoleDB MakeRoleDB) {
	tests := []struct {
		name       string
		users      []app.User
		cachedTest bool
		test       func(ctx context.Context, t *testing.T, userDB app.UserDB, users ...app.User)
	}{
		{
			name: "create a user and read it back",
			users: []app.User{
				{
					Username: "user",
					Password: "pw",
				},
			},
			test: func(ctx context.Context, t *testing.T, db app.UserDB, users ...app.User) {
				err := db.CreateUser(ctx, users[0])
				assert.NoError(t, err)

				usersCreated, err := db.GetUsers(ctx)
				assert.NoError(t, err)

				assert.Equal(t, len(usersCreated), 1)
				assert.Equal(t, usersCreated[0].Username, users[0].Username)
				assert.Equal(t, usersCreated[0].Password, users[0].Password)
			},
		},
		{
			name: "get all users",
			users: []app.User{
				{
					Username: "user1",
					Password: "pw",
				},
				{
					Username: "user2",
					Password: "pw",
				},
			},
			test: func(ctx context.Context, t *testing.T, db app.UserDB, users ...app.User) {
				for _, usr := range users {
					err := db.CreateUser(ctx, usr)
					assert.NoError(t, err)
				}

				res, err := db.GetUsers(ctx)
				assert.NoError(t, err)
				assert.Equal(t, len(res), 2)
			},
		},
		{
			name:  "no users",
			users: []app.User{},
			test: func(ctx context.Context, t *testing.T, db app.UserDB, users ...app.User) {
				res, err := db.GetUsers(ctx)
				assert.NoError(t, err)
				assert.Empty(t, res)
			},
		},
		{
			name: "getbyusername when exists",
			users: []app.User{
				{
					Username: "user1",
					Password: "pw",
				},
			},
			test: func(ctx context.Context, t *testing.T, db app.UserDB, users ...app.User) {
				err := db.CreateUser(ctx, users[0])
				assert.NoError(t, err)

				res, err := db.GetUserByUsername(ctx, users[0].Username)
				assert.NoError(t, err)
				assert.NotEmpty(t, res)
				assert.Equal(t, users[0].Username, res.Username)
				assert.Equal(t, users[0].Password, res.Password)

			},
		},
		{
			name: "getbyusername when doesnt exist",
			users: []app.User{
				{
					Username: "user1",
					Password: "pw",
				},
			},
			test: func(ctx context.Context, t *testing.T, db app.UserDB, users ...app.User) {
				res, err := db.GetUserByUsername(ctx, users[0].Username)
				assert.NoError(t, err)
				assert.Empty(t, res)
			},
		},
		{
			name: "getuserbyids when exists",
			users: []app.User{
				{
					Username: "user1",
					Password: "pw",
				},
				{
					Username: "user2",
					Password: "pw",
				},
			},
			test: func(ctx context.Context, t *testing.T, userDB app.UserDB, users ...app.User) {
				for _, user := range users {
					err := userDB.CreateUser(ctx, user)
					assert.NoError(t, err)
				}

				res, err := userDB.GetUsers(ctx)
				assert.NoError(t, err)
				assert.Equal(t, len(res), 2)

				userIDs := []string{}
				for _, r := range res {
					userIDs = append(userIDs, r.UserID)
				}

				usersByIDs, err := userDB.GetUsersByIDs(ctx, userIDs)
				assert.NoError(t, err)
				assert.Len(t, usersByIDs, 2)
			},
		},
		{
			name:  "getuserbyids when doesnt exists",
			users: []app.User{},
			test: func(ctx context.Context, t *testing.T, userDB app.UserDB, users ...app.User) {
				usersByIDs, err := userDB.GetUsersByIDs(ctx, []string{"1", "2"})
				assert.NoError(t, err)
				assert.Empty(t, usersByIDs)
			},
		},
		{
			name: "getuserbyroleid when exists",
			users: []app.User{
				{
					Username: "user1",
					Password: "pw",
				},
			},
			test: func(ctx context.Context, t *testing.T, userDB app.UserDB, users ...app.User) {
				for _, user := range users {
					err := userDB.CreateUser(ctx, user)
					assert.NoError(t, err)
				}
				usersByRoleID, err := userDB.GetUsersByRoleID(ctx, 1)
				assert.NoError(t, err)
				assert.Len(t, usersByRoleID, 1)
			},
		},
		{
			name:  "getuserbyroleid when doesnt exists",
			users: []app.User{},
			test: func(ctx context.Context, t *testing.T, userDB app.UserDB, users ...app.User) {
				usersByRoleID, err := userDB.GetUsersByRoleID(ctx, 1)
				assert.NoError(t, err)
				assert.Empty(t, usersByRoleID)
			},
		},
		{
			name: "GetUsersByRoleIDs when exists",
			users: []app.User{
				{
					Username: "user1",
					Password: "pw",
				},
			},
			test: func(ctx context.Context, t *testing.T, userDB app.UserDB, users ...app.User) {
				for _, user := range users {
					err := userDB.CreateUser(ctx, user)
					assert.NoError(t, err)
				}
				usersByRoleIDs, err := userDB.GetUsersByRoleIDs(ctx, []int{1})
				assert.NoError(t, err)
				assert.Len(t, usersByRoleIDs, 1)
				assert.NotEmpty(t, usersByRoleIDs)
			},
		},
		{
			name: "GetUsersByRoleIDs when doesnt exists",
			users: []app.User{
				{
					Username: "user1",
					Password: "pw",
				},
			},
			test: func(ctx context.Context, t *testing.T, userDB app.UserDB, users ...app.User) {
				usersByRoleIDs, err := userDB.GetUsersByRoleIDs(ctx, []int{1})
				assert.NoError(t, err)
				assert.Empty(t, usersByRoleIDs)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			userDB, userDBTeardown, err := makeUserDB()
			require.NoError(t, err)

			roleDB, roleDBTeardown, err := makeRoleDB()
			require.NoError(t, err)

			if u, ok := userDB.(local.LocalUserDB); ok {
				u.SetRoleDB(roleDB)
			}

			defer func() {
				require.NoError(t, userDBTeardown())
				require.NoError(t, roleDBTeardown())
			}()

			test.test(ctx, t, userDB, test.users...)
		})
	}
}
