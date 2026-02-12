package middleware

import (
	"github.com/cstone-io/twine/pkg/kit"
)

// Middleware wraps a HandlerFunc to add functionality
type Middleware func(kit.HandlerFunc) kit.HandlerFunc

// ApplyMiddlewares chains multiple middlewares together
func ApplyMiddlewares(h kit.HandlerFunc, middlewares ...Middleware) kit.HandlerFunc {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

// Chain combines multiple middlewares into a single middleware
// Useful for composing middlewares in layout files
func Chain(middlewares ...Middleware) Middleware {
	return func(next kit.HandlerFunc) kit.HandlerFunc {
		return ApplyMiddlewares(next, middlewares...)
	}
}
