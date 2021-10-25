package loaders

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strconv"

	"github.com/graph-gophers/dataloader"
	"github.com/jmoiron/sqlx"
	app "github.com/rislah/fakes/internal"
)

const usersByIDs contextKey = "usersByIDs"
const usersByRoleID contextKey = "usersByRoleID"

func newUsersByIDsLoader(db *sqlx.DB) LoaderDetails {
	return LoaderDetails{
		batchLoadFn: func(c context.Context, k dataloader.Keys) []*dataloader.Result {
			keys := k.Keys()
			query, args, err := sqlx.In(`select user_id, username from users where user_id in (?);`, keys)
			if err != nil {
				log.Fatal(err)
			}

			query = db.Rebind(query)
			rows, err := db.Query(query, args...)
			if err != nil {
				log.Fatal(err)
			}

			users := []app.User{}
			for rows.Next() {
				user := app.User{}
				if err := rows.Scan(&user.UserID, &user.Username); err != nil {
					if err == sql.ErrNoRows {
						break
					} else {
						log.Fatal(err)
					}
				}

				users = append(users, user)
			}

			m := map[string]*dataloader.Result{}
			for _, user := range users {
				m[user.UserID] = &dataloader.Result{Data: user}
			}

			results := make([]*dataloader.Result, 0, len(k))
			for _, key := range k {
				result, found := m[key.String()]
				if !found {
					result = &dataloader.Result{}
				}
				results = append(results, result)
			}

			return results
		},
	}
}

func PrimeUsers(ctx context.Context, usr *app.User) {
	loader, _ := extractLoader(ctx, usersByIDs)
	if loader == nil || usr == nil {
		return
	}

	loader.Prime(ctx, dataloader.StringKey(usr.UserID), usr)
}

func newUsersByRoleIDLoader(userDB app.UserDB) LoaderDetails {
	return LoaderDetails{
		batchLoadFn: func(ctx context.Context, k dataloader.Keys) []*dataloader.Result {
			results := make([]*dataloader.Result, 0, len(k))

			keysInt := []int{}
			for _, k := range k.Keys() {
				keyInt, _ := strconv.Atoi(k)
				keysInt = append(keysInt, keyInt)
			}

			users, err := userDB.GetUsersByRoleIDs(ctx, keysInt)
			if err != nil {
				results = append(results, &dataloader.Result{Error: err})
				return results
			}

			m := map[int][]*dataloader.Result{}
			for _, user := range users {
				m[user.Role.ID] = append(m[user.Role.ID], &dataloader.Result{Data: user})
			}

			for _, key := range keysInt {
				result, found := m[key]
				if !found {
					result = []*dataloader.Result{}
				}

				mergedData := []*app.UserRole{}
				for _, r := range result {
					data := r.Data.(*app.UserRole)
					mergedData = append(mergedData, data)
				}

				results = append(results, &dataloader.Result{Data: mergedData})
			}

			return results
		},
	}
}

func LoadUsersByRoleID(ctx context.Context, roleID int) ([]*app.User, error) {
	loader, err := extractLoader(ctx, usersByRoleID)
	if err != nil {
		return nil, err
	}

	if loader == nil {
		return nil, fmt.Errorf("loader is nil")
	}

	resp, err := loader.Load(ctx, dataloader.StringKey(strconv.Itoa(roleID)))()
	if err != nil {
		return nil, err
	}

	userRoles, ok := resp.([]*app.UserRole)
	if !ok {
		return nil, fmt.Errorf("wrong type: %s", reflect.TypeOf(resp))
	}

	users := make([]*app.User, 0, len(userRoles))
	for _, ur := range userRoles {
		users = append(users, &ur.User)
	}
	return users, nil
}
