package router

import (
	"net/http"
	"strings"

	"github.com/cstone-io/twine/pkg/kit"
)

// Method represents an HTTP method with trailing space for ServeMux pattern matching
type Method string

const (
	GET    Method = "GET "
	POST   Method = "POST "
	PUT    Method = "PUT "
	DELETE Method = "DELETE "
)

// Route represents an HTTP route with handler and metadata
type Route struct {
	HTTPHandler http.HandlerFunc
	Handler     kit.HandlerFunc
	Method      Method
	Prefix      string
	Pattern     string
}

// Path returns the combined prefix and pattern
func (r *Route) Path() string {
	return r.Prefix + r.Pattern
}

// FullPath returns the method-prefixed path for ServeMux registration
func (r *Route) FullPath() string {
	return string(r.Method) + r.Path()
}

// Builder returns a RouteBuilder initialized with this route's values
func (r *Route) Builder() *RouteBuilder {
	return &RouteBuilder{
		httpHandler: r.HTTPHandler,
		handler:     r.Handler,
		method:      r.Method,
		prefix:      r.Prefix,
		pattern:     r.Pattern,
	}
}

// RouteBuilder provides a fluent interface for building Routes
type RouteBuilder struct {
	httpHandler http.HandlerFunc
	handler     kit.HandlerFunc
	method      Method
	prefix      string
	pattern     string
}

// NewRouteBuilder creates a new RouteBuilder instance
func NewRouteBuilder() *RouteBuilder {
	return &RouteBuilder{}
}

// HTTPHandler sets the http.HandlerFunc for this route
func (b *RouteBuilder) HTTPHandler(httpHandler http.HandlerFunc) *RouteBuilder {
	b.httpHandler = httpHandler
	return b
}

// Handler sets the kit.HandlerFunc for this route
func (b *RouteBuilder) Handler(handler kit.HandlerFunc) *RouteBuilder {
	b.handler = handler
	return b
}

// Method sets the HTTP method for this route
func (b *RouteBuilder) Method(method Method) *RouteBuilder {
	b.method = method
	return b
}

// Prefix sets the URL prefix for this route
func (b *RouteBuilder) Prefix(prefix string) *RouteBuilder {
	b.prefix = prefix
	return b
}

// Pattern sets the URL pattern for this route
func (b *RouteBuilder) Pattern(pattern string) *RouteBuilder {
	b.pattern = pattern
	return b
}

// Build constructs and returns the final Route
func (b *RouteBuilder) Build() *Route {
	return &Route{
		HTTPHandler: b.httpHandler,
		Handler:     b.handler,
		Method:      b.method,
		Prefix:      b.prefix,
		Pattern:     b.pattern,
	}
}

func trim(s string) string {
	return strings.TrimSuffix(s, "/")
}
