package loaders

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"

	"github.com/jmoiron/sqlx"

	"github.com/graph-gophers/dataloader"
	app "github.com/rislah/fakes/internal"
)

const roleByUserID contextKey = "roleByUserID"
const rolesByNames contextKey = "rolesByNames"
const rolesByIDs contextKey = "rolesByIDs"

func newRoleByUserID(userBackend app.UserBackend) LoaderDetails {
	return LoaderDetails{
		batchLoadFn: func(ctx context.Context, k dataloader.Keys) []*dataloader.Result {
			keys := k.Keys()
			results := make([]*dataloader.Result, 0, len(k))

			roles, err := userBackend.GetUserRolesByUserIDs(ctx, keys)
			if err != nil {
				for i := range results {
					results[i] = &dataloader.Result{Error: err}
				}

				return results
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

func newRolesByNamesLoader(db *sqlx.DB) LoaderDetails {
	return LoaderDetails{
		options: []dataloader.Option{
			dataloader.WithBatchCapacity(100),
		},
		batchLoadFn: func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
			var (
				results = make([]*dataloader.Result, 0, len(keys))
				roles   = []*app.Role{}
			)
			query, args, err := sqlx.In(`
				select id as role_id, name
				from role
				where name in (?);
		`, keys.Keys())
			if err != nil {
				results = append(results, &dataloader.Result{Error: err})
				return results
			}

			query = db.Rebind(query)
			rows, err := db.Query(query, args...)
			if err != nil {
				results = append(results, &dataloader.Result{Error: err})
				return results
			}

			for rows.Next() {
				role := app.Role{}
				if err := rows.Scan(&role.ID, &role.Name); err != nil {
					if err == sql.ErrNoRows {
						return []*dataloader.Result{}
					}
					results = append(results, &dataloader.Result{Error: err})
					return results

				}
				roles = append(roles, &role)
			}

			m := map[string]*dataloader.Result{}
			for _, role := range roles {
				m[role.Name.String()] = &dataloader.Result{Data: role}
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

func newRolesByIDsLoader(db *sqlx.DB) LoaderDetails {
	return LoaderDetails{
		options: []dataloader.Option{
			dataloader.WithBatchCapacity(100),
		},
		batchLoadFn: func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
			var (
				results = make([]*dataloader.Result, 0, len(keys))
				roles   = []*app.Role{}
			)
			query, args, err := sqlx.In(`
				select id as role_id, name
				from role
				where id in (?);
		`, keys.Keys())
			if err != nil {
				results = append(results, &dataloader.Result{Error: err})
				return results
			}

			query = db.Rebind(query)
			rows, err := db.Query(query, args...)
			if err != nil {
				results = append(results, &dataloader.Result{Error: err})
				return results
			}

			for rows.Next() {
				role := app.Role{}
				if err := rows.Scan(&role.ID, &role.Name); err != nil {
					if err == sql.ErrNoRows {
						return []*dataloader.Result{}
					}
					results = append(results, &dataloader.Result{Error: err})
					return results

				}
				roles = append(roles, &role)
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
		return nil, fmt.Errorf("null")
	}

	res, err := loader.Load(ctx, dataloader.StringKey(name))()
	if err != nil {
		return nil, err
	}

	role, ok := res.(*app.Role)
	if !ok {
		return nil, fmt.Errorf("loadByName: wrong type: %s", reflect.TypeOf(res))
	}

	return role, nil
}

func LoadRoleByID(ctx context.Context, id string) (*app.Role, error) {
	loader, _ := extractLoader(ctx, rolesByIDs)
	if loader == nil {
		return nil, fmt.Errorf("null")
	}

	res, err := loader.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	role, ok := res.(*app.Role)
	if !ok {
		return nil, fmt.Errorf("loadRoleByID: wrong type: %s", reflect.TypeOf(res))
	}

	nameLoader, _ := extractLoader(ctx, rolesByNames)
	if nameLoader != nil && role != nil && role.Name != "" {
		nameLoader.Prime(ctx, dataloader.StringKey(role.Name), role)
	}

	// loadByType, _ := ExtractLoader(ctx, roleByRoleType)
	// loadByType.Prime(ctx, dataloader.StringKey(role.Name), *role)

	return role, nil
}

func LoadRoleByUserID(ctx context.Context, userID string) (app.Role, error) {
	loader, err := extractLoader(ctx, roleByUserID)
	if err != nil {
		return app.Role{}, err
	}
	if loader == nil || userID == "" {
		return app.Role{}, fmt.Errorf("null")
	}

	res, err := loader.Load(ctx, dataloader.StringKey(userID))()
	if err != nil {
		return app.Role{}, err
	}

	role, ok := res.(app.Role)
	if !ok {
		return app.Role{}, fmt.Errorf("loadRoleByUseRID: wrong type: %s", reflect.TypeOf(res))
	}

	// loader.Prime(ctx, dataloader.StringKey(role.), res)

	return role, nil
}
