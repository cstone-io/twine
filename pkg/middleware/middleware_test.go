package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/cstone-io/twine/pkg/kit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestApplyMiddlewares tests middleware application
func TestApplyMiddlewares(t *testing.T) {
	t.Run("applies single middleware", func(t *testing.T) {
		called := false

		mw := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				called = true
				return next(k)
			}
		}

		handler := func(k *kit.Kit) error {
			return k.Text(200, "ok")
		}

		wrapped := ApplyMiddlewares(handler, mw)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("applies multiple middlewares in order", func(t *testing.T) {
		order := []string{}

		mw1 := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				order = append(order, "mw1-before")
				err := next(k)
				order = append(order, "mw1-after")
				return err
			}
		}

		mw2 := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				order = append(order, "mw2-before")
				err := next(k)
				order = append(order, "mw2-after")
				return err
			}
		}

		handler := func(k *kit.Kit) error {
			order = append(order, "handler")
			return nil
		}

		// Apply mw1, then mw2
		wrapped := ApplyMiddlewares(handler, mw1, mw2)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)

		// mw2 is outermost (applied last), so executes first
		expected := []string{
			"mw2-before",
			"mw1-before",
			"handler",
			"mw1-after",
			"mw2-after",
		}
		assert.Equal(t, expected, order)
	})

	t.Run("middleware can modify context", func(t *testing.T) {
		mw := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				k.SetContext("modified", "true")
				return next(k)
			}
		}

		handler := func(k *kit.Kit) error {
			value := k.GetContext("modified")
			return k.Text(200, value)
		}

		wrapped := ApplyMiddlewares(handler, mw)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
		assert.Equal(t, "true", w.Body.String())
	})

	t.Run("middleware can short-circuit", func(t *testing.T) {
		handlerCalled := false

		mw := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				// Short-circuit without calling next
				return k.Text(403, "Forbidden")
			}
		}

		handler := func(k *kit.Kit) error {
			handlerCalled = true
			return k.Text(200, "ok")
		}

		wrapped := ApplyMiddlewares(handler, mw)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
		assert.False(t, handlerCalled)
		assert.Equal(t, 403, w.Code)
		assert.Equal(t, "Forbidden", w.Body.String())
	})

	t.Run("works with no middlewares", func(t *testing.T) {
		handler := func(k *kit.Kit) error {
			return k.Text(200, "ok")
		}

		wrapped := ApplyMiddlewares(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
		assert.Equal(t, "ok", w.Body.String())
	})

	t.Run("middleware can handle errors", func(t *testing.T) {
		mw := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				err := next(k)
				if err != nil {
					return k.Text(500, "Error handled by middleware")
				}
				return nil
			}
		}

		handler := func(k *kit.Kit) error {
			return assert.AnError
		}

		wrapped := ApplyMiddlewares(handler, mw)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
		assert.Equal(t, "Error handled by middleware", w.Body.String())
	})
}

// TestChain tests middleware chaining
func TestChain(t *testing.T) {
	t.Run("chains multiple middlewares", func(t *testing.T) {
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

		handler := func(k *kit.Kit) error {
			order = append(order, "handler")
			return nil
		}

		chained := Chain(mw1, mw2)
		wrapped := chained(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)

		// mw2 executes first (outer), then mw1, then handler
		assert.Equal(t, []string{"mw2", "mw1", "handler"}, order)
	})

	t.Run("chained middleware can be used as single middleware", func(t *testing.T) {
		callCount := 0

		mw1 := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				callCount++
				return next(k)
			}
		}

		mw2 := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				callCount++
				return next(k)
			}
		}

		handler := func(k *kit.Kit) error {
			return nil
		}

		// Create a chained middleware
		chained := Chain(mw1, mw2)

		// Apply it like any other middleware
		wrapped := ApplyMiddlewares(handler, chained)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
		assert.Equal(t, 2, callCount)
	})

	t.Run("empty chain does nothing", func(t *testing.T) {
		handler := func(k *kit.Kit) error {
			return k.Text(200, "ok")
		}

		chained := Chain()
		wrapped := chained(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
		assert.Equal(t, "ok", w.Body.String())
	})
}

// TestMiddleware_Integration tests realistic middleware scenarios
func TestMiddleware_Integration(t *testing.T) {
	t.Run("auth and logging middleware together", func(t *testing.T) {
		logged := false
		authenticated := false

		authMW := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				authenticated = true
				k.SetContext("user_id", "123")
				return next(k)
			}
		}

		loggingMW := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				logged = true
				return next(k)
			}
		}

		handler := func(k *kit.Kit) error {
			userID := k.GetContext("user_id")
			return k.Text(200, "User: "+userID)
		}

		wrapped := ApplyMiddlewares(handler, authMW, loggingMW)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
		assert.True(t, logged)
		assert.True(t, authenticated)
		assert.Equal(t, "User: 123", w.Body.String())
	})

	t.Run("middleware stack with error handling", func(t *testing.T) {
		recoverMW := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				err := next(k)
				if err != nil {
					return k.Text(500, "Recovered from error")
				}
				return nil
			}
		}

		handler := func(k *kit.Kit) error {
			return assert.AnError
		}

		wrapped := ApplyMiddlewares(handler, recoverMW)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
		assert.Equal(t, "Recovered from error", w.Body.String())
	})
}
