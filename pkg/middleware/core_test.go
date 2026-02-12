package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cstone-io/twine/pkg/kit"
)

// TestLoggingMiddleware tests request logging middleware
func TestLoggingMiddleware(t *testing.T) {
	t.Run("logs request method and path", func(t *testing.T) {
		handlerCalled := false

		mw := LoggingMiddleware()
		handler := func(k *kit.Kit) error {
			handlerCalled = true
			return k.Text(200, "ok")
		}

		wrapped := mw(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test/path", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
		assert.True(t, handlerCalled)
	})

	t.Run("logs different HTTP methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		mw := LoggingMiddleware()
		handler := func(k *kit.Kit) error {
			return k.Text(200, "ok")
		}

		wrapped := mw(handler)

		for _, method := range methods {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(method, "/api/users", nil)
			k := &kit.Kit{Response: w, Request: r}

			err := wrapped(k)
			require.NoError(t, err)
		}
	})

	t.Run("calls next handler after logging", func(t *testing.T) {
		handlerCalled := false

		mw := LoggingMiddleware()
		handler := func(k *kit.Kit) error {
			handlerCalled = true
			return nil
		}

		wrapped := mw(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
		assert.True(t, handlerCalled)
	})

	t.Run("logs even when handler returns error", func(t *testing.T) {
		mw := LoggingMiddleware()
		handler := func(k *kit.Kit) error {
			return assert.AnError
		}

		wrapped := mw(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		assert.Error(t, err) // Error should propagate
	})

	t.Run("logs requests with query parameters", func(t *testing.T) {
		mw := LoggingMiddleware()
		handler := func(k *kit.Kit) error {
			return k.Text(200, "ok")
		}

		wrapped := mw(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/users?page=1&limit=10", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
	})
}

// TestTimeoutMiddleware tests request timeout middleware
func TestTimeoutMiddleware(t *testing.T) {
	t.Run("allows fast requests to complete", func(t *testing.T) {
		handlerCalled := false

		mw := TimeoutMiddleware(100 * time.Millisecond)
		handler := func(k *kit.Kit) error {
			handlerCalled = true
			// Fast operation
			return k.Text(200, "ok")
		}

		wrapped := mw(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
		assert.True(t, handlerCalled)
		assert.Equal(t, "ok", w.Body.String())
	})

	t.Run("sets timeout context", func(t *testing.T) {
		var ctxHasDeadline bool

		mw := TimeoutMiddleware(1 * time.Second)
		handler := func(k *kit.Kit) error {
			_, ctxHasDeadline = k.Request.Context().Deadline()
			return k.Text(200, "ok")
		}

		wrapped := mw(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
		assert.True(t, ctxHasDeadline)
	})

	t.Run("handler can check context timeout", func(t *testing.T) {
		mw := TimeoutMiddleware(50 * time.Millisecond)
		handler := func(k *kit.Kit) error {
			// Simulate slow operation
			select {
			case <-time.After(200 * time.Millisecond):
				return k.Text(200, "completed")
			case <-k.Request.Context().Done():
				return k.Text(408, "timeout")
			}
		}

		wrapped := mw(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
		// Handler detected timeout and returned 408
		assert.Equal(t, 408, w.Code)
	})

	t.Run("different timeout durations work", func(t *testing.T) {
		durations := []time.Duration{
			10 * time.Millisecond,
			100 * time.Millisecond,
			1 * time.Second,
		}

		handler := func(k *kit.Kit) error {
			return k.Text(200, "ok")
		}

		for _, d := range durations {
			mw := TimeoutMiddleware(d)
			wrapped := mw(handler)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			k := &kit.Kit{Response: w, Request: r}

			err := wrapped(k)
			require.NoError(t, err)
		}
	})

	t.Run("preserves existing context values", func(t *testing.T) {
		var capturedValue string

		mw := TimeoutMiddleware(1 * time.Second)
		handler := func(k *kit.Kit) error {
			capturedValue = k.GetContext("key")
			return k.Text(200, "ok")
		}

		wrapped := mw(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		k := &kit.Kit{Response: w, Request: r}

		// Set context value before middleware
		k.SetContext("key", "value")

		err := wrapped(k)
		require.NoError(t, err)
		assert.Equal(t, "value", capturedValue)
	})
}

// TestCoreMiddleware_Integration tests realistic middleware scenarios
func TestCoreMiddleware_Integration(t *testing.T) {
	t.Run("logging and timeout together", func(t *testing.T) {
		logged := false

		loggingMW := LoggingMiddleware()
		timeoutMW := TimeoutMiddleware(1 * time.Second)

		handler := func(k *kit.Kit) error {
			logged = true
			return k.Text(200, "ok")
		}

		// Apply both middlewares
		wrapped := ApplyMiddlewares(handler, loggingMW, timeoutMW)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/test", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
		assert.True(t, logged)
		assert.Equal(t, "ok", w.Body.String())
	})

	t.Run("middleware stack for API endpoint", func(t *testing.T) {
		loggingMW := LoggingMiddleware()
		timeoutMW := TimeoutMiddleware(5 * time.Second)

		handler := func(k *kit.Kit) error {
			data := map[string]string{"status": "ok"}
			return k.JSON(200, data)
		}

		wrapped := ApplyMiddlewares(handler, loggingMW, timeoutMW)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/status", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
		assert.Contains(t, w.Body.String(), `"status":"ok"`)
	})

	t.Run("timeout prevents long-running operations", func(t *testing.T) {
		timeoutMW := TimeoutMiddleware(10 * time.Millisecond)

		handler := func(k *kit.Kit) error {
			// Check if context is already done
			select {
			case <-k.Request.Context().Done():
				return k.Text(408, "Request timeout")
			default:
				// Simulate slow operation
				time.Sleep(50 * time.Millisecond)

				// Check again after operation
				select {
				case <-k.Request.Context().Done():
					return k.Text(408, "Request timeout")
				default:
					return k.Text(200, "ok")
				}
			}
		}

		wrapped := timeoutMW(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/slow", nil)
		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)
		// Context should be cancelled during the operation
		assert.Equal(t, 408, w.Code)
	})
}
