package loaders

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/graph-gophers/dataloader"
	app "github.com/rislah/fakes/internal"
)

const usersByIDs contextKey = "usersByIDs"
const usersByRoleID contextKey = "usersByRoleID"

func newUsersByIDsLoader(user app.UserBackend) LoaderDetails {
	return LoaderDetails{
		batchLoadFn: func(ctx context.Context, k dataloader.Keys) []*dataloader.Result {
			args := app.UsersQueryArgs{
				UserIDs: k.Keys(),
			}
			users, err := user.GetUsers(ctx, args)
			if err != nil {
				return fillKeysWithError(k, err)
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

func newUsersByRoleIDLoader(user app.UserBackend) LoaderDetails {
	return LoaderDetails{
		batchLoadFn: func(ctx context.Context, k dataloader.Keys) []*dataloader.Result {
			keysInt := []int{}
			for _, k := range k.Keys() {
				keyInt, _ := strconv.Atoi(k)
				keysInt = append(keysInt, keyInt)
			}

			users, err := user.GetUsersByRoleIDs(ctx, keysInt)
			if err != nil {
				return fillKeysWithError(k, err)
			}

			m := map[int]*dataloader.Result{}
			for roleID, roleUsers := range users {
				m[roleID] = &dataloader.Result{Data: roleUsers}
			}

			results := make([]*dataloader.Result, 0, len(k))
			for _, key := range keysInt {
				result, found := m[key]
				if !found {
					result = &dataloader.Result{}
				}

				results = append(results, result)
			}

			return results
		},
	}
}

func LoadUsersByRoleID(ctx context.Context, roleID int) ([]app.User, error) {
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

	users, ok := resp.([]app.User)
	if !ok {
		return nil, fmt.Errorf("wrong type: %s", reflect.TypeOf(resp))
	}

	return users, nil
}
