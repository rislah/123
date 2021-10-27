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
	data    *app.Data
	queries *queries.QueryResolver
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
			scenario: "return all users by role",
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
		{
			scenario: "return by name given proper args",
			test: func(ctx context.Context, testCase roleTestCase) {
				guestRole := app.GuestRoleType

				role, err := testCase.queries.Role(ctx, queries.QueryRoleArgs{Name: &guestRole})
				assert.NoError(t, err)

				assert.Equal(t, "GUEST", role.Name())
				assert.Equal(t, graphql.ID("1"), role.ID())
			},
		},
		{
			scenario: "return by id given proper args",
			test: func(ctx context.Context, testCase roleTestCase) {
				id := graphql.ID("1")

				role, err := testCase.queries.Role(ctx, queries.QueryRoleArgs{ID: &id})
				assert.NoError(t, err)

				assert.Equal(t, "GUEST", role.Name())
				assert.Equal(t, id, role.ID())
			},
		},
		{
			scenario: "return nil if no args",
			test: func(ctx context.Context, testCase roleTestCase) {
				role, err := testCase.queries.Role(ctx, queries.QueryRoleArgs{})
				assert.Nil(t, role)
				assert.NoError(t, err)
			},
		},
		{
			scenario: "return all roles resolver",
			test: func(ctx context.Context, testCase roleTestCase) {
				roles, err := testCase.queries.Roles(ctx)
				assert.NoError(t, err)
				assert.Len(t, roles, 1)
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

			lds := loaders.New(data)
			ctxWithLoaders := lds.Attach(ctx)

			defer func() {
				cancel()
				require.NoError(t, roleDBTeardown())
				require.NoError(t, userDBTeardown())
			}()

			test.test(ctxWithLoaders, roleTestCase{
				data:    data,
				queries: &queries.QueryResolver{Data: data},
			})
		})
	}
}

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
			loaders := loaders.New(data)
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
