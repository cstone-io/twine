package auth

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	twineerrors "github.com/cstone-io/twine/pkg/errors"
)

// TestHashPassword tests password hashing
func TestHashPassword(t *testing.T) {
	t.Run("hashes password successfully", func(t *testing.T) {
		password := "mySecurePassword123"

		hash, err := HashPassword(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
	})

	t.Run("hash is different from password", func(t *testing.T) {
		password := "plaintext"

		hash, err := HashPassword(password)
		require.NoError(t, err)

		assert.NotEqual(t, password, hash)
	})

	t.Run("hash starts with bcrypt prefix", func(t *testing.T) {
		password := "test123"

		hash, err := HashPassword(password)
		require.NoError(t, err)

		// Bcrypt hashes start with $2a$, $2b$, or $2y$
		assert.True(t, strings.HasPrefix(hash, "$2"))
	})

	t.Run("generates unique hashes for same password", func(t *testing.T) {
		password := "samePassword"

		hash1, err := HashPassword(password)
		require.NoError(t, err)

		hash2, err := HashPassword(password)
		require.NoError(t, err)

		// Same password should produce different hashes (bcrypt uses salt)
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("hashes different passwords differently", func(t *testing.T) {
		password1 := "password1"
		password2 := "password2"

		hash1, err := HashPassword(password1)
		require.NoError(t, err)

		hash2, err := HashPassword(password2)
		require.NoError(t, err)

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("handles empty password", func(t *testing.T) {
		password := ""

		hash, err := HashPassword(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
	})

	t.Run("handles very long password", func(t *testing.T) {
		// Bcrypt has a 72-byte limit
		password := strings.Repeat("a", 72)

		hash, err := HashPassword(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
	})

	t.Run("rejects password exceeding 72 bytes", func(t *testing.T) {
		// Bcrypt has a 72-byte limit
		password := strings.Repeat("a", 100)

		_, err := HashPassword(password)
		assert.Error(t, err)
	})

	t.Run("handles special characters", func(t *testing.T) {
		password := "p@ssw0rd!#$%^&*()"

		hash, err := HashPassword(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
	})

	t.Run("handles unicode characters", func(t *testing.T) {
		password := "пароль密码🔒"

		hash, err := HashPassword(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
	})

	t.Run("hash can be verified with bcrypt", func(t *testing.T) {
		password := "testPassword"

		hash, err := HashPassword(password)
		require.NoError(t, err)

		// Verify using bcrypt directly
		err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		assert.NoError(t, err)
	})
}

// TestCredentials_Authenticate tests password authentication
func TestCredentials_Authenticate(t *testing.T) {
	t.Run("authenticates with correct password", func(t *testing.T) {
		password := "correctPassword"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		creds := Credentials{
			Email:    "user@example.com",
			Password: password,
		}

		err = creds.Authenticate(hash)
		assert.NoError(t, err)
	})

	t.Run("rejects incorrect password", func(t *testing.T) {
		correctPassword := "correct"
		hash, err := HashPassword(correctPassword)
		require.NoError(t, err)

		creds := Credentials{
			Email:    "user@example.com",
			Password: "wrong",
		}

		err = creds.Authenticate(hash)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, twineerrors.ErrAuthInvalidCredentials))
	})

	t.Run("rejects empty password against valid hash", func(t *testing.T) {
		password := "validPassword"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		creds := Credentials{
			Email:    "user@example.com",
			Password: "",
		}

		err = creds.Authenticate(hash)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, twineerrors.ErrAuthInvalidCredentials))
	})

	t.Run("rejects valid password against empty hash", func(t *testing.T) {
		creds := Credentials{
			Email:    "user@example.com",
			Password: "validPassword",
		}

		err := creds.Authenticate("")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, twineerrors.ErrAuthInvalidCredentials))
	})

	t.Run("rejects password against invalid hash", func(t *testing.T) {
		creds := Credentials{
			Email:    "user@example.com",
			Password: "password",
		}

		err := creds.Authenticate("not-a-valid-bcrypt-hash")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, twineerrors.ErrAuthInvalidCredentials))
	})

	t.Run("password is case-sensitive", func(t *testing.T) {
		password := "Password"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		creds := Credentials{
			Email:    "user@example.com",
			Password: "password", // lowercase
		}

		err = creds.Authenticate(hash)
		assert.Error(t, err)
	})

	t.Run("handles whitespace in password", func(t *testing.T) {
		password := "pass word"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		// Correct with whitespace
		creds1 := Credentials{Password: password}
		err = creds1.Authenticate(hash)
		assert.NoError(t, err)

		// Wrong without whitespace
		creds2 := Credentials{Password: "password"}
		err = creds2.Authenticate(hash)
		assert.Error(t, err)
	})

	t.Run("handles special characters", func(t *testing.T) {
		password := "p@ss!w0rd#"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		creds := Credentials{Password: password}
		err = creds.Authenticate(hash)
		assert.NoError(t, err)
	})

	t.Run("handles unicode passwords", func(t *testing.T) {
		password := "密码🔐"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		creds := Credentials{Password: password}
		err = creds.Authenticate(hash)
		assert.NoError(t, err)
	})
}

// TestCredentials_Struct tests the Credentials struct
func TestCredentials_Struct(t *testing.T) {
	t.Run("stores email and password", func(t *testing.T) {
		creds := Credentials{
			Email:    "test@example.com",
			Password: "password123",
		}

		assert.Equal(t, "test@example.com", creds.Email)
		assert.Equal(t, "password123", creds.Password)
	})

	t.Run("has json tags", func(t *testing.T) {
		// This test just documents that the struct has json tags
		// The actual tag values are checked by looking at the struct definition
		creds := Credentials{
			Email:    "user@example.com",
			Password: "secret",
		}

		// Verify fields are accessible
		assert.NotEmpty(t, creds.Email)
		assert.NotEmpty(t, creds.Password)
	})

	t.Run("has form tags", func(t *testing.T) {
		// This test documents that the struct has form tags
		// The actual parsing is tested in kit package
		creds := Credentials{
			Email:    "user@example.com",
			Password: "secret",
		}

		assert.NotEmpty(t, creds.Email)
		assert.NotEmpty(t, creds.Password)
	})
}

// TestAuth_FullLifecycle tests complete authentication workflow
func TestAuth_FullLifecycle(t *testing.T) {
	t.Run("complete user registration and login flow", func(t *testing.T) {
		// Registration: user provides password
		originalPassword := "mySecurePassword123"

		// Hash and store password
		hashedPassword, err := HashPassword(originalPassword)
		require.NoError(t, err)

		// Login: user provides credentials
		loginCreds := Credentials{
			Email:    "user@example.com",
			Password: originalPassword,
		}

		// Authenticate
		err = loginCreds.Authenticate(hashedPassword)
		assert.NoError(t, err)
	})

	t.Run("failed login attempt", func(t *testing.T) {
		// Registration with one password
		registrationPassword := "correctPassword"
		hashedPassword, err := HashPassword(registrationPassword)
		require.NoError(t, err)

		// Login with wrong password
		loginCreds := Credentials{
			Email:    "user@example.com",
			Password: "wrongPassword",
		}

		// Should fail
		err = loginCreds.Authenticate(hashedPassword)
		assert.Error(t, err)
	})

	t.Run("password change flow", func(t *testing.T) {
		// Old password
		oldPassword := "oldPassword"
		oldHash, err := HashPassword(oldPassword)
		require.NoError(t, err)

		// Verify old password works
		oldCreds := Credentials{Password: oldPassword}
		err = oldCreds.Authenticate(oldHash)
		require.NoError(t, err)

		// User changes password
		newPassword := "newPassword"
		newHash, err := HashPassword(newPassword)
		require.NoError(t, err)

		// Old password should not work with new hash
		err = oldCreds.Authenticate(newHash)
		assert.Error(t, err)

		// New password should work
		newCreds := Credentials{Password: newPassword}
		err = newCreds.Authenticate(newHash)
		assert.NoError(t, err)
	})

	t.Run("multiple users have independent passwords", func(t *testing.T) {
		// User 1
		password1 := "user1password"
		hash1, err := HashPassword(password1)
		require.NoError(t, err)

		// User 2
		password2 := "user2password"
		hash2, err := HashPassword(password2)
		require.NoError(t, err)

		// Each user can auth with their own password
		creds1 := Credentials{Email: "user1@example.com", Password: password1}
		err = creds1.Authenticate(hash1)
		assert.NoError(t, err)

		creds2 := Credentials{Email: "user2@example.com", Password: password2}
		err = creds2.Authenticate(hash2)
		assert.NoError(t, err)

		// But not with each other's passwords
		err = creds1.Authenticate(hash2)
		assert.Error(t, err)

		err = creds2.Authenticate(hash1)
		assert.Error(t, err)
	})
}

// TestAuth_ThreadSafety tests concurrent authentication operations
func TestAuth_ThreadSafety(t *testing.T) {
	t.Run("concurrent password hashing", func(t *testing.T) {
		const goroutines = 100
		var wg sync.WaitGroup
		hashes := make([]string, goroutines)
		errors := make([]error, goroutines)

		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func(index int) {
				defer wg.Done()
				hashes[index], errors[index] = HashPassword("password")
			}(i)
		}

		wg.Wait()

		// All should succeed
		for i, err := range errors {
			require.NoError(t, err, "Hash %d failed", i)
			assert.NotEmpty(t, hashes[i])
		}

		// All hashes should be unique (bcrypt uses random salt)
		seen := make(map[string]bool)
		for _, hash := range hashes {
			assert.False(t, seen[hash], "Found duplicate hash")
			seen[hash] = true
		}
	})

	t.Run("concurrent authentication", func(t *testing.T) {
		password := "testPassword"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		const goroutines = 100
		var wg sync.WaitGroup
		errors := make([]error, goroutines)

		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func(index int) {
				defer wg.Done()
				creds := Credentials{Password: password}
				errors[index] = creds.Authenticate(hash)
			}(i)
		}

		wg.Wait()

		// All should succeed
		for i, err := range errors {
			require.NoError(t, err, "Authentication %d failed", i)
		}
	})

	t.Run("concurrent hash and auth", func(t *testing.T) {
		password := "concurrentTest"

		const goroutines = 50
		var wg sync.WaitGroup

		// Hash concurrently
		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				hash, err := HashPassword(password)
				require.NoError(t, err)

				// Immediately authenticate
				creds := Credentials{Password: password}
				err = creds.Authenticate(hash)
				require.NoError(t, err)
			}()
		}

		wg.Wait()
	})
}

