package queries

import (
	"context"
	"strconv"
	"strings"

	"github.com/graph-gophers/graphql-go"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/loaders"
)

type RoleResolver struct {
	role    *app.Role
	backend *app.Backend
}

type QueryRoleArgs struct {
	ID   *graphql.ID
	Name *app.RoleType
}

func NewRoleResolver(r *app.Role, data *app.Backend) *RoleResolver {
	return &RoleResolver{role: r, backend: data}
}

func (r *QueryResolver) Roles(ctx context.Context) ([]*RoleResolver, error) {
	return NewRoleListResolver(ctx, r.Backend)
}

func (q *QueryResolver) Role(ctx context.Context, args QueryRoleArgs) (*RoleResolver, error) {
	if args.Name != nil {
		return NewRoleResolverByName(ctx, q.Backend, strings.ToLower(args.Name.String()))
	}

	if args.ID != nil {
		return NewRoleResolverByID(ctx, q.Backend, string(*args.ID))
	}

	return nil, nil
}

func NewRoleListResolver(ctx context.Context, data *app.Backend) ([]*RoleResolver, error) {
	roles, err := data.Role.GetRoles(ctx, app.RolesQueryArgs{})
	if err != nil {
		return nil, err
	}

	loaders.PrimeRoles(ctx, roles)

	roleResolvers := make([]*RoleResolver, 0, len(roles))
	for _, role := range roles {
		roleResolvers = append(roleResolvers, NewRoleResolver(&role, data))
	}

	return roleResolvers, nil
}

func NewRoleResolverByName(ctx context.Context, data *app.Backend, name string) (*RoleResolver, error) {
	role, err := loaders.LoadRoleByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if role == nil {
		return nil, nil
	}

	return NewRoleResolver(role, data), nil
}

func NewRoleResolverByID(ctx context.Context, data *app.Backend, id string) (*RoleResolver, error) {
	role, err := loaders.LoadRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role == nil {
		return nil, nil
	}

	return NewRoleResolver(role, data), nil
}

func NewRoleResolverByUserID(ctx context.Context, data *app.Backend, userID string) (*RoleResolver, error) {
	role, err := loaders.LoadRoleByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if role == nil {
		return nil, nil
	}

	return NewRoleResolver(role, data), nil
}

func (r *RoleResolver) ID() graphql.ID {
	return graphql.ID(strconv.Itoa(r.role.ID))
}

func (r *RoleResolver) Name() string {
	return strings.ToUpper(r.role.Name.String())
}

func (r *RoleResolver) Users(ctx context.Context) (*[]*UserResolver, error) {
	return NewUsersByRoleIDResolver(ctx, r.backend, r.role.ID)
}
