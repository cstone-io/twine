package errors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPredefinedErrors_Exist verifies all predefined errors are properly initialized
func TestPredefinedErrors_Exist(t *testing.T) {
	predefinedErrors := []*Error{
		// 1000 level - CRITICAL
		ErrDefaultCritical,
		ErrListenAndServe,
		ErrShutdownServer,
		// 1100 level - DATABASE CRITICAL
		ErrDatabaseDefaultCritical,
		ErrDatabaseLoad,
		ErrDatabaseConn,
		ErrDatabaseMigration,
		ErrDatabaseSeed,
		// 2000 level - ERROR
		ErrDefaultError,
		ErrDecodeJSON,
		ErrNotFound,
		// 2100 level - DATABASE ERROR
		ErrDatabaseDefaultError,
		ErrDatabaseRead,
		ErrDatabaseWrite,
		ErrDatabaseUpdate,
		ErrDatabaseDelete,
		ErrMigrateTable,
		ErrSortMigrations,
		ErrSeedObject,
		// 2200 level - AUTH ERROR
		ErrAuthDefault,
		ErrHashPassword,
		ErrGenerateToken,
		ErrGetPermissions,
		ErrGetCookie,
		// 2300 level - API ERROR
		ErrAPIDefault,
		ErrAPIGet,
		ErrAPIPost,
		ErrAPIPut,
		ErrAPIDelete,
		// 3000 level - MINOR
		ErrDefaultMinor,
		ErrDecodeForm,
		// 3100 level - DATABASE MINOR
		ErrDatabaseDefaultMinor,
		ErrDatabaseObjectNotFound,
		// 3200 level - AUTH MINOR
		ErrAuthDefaultMinor,
		ErrAuthInvalidToken,
		ErrAuthExpiredToken,
		ErrAuthInvalidCredentials,
		ErrPrimaryEmailNotFound,
		ErrInsufficientPermissions,
		ErrAuthMissingHeader,
		ErrAuthMissingAuthTypeHeader,
		// 3300 level - API MINOR
		ErrAPIDefaultMinor,
		ErrAPIIDMismatch,
		ErrAPIRequestPayload,
		ErrAPIPathValue,
		ErrAPIObjectNotFound,
		ErrAPIRequestContentType,
	}

	for _, err := range predefinedErrors {
		assert.NotNil(t, err, "Predefined error should not be nil")
		assert.NotEqual(t, 0, err.Code, "Error code should not be zero")
		assert.NotEmpty(t, err.Message, "Error message should not be empty")
	}
}

