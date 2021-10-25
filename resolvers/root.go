package resolvers

import (
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/resolvers/mutations"
	"github.com/rislah/fakes/resolvers/queries"
)

type RootResolver struct {
	*queries.QueryResolver
	*mutations.MutationResolver
}

func NewRootResolver(data *app.Data) *RootResolver {
	return &RootResolver{
		&queries.QueryResolver{Data: data},
		&mutations.MutationResolver{Data: data},
	}
}
