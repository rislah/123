package queries

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/graph-gophers/graphql-go"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/loaders"
)

type RoleResolver struct {
	role *app.Role
	data *app.Data

	id string
}

type QueryRoleArgs struct {
	ID   *graphql.ID
	Name *app.RoleType
}

func NewRoleResolver(r *app.Role, data *app.Data) *RoleResolver {
	return &RoleResolver{role: r, data: data, id: strconv.Itoa(r.ID)}
}

func (r *QueryResolver) Roles(ctx context.Context) ([]*RoleResolver, error) {
	var roles []app.Role

	err := r.Data.DB.SelectContext(ctx, &roles, `select id as role_id, name from role`)
	if err != nil {
		return nil, err
	}

	roleResolvers := make([]*RoleResolver, 0, len(roles))
	for _, role := range roles {
		role := role
		resolver := NewRoleResolver(&role, r.Data)
		roleResolvers = append(roleResolvers, resolver)
	}

	return roleResolvers, nil
}

func (q *QueryResolver) Role(ctx context.Context, args QueryRoleArgs) (*RoleResolver, error) {
	if args.Name != nil {
		return NewRoleResolverByName(ctx, q.Data, strings.ToLower(args.Name.String()))
	}

	if args.ID != nil {
		return NewRoleResolverByID(ctx, q.Data, string(*args.ID))
	}

	return nil, nil
}

func NewRoleResolverByName(ctx context.Context, data *app.Data, name string) (*RoleResolver, error) {
	role, err := loaders.LoadRoleByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return NewRoleResolver(role, data), nil
}

func NewRoleResolverByID(ctx context.Context, data *app.Data, id string) (*RoleResolver, error) {
	role, err := loaders.LoadRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role == nil {
		return nil, nil
	}

	return NewRoleResolver(role, data), nil
}

func NewRoleResolverByUserID(ctx context.Context, data *app.Data, userID string) *RoleResolver {
	role, err := loaders.LoadRoleByUserID(ctx, userID)
	if err != nil {
		log.Fatal(err)
	}

	return NewRoleResolver(&role, data)
}

func (r *RoleResolver) ID() graphql.ID {
	return graphql.ID(r.id)
}

func (r *RoleResolver) Name() string {
	return strings.ToUpper(r.role.Name.String())
}

func (r *RoleResolver) Users(ctx context.Context) (*[]*UserResolver, error) {
	return NewUsersByRoleIDResolver(ctx, r.data, r.role.ID)
}