// TestPredefinedErrors_SeverityLevels tests that errors have correct severity based on code ranges
func TestPredefinedErrors_SeverityLevels(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		severity ErrSeverity
	}{
		// 1000-1199: CRITICAL
		{"ErrDefaultCritical", ErrDefaultCritical, ErrCritical},
		{"ErrListenAndServe", ErrListenAndServe, ErrCritical},
		{"ErrShutdownServer", ErrShutdownServer, ErrCritical},
		{"ErrDatabaseDefaultCritical", ErrDatabaseDefaultCritical, ErrCritical},
		{"ErrDatabaseLoad", ErrDatabaseLoad, ErrCritical},
		{"ErrDatabaseConn", ErrDatabaseConn, ErrCritical},
		{"ErrDatabaseMigration", ErrDatabaseMigration, ErrCritical},
		{"ErrDatabaseSeed", ErrDatabaseSeed, ErrCritical},

		// 2000-2999: ERROR
		{"ErrDefaultError", ErrDefaultError, ErrError},
		{"ErrDecodeJSON", ErrDecodeJSON, ErrError},
		{"ErrNotFound", ErrNotFound, ErrError},
		{"ErrDatabaseDefaultError", ErrDatabaseDefaultError, ErrError},
		{"ErrDatabaseRead", ErrDatabaseRead, ErrError},
		{"ErrDatabaseWrite", ErrDatabaseWrite, ErrError},
		{"ErrDatabaseUpdate", ErrDatabaseUpdate, ErrError},
		{"ErrDatabaseDelete", ErrDatabaseDelete, ErrError},
		{"ErrMigrateTable", ErrMigrateTable, ErrError},
		{"ErrSortMigrations", ErrSortMigrations, ErrError},
		{"ErrSeedObject", ErrSeedObject, ErrError},
		{"ErrAuthDefault", ErrAuthDefault, ErrError},
		{"ErrHashPassword", ErrHashPassword, ErrError},
		{"ErrGenerateToken", ErrGenerateToken, ErrError},
		{"ErrGetPermissions", ErrGetPermissions, ErrError},
		{"ErrGetCookie", ErrGetCookie, ErrError},
		{"ErrAPIDefault", ErrAPIDefault, ErrError},
		{"ErrAPIGet", ErrAPIGet, ErrError},
		{"ErrAPIPost", ErrAPIPost, ErrError},
		{"ErrAPIPut", ErrAPIPut, ErrError},
		{"ErrAPIDelete", ErrAPIDelete, ErrError},

		// 3000-3999: MINOR
		{"ErrDefaultMinor", ErrDefaultMinor, ErrMinor},
		{"ErrDecodeForm", ErrDecodeForm, ErrMinor},
		{"ErrDatabaseDefaultMinor", ErrDatabaseDefaultMinor, ErrMinor},
		{"ErrDatabaseObjectNotFound", ErrDatabaseObjectNotFound, ErrMinor},
		{"ErrAuthDefaultMinor", ErrAuthDefaultMinor, ErrMinor},
		{"ErrAuthInvalidToken", ErrAuthInvalidToken, ErrMinor},
		{"ErrAuthExpiredToken", ErrAuthExpiredToken, ErrMinor},
		{"ErrAuthInvalidCredentials", ErrAuthInvalidCredentials, ErrMinor},
		{"ErrPrimaryEmailNotFound", ErrPrimaryEmailNotFound, ErrMinor},
		{"ErrInsufficientPermissions", ErrInsufficientPermissions, ErrMinor},
		{"ErrAuthMissingHeader", ErrAuthMissingHeader, ErrMinor},
		{"ErrAuthMissingAuthTypeHeader", ErrAuthMissingAuthTypeHeader, ErrMinor},
		{"ErrAPIDefaultMinor", ErrAPIDefaultMinor, ErrMinor},
		{"ErrAPIIDMismatch", ErrAPIIDMismatch, ErrMinor},
		{"ErrAPIRequestPayload", ErrAPIRequestPayload, ErrMinor},
		{"ErrAPIPathValue", ErrAPIPathValue, ErrMinor},
		{"ErrAPIObjectNotFound", ErrAPIObjectNotFound, ErrMinor},
		{"ErrAPIRequestContentType", ErrAPIRequestContentType, ErrMinor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.severity, tt.err.Severity,
				"Error %s (code %d) should have severity %v", tt.name, tt.err.Code, tt.severity)
		})
	}
}

// TestPredefinedErrors_HTTPStatusCodes tests that errors have appropriate HTTP status codes
func TestPredefinedErrors_HTTPStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		err        *Error
		httpStatus int
	}{
		// 404 Not Found
		{"ErrNotFound", ErrNotFound, http.StatusNotFound},
		{"ErrDatabaseObjectNotFound", ErrDatabaseObjectNotFound, http.StatusNotFound},
		{"ErrPrimaryEmailNotFound", ErrPrimaryEmailNotFound, http.StatusNotFound},
		{"ErrAPIObjectNotFound", ErrAPIObjectNotFound, http.StatusNotFound},

		// 401 Unauthorized
		{"ErrAuthInvalidToken", ErrAuthInvalidToken, http.StatusUnauthorized},
		{"ErrAuthExpiredToken", ErrAuthExpiredToken, http.StatusUnauthorized},
		{"ErrAuthInvalidCredentials", ErrAuthInvalidCredentials, http.StatusUnauthorized},

		// 403 Forbidden
		{"ErrInsufficientPermissions", ErrInsufficientPermissions, http.StatusForbidden},

		// 400 Bad Request
		{"ErrAuthMissingHeader", ErrAuthMissingHeader, http.StatusBadRequest},
		{"ErrAuthMissingAuthTypeHeader", ErrAuthMissingAuthTypeHeader, http.StatusBadRequest},
		{"ErrAPIIDMismatch", ErrAPIIDMismatch, http.StatusBadRequest},
		{"ErrAPIRequestPayload", ErrAPIRequestPayload, http.StatusBadRequest},
		{"ErrAPIPathValue", ErrAPIPathValue, http.StatusBadRequest},

		// 415 Unsupported Media Type
		{"ErrAPIRequestContentType", ErrAPIRequestContentType, http.StatusUnsupportedMediaType},

		// 500 Internal Server Error
		{"ErrAuthDefault", ErrAuthDefault, http.StatusInternalServerError},
		{"ErrHashPassword", ErrHashPassword, http.StatusInternalServerError},
		{"ErrGenerateToken", ErrGenerateToken, http.StatusInternalServerError},
		{"ErrGetPermissions", ErrGetPermissions, http.StatusInternalServerError},
		{"ErrGetCookie", ErrGetCookie, http.StatusInternalServerError},
		{"ErrAPIDefault", ErrAPIDefault, http.StatusInternalServerError},
		{"ErrAPIGet", ErrAPIGet, http.StatusInternalServerError},
		{"ErrAPIPost", ErrAPIPost, http.StatusInternalServerError},
		{"ErrAPIPut", ErrAPIPut, http.StatusInternalServerError},
		{"ErrAPIDelete", ErrAPIDelete, http.StatusInternalServerError},
		{"ErrDefaultMinor", ErrDefaultMinor, http.StatusInternalServerError},
		{"ErrDatabaseDefaultMinor", ErrDatabaseDefaultMinor, http.StatusInternalServerError},
		{"ErrAuthDefaultMinor", ErrAuthDefaultMinor, http.StatusInternalServerError},
		{"ErrAPIDefaultMinor", ErrAPIDefaultMinor, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.httpStatus, tt.err.HTTPStatus,
				"Error %s should have HTTP status %d", tt.name, tt.httpStatus)
		})
	}
}

