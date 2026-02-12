package router

import (
	"net/http"
	"testing"

	"github.com/cstone-io/twine/pkg/kit"
	"github.com/stretchr/testify/assert"
)

// TestMethod_Constants tests HTTP method constants
func TestMethod_Constants(t *testing.T) {
	t.Run("method constants have trailing space", func(t *testing.T) {
		assert.Equal(t, "GET ", string(GET))
		assert.Equal(t, "POST ", string(POST))
		assert.Equal(t, "PUT ", string(PUT))
		assert.Equal(t, "DELETE ", string(DELETE))
	})

	t.Run("methods are unique", func(t *testing.T) {
		methods := []Method{GET, POST, PUT, DELETE}
		seen := make(map[Method]bool)

		for _, m := range methods {
			assert.False(t, seen[m], "Method %s should be unique", m)
			seen[m] = true
		}
	})
}

// TestRoute_Path tests the Path method
func TestRoute_Path(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		pattern  string
		expected string
	}{
		{
			name:     "simple path",
			prefix:   "",
			pattern:  "/users",
			expected: "/users",
		},
		{
			name:     "path with prefix",
			prefix:   "/api",
			pattern:  "/users",
			expected: "/api/users",
		},
		{
			name:     "empty prefix and pattern",
			prefix:   "",
			pattern:  "",
			expected: "",
		},
		{
			name:     "path parameter",
			prefix:   "/api",
			pattern:  "/users/{id}",
			expected: "/api/users/{id}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := Route{
				Prefix:  tt.prefix,
				Pattern: tt.pattern,
			}

			result := route.Path()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRoute_FullPath tests the FullPath method
func TestRoute_FullPath(t *testing.T) {
	tests := []struct {
		name     string
		method   Method
		prefix   string
		pattern  string
		expected string
	}{
		{
			name:     "GET request",
			method:   GET,
			prefix:   "",
			pattern:  "/users",
			expected: "GET /users",
		},
		{
			name:     "POST request with prefix",
			method:   POST,
			prefix:   "/api",
			pattern:  "/users",
			expected: "POST /api/users",
		},
		{
			name:     "PUT request with path parameter",
			method:   PUT,
			prefix:   "/api",
			pattern:  "/users/{id}",
			expected: "PUT /api/users/{id}",
		},
		{
			name:     "DELETE request",
			method:   DELETE,
			prefix:   "",
			pattern:  "/users/{id}",
			expected: "DELETE /users/{id}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := Route{
				Method:  tt.method,
				Prefix:  tt.prefix,
				Pattern: tt.pattern,
			}

			result := route.FullPath()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRoute_Builder tests the Builder method
func TestRoute_Builder(t *testing.T) {
	t.Run("creates builder from route", func(t *testing.T) {
		handler := func(k *kit.Kit) error { return nil }
		httpHandler := func(w http.ResponseWriter, r *http.Request) {}

		route := Route{
			HTTPHandler: httpHandler,
			Handler:     handler,
			Method:      GET,
			Prefix:      "/api",
			Pattern:     "/users",
		}

		builder := route.Builder()

		assert.NotNil(t, builder)
		assert.Equal(t, GET, builder.method)
		assert.Equal(t, "/api", builder.prefix)
		assert.Equal(t, "/users", builder.pattern)
	})

	t.Run("builder can modify route", func(t *testing.T) {
		originalRoute := Route{
			Method:  GET,
			Prefix:  "/api",
			Pattern: "/users",
		}

		modifiedRoute := originalRoute.Builder().
			Prefix("/admin").
			Build()

		assert.Equal(t, "/admin", modifiedRoute.Prefix)
		assert.Equal(t, "/users", modifiedRoute.Pattern)
		assert.Equal(t, GET, modifiedRoute.Method)
	})
}

// TestRouteBuilder_NewRouteBuilder tests builder creation
func TestRouteBuilder_NewRouteBuilder(t *testing.T) {
	t.Run("creates empty builder", func(t *testing.T) {
		builder := NewRouteBuilder()

		assert.NotNil(t, builder)
	})
}

// TestRouteBuilder_Methods tests all builder methods
func TestRouteBuilder_Methods(t *testing.T) {
	t.Run("HTTPHandler sets handler", func(t *testing.T) {
		httpHandler := func(w http.ResponseWriter, r *http.Request) {}

		builder := NewRouteBuilder()
		result := builder.HTTPHandler(httpHandler)

		assert.Equal(t, builder, result, "Should return builder for chaining")
	})

	t.Run("Handler sets handler", func(t *testing.T) {
		handler := func(k *kit.Kit) error { return nil }

		builder := NewRouteBuilder()
		result := builder.Handler(handler)

		assert.Equal(t, builder, result, "Should return builder for chaining")
	})

	t.Run("Method sets method", func(t *testing.T) {
		builder := NewRouteBuilder()
		result := builder.Method(POST)

		assert.Equal(t, builder, result, "Should return builder for chaining")

		route := builder.Build()
		assert.Equal(t, POST, route.Method)
	})

	t.Run("Prefix sets prefix", func(t *testing.T) {
		builder := NewRouteBuilder()
		result := builder.Prefix("/api")

		assert.Equal(t, builder, result, "Should return builder for chaining")

		route := builder.Build()
		assert.Equal(t, "/api", route.Prefix)
	})

	t.Run("Pattern sets pattern", func(t *testing.T) {
		builder := NewRouteBuilder()
		result := builder.Pattern("/users")

		assert.Equal(t, builder, result, "Should return builder for chaining")

		route := builder.Build()
		assert.Equal(t, "/users", route.Pattern)
	})
}

// TestRouteBuilder_Build tests route building
func TestRouteBuilder_Build(t *testing.T) {
	t.Run("builds complete route", func(t *testing.T) {
		handler := func(k *kit.Kit) error { return nil }
		httpHandler := func(w http.ResponseWriter, r *http.Request) {}

		route := NewRouteBuilder().
			Handler(handler).
			HTTPHandler(httpHandler).
			Method(GET).
			Prefix("/api").
			Pattern("/users").
			Build()

		assert.NotNil(t, route)
		assert.NotNil(t, route.Handler)
		assert.NotNil(t, route.HTTPHandler)
		assert.Equal(t, GET, route.Method)
		assert.Equal(t, "/api", route.Prefix)
		assert.Equal(t, "/users", route.Pattern)
	})

	t.Run("builds minimal route", func(t *testing.T) {
		route := NewRouteBuilder().Build()

		assert.NotNil(t, route)
		assert.Equal(t, Method(""), route.Method)
		assert.Equal(t, "", route.Prefix)
		assert.Equal(t, "", route.Pattern)
	})

	t.Run("builder is reusable", func(t *testing.T) {
		builder := NewRouteBuilder().
			Method(GET).
			Prefix("/api")

		route1 := builder.Pattern("/users").Build()
		route2 := builder.Pattern("/posts").Build()

		assert.Equal(t, "/users", route1.Pattern)
		assert.Equal(t, "/posts", route2.Pattern)
	})
}

// TestRouteBuilder_Chaining tests method chaining
func TestRouteBuilder_Chaining(t *testing.T) {
	t.Run("all methods support chaining", func(t *testing.T) {
		handler := func(k *kit.Kit) error { return nil }
		httpHandler := func(w http.ResponseWriter, r *http.Request) {}

		builder := NewRouteBuilder()

		// Chain all methods
		result := builder.
			Handler(handler).
			HTTPHandler(httpHandler).
			Method(POST).
			Prefix("/api").
			Pattern("/users")

		// All should return the same builder
		assert.Equal(t, builder, result)
	})

	t.Run("complex chaining scenario", func(t *testing.T) {
		route := NewRouteBuilder().
			Method(PUT).
			Prefix("/api").
			Prefix("/admin"). // Overwrite
			Pattern("/users/{id}").
			Method(DELETE). // Overwrite
			Build()

		assert.Equal(t, DELETE, route.Method)
		assert.Equal(t, "/admin", route.Prefix)
		assert.Equal(t, "/users/{id}", route.Pattern)
	})
}

// TestTrim tests the trim helper function
func TestTrim(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes trailing slash",
			input:    "/api/",
			expected: "/api",
		},
		{
			name:     "no trailing slash",
			input:    "/api",
			expected: "/api",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only slash",
			input:    "/",
			expected: "",
		},
		{
			name:     "multiple trailing slashes",
			input:    "/api//",
			expected: "/api/",
		},
		{
			name:     "leading slash preserved",
			input:    "/api/v1/",
			expected: "/api/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trim(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRoute_Struct tests the Route struct
func TestRoute_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		handler := func(k *kit.Kit) error { return nil }
		httpHandler := func(w http.ResponseWriter, r *http.Request) {}

		route := Route{
			HTTPHandler: httpHandler,
			Handler:     handler,
			Method:      GET,
			Prefix:      "/api",
			Pattern:     "/users",
		}

		assert.NotNil(t, route.HTTPHandler)
		assert.NotNil(t, route.Handler)
		assert.Equal(t, GET, route.Method)
		assert.Equal(t, "/api", route.Prefix)
		assert.Equal(t, "/users", route.Pattern)
	})

	t.Run("can be created with struct literal", func(t *testing.T) {
		route := Route{
			Method:  POST,
			Prefix:  "/admin",
			Pattern: "/users",
		}

		assert.Equal(t, POST, route.Method)
		assert.Equal(t, "/admin/users", route.Path())
		assert.Equal(t, "POST /admin/users", route.FullPath())
	})
}

// TestRoute_Integration tests realistic route usage
func TestRoute_Integration(t *testing.T) {
	t.Run("REST API route creation", func(t *testing.T) {
		// Simulate creating routes for a REST API
		routes := []Route{
			*NewRouteBuilder().
				Method(GET).
				Prefix("/api/v1").
				Pattern("/users").
				Build(),
			*NewRouteBuilder().
				Method(GET).
				Prefix("/api/v1").
				Pattern("/users/{id}").
				Build(),
			*NewRouteBuilder().
				Method(POST).
				Prefix("/api/v1").
				Pattern("/users").
				Build(),
			*NewRouteBuilder().
				Method(PUT).
				Prefix("/api/v1").
				Pattern("/users/{id}").
				Build(),
			*NewRouteBuilder().
				Method(DELETE).
				Prefix("/api/v1").
				Pattern("/users/{id}").
				Build(),
		}

		assert.Len(t, routes, 5)

		// Verify full paths
		assert.Equal(t, "GET /api/v1/users", routes[0].FullPath())
		assert.Equal(t, "GET /api/v1/users/{id}", routes[1].FullPath())
		assert.Equal(t, "POST /api/v1/users", routes[2].FullPath())
		assert.Equal(t, "PUT /api/v1/users/{id}", routes[3].FullPath())
		assert.Equal(t, "DELETE /api/v1/users/{id}", routes[4].FullPath())
	})

	t.Run("route modification via builder", func(t *testing.T) {
		// Create a base route
		baseRoute := Route{
			Method:  GET,
			Prefix:  "/api",
			Pattern: "/users",
		}

		// Modify it for different environment
		prodRoute := baseRoute.Builder().
			Prefix("/prod/api").
			Build()

		// Original unchanged
		assert.Equal(t, "/api", baseRoute.Prefix)

		// Modified version
		assert.Equal(t, "/prod/api", prodRoute.Prefix)
		assert.Equal(t, "/users", prodRoute.Pattern)
		assert.Equal(t, GET, prodRoute.Method)
	})

	t.Run("building routes with different patterns", func(t *testing.T) {
		patterns := []string{
			"/users",
			"/users/{id}",
			"/users/{id}/posts",
			"/users/{id}/posts/{postId}",
		}

		routes := make([]Route, len(patterns))
		for i, pattern := range patterns {
			routes[i] = *NewRouteBuilder().
				Method(GET).
				Prefix("/api").
				Pattern(pattern).
				Build()
		}

		for i, route := range routes {
			assert.Equal(t, "/api"+patterns[i], route.Path())
		}
	})
}
