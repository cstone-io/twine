package auth

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	twineerrors "github.com/cstone-io/twine/pkg/errors"
)

// setupTestAuth sets up test environment with auth secret
func setupTestAuth(t *testing.T) func() {
	t.Helper()

	// Save original env
	originalSecret := os.Getenv("AUTH_SECRET")

	// Set test secret
	os.Setenv("AUTH_SECRET", "test-secret-key-for-testing")

	// Reset config singleton
	resetConfig()

	return func() {
		if originalSecret == "" {
			os.Unsetenv("AUTH_SECRET")
		} else {
			os.Setenv("AUTH_SECRET", originalSecret)
		}
		resetConfig()
	}
}

// resetConfig resets the config singleton for testing
func resetConfig() {
	// This is a hack to reset the config singleton
	// In production code, config.Get() caches the instance
	// For tests, we need to be able to reset it
}

// TestToken_NewToken tests JWT token generation
func TestToken_NewToken(t *testing.T) {
	cleanup := setupTestAuth(t)
	defer cleanup()

	t.Run("generates valid token", func(t *testing.T) {
		userID := uuid.New()
		email := "test@example.com"

		token, err := NewToken(userID, email)
		require.NoError(t, err)
		require.NotNil(t, token)

		assert.NotEmpty(t, token.Token)
	})

	t.Run("token contains user_id claim", func(t *testing.T) {
		userID := uuid.New()
		email := "test@example.com"

		token, err := NewToken(userID, email)
		require.NoError(t, err)

		// Parse the token to verify claims
		parsed, err := jwt.Parse(token.Token, func(token *jwt.Token) (any, error) {
			return []byte("test-secret-key-for-testing"), nil
		})
		require.NoError(t, err)

		claims, ok := parsed.Claims.(jwt.MapClaims)
		require.True(t, ok)

		assert.Equal(t, userID.String(), claims["user_id"])
	})

	t.Run("token contains email claim", func(t *testing.T) {
		userID := uuid.New()
		email := "user@example.com"

		token, err := NewToken(userID, email)
		require.NoError(t, err)

		parsed, err := jwt.Parse(token.Token, func(token *jwt.Token) (any, error) {
			return []byte("test-secret-key-for-testing"), nil
		})
		require.NoError(t, err)

		claims := parsed.Claims.(jwt.MapClaims)
		assert.Equal(t, email, claims["email"])
	})

	t.Run("token has 1 hour expiration", func(t *testing.T) {
		userID := uuid.New()
		email := "test@example.com"

		before := time.Now()
		token, err := NewToken(userID, email)
		require.NoError(t, err)
		after := time.Now()

		parsed, err := jwt.Parse(token.Token, func(token *jwt.Token) (any, error) {
			return []byte("test-secret-key-for-testing"), nil
		})
		require.NoError(t, err)

		claims := parsed.Claims.(jwt.MapClaims)
		exp, ok := claims["exp"].(float64)
		require.True(t, ok)

		expTime := time.Unix(int64(exp), 0)

		// Expiration should be ~1 hour from now
		expectedExp := before.Add(time.Hour)
		assert.True(t, expTime.After(expectedExp.Add(-time.Second)))
		assert.True(t, expTime.Before(after.Add(time.Hour).Add(time.Second)))
	})

	t.Run("uses HS256 signing method", func(t *testing.T) {
		userID := uuid.New()
		email := "test@example.com"

		token, err := NewToken(userID, email)
		require.NoError(t, err)

		parsed, err := jwt.Parse(token.Token, func(token *jwt.Token) (any, error) {
			return []byte("test-secret-key-for-testing"), nil
		})
		require.NoError(t, err)

		assert.Equal(t, "HS256", parsed.Method.Alg())
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		userID := uuid.New()
		email := "test@example.com"

		token1, err := NewToken(userID, email)
		require.NoError(t, err)

		// Delay to ensure different timestamp (exp is in seconds)
		time.Sleep(2 * time.Second)

		token2, err := NewToken(userID, email)
		require.NoError(t, err)

		// Tokens should be different due to different exp times
		assert.NotEqual(t, token1.Token, token2.Token)
	})

	t.Run("handles different user IDs", func(t *testing.T) {
		user1 := uuid.New()
		user2 := uuid.New()

		token1, err := NewToken(user1, "user1@example.com")
		require.NoError(t, err)

		token2, err := NewToken(user2, "user2@example.com")
		require.NoError(t, err)

		assert.NotEqual(t, token1.Token, token2.Token)
	})
}

