package local

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	app "github.com/rislah/fakes/internal"
)

type localRoleDB struct {
	roles     []*app.Role
	userRoles []*app.Role
	mu        sync.RWMutex
}

func NewRoleDB() *localRoleDB {
	return &localRoleDB{}
}

func MakeRoleDB() (app.RoleDB, func() error, error) {
	db := NewRoleDB()
	db.CreateRole(context.Background(), app.Role{ID: 1, Name: app.GuestRoleType})
	return db, db.flushAll, nil
}

var _ app.RoleDB = &localRoleDB{}

func (l *localRoleDB) CreateRole(ctx context.Context, role app.Role) error {
	var found bool
	for _, value := range l.roles {
		if value.Name == role.Name {
			found = true
			break
		}
	}

	if found {
		return errors.New("unique constraint error")
	}

	l.roles = append(l.roles, &role)
	return nil
}

func (l *localRoleDB) CreateUserRole(ctx context.Context, userID string, roleID int) error {
	var found bool
	for _, value := range l.userRoles {
		if value.UserID == userID {
			found = true
			break
		}
	}

	if found {
		return errors.New("unique constraint error")
	}

	l.userRoles = append(l.userRoles, &app.Role{
		ID:     roleID,
		UserID: userID,
	})

	return nil
}

func (l *localRoleDB) GetRoles(ctx context.Context) ([]*app.Role, error) {
	return l.roles, nil
}

func (l *localRoleDB) GetRolesByIDs(ctx context.Context, ids []int) ([]*app.Role, error) {
	roles := []*app.Role{}
	for _, role := range l.roles {
		for _, id := range ids {
			if role.ID == id {
				roles = append(roles, role)
			}
		}
	}

	return roles, nil
}

func (l *localRoleDB) GetRolesByNames(ctx context.Context, names []string) ([]*app.Role, error) {
	roles := []*app.Role{}
	for _, role := range l.roles {
		for _, name := range names {
			if role.Name.String() == name {
				roles = append(roles, role)
			}
		}
	}

	return roles, nil
}

func (l *localRoleDB) GetRolesByUserIDs(ctx context.Context, userIDs []string) ([]*app.Role, error) {
	roles := []*app.Role{}
	for _, role := range l.roles {
		for _, userID := range userIDs {
			if role.UserID == userID {
				roles = append(roles, role)
			}
		}
	}

	return roles, nil
}

func (l *localRoleDB) GetUserRoleByUserID(ctx context.Context, userID string) (*app.Role, error) {
	for _, userRole := range l.userRoles {
		if userRole.UserID == userID {
			return userRole, nil
		}
	}
	return nil, nil
}

func (l *localRoleDB) GetUserRolesByUserIDs(ctx context.Context, userIDs []string) ([]*app.Role, error) {
	roles := []*app.Role{}
	for _, userRole := range l.userRoles {
		for _, userID := range userIDs {
			if userRole.UserID == userID {
				roles = append(roles, userRole)
			}
		}
	}

	for _, role := range roles {
		for _, r := range l.roles {
			if role.ID == r.ID {
				role.Name = r.Name
			}
		}
	}

	return roles, nil
}

func (l *localRoleDB) flushAll() error {
	l.roles = l.roles[:0]
	return nil
}
