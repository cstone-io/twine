package kit

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKit_JSON tests JSON response writing
func TestKit_JSON(t *testing.T) {
	t.Run("writes JSON response", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Response: w, Request: r}

		data := map[string]string{"message": "hello"}
		err := k.JSON(200, data)
		require.NoError(t, err)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), `"message":"hello"`)
	})

	t.Run("handles different status codes", func(t *testing.T) {
		testCases := []struct {
			name   string
			status int
		}{
			{"200 OK", 200},
			{"201 Created", 201},
			{"400 Bad Request", 400},
			{"404 Not Found", 404},
			{"500 Internal Server Error", 500},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/", nil)

				k := &Kit{Response: w, Request: r}

				err := k.JSON(tc.status, map[string]string{"status": tc.name})
				require.NoError(t, err)
				assert.Equal(t, tc.status, w.Code)
			})
		}
	})

	t.Run("encodes complex structures", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Response: w, Request: r}

		data := map[string]any{
			"user": map[string]any{
				"id":    123,
				"name":  "Alice",
				"email": "alice@example.com",
			},
			"tags": []string{"golang", "testing"},
		}

		err := k.JSON(200, data)
		require.NoError(t, err)
		assert.Contains(t, w.Body.String(), `"id":123`)
		assert.Contains(t, w.Body.String(), `"name":"Alice"`)
	})
}

// TestKit_Text tests plain text response writing
func TestKit_Text(t *testing.T) {
	t.Run("writes text response", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Response: w, Request: r}

		err := k.Text(200, "Hello, World!")
		require.NoError(t, err)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
		assert.Equal(t, "Hello, World!", w.Body.String())
	})

	t.Run("handles empty text", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Response: w, Request: r}

		err := k.Text(200, "")
		require.NoError(t, err)
		assert.Equal(t, "", w.Body.String())
	})

	t.Run("handles multiline text", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Response: w, Request: r}

		text := "Line 1\nLine 2\nLine 3"
		err := k.Text(200, text)
		require.NoError(t, err)
		assert.Equal(t, text, w.Body.String())
	})
}

// TestKit_Bytes tests raw bytes response
func TestKit_Bytes(t *testing.T) {
	t.Run("writes byte response", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Response: w, Request: r}

		data := []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f} // "Hello"
		err := k.Bytes(200, data)
		require.NoError(t, err)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
		assert.Equal(t, "Hello", w.Body.String())
	})

	t.Run("handles empty bytes", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Response: w, Request: r}

		err := k.Bytes(200, []byte{})
		require.NoError(t, err)
		assert.Equal(t, 0, w.Body.Len())
	})

	t.Run("handles binary data", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Response: w, Request: r}

		binaryData := []byte{0x00, 0xFF, 0xAB, 0xCD}
		err := k.Bytes(200, binaryData)
		require.NoError(t, err)
		assert.Equal(t, binaryData, w.Body.Bytes())
	})
}

// TestKit_HTML tests HTML response writing
func TestKit_HTML(t *testing.T) {
	t.Run("writes HTML response", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Response: w, Request: r}

		html := "<h1>Title</h1><p>Content</p>"
		err := k.HTML(200, html)
		require.NoError(t, err)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
		assert.Equal(t, html, w.Body.String())
	})

	t.Run("handles complex HTML", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Response: w, Request: r}

		html := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body><div class="container">Content</div></body>
</html>`

		err := k.HTML(200, html)
		require.NoError(t, err)
		assert.Contains(t, w.Body.String(), "<!DOCTYPE html>")
		assert.Contains(t, w.Body.String(), `class="container"`)
	})
}

// TestKit_NoContent tests 204 No Content response
func TestKit_NoContent(t *testing.T) {
	t.Run("writes 204 No Content", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Response: w, Request: r}

		err := k.NoContent()
		require.NoError(t, err)

		assert.Equal(t, 204, w.Code)
		assert.Equal(t, 0, w.Body.Len())
	})
}

// TestKit_IsAjax tests Ajax request detection
func TestKit_IsAjax(t *testing.T) {
	t.Run("detects Ajax request", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("X-Alpine-Request", "true")

		k := &Kit{Request: r}

		assert.True(t, k.IsAjax())
	})

	t.Run("returns false for non-Ajax request", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Request: r}

		assert.False(t, k.IsAjax())
	})

	t.Run("detects Ajax with any header value", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("X-Alpine-Request", "anything")

		k := &Kit{Request: r}

		assert.True(t, k.IsAjax())
	})

	t.Run("empty X-Alpine-Request header is not Ajax", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("X-Alpine-Request", "")

		k := &Kit{Request: r}

		assert.False(t, k.IsAjax())
	})
}

// TestKit_Redirect tests HTTP redirects
func TestKit_Redirect(t *testing.T) {
	t.Run("standard redirect", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Response: w, Request: r}

		err := k.Redirect("/login")
		require.NoError(t, err)

		assert.Equal(t, 303, w.Code)
		assert.Equal(t, "/login", w.Header().Get("Location"))
	})

	t.Run("Ajax redirect uses standard Location header", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("X-Alpine-Request", "true")

		k := &Kit{Response: w, Request: r}

		err := k.Redirect("/dashboard")
		require.NoError(t, err)

		assert.Equal(t, 303, w.Code)
		assert.Equal(t, "/dashboard", w.Header().Get("Location"))
	})

	t.Run("redirects to different paths", func(t *testing.T) {
		testCases := []string{
			"/",
			"/home",
			"/users/123",
			"/admin/dashboard",
			"https://example.com",
		}

		for _, path := range testCases {
			t.Run(path, func(t *testing.T) {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/", nil)

				k := &Kit{Response: w, Request: r}

				err := k.Redirect(path)
				require.NoError(t, err)
				assert.Equal(t, path, w.Header().Get("Location"))
			})
		}
	})
}

// TestKit_ResponseIntegration tests realistic response scenarios
func TestKit_ResponseIntegration(t *testing.T) {
	t.Run("API endpoint returns JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/users/123", nil)

		k := &Kit{Response: w, Request: r}

		user := map[string]any{
			"id":    123,
			"name":  "Alice",
			"email": "alice@example.com",
		}

		err := k.JSON(200, user)
		require.NoError(t, err)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), `"name":"Alice"`)
	})

	t.Run("error response returns JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/users/999", nil)

		k := &Kit{Response: w, Request: r}

		errorResp := map[string]any{
			"error":   "User not found",
			"code":    404,
			"message": "The requested user does not exist",
		}

		err := k.JSON(404, errorResp)
		require.NoError(t, err)

		assert.Equal(t, 404, w.Code)
		assert.Contains(t, w.Body.String(), `"error":"User not found"`)
	})

	t.Run("DELETE returns no content", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("DELETE", "/api/users/123", nil)

		k := &Kit{Response: w, Request: r}

		err := k.NoContent()
		require.NoError(t, err)

		assert.Equal(t, 204, w.Code)
		assert.Empty(t, w.Body.String())
	})

	t.Run("health check returns plain text", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/health", nil)

		k := &Kit{Response: w, Request: r}

		err := k.Text(200, "OK")
		require.NoError(t, err)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "OK", w.Body.String())
	})
}
