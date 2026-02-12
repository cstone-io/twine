package kit

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_Conversion tests HandlerFunc to http.HandlerFunc conversion
func TestHandler_Conversion(t *testing.T) {
	t.Run("converts HandlerFunc successfully", func(t *testing.T) {
		called := false

		h := Handler(func(k *Kit) error {
			called = true
			return k.Text(200, "success")
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		h(w, r)

		assert.True(t, called)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "success", w.Body.String())
	})

	t.Run("creates Kit with Response and Request", func(t *testing.T) {
		var capturedKit *Kit

		h := Handler(func(k *Kit) error {
			capturedKit = k
			return nil
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		h(w, r)

		require.NotNil(t, capturedKit)
		assert.NotNil(t, capturedKit.Response)
		assert.NotNil(t, capturedKit.Request)
		assert.Equal(t, "/test", capturedKit.Request.URL.Path)
	})

	t.Run("calls default error handler on error", func(t *testing.T) {
		h := Handler(func(k *Kit) error {
			return errors.New("test error")
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		h(w, r)

		// Default error handler should return 500
		assert.Equal(t, 500, w.Code)
	})

	t.Run("returns success without error", func(t *testing.T) {
		h := Handler(func(k *Kit) error {
			return k.Text(200, "ok")
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		h(w, r)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "ok", w.Body.String())
	})

	t.Run("allows Kit methods to be used", func(t *testing.T) {
		h := Handler(func(k *Kit) error {
			data := map[string]string{"message": "hello"}
			return k.JSON(200, data)
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		h(w, r)

		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), `"message":"hello"`)
	})
}

// TestHandler_ErrorHandling tests error handling behavior
func TestHandler_ErrorHandling(t *testing.T) {
	t.Run("nil error does not trigger error handler", func(t *testing.T) {
		h := Handler(func(k *Kit) error {
			return nil
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		h(w, r)

		// Should not write any error response
		assert.Equal(t, 200, w.Code)
	})

	t.Run("error triggers error handler", func(t *testing.T) {
		h := Handler(func(k *Kit) error {
			return errors.New("something went wrong")
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		h(w, r)

		assert.Equal(t, 500, w.Code)
		assert.NotEmpty(t, w.Body.String())
	})
}

// TestHandler_Integration tests realistic handler scenarios
func TestHandler_Integration(t *testing.T) {
	t.Run("GET handler returns data", func(t *testing.T) {
		h := Handler(func(k *Kit) error {
			users := []map[string]any{
				{"id": 1, "name": "Alice"},
				{"id": 2, "name": "Bob"},
			}
			return k.JSON(200, users)
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/users", nil)

		h(w, r)

		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Alice")
		assert.Contains(t, w.Body.String(), "Bob")
	})

	t.Run("POST handler processes data", func(t *testing.T) {
		h := Handler(func(k *Kit) error {
			// Simulate creating resource
			return k.JSON(201, map[string]any{
				"id":      123,
				"message": "Created",
			})
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/users", nil)

		h(w, r)

		assert.Equal(t, 201, w.Code)
		assert.Contains(t, w.Body.String(), `"id":123`)
	})

	t.Run("DELETE handler returns no content", func(t *testing.T) {
		h := Handler(func(k *Kit) error {
			return k.NoContent()
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("DELETE", "/users/123", nil)

		h(w, r)

		assert.Equal(t, 204, w.Code)
		assert.Empty(t, w.Body.String())
	})

	t.Run("handler uses path parameters", func(t *testing.T) {
		h := Handler(func(k *Kit) error {
			id := k.PathValue("id")
			return k.JSON(200, map[string]string{"id": id})
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/users/456", nil)
		r.SetPathValue("id", "456")

		h(w, r)

		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), `"id":"456"`)
	})

	t.Run("handler uses context", func(t *testing.T) {
		h := Handler(func(k *Kit) error {
			k.SetContext("user_id", "789")
			userID := k.GetContext("user_id")
			return k.JSON(200, map[string]string{"user_id": userID})
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/profile", nil)

		h(w, r)

		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), `"user_id":"789"`)
	})
}
