package kit

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	twineerrors "github.com/cstone-io/twine/pkg/errors"
)

// TestKit_Decode tests request body decoding
func TestKit_Decode(t *testing.T) {
	t.Run("decodes JSON successfully", func(t *testing.T) {
		type Payload struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		payload := Payload{Name: "John", Email: "john@example.com"}
		body, err := json.Marshal(payload)
		require.NoError(t, err)

		r := httptest.NewRequest("POST", "/", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")

		k := &Kit{Request: r}

		var result Payload
		err = k.Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "John", result.Name)
		assert.Equal(t, "john@example.com", result.Email)
	})

	t.Run("decodes form data successfully", func(t *testing.T) {
		type FormData struct {
			Email    string `form:"email"`
			Password string `form:"password"`
		}

		form := url.Values{}
		form.Add("email", "user@example.com")
		form.Add("password", "secret123")

		r := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		k := &Kit{Request: r}

		var result FormData
		err := k.Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "user@example.com", result.Email)
		assert.Equal(t, "secret123", result.Password)
	})

	t.Run("returns error for unsupported content type", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/", nil)
		r.Header.Set("Content-Type", "application/xml")

		k := &Kit{Request: r}

		var result struct{}
		err := k.Decode(&result)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, twineerrors.ErrAPIRequestContentType))
	})

	t.Run("returns error for malformed JSON", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/", strings.NewReader("{invalid json}"))
		r.Header.Set("Content-Type", "application/json")

		k := &Kit{Request: r}

		var result struct{}
		err := k.Decode(&result)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, twineerrors.ErrDecodeJSON))
	})

	t.Run("decodes form with string field", func(t *testing.T) {
		type Form struct {
			Name string `form:"name"`
		}

		form := url.Values{}
		form.Add("name", "Alice")

		r := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		k := &Kit{Request: r}

		var result Form
		err := k.Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "Alice", result.Name)
	})

	t.Run("decodes form with empty optional field", func(t *testing.T) {
		type Form struct {
			Name     string `form:"name"`
			Optional string `form:"optional"`
		}

		form := url.Values{}
		form.Add("name", "Bob")
		// optional is not provided

		r := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		k := &Kit{Request: r}

		var result Form
		err := k.Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "Bob", result.Name)
		assert.Equal(t, "", result.Optional)
	})

	t.Run("ignores fields without form tags", func(t *testing.T) {
		type Form struct {
			Tagged   string `form:"tagged"`
			Untagged string
		}

		form := url.Values{}
		form.Add("tagged", "value1")
		form.Add("Untagged", "value2")

		r := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		k := &Kit{Request: r}

		var result Form
		err := k.Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "value1", result.Tagged)
		assert.Equal(t, "", result.Untagged) // Should remain empty
	})

	t.Run("decodes form with slice field", func(t *testing.T) {
		type Form struct {
			Tags []string `form:"tags"`
		}

		form := url.Values{}
		form.Add("tags", "golang")
		form.Add("tags", "testing")
		form.Add("tags", "web")

		r := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		k := &Kit{Request: r}

		var result Form
		err := k.Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, []string{"golang", "testing", "web"}, result.Tags)
	})

	t.Run("decodes form with nested struct", func(t *testing.T) {
		type Address struct {
			City string `form:"city"`
		}

		type Form struct {
			Name    string  `form:"name"`
			Address Address `form:"address"`
		}

		form := url.Values{}
		form.Add("name", "John")
		form.Add("city", "NYC")

		r := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		k := &Kit{Request: r}

		var result Form
		err := k.Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "John", result.Name)
		assert.Equal(t, "NYC", result.Address.City)
	})
}

// TestKit_PathValue tests path parameter extraction
func TestKit_PathValue(t *testing.T) {
	t.Run("extracts path parameter", func(t *testing.T) {
		// Create a request with path parameters
		r := httptest.NewRequest("GET", "/users/123", nil)
		r.SetPathValue("id", "123")

		k := &Kit{Request: r}

		id := k.PathValue("id")
		assert.Equal(t, "123", id)
	})

	t.Run("returns empty for missing parameter", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Request: r}

		id := k.PathValue("id")
		assert.Equal(t, "", id)
	})

	t.Run("extracts multiple path parameters", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/users/456/posts/789", nil)
		r.SetPathValue("userId", "456")
		r.SetPathValue("postId", "789")

		k := &Kit{Request: r}

		assert.Equal(t, "456", k.PathValue("userId"))
		assert.Equal(t, "789", k.PathValue("postId"))
	})
}

