package local

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	app "github.com/rislah/fakes/internal"
)

type LocalUserDB interface {
	app.UserDB
	SetRoleDB(app.RoleDB)
}

type localDB struct {
	roleDB *localRoleDB
	users  []app.User
	// userRoles []app.Role
	// roles     []app.Role
}

func NewUserDB() *localDB {
	// roleDB := NewRoleDB()
	// roles := []app.Role{}
	// roles = append(roles, app.Role{ID: 0, Name: "guest"})
	return &localDB{
		// roles:  roles,
		// roleDB: roleDB,
	}
}

func MakeUserDB() (app.UserDB, func() error, error) {
	db := NewUserDB()
	db.roleDB = NewRoleDB()
	return db, db.flushAll, nil
}

func (ld *localDB) SetRoleDB(roleDB app.RoleDB) {
	if r, ok := roleDB.(*localRoleDB); ok {
		ld.roleDB = r
	}
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
	ld.users = append(ld.users, user)
	return ld.roleDB.CreateUserRole(ctx, user.UserID, 1)
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

func (ld *localDB) GetUsersByRoleID(ctx context.Context, roleID int) ([]app.User, error) {
	var (
		userRoles []app.Role
		users     []app.User
	)

	for _, userRole := range ld.roleDB.userRoles {
		if userRole.ID == roleID {
			userRoles = append(userRoles, *userRole)
		}
	}

	for _, userRole := range userRoles {
		for _, user := range ld.users {
			if user.UserID == userRole.UserID {
				users = append(users, user)
			}
		}
	}

	return users, nil
}

func (ld *localDB) GetUsersByRoleIDs(ctx context.Context, roleIDs []int) (map[int][]app.User, error) {
	results := map[int][]app.User{}
	userIDsFound := map[int][]string{}

	for _, userRole := range ld.roleDB.userRoles {
		for _, roleID := range roleIDs {
			if userRole.ID == roleID {
				userIDsFound[userRole.ID] = append(userIDsFound[userRole.ID], userRole.UserID)
			}
		}
	}

	for _, roleID := range roleIDs {
		userIDs, ok := userIDsFound[roleID]
		if !ok {
			continue
		}

		for _, userID := range userIDs {
			for _, user := range ld.users {
				if user.UserID == userID {
					results[roleID] = append(results[roleID], user)
				}
			}
		}
	}

	// for roleID, userIDs := range userIDsFound {
	// 	for _, user := range ld.users {
	// 		for _, userID := range userIDs {
	// 			if user.UserID == userID {
	// 				results[roleID] = append(results[roleID], user)
	// 			}
	// 		}
	// 	}
	// }

	return results, nil
}

func (ld *localDB) flushAll() error {
	ld.users = ld.users[:0]
	return nil
}