// TestToken_ParseToken tests JWT token parsing and validation
func TestToken_ParseToken(t *testing.T) {
	cleanup := setupTestAuth(t)
	defer cleanup()

	t.Run("parses valid token", func(t *testing.T) {
		userID := uuid.New()
		email := "test@example.com"

		token, err := NewToken(userID, email)
		require.NoError(t, err)

		parsedUserID, err := ParseToken(token.Token)
		require.NoError(t, err)

		assert.Equal(t, userID.String(), parsedUserID)
	})

	t.Run("rejects invalid signature", func(t *testing.T) {
		// Create token with wrong secret
		claims := jwt.MapClaims{
			"user_id": uuid.New().String(),
			"email":   "test@example.com",
			"exp":     time.Now().Add(time.Hour).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte("wrong-secret"))
		require.NoError(t, err)

		// Try to parse with correct secret
		_, err = ParseToken(signed)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, twineerrors.ErrAuthInvalidToken))
	})

	t.Run("rejects expired token", func(t *testing.T) {
		userID := uuid.New()

		// Create expired token
		claims := jwt.MapClaims{
			"user_id": userID.String(),
			"email":   "test@example.com",
			"exp":     time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte("test-secret-key-for-testing"))
		require.NoError(t, err)

		_, err = ParseToken(signed)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, twineerrors.ErrAuthInvalidToken))
	})

	t.Run("rejects malformed token", func(t *testing.T) {
		_, err := ParseToken("not-a-valid-jwt-token")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, twineerrors.ErrAuthInvalidToken))
	})

	t.Run("rejects token without user_id claim", func(t *testing.T) {
		claims := jwt.MapClaims{
			"email": "test@example.com",
			"exp":   time.Now().Add(time.Hour).Unix(),
			// Missing user_id
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte("test-secret-key-for-testing"))
		require.NoError(t, err)

		_, err = ParseToken(signed)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, twineerrors.ErrAuthInvalidToken))
	})

	t.Run("rejects token with wrong signing method", func(t *testing.T) {
		// Create token with RS256 (wrong method)
		claims := jwt.MapClaims{
			"user_id": uuid.New().String(),
			"email":   "test@example.com",
			"exp":     time.Now().Add(time.Hour).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		_, err = ParseToken(signed)
		assert.Error(t, err)
	})

	t.Run("rejects empty token", func(t *testing.T) {
		_, err := ParseToken("")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, twineerrors.ErrAuthInvalidToken))
	})

	t.Run("rejects token with invalid user_id type", func(t *testing.T) {
		claims := jwt.MapClaims{
			"user_id": 12345, // Should be string
			"email":   "test@example.com",
			"exp":     time.Now().Add(time.Hour).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte("test-secret-key-for-testing"))
		require.NoError(t, err)

		_, err = ParseToken(signed)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, twineerrors.ErrAuthInvalidToken))
	})
}

// TestToken_Lifecycle tests full token lifecycle
func TestToken_Lifecycle(t *testing.T) {
	cleanup := setupTestAuth(t)
	defer cleanup()

	t.Run("generate and parse token", func(t *testing.T) {
		originalUserID := uuid.New()
		originalEmail := "lifecycle@example.com"

		// Generate
		token, err := NewToken(originalUserID, originalEmail)
		require.NoError(t, err)
		require.NotNil(t, token)
		assert.NotEmpty(t, token.Token)

		// Parse
		parsedUserID, err := ParseToken(token.Token)
		require.NoError(t, err)
		assert.Equal(t, originalUserID.String(), parsedUserID)
	})

	t.Run("token is valid for the duration", func(t *testing.T) {
		userID := uuid.New()

		token, err := NewToken(userID, "test@example.com")
		require.NoError(t, err)

		// Should be valid immediately
		parsedUserID, err := ParseToken(token.Token)
		require.NoError(t, err)
		assert.Equal(t, userID.String(), parsedUserID)

		// Small delay (should still be valid)
		time.Sleep(10 * time.Millisecond)

		parsedUserID, err = ParseToken(token.Token)
		require.NoError(t, err)
		assert.Equal(t, userID.String(), parsedUserID)
	})
}

