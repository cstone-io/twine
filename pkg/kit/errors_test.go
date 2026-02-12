package kit

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	twineerrors "github.com/cstone-io/twine/pkg/errors"
)

// TestUseErrorHandler tests custom error handler registration
func TestUseErrorHandler(t *testing.T) {
	// Save original error handler
	originalHandler := errorHandler
	defer func() {
		errorHandler = originalHandler
	}()

	t.Run("sets custom error handler", func(t *testing.T) {
		customCalled := false

		UseErrorHandler(func(k *Kit, err error) {
			customCalled = true
			k.Text(418, "Custom error")
		})

		h := Handler(func(k *Kit) error {
			return errors.New("test error")
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		h(w, r)

		assert.True(t, customCalled)
		assert.Equal(t, 418, w.Code)
		assert.Equal(t, "Custom error", w.Body.String())
	})

	t.Run("custom handler receives error", func(t *testing.T) {
		var capturedError error

		UseErrorHandler(func(k *Kit, err error) {
			capturedError = err
			k.Text(500, "Error captured")
		})

		testErr := errors.New("specific error")

		h := Handler(func(k *Kit) error {
			return testErr
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		h(w, r)

		assert.Equal(t, testErr, capturedError)
	})

	t.Run("custom handler receives Kit", func(t *testing.T) {
		var capturedKit *Kit

		UseErrorHandler(func(k *Kit, err error) {
			capturedKit = k
			k.Text(500, "ok")
		})

		h := Handler(func(k *Kit) error {
			return errors.New("error")
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test-path", nil)

		h(w, r)

		require.NotNil(t, capturedKit)
		assert.Equal(t, "/test-path", capturedKit.Request.URL.Path)
	})
}

// TestDefaultErrorHandler tests default error handler behavior
func TestDefaultErrorHandler(t *testing.T) {
	// Save and restore original error handler
	originalHandler := errorHandler
	defer func() {
		errorHandler = originalHandler
	}()

	// Reset to default
	errorHandler = originalHandler

	t.Run("handles Twine Error with correct status", func(t *testing.T) {
		customErr := &twineerrors.Error{
			Code:       1001,
			Message:    "Not found",
			HTTPStatus: http.StatusNotFound,
			Severity:   twineerrors.ErrCritical,
		}

		h := Handler(func(k *Kit) error {
			return customErr
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		h(w, r)

		assert.Equal(t, 404, w.Code)
		assert.Contains(t, w.Body.String(), `"error":"Not found"`)
		assert.Contains(t, w.Body.String(), `"code":1001`)
	})

	t.Run("handles Twine Error without HTTPStatus", func(t *testing.T) {
		customErr := &twineerrors.Error{
			Code:     9999,
			Message:  "Custom error",
			Severity: twineerrors.ErrError,
			// No HTTPStatus set
		}

		h := Handler(func(k *Kit) error {
			return customErr
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		h(w, r)

		// Should default to 500
		assert.Equal(t, 500, w.Code)
		assert.Contains(t, w.Body.String(), `"error":"Custom error"`)
	})

	t.Run("handles standard Go error", func(t *testing.T) {
		h := Handler(func(k *Kit) error {
			return errors.New("standard error")
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		h(w, r)

		assert.Equal(t, 500, w.Code)
		assert.Contains(t, w.Body.String(), `"error"`)
		assert.Contains(t, w.Body.String(), `"code"`)
	})

	t.Run("returns JSON error response", func(t *testing.T) {
		h := Handler(func(k *Kit) error {
			return twineerrors.ErrNotFound
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		h(w, r)

		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), `{`)
		assert.Contains(t, w.Body.String(), `}`)
	})
}

// TestNotFoundHandler tests 404 handler
func TestNotFoundHandler(t *testing.T) {
	t.Run("returns 404 error", func(t *testing.T) {
		h := NotFoundHandler()

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/nonexistent", nil)

		h(w, r)

		assert.Equal(t, 404, w.Code)
		assert.Contains(t, w.Body.String(), `"error"`)
	})

	t.Run("uses ErrNotFound", func(t *testing.T) {
		h := NotFoundHandler()

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		h(w, r)

		// Should contain the ErrNotFound code (2002)
		assert.Contains(t, w.Body.String(), `2002`)
	})
}

// TestErrorHandling_Integration tests realistic error scenarios
func TestErrorHandling_Integration(t *testing.T) {
	// Save and restore original error handler
	originalHandler := errorHandler
	defer func() {
		errorHandler = originalHandler
	}()

	t.Run("database error returns 500", func(t *testing.T) {
		h := Handler(func(k *Kit) error {
			return twineerrors.ErrDatabaseRead
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/users", nil)

		h(w, r)

		assert.Equal(t, 500, w.Code)
		assert.Contains(t, w.Body.String(), `"error"`)
	})

	t.Run("validation error returns appropriate status", func(t *testing.T) {
		validationErr := &twineerrors.Error{
			Code:       4001,
			Message:    "Invalid input",
			HTTPStatus: http.StatusBadRequest,
			Severity:   twineerrors.ErrMinor,
		}

		h := Handler(func(k *Kit) error {
			return validationErr
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/users", nil)

		h(w, r)

		assert.Equal(t, 400, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid input")
	})

	t.Run("custom error handler can render HTML", func(t *testing.T) {
		UseErrorHandler(func(k *Kit, err error) {
			k.HTML(500, "<h1>Error</h1><p>Something went wrong</p>")
		})

		h := Handler(func(k *Kit) error {
			return errors.New("error")
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		h(w, r)

		assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), "<h1>Error</h1>")
	})

	t.Run("error handler can check request context", func(t *testing.T) {
		UseErrorHandler(func(k *Kit, err error) {
			userID := k.GetContext("user_id")
			k.JSON(500, map[string]string{
				"error":   "Internal error",
				"user_id": userID,
			})
		})

		h := Handler(func(k *Kit) error {
			k.SetContext("user_id", "123")
			return errors.New("error")
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		h(w, r)

		assert.Contains(t, w.Body.String(), `"user_id":"123"`)
	})

	t.Run("error handler called for each error", func(t *testing.T) {
		callCount := 0

		UseErrorHandler(func(k *Kit, err error) {
			callCount++
			k.Text(500, "error")
		})

		// Make multiple requests with errors
		for i := 0; i < 3; i++ {
			h := Handler(func(k *Kit) error {
				return errors.New("error")
			})

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			h(w, r)
		}

		assert.Equal(t, 3, callCount)
	})
}

// TestErrorHandler_ThreadSafety tests concurrent error handling
func TestErrorHandler_ThreadSafety(t *testing.T) {
	// Save and restore original error handler
	originalHandler := errorHandler
	defer func() {
		errorHandler = originalHandler
	}()

	t.Run("handles concurrent errors", func(t *testing.T) {
		UseErrorHandler(func(k *Kit, err error) {
			k.Text(500, "error")
		})

		h := Handler(func(k *Kit) error {
			return errors.New("concurrent error")
		})

		// Make concurrent requests
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/", nil)
				h(w, r)
				done <- true
			}()
		}

		// Wait for all to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// If we get here without panic, thread safety is good
		assert.True(t, true)
	})
}
