package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/jwt"
	"github.com/rislah/fakes/internal/logger"
)

type RouteModule struct {
	routes     []*Route
	jwtWrapper jwt.Wrapper
	log        *logger.Logger
}

func NewRouteModule(jwtWrapper jwt.Wrapper) *RouteModule {
	return &RouteModule{
		jwtWrapper: jwtWrapper,
	}
}

func (r *RouteModule) InjectRoutes(mux *mux.Router) {
	for _, route := range r.routes {
		var handler http.Handler
		handler = r.wrap(route.handler, r.log)

		if len(route.permissions) != 0 {
			handler = route.authMiddleware(handler, r.jwtWrapper)
		}

		mux.Handle(route.path, handler).Methods(route.method)
	}
}

func (r *RouteModule) Get(path string, handler ApiFunc) *Route {
	route := &Route{
		handler: handler,
		path:    path,
		method:  "GET",
		module:  r,
	}
	r.routes = append(r.routes, route)
	return route
}

func (r *RouteModule) Post(path string, handler ApiFunc) *Route {
	route := &Route{
		handler: handler,
		path:    path,
		method:  "POST",
		module:  r,
	}
	r.routes = append(r.routes, route)
	return route
}

func (r *RouteModule) Put(path string, handler ApiFunc) *Route {
	route := &Route{
		handler: handler,
		path:    path,
		method:  "PUT",
		module:  r,
	}
	r.routes = append(r.routes, route)
	return route
}

func (r *RouteModule) wrap(handler ApiFunc, log *logger.Logger) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		resp := &Response{ResponseWriter: rw}
		err := handler(r.Context(), resp, r)
		if err != nil {
			if !resp.WasWritten() {
				resp.WriteHeader(http.StatusInternalServerError)
			}

			switch resp.Status() {
			case http.StatusInternalServerError:
				log.LogRequestError(err, r)
				resp.WriteJSON(errors.NewErrorResponse("Internal server error has occured", http.StatusInternalServerError))
			}

			return
		}

	})
}