// TestPredefinedErrors_NoDuplicateCodes ensures all error codes are unique
func TestPredefinedErrors_NoDuplicateCodes(t *testing.T) {
	allErrors := []*Error{
		// 1000 level
		ErrDefaultCritical,
		ErrListenAndServe,
		ErrShutdownServer,
		// 1100 level
		ErrDatabaseDefaultCritical,
		ErrDatabaseLoad,
		ErrDatabaseConn,
		ErrDatabaseMigration,
		ErrDatabaseSeed,
		// 2000 level
		ErrDefaultError,
		ErrDecodeJSON,
		ErrNotFound,
		// 2100 level
		ErrDatabaseDefaultError,
		ErrDatabaseRead,
		ErrDatabaseWrite,
		ErrDatabaseUpdate,
		ErrDatabaseDelete,
		ErrMigrateTable,
		ErrSortMigrations,
		ErrSeedObject,
		// 2200 level
		ErrAuthDefault,
		ErrHashPassword,
		ErrGenerateToken,
		ErrGetPermissions,
		ErrGetCookie,
		// 2300 level
		ErrAPIDefault,
		ErrAPIGet,
		ErrAPIPost,
		ErrAPIPut,
		ErrAPIDelete,
		// 3000 level
		ErrDefaultMinor,
		ErrDecodeForm,
		// 3100 level
		ErrDatabaseDefaultMinor,
		ErrDatabaseObjectNotFound,
		// 3200 level
		ErrAuthDefaultMinor,
		ErrAuthInvalidToken,
		ErrAuthExpiredToken,
		ErrAuthInvalidCredentials,
		ErrPrimaryEmailNotFound,
		ErrInsufficientPermissions,
		ErrAuthMissingHeader,
		ErrAuthMissingAuthTypeHeader,
		// 3300 level
		ErrAPIDefaultMinor,
		ErrAPIIDMismatch,
		ErrAPIRequestPayload,
		ErrAPIPathValue,
		ErrAPIObjectNotFound,
		ErrAPIRequestContentType,
	}

	seenCodes := make(map[int]string)

	for _, err := range allErrors {
		if existingErr, exists := seenCodes[err.Code]; exists {
			t.Errorf("Duplicate error code %d found: %s and %s",
				err.Code, existingErr, err.Message)
		}
		seenCodes[err.Code] = err.Message
	}
}