// TestAuth_Security tests security properties
func TestAuth_Security(t *testing.T) {
	t.Run("hash is not reversible", func(t *testing.T) {
		password := "secretPassword"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		// Hash should not contain the password
		assert.NotContains(t, hash, password)
	})

	t.Run("timing attack resistance", func(t *testing.T) {
		// Bcrypt should have consistent timing regardless of password
		// This is a basic test - full timing attack testing requires more sophisticated analysis

		password := "password"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		// Try multiple wrong passwords
		wrongPasswords := []string{
			"wrong1",
			"wrong2",
			"completely different",
			"p", // Very short
			strings.Repeat("a", 100), // Very long
		}

		for _, wrong := range wrongPasswords {
			creds := Credentials{Password: wrong}
			err := creds.Authenticate(hash)
			assert.Error(t, err, "Should reject wrong password: %s", wrong)
		}
	})

	t.Run("salt makes rainbow tables ineffective", func(t *testing.T) {
		// Same password hashed multiple times produces different hashes
		password := "commonPassword123"

		hashes := make([]string, 10)
		for i := 0; i < 10; i++ {
			var err error
			hashes[i], err = HashPassword(password)
			require.NoError(t, err)
		}

		// All hashes should be different
		for i := 0; i < len(hashes); i++ {
			for j := i + 1; j < len(hashes); j++ {
				assert.NotEqual(t, hashes[i], hashes[j],
					"Hashes %d and %d should be different", i, j)
			}
		}

		// But all should verify against the same password
		for _, hash := range hashes {
			creds := Credentials{Password: password}
			err := creds.Authenticate(hash)
			assert.NoError(t, err)
		}
	})
}

