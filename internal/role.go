package app

import (
	"context"
)

type RoleType string

func (r RoleType) String() string {
	return string(r)
}

const (
	DeveloperRoleType RoleType = "developer"
	UserRoleType      RoleType = "user"
	GuestRoleType     RoleType = "guest"
	AdminRoleType     RoleType = "admin"
)

type Role struct {
	ID     int      `db:"role_id"`
	UserID string   `db:"user_id"`
	Name   RoleType `db:"name"`
}

type RoleBackend interface {
	CreateUserRole(ctx context.Context, userID string, roleID int) error
	GetRoles(ctx context.Context, args RolesQueryArgs) ([]Role, error)
	GetUserRoleByUserID(ctx context.Context, userID string) (Role, error)
}

type RoleDB interface {
	CreateUserRole(ctx context.Context, userID string, roleID int) error
	GetRoles(ctx context.Context) ([]Role, error)
	GetRolesByIDs(ctx context.Context, ids []int) ([]Role, error)
	GetRolesByNames(ctx context.Context, names []string) ([]Role, error)
	GetRolesByUserIDs(ctx context.Context, userIDs []string) ([]Role, error)
	GetUserRoleByUserID(ctx context.Context, userID string) (Role, error)
	GetUserRolesByUserIDs(ctx context.Context, userIDs []string) ([]Role, error)
}

type RolesQueryArgs struct {
	IDs             []int
	Names           []string
	UserIDs         []string
	UserRoleUserIDs []string
	UserRoleUserID  string
}

type roleImpl struct {
	roleDB RoleDB
}

func NewRoleBackend(roleDB RoleDB) *roleImpl {
	return &roleImpl{
		roleDB,
	}
}
