package tests

import (
	"context"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/loaders"
	"github.com/rislah/fakes/resolvers/queries"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roleTestCase struct {
	data *app.Data
}

func TestRoles(t *testing.T, makeRoleDB MakeRoleDB, makeUserDB MakeUserDB) {
	tests := []struct {
		scenario string
		test     func(ctx context.Context, testCase roleTestCase)
	}{
		{
			scenario: "return all",
			test: func(ctx context.Context, testCase roleTestCase) {
				roles, err := queries.NewRoleListResolver(ctx, testCase.data)
				assert.NoError(t, err)
				assert.NotEmpty(t, roles)
			},
		},
		{
			scenario: "return by name if exists",
			test: func(ctx context.Context, testCase roleTestCase) {
				role, err := queries.NewRoleResolverByName(ctx, testCase.data, "guest")
				assert.NoError(t, err)
				assert.NotNil(t, role)
				assert.Equal(t, "GUEST", role.Name())
			},
		},
		{
			scenario: "return by name if not exists",
			test: func(ctx context.Context, testCase roleTestCase) {
				role, err := queries.NewRoleResolverByName(ctx, testCase.data, "guesta")
				assert.NoError(t, err)
				assert.Nil(t, role)
			},
		},
		{
			scenario: "return by id if exists",
			test: func(ctx context.Context, testCase roleTestCase) {
				role, err := queries.NewRoleResolverByID(ctx, testCase.data, "1")
				assert.NoError(t, err)
				assert.Equal(t, "GUEST", role.Name())
			},
		},
		{
			scenario: "return by id if not exists",
			test: func(ctx context.Context, testCase roleTestCase) {
				role, err := queries.NewRoleResolverByID(ctx, testCase.data, "0")
				assert.NoError(t, err)
				assert.Nil(t, role)
			},
		},
		{
			scenario: "return by userid if exists",
			test: func(ctx context.Context, testCase roleTestCase) {
				err := testCase.data.RoleDB.CreateUserRole(ctx, "123", 1)
				assert.NoError(t, err)
				role, err := queries.NewRoleResolverByUserID(ctx, testCase.data, "123")
				assert.NoError(t, err)
				assert.Equal(t, graphql.ID("1"), role.ID())
			},
		},
		{
			scenario: "return by userid if not exists",
			test: func(ctx context.Context, testCase roleTestCase) {
				role, err := queries.NewRoleResolverByUserID(ctx, testCase.data, "1233")
				assert.NoError(t, err)
				assert.Nil(t, role)
			},
		},
		{
			scenario: "return all users",
			test: func(ctx context.Context, testCase roleTestCase) {
				testCase.data.UserDB.CreateUser(ctx, app.User{
					Username: "kasutaja",
					Password: "parool",
				})

				role, err := queries.NewRoleResolverByName(ctx, testCase.data, "guest")
				assert.NoError(t, err)
				assert.NotEmpty(t, role)

				users, err := role.Users(ctx)
				assert.NoError(t, err)
				assert.Len(t, *users, 1)

				for _, user := range *users {
					assert.NotEmpty(t, user)
					assert.Equal(t, "kasutaja", user.Username())
				}
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
			}

			if u, ok := userDB.(local.LocalUserDB); ok {
				u.SetRoleDB(roleDB)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			lds := loaders.New(data, nil, nil, nil)
			ctxWithLoaders := lds.Attach(ctx)

			defer func() {
				cancel()
				require.NoError(t, roleDBTeardown())
				require.NoError(t, userDBTeardown())
			}()

			test.test(ctxWithLoaders, roleTestCase{
				data: data,
			})
		})
	}
}