// TestAuth_EdgeCases tests edge cases and boundary conditions
func TestAuth_EdgeCases(t *testing.T) {
	t.Run("password at 72 byte limit", func(t *testing.T) {
		// Bcrypt has a max length of 72 bytes
		password := strings.Repeat("a", 72)

		hash, err := HashPassword(password)
		require.NoError(t, err)

		creds := Credentials{Password: password}
		err = creds.Authenticate(hash)
		assert.NoError(t, err)
	})

	t.Run("password with null bytes", func(t *testing.T) {
		password := "pass\x00word"

		hash, err := HashPassword(password)
		require.NoError(t, err)

		creds := Credentials{Password: password}
		err = creds.Authenticate(hash)
		assert.NoError(t, err)
	})

	t.Run("many printable ASCII characters", func(t *testing.T) {
		// Password with many printable ASCII (within 72-byte limit)
		password := "!\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]"

		hash, err := HashPassword(password)
		require.NoError(t, err)

		creds := Credentials{Password: password}
		err = creds.Authenticate(hash)
		assert.NoError(t, err)
	})

	t.Run("password that looks like a hash", func(t *testing.T) {
		// Password that starts like a bcrypt hash
		password := "$2a$10$fakeHashLikingPassword"

		hash, err := HashPassword(password)
		require.NoError(t, err)

		creds := Credentials{Password: password}
		err = creds.Authenticate(hash)
		assert.NoError(t, err)
	})
}
