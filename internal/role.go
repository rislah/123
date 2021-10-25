package app

import "context"

type RoleType string

func (r RoleType) String() string {
	return string(r)
}

const (
	DeveloperRoleType RoleType = "Developer"
	UserRoleType      RoleType = "User"
	GuestRoleType     RoleType = "Guest"
	AdminRoleType     RoleType = "Admin"
)

type Role struct {
	ID     int      `db:"role_id"`
	UserID string   `db:"user_id"`
	Name   RoleType `db:"name"`
}

type RoleDB interface {
	GetRoles(ctx context.Context) ([]*Role, error)
}
