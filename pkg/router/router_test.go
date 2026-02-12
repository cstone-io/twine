package router

import (
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/cstone-io/twine/pkg/kit"
	"github.com/stretchr/testify/assert"
)

// TestRouter_NewRouter tests router creation
func TestRouter_NewRouter(t *testing.T) {
	t.Run("creates router with prefix", func(t *testing.T) {
		r := NewRouter("/api")

		assert.NotNil(t, r)
		assert.Equal(t, "/api", r.Prefix)
		assert.Empty(t, r.Routes)
		assert.Empty(t, r.Middlewares)
		assert.Empty(t, r.Children)
	})

	t.Run("creates router with empty prefix", func(t *testing.T) {
		r := NewRouter("")

		assert.NotNil(t, r)
		assert.Equal(t, "", r.Prefix)
	})

	t.Run("trims trailing slash from prefix", func(t *testing.T) {
		r := NewRouter("/api/")

		assert.Equal(t, "/api", r.Prefix)
	})
}

// TestRouter_Get tests GET route registration
func TestRouter_Get(t *testing.T) {
	t.Run("registers GET route", func(t *testing.T) {
		r := NewRouter("")

		handler := func(k *kit.Kit) error {
			return k.Text(200, "GET handler")
		}

		r.Get("/users", handler)

		assert.Len(t, r.Routes, 1)
		assert.Equal(t, GET, r.Routes[0].Method)
		assert.Equal(t, "/users", r.Routes[0].Pattern)
	})

	t.Run("registers multiple GET routes", func(t *testing.T) {
		r := NewRouter("")

		r.Get("/users", func(k *kit.Kit) error { return nil })
		r.Get("/posts", func(k *kit.Kit) error { return nil })

		assert.Len(t, r.Routes, 2)
		assert.Equal(t, "/users", r.Routes[0].Pattern)
		assert.Equal(t, "/posts", r.Routes[1].Pattern)
	})
}

// TestRouter_Post tests POST route registration
func TestRouter_Post(t *testing.T) {
	t.Run("registers POST route", func(t *testing.T) {
		r := NewRouter("")

		r.Post("/users", func(k *kit.Kit) error {
			return k.Text(200, "POST handler")
		})

		assert.Len(t, r.Routes, 1)
		assert.Equal(t, POST, r.Routes[0].Method)
		assert.Equal(t, "/users", r.Routes[0].Pattern)
	})
}

// TestRouter_Put tests PUT route registration
func TestRouter_Put(t *testing.T) {
	t.Run("registers PUT route", func(t *testing.T) {
		r := NewRouter("")

		r.Put("/users/{id}", func(k *kit.Kit) error {
			return k.Text(200, "PUT handler")
		})

		assert.Len(t, r.Routes, 1)
		assert.Equal(t, PUT, r.Routes[0].Method)
		assert.Equal(t, "/users/{id}", r.Routes[0].Pattern)
	})
}

// TestRouter_Delete tests DELETE route registration
func TestRouter_Delete(t *testing.T) {
	t.Run("registers DELETE route", func(t *testing.T) {
		r := NewRouter("")

		r.Delete("/users/{id}", func(k *kit.Kit) error {
			return k.Text(200, "DELETE handler")
		})

		assert.Len(t, r.Routes, 1)
		assert.Equal(t, DELETE, r.Routes[0].Method)
		assert.Equal(t, "/users/{id}", r.Routes[0].Pattern)
	})
}

// TestRouter_Use tests middleware registration
func TestRouter_Use(t *testing.T) {
	t.Run("adds single middleware", func(t *testing.T) {
		r := NewRouter("")

		mw := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				return next(k)
			}
		}

		r.Use(mw)

		assert.Len(t, r.Middlewares, 1)
	})

	t.Run("adds multiple middlewares", func(t *testing.T) {
		r := NewRouter("")

		mw1 := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error { return next(k) }
		}
		mw2 := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error { return next(k) }
		}

		r.Use(mw1, mw2)

		assert.Len(t, r.Middlewares, 2)
	})

	t.Run("accumulates middlewares", func(t *testing.T) {
		r := NewRouter("")

		mw1 := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error { return next(k) }
		}
		mw2 := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error { return next(k) }
		}

		r.Use(mw1)
		r.Use(mw2)

		assert.Len(t, r.Middlewares, 2)
	})
}

