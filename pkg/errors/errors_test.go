package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestError_Error tests the Error() method string formatting
func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		error    *Error
		expected string
	}{
		{
			name: "simple error without cause or value",
			error: &Error{
				Code:    1000,
				Message: "test error",
			},
			expected: "1000: test error",
		},
		{
			name: "error with cause",
			error: &Error{
				Code:    2000,
				Message: "wrapped error",
				Cause:   fmt.Errorf("base error"),
			},
			expected: "2000: wrapped error: base error",
		},
		{
			name: "error with value",
			error: &Error{
				Code:    3000,
				Message: "error with context",
				Value:   map[string]string{"user": "john"},
			},
			expected: "3000: error with context, Value: map[user:john]",
		},
		{
			name: "error with both cause and value",
			error: &Error{
				Code:    4000,
				Message: "complex error",
				Cause:   fmt.Errorf("root cause"),
				Value:   "context data",
			},
			expected: "4000: complex error: root cause, Value: context data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.error.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestError_Wrap tests error wrapping functionality
func TestError_Wrap(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	originalErr := &Error{
		Code:       2001,
		Message:    "database error",
		HTTPStatus: 500,
		Severity:   ErrError,
	}

	wrappedErr := originalErr.Wrap(baseErr)

	// Verify all properties are copied
	assert.Equal(t, originalErr.Code, wrappedErr.Code)
	assert.Equal(t, originalErr.Message, wrappedErr.Message)
	assert.Equal(t, originalErr.HTTPStatus, wrappedErr.HTTPStatus)
	assert.Equal(t, originalErr.Severity, wrappedErr.Severity)
	assert.Equal(t, baseErr, wrappedErr.Cause)

	// Verify error message contains cause
	assert.Contains(t, wrappedErr.Error(), "base error")
}

// TestError_Unwrap tests the Unwrap method for errors.Is/As support
func TestError_Unwrap(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	wrappedErr := &Error{
		Code:    2001,
		Message: "wrapped",
		Cause:   baseErr,
	}

	unwrapped := wrappedErr.Unwrap()
	assert.Equal(t, baseErr, unwrapped)

	// Test with standard errors.Is
	assert.True(t, errors.Is(wrappedErr, baseErr))
}

// TestError_WithValue tests adding context values to errors
func TestError_WithValue(t *testing.T) {
	originalErr := &Error{
		Code:       3001,
		Message:    "test error",
		HTTPStatus: 400,
		Severity:   ErrMinor,
	}

	contextValue := map[string]interface{}{
		"user_id": 123,
		"action":  "delete",
	}

	errorWithValue := originalErr.WithValue(contextValue)

	// Verify all properties are copied
	assert.Equal(t, originalErr.Code, errorWithValue.Code)
	assert.Equal(t, originalErr.Message, errorWithValue.Message)
	assert.Equal(t, originalErr.HTTPStatus, errorWithValue.HTTPStatus)
	assert.Equal(t, originalErr.Severity, errorWithValue.Severity)
	assert.Equal(t, contextValue, errorWithValue.Value)

	// Verify error message contains value
	assert.Contains(t, errorWithValue.Error(), "user_id")
}

// TestError_WithValue_PreservesCause tests that WithValue preserves the cause chain
func TestError_WithValue_PreservesCause(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	wrappedErr := &Error{
		Code:    2001,
		Message: "wrapped",
		Cause:   baseErr,
	}

	errorWithValue := wrappedErr.WithValue("context")

	assert.Equal(t, baseErr, errorWithValue.Cause)
	assert.Contains(t, errorWithValue.Error(), "base error")
	assert.Contains(t, errorWithValue.Error(), "context")
}

// TestError_ErrorChain tests the full error chain display
func TestError_ErrorChain(t *testing.T) {
	// Create a chain of custom errors
	err1 := &Error{Code: 1000, Message: "first error"}
	err2 := &Error{Code: 2000, Message: "second error", Cause: err1}
	err3 := &Error{Code: 3000, Message: "third error", Cause: err2}

	chain := err3.ErrorChain()

	// Verify chain contains all errors
	assert.Contains(t, chain, "3000: third error")
	assert.Contains(t, chain, "2000: second error")
	assert.Contains(t, chain, "1000: first error")

	// Verify the chain has newlines
	assert.Contains(t, chain, "\n")
}

// TestError_ErrorChain_WithStandardError tests chain with standard Go error
func TestError_ErrorChain_WithStandardError(t *testing.T) {
	baseErr := fmt.Errorf("standard error")
	customErr := &Error{Code: 2000, Message: "custom error", Cause: baseErr}

	chain := customErr.ErrorChain()

	// Chain should include the custom error but stop at standard error
	assert.Contains(t, chain, "2000: custom error")
	// The standard error is shown as part of the cause in the Error() method
	assert.Contains(t, chain, "standard error")
}

// TestError_Is tests the Is method for error comparison
func TestError_Is(t *testing.T) {
	err1 := &Error{Code: 2001, Message: "database error"}
	err2 := &Error{Code: 2001, Message: "different message"}
	err3 := &Error{Code: 2002, Message: "database error"}
	standardErr := fmt.Errorf("standard error")

	tests := []struct {
		name     string
		err      *Error
		target   error
		expected bool
	}{
		{
			name:     "same error code matches",
			err:      err1,
			target:   err2,
			expected: true,
		},
		{
			name:     "different error code does not match",
			err:      err1,
			target:   err3,
			expected: false,
		},
		{
			name:     "standard error does not match",
			err:      err1,
			target:   standardErr,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Is(tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestError_DisplayStatus tests HTTP status display
func TestError_DisplayStatus(t *testing.T) {
	tests := []struct {
		name       string
		httpStatus int
		expected   string
	}{
		{
			name:       "with HTTP status",
			httpStatus: 404,
			expected:   "404",
		},
		{
			name:       "without HTTP status defaults to 500",
			httpStatus: 0,
			expected:   "500",
		},
		{
			name:       "with 200 status",
			httpStatus: 200,
			expected:   "200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{HTTPStatus: tt.httpStatus}
			result := err.DisplayStatus()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestError_DisplayCode tests error code display
func TestError_DisplayCode(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected string
	}{
		{
			name:     "standard code",
			code:     2001,
			expected: "Error Code #2001",
		},
		{
			name:     "zero code",
			code:     0,
			expected: "Error Code #0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{Code: tt.code}
			result := err.DisplayCode()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestErrorBuilder_Complete tests the full builder pattern
func TestErrorBuilder_Complete(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	contextValue := "test context"

	err := NewErrorBuilder().
		Code(2001).
		Message("test error").
		HTTPStatus(404).
		Severity(ErrError).
		Cause(baseErr).
		Value(contextValue).
		Build()

	assert.Equal(t, 2001, err.Code)
	assert.Equal(t, "test error", err.Message)
	assert.Equal(t, 404, err.HTTPStatus)
	assert.Equal(t, ErrError, err.Severity)
	assert.Equal(t, baseErr, err.Cause)
	assert.Equal(t, contextValue, err.Value)
}

// TestErrorBuilder_Partial tests builder with only some fields
func TestErrorBuilder_Partial(t *testing.T) {
	err := NewErrorBuilder().
		Code(1000).
		Message("minimal error").
		Build()

	assert.Equal(t, 1000, err.Code)
	assert.Equal(t, "minimal error", err.Message)
	assert.Equal(t, 0, err.HTTPStatus)
	assert.Equal(t, ErrSeverity(0), err.Severity)
	assert.Nil(t, err.Cause)
	assert.Nil(t, err.Value)
}

// TestErrorBuilder_Chaining tests builder method chaining
func TestErrorBuilder_Chaining(t *testing.T) {
	builder := NewErrorBuilder()

	// Verify each method returns the builder
	result := builder.Code(1000)
	assert.Equal(t, builder, result)

	result = builder.Message("test")
	assert.Equal(t, builder, result)

	result = builder.HTTPStatus(500)
	assert.Equal(t, builder, result)

	result = builder.Severity(ErrCritical)
	assert.Equal(t, builder, result)

	result = builder.Cause(fmt.Errorf("cause"))
	assert.Equal(t, builder, result)

	result = builder.Value("value")
	assert.Equal(t, builder, result)
}

// TestErrSeverity_Values tests severity level constants
func TestErrSeverity_Values(t *testing.T) {
	// Verify severity levels are in expected order
	assert.Equal(t, ErrSeverity(0), ErrMinor)
	assert.Equal(t, ErrSeverity(1), ErrError)
	assert.Equal(t, ErrSeverity(2), ErrCritical)

	// Verify ordering
	assert.Less(t, int(ErrMinor), int(ErrError))
	assert.Less(t, int(ErrError), int(ErrCritical))
}

// TestError_Integration tests realistic error usage patterns
func TestError_Integration(t *testing.T) {
	t.Run("database operation error chain", func(t *testing.T) {
		// Simulate a database operation that fails
		sqlErr := fmt.Errorf("connection refused")
		dbErr := ErrDatabaseRead.Wrap(sqlErr)
		apiErr := ErrAPIGet.Wrap(dbErr)

		// Verify error chain
		assert.Contains(t, apiErr.Error(), "connection refused")
		assert.True(t, errors.Is(apiErr, sqlErr))
		assert.True(t, errors.Is(apiErr, dbErr))

		// Verify error chain display
		chain := apiErr.ErrorChain()
		assert.Contains(t, chain, "Failed to GET data")
		assert.Contains(t, chain, "Failed to read from database")
		assert.Contains(t, chain, "connection refused")
	})

	t.Run("error with debugging context", func(t *testing.T) {
		userID := "user-123"
		err := ErrNotFound.WithValue(map[string]string{
			"user_id": userID,
			"resource": "profile",
		})

		// Verify context is preserved
		assert.Contains(t, err.Error(), userID)
		assert.Contains(t, err.Error(), "profile")
	})

	t.Run("error unwrapping for standard errors.Is", func(t *testing.T) {
		rootErr := fmt.Errorf("root cause")
		wrapped1 := &Error{Code: 1000, Message: "wrap1", Cause: rootErr}
		wrapped2 := &Error{Code: 2000, Message: "wrap2", Cause: wrapped1}

		// Standard errors.Is should work through the chain
		assert.True(t, errors.Is(wrapped2, rootErr))
		assert.True(t, errors.Is(wrapped2, wrapped1))
	})
}
