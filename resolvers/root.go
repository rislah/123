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

func NewRootResolver(backend *app.Backend) *RootResolver {
	return &RootResolver{
		&queries.QueryResolver{Backend: backend},
		&mutations.MutationResolver{Backend: backend},
	}
}