// TestRouter_Sub tests sub-router mounting
func TestRouter_Sub(t *testing.T) {
	t.Run("adds sub-router", func(t *testing.T) {
		r := NewRouter("")
		sub := NewRouter("/api")

		r.Sub(sub)

		assert.Len(t, r.Children, 1)
		assert.Equal(t, sub, r.Children[0])
	})

	t.Run("adds multiple sub-routers", func(t *testing.T) {
		r := NewRouter("")
		sub1 := NewRouter("/api")
		sub2 := NewRouter("/admin")

		r.Sub(sub1)
		r.Sub(sub2)

		assert.Len(t, r.Children, 2)
		assert.Equal(t, sub1, r.Children[0])
		assert.Equal(t, sub2, r.Children[1])
	})
}

// TestRouter_InitializeAsRoot tests router initialization
func TestRouter_InitializeAsRoot(t *testing.T) {
	t.Run("initializes simple router", func(t *testing.T) {
		r := NewRouter("")

		r.Get("/users", func(k *kit.Kit) error {
			return k.Text(200, "users")
		})

		mux := r.InitializeAsRoot()

		assert.NotNil(t, mux)
		assert.Len(t, r.Routes, 1)
	})

	t.Run("concatenates prefix with pattern", func(t *testing.T) {
		r := NewRouter("/api")

		r.Get("/users", func(k *kit.Kit) error {
			return k.Text(200, "users")
		})

		r.InitializeAsRoot()

		assert.Equal(t, "/api", r.Routes[0].Prefix)
		assert.Equal(t, "/users", r.Routes[0].Pattern)
		assert.Equal(t, "/api/users", r.Routes[0].Path())
	})

	t.Run("sorts routes by path length", func(t *testing.T) {
		r := NewRouter("")

		r.Get("/a", func(k *kit.Kit) error { return nil })
		r.Get("/abc", func(k *kit.Kit) error { return nil })
		r.Get("/ab", func(k *kit.Kit) error { return nil })

		r.InitializeAsRoot()

		// Longest first
		assert.Equal(t, "/abc", r.Routes[0].Path())
		assert.Equal(t, "/ab", r.Routes[1].Path())
		assert.Equal(t, "/a", r.Routes[2].Path())
	})

	t.Run("handles sub-routers", func(t *testing.T) {
		root := NewRouter("")
		api := NewRouter("/api")

		api.Get("/users", func(k *kit.Kit) error {
			return k.Text(200, "users")
		})

		root.Sub(api)
		root.InitializeAsRoot()

		// Should have one route from sub-router
		assert.Len(t, root.Routes, 1)
		assert.Equal(t, "/api/users", root.Routes[0].Path())
	})

	t.Run("handles nested sub-routers", func(t *testing.T) {
		root := NewRouter("")
		api := NewRouter("/api")
		v1 := NewRouter("/v1")

		v1.Get("/users", func(k *kit.Kit) error {
			return k.Text(200, "users")
		})

		api.Sub(v1)
		root.Sub(api)
		root.InitializeAsRoot()

		assert.Len(t, root.Routes, 1)
		assert.Equal(t, "/api/v1/users", root.Routes[0].Path())
	})

	t.Run("middleware inheritance from parent", func(t *testing.T) {
		root := NewRouter("")
		api := NewRouter("/api")

		// Parent middleware
		parentCalled := false
		parentMW := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				parentCalled = true
				return next(k)
			}
		}

		root.Use(parentMW)

		api.Get("/users", func(k *kit.Kit) error {
			return k.Text(200, "users")
		})

		root.Sub(api)
		mux := root.InitializeAsRoot()

		// Test the route
		req := httptest.NewRequest("GET", "/api/users", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assert.True(t, parentCalled, "Parent middleware should be called")
	})

	t.Run("middleware order preserved", func(t *testing.T) {
		r := NewRouter("")

		order := []string{}
		mw1 := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				order = append(order, "mw1")
				return next(k)
			}
		}
		mw2 := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				order = append(order, "mw2")
				return next(k)
			}
		}

		r.Use(mw1, mw2)
		r.Get("/test", func(k *kit.Kit) error {
			order = append(order, "handler")
			return k.Text(200, "ok")
		})

		mux := r.InitializeAsRoot()

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		// Middleware executes in reverse order (last added runs first)
		assert.Equal(t, []string{"mw2", "mw1", "handler"}, order)
	})
}

