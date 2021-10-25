package local

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	app "github.com/rislah/fakes/internal"
)

type localDB struct {
	users     []app.User
	userRoles []app.Role
	roles     []app.Role
}

func NewUserDB() *localDB {
	roles := []app.Role{}
	roles = append(roles, app.Role{ID: 0, Name: "guest"})
	return &localDB{
		roles: roles,
	}
}

func MakeUserDB() (app.UserDB, func() error, error) {
	db := NewUserDB()
	return NewUserDB(), db.flushAll, nil
}

var _ app.UserDB = &localDB{}

func (ld *localDB) CreateUser(ctx context.Context, user app.User) error {
	var found bool
	for _, value := range ld.users {
		if value.Username == user.Username {
			found = true
			break
		}
	}

	if found {
		return errors.New("unique constraint error")
	}

	user.UserID = uuid.NewString()
	ld.userRoles = append(ld.userRoles, app.Role{ID: 0, UserID: user.UserID})
	ld.users = append(ld.users, user)
	return nil
}

func (ld *localDB) GetUsersByIDs(ctx context.Context, userIDs []string) ([]app.User, error) {
	var users []app.User
	for _, usr := range ld.users {
		for _, userID := range userIDs {
			if usr.UserID == userID {
				users = append(users, usr)
			}
		}
	}

	return users, nil
}

func (ld *localDB) GetUserRoleByUserID(ctx context.Context, userID string) (app.Role, error) {
	var userRole app.Role

	for _, urs := range ld.userRoles {
		if urs.UserID == userID {
			userRole = urs
		}
	}

	return userRole, nil
}

func (ld *localDB) GetUserRolesByUserIDs(ctx context.Context, userIDs []string) ([]app.Role, error) {
	var userRoles []app.Role

	for _, userRole := range ld.userRoles {
		for _, userID := range userIDs {
			if userRole.UserID == userID {
				userRoles = append(userRoles, userRole)
			}
		}
	}

	return userRoles, nil
}

func (ld *localDB) GetUserByUsername(ctx context.Context, username string) (app.User, error) {
	for _, value := range ld.users {
		if value.Username == username {
			return value, nil
		}
	}

	return app.User{}, nil
}

func (ld *localDB) GetUsers(ctx context.Context) ([]app.User, error) {
	return ld.users, nil
}

func (ld *localDB) GetUsersByRoleID(ctx context.Context, roleID int) ([]*app.User, error) {
	var (
		userRoles []app.Role
		users     []*app.User
	)

	for _, userRole := range ld.userRoles {
		if userRole.ID == roleID {
			userRoles = append(userRoles, userRole)
		}
	}

	for _, userRole := range userRoles {
		for _, user := range ld.users {
			if user.UserID == userRole.UserID {
				users = append(users, &user)
			}
		}
	}

	return users, nil
}

func (ld *localDB) GetUsersByRoleIDs(ctx context.Context, roleIDs []int) ([]*app.UserRole, error) {
	return nil, nil
}

func (ld *localDB) flushAll() error {
	ld.users = ld.users[:0]
	return nil
}