// TestKit_Authorization tests authorization token extraction
func TestKit_Authorization(t *testing.T) {
	t.Run("extracts token from cookie", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{Name: "token", Value: "cookie-token-123"})

		k := &Kit{Request: r}

		token, err := k.Authorization()
		require.NoError(t, err)
		assert.Equal(t, "cookie-token-123", token)
	})

	t.Run("extracts token from Bearer header", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Authorization", "Bearer header-token-456")

		k := &Kit{Request: r}

		token, err := k.Authorization()
		require.NoError(t, err)
		assert.Equal(t, "header-token-456", token)
	})

	t.Run("prefers cookie over header", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{Name: "token", Value: "cookie-token"})
		r.Header.Set("Authorization", "Bearer header-token")

		k := &Kit{Request: r}

		token, err := k.Authorization()
		require.NoError(t, err)
		assert.Equal(t, "cookie-token", token) // Cookie should win
	})

	t.Run("returns error when missing", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Request: r}

		_, err := k.Authorization()
		assert.Error(t, err)
		assert.True(t, errors.Is(err, twineerrors.ErrAuthMissingHeader))
	})

	t.Run("returns error for invalid Bearer format", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Authorization", "InvalidFormat token123")

		k := &Kit{Request: r}

		_, err := k.Authorization()
		assert.Error(t, err)
		assert.True(t, errors.Is(err, twineerrors.ErrAuthInvalidToken))
	})

	t.Run("returns empty string for Bearer with no token", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Authorization", "Bearer ")

		k := &Kit{Request: r}

		token, err := k.Authorization()
		require.NoError(t, err)
		assert.Equal(t, "", token) // Empty token is valid
	})

	t.Run("handles token without Bearer prefix", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Authorization", "just-a-token")

		k := &Kit{Request: r}

		_, err := k.Authorization()
		assert.Error(t, err)
		assert.True(t, errors.Is(err, twineerrors.ErrAuthInvalidToken))
	})
}

// TestKit_GetHeader tests header retrieval
func TestKit_GetHeader(t *testing.T) {
	t.Run("retrieves header value", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("X-Custom-Header", "custom-value")

		k := &Kit{Request: r}

		value := k.GetHeader("X-Custom-Header")
		assert.Equal(t, "custom-value", value)
	})

	t.Run("returns empty for missing header", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Request: r}

		value := k.GetHeader("X-Missing")
		assert.Equal(t, "", value)
	})

	t.Run("is case-insensitive", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Content-Type", "application/json")

		k := &Kit{Request: r}

		assert.Equal(t, "application/json", k.GetHeader("Content-Type"))
		assert.Equal(t, "application/json", k.GetHeader("content-type"))
	})
}

// TestKit_Context tests context value storage and retrieval
func TestKit_Context(t *testing.T) {
	t.Run("sets and gets context value", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Request: r}

		k.SetContext("user_id", "12345")
		value := k.GetContext("user_id")

		assert.Equal(t, "12345", value)
	})

	t.Run("returns empty for missing context key", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Request: r}

		value := k.GetContext("missing_key")
		assert.Equal(t, "", value)
	})

	t.Run("handles multiple context values", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Request: r}

		k.SetContext("key1", "value1")
		k.SetContext("key2", "value2")
		k.SetContext("key3", "value3")

		assert.Equal(t, "value1", k.GetContext("key1"))
		assert.Equal(t, "value2", k.GetContext("key2"))
		assert.Equal(t, "value3", k.GetContext("key3"))
	})

	t.Run("overwrites context value", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Request: r}

		k.SetContext("key", "original")
		k.SetContext("key", "updated")

		assert.Equal(t, "updated", k.GetContext("key"))
	})
}

// TestKit_Cookies tests cookie operations
func TestKit_Cookies(t *testing.T) {
	t.Run("sets and gets cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Response: w, Request: r}

		k.SetCookie("session", "abc123")

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "session", cookies[0].Name)
		assert.Equal(t, "abc123", cookies[0].Value)
		assert.Equal(t, "/", cookies[0].Path)
		assert.True(t, cookies[0].HttpOnly)
	})

	t.Run("gets cookie value", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{Name: "session", Value: "xyz789"})

		k := &Kit{Request: r}

		value, err := k.GetCookie("session")
		require.NoError(t, err)
		assert.Equal(t, "xyz789", value)
	})

	t.Run("returns error for missing cookie", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Request: r}

		_, err := k.GetCookie("missing")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, twineerrors.ErrGetCookie))
	})

	t.Run("handles multiple cookies", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{Name: "cookie1", Value: "value1"})
		r.AddCookie(&http.Cookie{Name: "cookie2", Value: "value2"})

		k := &Kit{Request: r}

		value1, err1 := k.GetCookie("cookie1")
		value2, err2 := k.GetCookie("cookie2")

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Equal(t, "value1", value1)
		assert.Equal(t, "value2", value2)
	})

	t.Run("cookie attributes are correct", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		k := &Kit{Response: w, Request: r}

		k.SetCookie("test", "value")

		cookie := w.Result().Cookies()[0]
		assert.Equal(t, "/", cookie.Path)
		assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
		assert.False(t, cookie.Secure) // Dev mode
		assert.True(t, cookie.HttpOnly)
		assert.False(t, cookie.Expires.IsZero())
	})
}
