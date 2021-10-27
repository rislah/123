package tests

import (
	"context"
	"testing"
	"time"

	app "github.com/rislah/fakes/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleDB(t *testing.T, makeRoleDB MakeRoleDB) {
	tests := []struct {
		scenario string
		role     app.Role
		test     func(ctx context.Context, t *testing.T, roleDB app.RoleDB)
	}{
		{
			scenario: "CreateUserRole",
			test: func(ctx context.Context, t *testing.T, roleDB app.RoleDB) {
				userID := "1"
				err := roleDB.CreateUserRole(ctx, userID, 1)
				assert.NoError(t, err)
			},
		},
		{
			scenario: "GetRoles when exists",
			test: func(ctx context.Context, t *testing.T, roleDB app.RoleDB) {
				roles, err := roleDB.GetRoles(ctx)
				assert.NoError(t, err)
				assert.NotEmpty(t, roles)
			},
		},
		{
			scenario: "GetRolesByIDs when exists",
			test: func(ctx context.Context, t *testing.T, roleDB app.RoleDB) {
				roles, err := roleDB.GetRolesByIDs(ctx, []int{1})
				assert.NoError(t, err)
				assert.NotEmpty(t, roles)
			},
		},
		{
			scenario: "GetRolesByIDs when doesnt exists",
			test: func(ctx context.Context, t *testing.T, roleDB app.RoleDB) {
				roles, err := roleDB.GetRolesByIDs(ctx, []int{52})
				assert.NoError(t, err)
				assert.Empty(t, roles)
			},
		},
		{
			scenario: "GetRolesByNames when exists",
			test: func(ctx context.Context, t *testing.T, roleDB app.RoleDB) {
				roles, err := roleDB.GetRolesByNames(ctx, []string{"guest"})
				assert.NoError(t, err)
				assert.NotEmpty(t, roles)
			},
		},
		{
			scenario: "GetRolesByNames when doesnt exists",
			test: func(ctx context.Context, t *testing.T, roleDB app.RoleDB) {
				roles, err := roleDB.GetRolesByNames(ctx, []string{"asdfsdf"})
				assert.NoError(t, err)
				assert.Len(t, roles, 0)
			},
		},
		{
			scenario: "GetRolesByUserIDs when exists",
			test: func(ctx context.Context, t *testing.T, roleDB app.RoleDB) {
				userID := "1"
				err := roleDB.CreateUserRole(ctx, userID, 1)
				assert.NoError(t, err)

				roles, err := roleDB.GetRolesByUserIDs(ctx, []string{userID})
				assert.NoError(t, err)
				assert.NotEmpty(t, roles)
			},
		},
		{
			scenario: "GetRolesByUserIDs when doesnt exists",
			test: func(ctx context.Context, t *testing.T, roleDB app.RoleDB) {
				roles, err := roleDB.GetRolesByUserIDs(ctx, []string{"asd"})
				assert.NoError(t, err)
				assert.Empty(t, roles)
			},
		},
		{
			scenario: "GetUserRoleByUserID when doesnt exists",
			test: func(ctx context.Context, t *testing.T, roleDB app.RoleDB) {
				roles, err := roleDB.GetUserRoleByUserID(ctx, "1")
				assert.NoError(t, err)
				assert.Empty(t, roles)
			},
		},
		{
			scenario: "GetUserRoleByUserID when exists",
			test: func(ctx context.Context, t *testing.T, roleDB app.RoleDB) {
				userID := "1"
				err := roleDB.CreateUserRole(ctx, userID, 1)
				assert.NoError(t, err)

				roles, err := roleDB.GetUserRoleByUserID(ctx, "1")
				assert.NoError(t, err)
				assert.NotEmpty(t, roles)
			},
		},
		{
			scenario: "GetUserRolesByUserIDs when exists",
			test: func(ctx context.Context, t *testing.T, roleDB app.RoleDB) {
				userID := "1"
				err := roleDB.CreateUserRole(ctx, userID, 1)
				assert.NoError(t, err)

				roles, err := roleDB.GetUserRolesByUserIDs(ctx, []string{userID})
				assert.NoError(t, err)
				assert.NotEmpty(t, roles)
			},
		},
		{
			scenario: "GetUserRolesByUserIDs when doesnt exists",
			test: func(ctx context.Context, t *testing.T, roleDB app.RoleDB) {
				roles, err := roleDB.GetUserRolesByUserIDs(ctx, []string{"1"})
				assert.NoError(t, err)
				assert.Len(t, roles, 0)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			roleDB, roleDBTeardown, err := makeRoleDB()
			require.NoError(t, err)

			defer roleDBTeardown()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			test.test(ctx, t, roleDB)
		})
	}
}