// TestRouter_HTTPIntegration tests actual HTTP requests
func TestRouter_HTTPIntegration(t *testing.T) {
	t.Run("GET request works", func(t *testing.T) {
		r := NewRouter("")

		r.Get("/hello", func(k *kit.Kit) error {
			return k.Text(200, "Hello, World!")
		})

		mux := r.InitializeAsRoot()

		req := httptest.NewRequest("GET", "/hello", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "Hello, World!", w.Body.String())
	})

	t.Run("POST request works", func(t *testing.T) {
		r := NewRouter("")

		r.Post("/users", func(k *kit.Kit) error {
			return k.Text(201, "Created")
		})

		mux := r.InitializeAsRoot()

		req := httptest.NewRequest("POST", "/users", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, 201, w.Code)
		assert.Equal(t, "Created", w.Body.String())
	})

	t.Run("different methods on same path", func(t *testing.T) {
		r := NewRouter("")

		r.Get("/users", func(k *kit.Kit) error {
			return k.Text(200, "GET")
		})

		r.Post("/users", func(k *kit.Kit) error {
			return k.Text(200, "POST")
		})

		mux := r.InitializeAsRoot()

		// GET request
		req := httptest.NewRequest("GET", "/users", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		assert.Equal(t, "GET", w.Body.String())

		// POST request
		req = httptest.NewRequest("POST", "/users", nil)
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		assert.Equal(t, "POST", w.Body.String())
	})

	t.Run("path parameters work", func(t *testing.T) {
		r := NewRouter("")

		r.Get("/users/{id}", func(k *kit.Kit) error {
			id := k.PathValue("id")
			return k.Text(200, "User ID: "+id)
		})

		mux := r.InitializeAsRoot()

		req := httptest.NewRequest("GET", "/users/123", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "User ID: 123", w.Body.String())
	})

	t.Run("sub-router requests work", func(t *testing.T) {
		root := NewRouter("")
		api := NewRouter("/api")

		api.Get("/status", func(k *kit.Kit) error {
			return k.Text(200, "OK")
		})

		root.Sub(api)
		mux := root.InitializeAsRoot()

		req := httptest.NewRequest("GET", "/api/status", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "OK", w.Body.String())
	})
}

// TestRouter_ThreadSafety tests concurrent operations
func TestRouter_ThreadSafety(t *testing.T) {
	t.Run("concurrent route registration", func(t *testing.T) {
		r := NewRouter("")

		const goroutines = 100
		var wg sync.WaitGroup

		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func(index int) {
				defer wg.Done()
				r.Get("/route", func(k *kit.Kit) error {
					return k.Text(200, "ok")
				})
			}(i)
		}

		wg.Wait()

		assert.Len(t, r.Routes, goroutines)
	})

	t.Run("concurrent middleware registration", func(t *testing.T) {
		r := NewRouter("")

		const goroutines = 100
		var wg sync.WaitGroup

		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				mw := func(next kit.HandlerFunc) kit.HandlerFunc {
					return func(k *kit.Kit) error { return next(k) }
				}
				r.Use(mw)
			}()
		}

		wg.Wait()

		assert.Len(t, r.Middlewares, goroutines)
	})

	t.Run("concurrent sub-router addition", func(t *testing.T) {
		r := NewRouter("")

		const goroutines = 100
		var wg sync.WaitGroup

		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				sub := NewRouter("/api")
				r.Sub(sub)
			}()
		}

		wg.Wait()

		assert.Len(t, r.Children, goroutines)
	})
}

