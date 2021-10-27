package tests

import (
	"context"
	"testing"
	"time"

	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/loaders"
	"github.com/rislah/fakes/resolvers/queries"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type usersTestCase struct {
	user app.User

	db            app.UserDB
	queryResolver *queries.QueryResolver
}

type MakeRoleDB func() (app.RoleDB, func() error, error)

func TestUsers(t *testing.T, makeUserDB MakeUserDB, makeRoleDB MakeRoleDB) {
	tests := []struct {
		scenario string
		roleID   int
		user     app.User
		test     func(ctx context.Context, t *testing.T, testCase usersTestCase)
	}{
		{
			scenario: "return users when users exist",
			user: app.User{
				UserID:   "1",
				Username: "kasutaja",
				Password: "parool",
			},
			test: func(ctx context.Context, t *testing.T, testCase usersTestCase) {
				err := testCase.db.CreateUser(ctx, testCase.user)
				assert.NoError(t, err)

				userResolvers, err := testCase.queryResolver.Users(ctx)
				assert.NoError(t, err)
				assert.NotNil(t, userResolvers)
				assert.Len(t, *userResolvers, 1)
			},
		},
		{
			scenario: "return no users when users doesnt exist",
			test: func(ctx context.Context, t *testing.T, testCase usersTestCase) {
				userResolvers, err := testCase.queryResolver.Users(ctx)
				assert.NoError(t, err)
				assert.Len(t, *userResolvers, 0)
			},
		},
		{
			scenario: "return user role",
			user: app.User{
				UserID:   "1",
				Username: "kasutaja",
				Password: "parool",
			},
			test: func(ctx context.Context, t *testing.T, testCase usersTestCase) {
				err := testCase.db.CreateUser(ctx, testCase.user)
				assert.NoError(t, err)

				userResolvers, err := testCase.queryResolver.Users(ctx)
				assert.NoError(t, err)
				assert.NotNil(t, userResolvers)
				assert.Len(t, *userResolvers, 1)

				for _, userResolver := range *userResolvers {
					roleResolver, err := userResolver.Role(ctx)
					assert.NoError(t, err)
					assert.NotNil(t, roleResolver)
					assert.Equal(t, "GUEST", roleResolver.Name())
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			userDB, userDBTeardown, err := makeUserDB()
			require.NoError(t, err)

			roleDB, roleDBTeardown, err := makeRoleDB()
			require.NoError(t, err)

			if u, ok := userDB.(local.LocalUserDB); ok {
				u.SetRoleDB(roleDB)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			data := &app.Data{
				UserDB: userDB,
				RoleDB: roleDB,
			}

			qr := &queries.QueryResolver{Data: data}
			loaders := loaders.New(data, nil, userDB, nil)
			ctxWithLoaders := loaders.Attach(ctx)

			defer func() {
				require.NoError(t, userDBTeardown())
				require.NoError(t, roleDBTeardown())
				cancel()
			}()

			test.test(ctxWithLoaders, t, usersTestCase{
				user:          test.user,
				db:            userDB,
				queryResolver: qr,
			})
		})
	}
}
