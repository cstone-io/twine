package errors

import (
	"fmt"
	"strconv"
)

// ErrSeverity represents the severity level of an error
type ErrSeverity int

const (
	ErrMinor ErrSeverity = iota
	ErrError
	ErrCritical
)

// Error represents a structured error with code, message, and context
type Error struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	HTTPStatus int         `json:"-"`
	Severity   ErrSeverity `json:"-"`
	Cause      error       `json:"-"`
	Value      any         `json:"-"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		if e.Value != nil {
			return fmt.Sprintf("%d: %s: %v, Value: %v", e.Code, e.Message, e.Cause, e.Value)
		}
		return fmt.Sprintf("%d: %s: %v", e.Code, e.Message, e.Cause)
	}
	if e.Value != nil {
		return fmt.Sprintf("%d: %s, Value: %v", e.Code, e.Message, e.Value)
	}
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// Wrap wraps another error with this error's context
func (e *Error) Wrap(cause error) *Error {
	return NewErrorBuilder().
		Code(e.Code).
		Message(e.Message).
		HTTPStatus(e.HTTPStatus).
		Severity(e.Severity).
		Cause(cause).
		Build()
}

// Unwrap returns the wrapped error for errors.Is/As support
func (e *Error) Unwrap() error {
	return e.Cause
}

// WithValue adds a value to the error for debugging
func (e *Error) WithValue(value any) *Error {
	return NewErrorBuilder().
		Code(e.Code).
		Message(e.Message).
		HTTPStatus(e.HTTPStatus).
		Severity(e.Severity).
		Cause(e.Cause).
		Value(value).
		Build()
}

// ErrorChain returns the full chain of wrapped errors
func (e *Error) ErrorChain() string {
	var chain string
	for err := e; err != nil; {
		chain += err.Error() + "\n"
		if cause, ok := err.Cause.(*Error); ok {
			err = cause
		} else {
			break
		}
	}
	return chain
}

// Is implements errors.Is interface
func (e *Error) Is(target error) bool {
	if err, ok := target.(*Error); ok {
		if e.Code == err.Code {
			return true
		}
	}
	return false
}

// DisplayStatus returns the HTTP status as a string
func (e *Error) DisplayStatus() string {
	if e.HTTPStatus == 0 {
		return "500"
	}
	return strconv.Itoa(e.HTTPStatus)
}

// DisplayCode returns a formatted error code string
func (e *Error) DisplayCode() string {
	return fmt.Sprintf("Error Code #%d", e.Code)
}

// ErrorBuilder provides a fluent interface for building errors
type ErrorBuilder struct {
	code       int
	message    string
	httpStatus int
	severity   ErrSeverity
	cause      error
	value      any
}

// NewErrorBuilder creates a new ErrorBuilder instance
func NewErrorBuilder() *ErrorBuilder {
	return &ErrorBuilder{}
}

// Code sets the error code
func (b *ErrorBuilder) Code(code int) *ErrorBuilder {
	b.code = code
	return b
}

// Message sets the error message
func (b *ErrorBuilder) Message(message string) *ErrorBuilder {
	b.message = message
	return b
}

// HTTPStatus sets the HTTP status code
func (b *ErrorBuilder) HTTPStatus(status int) *ErrorBuilder {
	b.httpStatus = status
	return b
}

// Severity sets the error severity
func (b *ErrorBuilder) Severity(severity ErrSeverity) *ErrorBuilder {
	b.severity = severity
	return b
}

// Cause sets the wrapped error
func (b *ErrorBuilder) Cause(cause error) *ErrorBuilder {
	b.cause = cause
	return b
}

// Value sets a debug value
func (b *ErrorBuilder) Value(value any) *ErrorBuilder {
	b.value = value
	return b
}

// Build constructs the final Error
func (b *ErrorBuilder) Build() *Error {
	return &Error{
		Code:       b.code,
		Message:    b.message,
		HTTPStatus: b.httpStatus,
		Severity:   b.severity,
		Cause:      b.cause,
		Value:      b.value,
	}
}