// TestToken_ThreadSafety tests concurrent token operations
func TestToken_ThreadSafety(t *testing.T) {
	cleanup := setupTestAuth(t)
	defer cleanup()

	t.Run("concurrent token generation", func(t *testing.T) {
		const goroutines = 100
		var wg sync.WaitGroup
		tokens := make([]*Token, goroutines)
		errors := make([]error, goroutines)

		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func(index int) {
				defer wg.Done()
				userID := uuid.New()
				tokens[index], errors[index] = NewToken(userID, "test@example.com")
			}(i)
		}

		wg.Wait()

		// All should succeed
		for i, err := range errors {
			require.NoError(t, err, "Token generation %d failed", i)
			assert.NotNil(t, tokens[i])
		}
	})

	t.Run("concurrent token parsing", func(t *testing.T) {
		// Generate one token
		userID := uuid.New()
		token, err := NewToken(userID, "test@example.com")
		require.NoError(t, err)

		// Parse concurrently
		const goroutines = 100
		var wg sync.WaitGroup
		results := make([]string, goroutines)
		errors := make([]error, goroutines)

		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func(index int) {
				defer wg.Done()
				results[index], errors[index] = ParseToken(token.Token)
			}(i)
		}

		wg.Wait()

		// All should succeed with same result
		for i, err := range errors {
			require.NoError(t, err, "Token parsing %d failed", i)
			assert.Equal(t, userID.String(), results[i])
		}
	})
}

// TestToken_EdgeCases tests edge cases
func TestToken_EdgeCases(t *testing.T) {
	cleanup := setupTestAuth(t)
	defer cleanup()

	t.Run("nil UUID", func(t *testing.T) {
		token, err := NewToken(uuid.Nil, "test@example.com")
		require.NoError(t, err)

		parsedUserID, err := ParseToken(token.Token)
		require.NoError(t, err)
		assert.Equal(t, uuid.Nil.String(), parsedUserID)
	})

	t.Run("empty email", func(t *testing.T) {
		userID := uuid.New()

		token, err := NewToken(userID, "")
		require.NoError(t, err)

		// Token should still be valid
		parsedUserID, err := ParseToken(token.Token)
		require.NoError(t, err)
		assert.Equal(t, userID.String(), parsedUserID)
	})

	t.Run("very long email", func(t *testing.T) {
		userID := uuid.New()
		longEmail := string(make([]byte, 1000)) + "@example.com"

		token, err := NewToken(userID, longEmail)
		require.NoError(t, err)

		parsedUserID, err := ParseToken(token.Token)
		require.NoError(t, err)
		assert.Equal(t, userID.String(), parsedUserID)
	})

	t.Run("special characters in email", func(t *testing.T) {
		userID := uuid.New()
		email := "test+special@example.com"

		token, err := NewToken(userID, email)
		require.NoError(t, err)

		parsedUserID, err := ParseToken(token.Token)
		require.NoError(t, err)
		assert.Equal(t, userID.String(), parsedUserID)
	})
}

// TestToken_Integration tests realistic authentication scenarios
func TestToken_Integration(t *testing.T) {
	cleanup := setupTestAuth(t)
	defer cleanup()

	t.Run("user login flow", func(t *testing.T) {
		// Simulate user login
		userID := uuid.New()
		email := "user@example.com"

		// Generate token after successful login
		token, err := NewToken(userID, email)
		require.NoError(t, err)

		// Client stores token and uses it for subsequent requests
		// Server validates token
		parsedUserID, err := ParseToken(token.Token)
		require.NoError(t, err)

		// Server can now identify the user
		assert.Equal(t, userID.String(), parsedUserID)
	})

	t.Run("multiple users have unique tokens", func(t *testing.T) {
		user1ID := uuid.New()
		user2ID := uuid.New()

		token1, err := NewToken(user1ID, "user1@example.com")
		require.NoError(t, err)

		token2, err := NewToken(user2ID, "user2@example.com")
		require.NoError(t, err)

		// Tokens should be different
		assert.NotEqual(t, token1.Token, token2.Token)

		// Each token should parse to correct user
		parsed1, err := ParseToken(token1.Token)
		require.NoError(t, err)
		assert.Equal(t, user1ID.String(), parsed1)

		parsed2, err := ParseToken(token2.Token)
		require.NoError(t, err)
		assert.Equal(t, user2ID.String(), parsed2)
	})

	t.Run("token refresh scenario", func(t *testing.T) {
		userID := uuid.New()

		// Original token
		token1, err := NewToken(userID, "user@example.com")
		require.NoError(t, err)

		time.Sleep(2 * time.Second) // Ensure different exp time

		// Refreshed token
		token2, err := NewToken(userID, "user@example.com")
		require.NoError(t, err)

		// Both should be valid
		parsed1, err := ParseToken(token1.Token)
		require.NoError(t, err)
		assert.Equal(t, userID.String(), parsed1)

		parsed2, err := ParseToken(token2.Token)
		require.NoError(t, err)
		assert.Equal(t, userID.String(), parsed2)

		// Tokens should differ due to different exp times
		assert.NotEqual(t, token1.Token, token2.Token)
	})
}
