package loaders

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/graph-gophers/dataloader"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/errors"
)

const rolesByUserIDs contextKey = "roleByUserID"
const rolesByNames contextKey = "rolesByNames"
const rolesByIDs contextKey = "rolesByIDs"

func newRolesByUserIDs(role app.RoleBackend) LoaderDetails {
	return LoaderDetails{
		batchLoadFn: func(ctx context.Context, k dataloader.Keys) []*dataloader.Result {
			results := make([]*dataloader.Result, 0, len(k))

			args := app.RolesQueryArgs{
				UserIDs: k.Keys(),
			}

			roles, err := role.GetRoles(ctx, args)
			if err != nil {
				return fillKeysWithError(k, err)
			}

			m := map[string]*dataloader.Result{}
			for _, role := range roles {
				m[role.UserID] = &dataloader.Result{Data: role}
			}

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

func newRolesByNamesLoader(role app.RoleBackend) LoaderDetails {
	return LoaderDetails{
		batchLoadFn: func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
			args := app.RolesQueryArgs{
				Names: keys.Keys(),
			}
			roles, err := role.GetRoles(ctx, args)
			if err != nil {
				return fillKeysWithError(keys, err)
			}

			m := map[string]*dataloader.Result{}
			for _, role := range roles {
				m[role.Name.String()] = &dataloader.Result{Data: role}
			}

			var results = make([]*dataloader.Result, 0, len(keys))
			for _, key := range keys {
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

func newRolesByIDsLoader(role app.RoleBackend) LoaderDetails {
	return LoaderDetails{
		options: []dataloader.Option{
			dataloader.WithBatchCapacity(100),
		},
		batchLoadFn: func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
			var (
				results = make([]*dataloader.Result, 0, len(keys))
				intKeys = make([]int, len(keys))
			)

			for i, key := range keys {
				keyInt, err := strconv.Atoi(key.String())
				if err != nil {
					return fillKeysWithError(keys, err)
				}
				intKeys[i] = keyInt
			}

			args := app.RolesQueryArgs{
				IDs: intKeys,
			}

			roles, err := role.GetRoles(ctx, args)
			if err != nil {
				return fillKeysWithError(keys, err)
			}

			m := map[string]*dataloader.Result{}
			for _, role := range roles {
				m[strconv.Itoa(role.ID)] = &dataloader.Result{Data: role}
			}

			for _, key := range keys {
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

func LoadRoleByName(ctx context.Context, name string) (*app.Role, error) {
	loader, _ := extractLoader(ctx, rolesByNames)
	if loader == nil || name == "" {
		return nil, nil
	}

	res, err := loader.Load(ctx, dataloader.StringKey(name))()
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	role, ok := res.(app.Role)
	if !ok {
		fmt.Println(reflect.TypeOf(res))
		return nil, errors.New("unexpected type cast error")
	}

	return &role, nil
}

func LoadRoleByID(ctx context.Context, id string) (*app.Role, error) {
	loader, _ := extractLoader(ctx, rolesByIDs)
	if loader == nil || id == "" {
		return nil, nil
	}

	res, err := loader.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	role, ok := res.(app.Role)
	if !ok {
		return nil, errors.New("unexpected type cast error")
	}

	rolesByNamesLoader, _ := extractLoader(ctx, rolesByNames)
	if rolesByNamesLoader != nil && role.Name != "" {
		rolesByNamesLoader.Prime(ctx, dataloader.StringKey(role.Name), role)
	}

	return &role, nil
}

func LoadRoleByUserID(ctx context.Context, userID string) (*app.Role, error) {
	loader, _ := extractLoader(ctx, rolesByUserIDs)
	if loader == nil || userID == "" {
		return nil, nil
	}

	res, err := loader.Load(ctx, dataloader.StringKey(userID))()
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	role, ok := res.(app.Role)
	if !ok {
		return nil, errors.New("unexpected type cast error")
	}

	return &role, nil
}

func PrimeRoles(ctx context.Context, roles []app.Role) {
	byUserIDLoader, _ := extractLoader(ctx, rolesByUserIDs)
	byNamesLoader, _ := extractLoader(ctx, rolesByNames)
	byIDsLoader, _ := extractLoader(ctx, rolesByIDs)
	for _, role := range roles {
		if role.Name.String() != "" {
			byNamesLoader.Prime(ctx, dataloader.StringKey(strings.ToLower(role.Name.String())), role)
		}

		if role.ID > 0 {
			byIDsLoader.Prime(ctx, dataloader.StringKey(strconv.Itoa(role.ID)), role)
		}

		if role.UserID != "" {
			byUserIDLoader.Prime(ctx, dataloader.StringKey(role.UserID), role)
		}
	}
}
