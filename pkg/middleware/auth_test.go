package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cstone-io/twine/pkg/auth"
	"github.com/cstone-io/twine/pkg/kit"
)

// setupTestAuth sets up test environment with auth secret
func setupTestAuth(t *testing.T) func() {
	t.Helper()

	// Save original env
	originalSecret := os.Getenv("AUTH_SECRET")

	// Set test secret
	os.Setenv("AUTH_SECRET", "test-secret-key-for-testing")

	return func() {
		if originalSecret == "" {
			os.Unsetenv("AUTH_SECRET")
		} else {
			os.Setenv("AUTH_SECRET", originalSecret)
		}
	}
}

// TestJWTMiddleware tests JWT authentication middleware
func TestJWTMiddleware(t *testing.T) {
	cleanup := setupTestAuth(t)
	defer cleanup()

	t.Run("allows request with valid token in header", func(t *testing.T) {
		userID := uuid.New()
		token, err := auth.NewToken(userID, "test@example.com")
		require.NoError(t, err)

		handlerCalled := false

		mw := JWTMiddleware()
		handler := func(k *kit.Kit) error {
			handlerCalled = true
			return k.Text(200, "ok")
		}

		wrapped := mw(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Authorization", "Bearer "+token.Token)

		k := &kit.Kit{Response: w, Request: r}

		err = wrapped(k)
		require.NoError(t, err)
		assert.True(t, handlerCalled)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("allows request with valid token in cookie", func(t *testing.T) {
		userID := uuid.New()
		token, err := auth.NewToken(userID, "test@example.com")
		require.NoError(t, err)

		handlerCalled := false

		mw := JWTMiddleware()
		handler := func(k *kit.Kit) error {
			handlerCalled = true
			return k.Text(200, "ok")
		}

		wrapped := mw(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{Name: "token", Value: token.Token})

		k := &kit.Kit{Response: w, Request: r}

		err = wrapped(k)
		require.NoError(t, err)
		assert.True(t, handlerCalled)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("redirects on missing token", func(t *testing.T) {
		handlerCalled := false

		mw := JWTMiddleware()
		handler := func(k *kit.Kit) error {
			handlerCalled = true
			return k.Text(200, "ok")
		}

		wrapped := mw(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/protected", nil)

		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)

		assert.False(t, handlerCalled)
		assert.Equal(t, 303, w.Code)
		assert.Equal(t, "/auth/login", w.Header().Get("Location"))
	})

	t.Run("redirects on invalid token", func(t *testing.T) {
		handlerCalled := false

		mw := JWTMiddleware()
		handler := func(k *kit.Kit) error {
			handlerCalled = true
			return k.Text(200, "ok")
		}

		wrapped := mw(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/protected", nil)
		r.Header.Set("Authorization", "Bearer invalid-token")

		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)

		assert.False(t, handlerCalled)
		assert.Equal(t, 303, w.Code)
		assert.Equal(t, "/auth/login", w.Header().Get("Location"))
	})

	t.Run("sets user context on success", func(t *testing.T) {
		userID := uuid.New()
		token, err := auth.NewToken(userID, "test@example.com")
		require.NoError(t, err)

		var capturedUserID string

		mw := JWTMiddleware()
		handler := func(k *kit.Kit) error {
			capturedUserID = k.GetContext("user")
			return k.Text(200, "ok")
		}

		wrapped := mw(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Authorization", "Bearer "+token.Token)

		k := &kit.Kit{Response: w, Request: r}

		err = wrapped(k)
		require.NoError(t, err)
		assert.Equal(t, userID.String(), capturedUserID)
	})

	t.Run("handles Ajax redirect on auth failure", func(t *testing.T) {
		mw := JWTMiddleware()
		handler := func(k *kit.Kit) error {
			return k.Text(200, "ok")
		}

		wrapped := mw(handler)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/protected", nil)
		r.Header.Set("X-Alpine-Request", "true")

		k := &kit.Kit{Response: w, Request: r}

		err := wrapped(k)
		require.NoError(t, err)

		assert.Equal(t, 303, w.Code)
		assert.Equal(t, "/auth/login", w.Header().Get("Location"))
	})
}

// TestJWTMiddleware_Integration tests realistic authentication scenarios
func TestJWTMiddleware_Integration(t *testing.T) {
	cleanup := setupTestAuth(t)
	defer cleanup()

	t.Run("protects multiple routes", func(t *testing.T) {
		userID := uuid.New()
		token, err := auth.NewToken(userID, "test@example.com")
		require.NoError(t, err)

		mw := JWTMiddleware()

		// Route 1
		handler1 := func(k *kit.Kit) error {
			return k.Text(200, "route1")
		}

		// Route 2
		handler2 := func(k *kit.Kit) error {
			return k.Text(200, "route2")
		}

		wrapped1 := mw(handler1)
		wrapped2 := mw(handler2)

		// Test route 1
		w1 := httptest.NewRecorder()
		r1 := httptest.NewRequest("GET", "/route1", nil)
		r1.Header.Set("Authorization", "Bearer "+token.Token)
		k1 := &kit.Kit{Response: w1, Request: r1}
		err = wrapped1(k1)
		require.NoError(t, err)
		assert.Equal(t, "route1", w1.Body.String())

		// Test route 2
		w2 := httptest.NewRecorder()
		r2 := httptest.NewRequest("GET", "/route2", nil)
		r2.Header.Set("Authorization", "Bearer "+token.Token)
		k2 := &kit.Kit{Response: w2, Request: r2}
		err = wrapped2(k2)
		require.NoError(t, err)
		assert.Equal(t, "route2", w2.Body.String())
	})

	t.Run("combines with other middleware", func(t *testing.T) {
		userID := uuid.New()
		token, err := auth.NewToken(userID, "test@example.com")
		require.NoError(t, err)

		logged := false

		loggingMW := func(next kit.HandlerFunc) kit.HandlerFunc {
			return func(k *kit.Kit) error {
				logged = true
				return next(k)
			}
		}

		authMW := JWTMiddleware()

		handler := func(k *kit.Kit) error {
			return k.Text(200, "ok")
		}

		// Apply both middlewares
		wrapped := ApplyMiddlewares(handler, authMW, loggingMW)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Authorization", "Bearer "+token.Token)
		k := &kit.Kit{Response: w, Request: r}

		err = wrapped(k)
		require.NoError(t, err)
		assert.True(t, logged)
		assert.Equal(t, "ok", w.Body.String())
	})

	t.Run("different users have different contexts", func(t *testing.T) {
		user1ID := uuid.New()
		user2ID := uuid.New()

		token1, err := auth.NewToken(user1ID, "user1@example.com")
		require.NoError(t, err)

		token2, err := auth.NewToken(user2ID, "user2@example.com")
		require.NoError(t, err)

		mw := JWTMiddleware()

		handler := func(k *kit.Kit) error {
			userID := k.GetContext("user")
			return k.Text(200, userID)
		}

		wrapped := mw(handler)

		// User 1
		w1 := httptest.NewRecorder()
		r1 := httptest.NewRequest("GET", "/", nil)
		r1.Header.Set("Authorization", "Bearer "+token1.Token)
		k1 := &kit.Kit{Response: w1, Request: r1}
		err = wrapped(k1)
		require.NoError(t, err)
		assert.Equal(t, user1ID.String(), w1.Body.String())

		// User 2
		w2 := httptest.NewRecorder()
		r2 := httptest.NewRequest("GET", "/", nil)
		r2.Header.Set("Authorization", "Bearer "+token2.Token)
		k2 := &kit.Kit{Response: w2, Request: r2}
		err = wrapped(k2)
		require.NoError(t, err)
		assert.Equal(t, user2ID.String(), w2.Body.String())
	})
}
