package api

import (
	"context"
	"net/http"

	app "github.com/rislah/fakes/internal"
)

type ApiFunc func(ctx context.Context, response *Response, request *http.Request) error

type Route struct {
	handler     ApiFunc
	method      string
	module      *RouteModule
	path        string
	permissions []string
	role        app.Role
}

func NewRoute(path string, handler ApiFunc, method string) *Route {
	return &Route{
		path:    path,
		handler: handler,
		method:  method,
	}
}

func (r *Route) Permissions(permissions ...string) *Route {
	r.permissions = permissions
	return r
}

func (r *Route) Role(role app.Role) *Route {
	r.role = role
	return r
}