// TestRouter_EdgeCases tests edge cases
func TestRouter_EdgeCases(t *testing.T) {
	t.Run("empty router initializes", func(t *testing.T) {
		r := NewRouter("")

		mux := r.InitializeAsRoot()

		assert.NotNil(t, mux)
		assert.Empty(t, r.Routes)
	})

	t.Run("router with only sub-routers", func(t *testing.T) {
		root := NewRouter("")
		sub := NewRouter("/api")

		sub.Get("/test", func(k *kit.Kit) error {
			return k.Text(200, "ok")
		})

		root.Sub(sub)
		root.InitializeAsRoot()

		assert.Len(t, root.Routes, 1)
	})

	t.Run("router with trailing slashes in paths", func(t *testing.T) {
		r := NewRouter("/api/")

		r.Get("/users/", func(k *kit.Kit) error {
			return k.Text(200, "users")
		})

		r.InitializeAsRoot()

		// Prefix should have trailing slash trimmed
		assert.Equal(t, "/api", r.Routes[0].Prefix)
	})

	t.Run("deeply nested routers", func(t *testing.T) {
		r1 := NewRouter("")
		r2 := NewRouter("/api")
		r3 := NewRouter("/v1")
		r4 := NewRouter("/users")

		r4.Get("/{id}", func(k *kit.Kit) error {
			return k.Text(200, "user")
		})

		r3.Sub(r4)
		r2.Sub(r3)
		r1.Sub(r2)

		r1.InitializeAsRoot()

		assert.Len(t, r1.Routes, 1)
		assert.Equal(t, "/api/v1/users/{id}", r1.Routes[0].Path())
	})

	t.Run("middleware on leaf and parent", func(t *testing.T) {
		root := NewRouter("")
		api := NewRouter("/api")

		order := []string{}

		rootMW := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				order = append(order, "root")
				return next(k)
			}
		}

		apiMW := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				order = append(order, "api")
				return next(k)
			}
		}

		root.Use(rootMW)
		api.Use(apiMW)

		api.Get("/test", func(k *kit.Kit) error {
			order = append(order, "handler")
			return k.Text(200, "ok")
		})

		root.Sub(api)
		mux := root.InitializeAsRoot()

		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		// Parent middleware runs first, then child, then handler
		assert.Equal(t, []string{"root", "api", "handler"}, order)
	})
}

// TestRouter_ComplexScenarios tests realistic routing scenarios
func TestRouter_ComplexScenarios(t *testing.T) {
	t.Run("REST API structure", func(t *testing.T) {
		root := NewRouter("")
		api := NewRouter("/api")
		v1 := NewRouter("/v1")

		v1.Get("/users", func(k *kit.Kit) error {
			return k.Text(200, "list users")
		})
		v1.Get("/users/{id}", func(k *kit.Kit) error {
			return k.Text(200, "get user")
		})
		v1.Post("/users", func(k *kit.Kit) error {
			return k.Text(201, "create user")
		})
		v1.Put("/users/{id}", func(k *kit.Kit) error {
			return k.Text(200, "update user")
		})
		v1.Delete("/users/{id}", func(k *kit.Kit) error {
			return k.Text(204, "delete user")
		})

		api.Sub(v1)
		root.Sub(api)
		mux := root.InitializeAsRoot()

		// Test all endpoints
		tests := []struct {
			method string
			path   string
			status int
		}{
			{"GET", "/api/v1/users", 200},
			{"GET", "/api/v1/users/123", 200},
			{"POST", "/api/v1/users", 201},
			{"PUT", "/api/v1/users/123", 200},
			{"DELETE", "/api/v1/users/123", 204},
		}

		for _, tt := range tests {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			assert.Equal(t, tt.status, w.Code, "Method: %s, Path: %s", tt.method, tt.path)
		}
	})

	t.Run("admin vs public routes", func(t *testing.T) {
		root := NewRouter("")
		public := NewRouter("/public")
		admin := NewRouter("/admin")

		adminCalled := false
		adminMW := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				adminCalled = true
				return next(k)
			}
		}

		admin.Use(adminMW)

		public.Get("/home", func(k *kit.Kit) error {
			return k.Text(200, "public home")
		})

		admin.Get("/dashboard", func(k *kit.Kit) error {
			return k.Text(200, "admin dashboard")
		})

		root.Sub(public)
		root.Sub(admin)
		mux := root.InitializeAsRoot()

		// Public route (no admin middleware)
		adminCalled = false
		req := httptest.NewRequest("GET", "/public/home", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		assert.False(t, adminCalled)

		// Admin route (admin middleware called)
		adminCalled = false
		req = httptest.NewRequest("GET", "/admin/dashboard", nil)
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		assert.True(t, adminCalled)
	})
}