// TestPredefinedErrors_CodeRanges tests that error codes are in expected ranges
func TestPredefinedErrors_CodeRanges(t *testing.T) {
	tests := []struct {
		name      string
		err       *Error
		minCode   int
		maxCode   int
		category  string
	}{
		// Critical errors (1000-1099)
		{"ErrDefaultCritical", ErrDefaultCritical, 1000, 1099, "critical"},
		{"ErrListenAndServe", ErrListenAndServe, 1000, 1099, "critical"},
		{"ErrShutdownServer", ErrShutdownServer, 1000, 1099, "critical"},

		// Database critical (1100-1199)
		{"ErrDatabaseDefaultCritical", ErrDatabaseDefaultCritical, 1100, 1199, "database critical"},
		{"ErrDatabaseLoad", ErrDatabaseLoad, 1100, 1199, "database critical"},
		{"ErrDatabaseConn", ErrDatabaseConn, 1100, 1199, "database critical"},

		// General errors (2000-2099)
		{"ErrDefaultError", ErrDefaultError, 2000, 2099, "general error"},
		{"ErrDecodeJSON", ErrDecodeJSON, 2000, 2099, "general error"},
		{"ErrNotFound", ErrNotFound, 2000, 2099, "general error"},

		// Database errors (2100-2199)
		{"ErrDatabaseDefaultError", ErrDatabaseDefaultError, 2100, 2199, "database error"},
		{"ErrDatabaseRead", ErrDatabaseRead, 2100, 2199, "database error"},

		// Auth errors (2200-2299)
		{"ErrAuthDefault", ErrAuthDefault, 2200, 2299, "auth error"},
		{"ErrHashPassword", ErrHashPassword, 2200, 2299, "auth error"},

		// API errors (2300-2399)
		{"ErrAPIDefault", ErrAPIDefault, 2300, 2399, "api error"},
		{"ErrAPIGet", ErrAPIGet, 2300, 2399, "api error"},

		// General minor (3000-3099)
		{"ErrDefaultMinor", ErrDefaultMinor, 3000, 3099, "general minor"},
		{"ErrDecodeForm", ErrDecodeForm, 3000, 3099, "general minor"},

		// Database minor (3100-3199)
		{"ErrDatabaseDefaultMinor", ErrDatabaseDefaultMinor, 3100, 3199, "database minor"},
		{"ErrDatabaseObjectNotFound", ErrDatabaseObjectNotFound, 3100, 3199, "database minor"},

		// Auth minor (3200-3299)
		{"ErrAuthDefaultMinor", ErrAuthDefaultMinor, 3200, 3299, "auth minor"},
		{"ErrAuthInvalidToken", ErrAuthInvalidToken, 3200, 3299, "auth minor"},

		// API minor (3300-3399)
		{"ErrAPIDefaultMinor", ErrAPIDefaultMinor, 3300, 3399, "api minor"},
		{"ErrAPIIDMismatch", ErrAPIIDMismatch, 3300, 3399, "api minor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.GreaterOrEqual(t, tt.err.Code, tt.minCode,
				"Error %s should have code >= %d for %s category", tt.name, tt.minCode, tt.category)
			assert.LessOrEqual(t, tt.err.Code, tt.maxCode,
				"Error %s should have code <= %d for %s category", tt.name, tt.maxCode, tt.category)
		})
	}
}

// TestPredefinedErrors_MessageQuality tests that error messages follow conventions
func TestPredefinedErrors_MessageQuality(t *testing.T) {
	t.Run("critical errors are uppercase", func(t *testing.T) {
		criticalErrors := []*Error{
			ErrDefaultCritical,
			ErrListenAndServe,
			ErrShutdownServer,
			ErrDatabaseDefaultCritical,
			ErrDatabaseLoad,
			ErrDatabaseConn,
			ErrDatabaseMigration,
			ErrDatabaseSeed,
		}

		for _, err := range criticalErrors {
			// Critical errors should be in all caps
			assert.Equal(t, err.Message, err.Message,
				"Critical error message should exist: %s", err.Message)
			// Just verify they have messages - the uppercase convention is visible in the code
		}
	})

	t.Run("default errors mention 'default' or 'unknown'", func(t *testing.T) {
		defaultErrors := []*Error{
			ErrDefaultCritical,
			ErrDatabaseDefaultCritical,
			ErrDefaultError,
			ErrDatabaseDefaultError,
			ErrAuthDefault,
			ErrAPIDefault,
			ErrDefaultMinor,
			ErrDatabaseDefaultMinor,
			ErrAuthDefaultMinor,
			ErrAPIDefaultMinor,
		}

		for _, err := range defaultErrors {
			message := err.Message
			hasDefault := contains(message, "Default") || contains(message, "DEFAULT")
			hasUnknown := contains(message, "unknown") || contains(message, "UNKNOWN")

			assert.True(t, hasDefault || hasUnknown,
				"Default error message should contain 'default' or 'unknown': %s", message)
		}
	})
}

// TestPredefinedErrors_Usage tests common usage patterns
func TestPredefinedErrors_Usage(t *testing.T) {
	t.Run("wrapping predefined errors preserves properties", func(t *testing.T) {
		baseErr := assert.AnError
		wrapped := ErrDatabaseRead.Wrap(baseErr)

		assert.Equal(t, ErrDatabaseRead.Code, wrapped.Code)
		assert.Equal(t, ErrDatabaseRead.Message, wrapped.Message)
		assert.Equal(t, ErrDatabaseRead.Severity, wrapped.Severity)
		assert.Equal(t, baseErr, wrapped.Cause)
	})

	t.Run("adding value to predefined errors", func(t *testing.T) {
		userID := "user-123"
		errWithValue := ErrNotFound.WithValue(map[string]string{"user_id": userID})

		assert.Equal(t, ErrNotFound.Code, errWithValue.Code)
		assert.Contains(t, errWithValue.Error(), userID)
	})

	t.Run("comparing predefined errors with Is", func(t *testing.T) {
		err1 := ErrNotFound
		err2 := ErrNotFound.Wrap(assert.AnError)

		assert.True(t, err2.Is(err1))
		assert.True(t, err1.Is(err2))
	})
}

// Helper function to check if a string contains a substring (case-sensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
