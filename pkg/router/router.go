package router

import (
	"net/http"
	"sort"
	"sync"

	"github.com/cstone-io/twine/pkg/kit"
	"github.com/cstone-io/twine/pkg/logger"
	"github.com/cstone-io/twine/pkg/middleware"
)

// Router provides hierarchical routing with middleware support
type Router struct {
	mu sync.Mutex

	Prefix      string
	Routes      []Route
	Middlewares []middleware.Middleware

	Children []*Router
}

// NewRouter creates a new Router with the given URL prefix
func NewRouter(prefix string) *Router {
	return &Router{
		Prefix: trim(prefix),
		Routes: []Route{},
	}
}

// Sub adds a child router to this router
func (r *Router) Sub(sub *Router) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Children = append(r.Children, sub)
}

// Use adds middleware to this router
func (r *Router) Use(middlewares ...middleware.Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Middlewares = append(r.Middlewares, middlewares...)
}

func (r *Router) handle(method Method, pattern string, h kit.HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	route := NewRouteBuilder().Handler(h).Method(method).Pattern(pattern).Build()
	r.Routes = append(r.Routes, *route)
}

// Get registers a GET route
func (r *Router) Get(pattern string, h kit.HandlerFunc) {
	r.handle(GET, pattern, h)
}

// Post registers a POST route
func (r *Router) Post(pattern string, h kit.HandlerFunc) {
	r.handle(POST, pattern, h)
}

// Put registers a PUT route
func (r *Router) Put(pattern string, h kit.HandlerFunc) {
	r.handle(PUT, pattern, h)
}

// Delete registers a DELETE route
func (r *Router) Delete(pattern string, h kit.HandlerFunc) {
	r.handle(DELETE, pattern, h)
}

func (r *Router) initializeRoutes(prefix string, routes *[]Route) {
	for _, sub := range r.Children {
		fullPrefix := trim(prefix) + trim(sub.Prefix)
		sub.Middlewares = append(sub.Middlewares, r.Middlewares...)
		sub.initializeRoutes(fullPrefix, routes)
	}

	for _, route := range r.Routes {
		finalHandler := kit.Handler(middleware.ApplyMiddlewares(route.Handler, r.Middlewares...))
		revisedRoute := route.Builder().Prefix(prefix + route.Prefix).HTTPHandler(finalHandler).Build()
		*routes = append(*routes, *revisedRoute)
	}
}

// InitializeAsRoot finalizes the router tree and returns an http.ServeMux
func (r *Router) InitializeAsRoot() *http.ServeMux {
	mux := http.NewServeMux()

	routes := []Route{}
	r.initializeRoutes(r.Prefix, &routes)

	// Sort routes by path length (longest first) for proper route matching
	sort.SliceStable(routes, func(a, b int) bool {
		return len(routes[a].Path()) > len(routes[b].Path())
	})

	r.Routes = routes

	for _, route := range routes {
		logger.Get().Debug("Registering route: %s", route.FullPath())
		mux.HandleFunc(route.FullPath(), route.HTTPHandler)
	}

	return mux
}
