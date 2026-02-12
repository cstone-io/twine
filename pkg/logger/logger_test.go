package logger

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/cstone-io/twine/pkg/config"
	"github.com/cstone-io/twine/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// resetLogger resets the singleton for testing
func resetLogger() {
	once = sync.Once{}
	instance = nil
}

// createTestLogger creates a logger with custom output for testing
func createTestLogger(output *bytes.Buffer, level config.LogLevel) *Logger {
	cfg := config.LoggerConfig{
		Level:       level,
		Output:      output,
		ErrorOutput: output,
	}
	initialize(cfg)
	return instance
}

// TestLogger_Get_Singleton tests that Get() returns the same instance
func TestLogger_Get_Singleton(t *testing.T) {
	// Note: We cannot easily reset the logger singleton without affecting config
	// So this test just verifies Get() returns a non-nil logger
	logger1 := Get()
	logger2 := Get()

	assert.NotNil(t, logger1)
	assert.NotNil(t, logger2)
	assert.Same(t, logger1, logger2, "Get() should return the same instance")
}

// TestLogger_Trace tests trace-level logging
func TestLogger_Trace(t *testing.T) {
	t.Run("trace logged when level is trace", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogTrace)

		logger.Trace("test trace message")

		output := buf.String()
		assert.Contains(t, output, "TRACE:")
		assert.Contains(t, output, "test trace message")
	})

	t.Run("trace not logged when level is debug", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogDebug)

		logger.Trace("test trace message")

		output := buf.String()
		assert.Empty(t, output)
	})

	t.Run("trace not logged when level is info", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogInfo)

		logger.Trace("test trace message")

		output := buf.String()
		assert.Empty(t, output)
	})
}

// TestLogger_Debug tests debug-level logging
func TestLogger_Debug(t *testing.T) {
	t.Run("debug logged when level is trace", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogTrace)

		logger.Debug("test debug message")

		output := buf.String()
		assert.Contains(t, output, "DEBUG:")
		assert.Contains(t, output, "test debug message")
	})

	t.Run("debug logged when level is debug", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogDebug)

		logger.Debug("test debug message")

		output := buf.String()
		assert.Contains(t, output, "DEBUG:")
		assert.Contains(t, output, "test debug message")
	})

	t.Run("debug not logged when level is info", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogInfo)

		logger.Debug("test debug message")

		output := buf.String()
		assert.Empty(t, output)
	})

	t.Run("debug not logged when level is warn", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogWarn)

		logger.Debug("test debug message")

		output := buf.String()
		assert.Empty(t, output)
	})
}

// TestLogger_Info tests info-level logging
func TestLogger_Info(t *testing.T) {
	t.Run("info logged when level is trace", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogTrace)

		logger.Info("test info message")

		output := buf.String()
		assert.Contains(t, output, "INFO:")
		assert.Contains(t, output, "test info message")
	})

	t.Run("info logged when level is info", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogInfo)

		logger.Info("test info message")

		output := buf.String()
		assert.Contains(t, output, "INFO:")
		assert.Contains(t, output, "test info message")
	})

	t.Run("info not logged when level is warn", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogWarn)

		logger.Info("test info message")

		output := buf.String()
		assert.Empty(t, output)
	})
}

// TestLogger_Warn tests warn-level logging
func TestLogger_Warn(t *testing.T) {
	t.Run("warn logged when level is info", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogInfo)

		logger.Warn("test warn message")

		output := buf.String()
		assert.Contains(t, output, "WARN:")
		assert.Contains(t, output, "test warn message")
	})

	t.Run("warn logged when level is warn", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogWarn)

		logger.Warn("test warn message")

		output := buf.String()
		assert.Contains(t, output, "WARN:")
		assert.Contains(t, output, "test warn message")
	})

	t.Run("warn not logged when level is error", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogError)

		logger.Warn("test warn message")

		output := buf.String()
		assert.Empty(t, output)
	})
}

// TestLogger_Error tests error-level logging
func TestLogger_Error(t *testing.T) {
	t.Run("error logged when level is warn", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogWarn)

		logger.Error("test error message")

		output := buf.String()
		assert.Contains(t, output, "ERROR:")
		assert.Contains(t, output, "test error message")
	})

	t.Run("error logged when level is error", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogError)

		logger.Error("test error message")

		output := buf.String()
		assert.Contains(t, output, "ERROR:")
		assert.Contains(t, output, "test error message")
	})

	t.Run("error not logged when level is critical", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogCritical)

		logger.Error("test error message")

		output := buf.String()
		assert.Empty(t, output)
	})
}

// TestLogger_Critical tests critical-level logging
func TestLogger_Critical(t *testing.T) {
	t.Run("critical always logged regardless of level", func(t *testing.T) {
		levels := []config.LogLevel{
			config.LogTrace,
			config.LogDebug,
			config.LogInfo,
			config.LogWarn,
			config.LogError,
			config.LogCritical,
		}

		for _, level := range levels {
			resetLogger()
			var buf bytes.Buffer
			logger := createTestLogger(&buf, level)

			logger.Critical("test critical message")

			output := buf.String()
			assert.Contains(t, output, "CRITICAL:", "Critical should be logged at level %v", level)
			assert.Contains(t, output, "test critical message")
		}
	})
}

// TestLogger_FormatString tests format string handling
func TestLogger_FormatString(t *testing.T) {
	t.Run("format with arguments", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogInfo)

		logger.Info("user %s logged in with ID %d", "john", 123)

		output := buf.String()
		assert.Contains(t, output, "user john logged in with ID 123")
	})

	t.Run("format with multiple types", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogInfo)

		logger.Info("string=%s int=%d float=%f bool=%t", "test", 42, 3.14, true)

		output := buf.String()
		assert.Contains(t, output, "string=test")
		assert.Contains(t, output, "int=42")
		assert.Contains(t, output, "float=3.14")
		assert.Contains(t, output, "bool=true")
	})

	t.Run("no arguments", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogInfo)

		logger.Info("simple message")

		output := buf.String()
		assert.Contains(t, output, "simple message")
	})
}

// TestLogger_CustomError tests CustomError method
func TestLogger_CustomError(t *testing.T) {
	t.Run("minor error routed to warn", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogTrace)

		minorErr := &errors.Error{
			Code:     3001,
			Message:  "minor error",
			Severity: errors.ErrMinor,
		}

		logger.CustomError(minorErr)

		output := buf.String()
		assert.Contains(t, output, "WARN:")
		assert.Contains(t, output, "3001:")
		assert.Contains(t, output, "minor error")
	})

	t.Run("error routed to error log", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogTrace)

		normalErr := &errors.Error{
			Code:     2001,
			Message:  "normal error",
			Severity: errors.ErrError,
		}

		logger.CustomError(normalErr)

		output := buf.String()
		assert.Contains(t, output, "ERROR:")
		assert.Contains(t, output, "2001:")
		assert.Contains(t, output, "normal error")
	})

	t.Run("critical error routed to critical log", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogTrace)

		criticalErr := &errors.Error{
			Code:     1001,
			Message:  "critical error",
			Severity: errors.ErrCritical,
		}

		logger.CustomError(criticalErr)

		output := buf.String()
		assert.Contains(t, output, "CRITICAL:")
		assert.Contains(t, output, "1001:")
		assert.Contains(t, output, "critical error")
	})

	t.Run("error chain is logged", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogTrace)

		baseErr := &errors.Error{Code: 1000, Message: "base error"}
		wrappedErr := &errors.Error{
			Code:     2000,
			Message:  "wrapped error",
			Severity: errors.ErrError,
			Cause:    baseErr,
		}

		logger.CustomError(wrappedErr)

		output := buf.String()
		assert.Contains(t, output, "wrapped error")
		assert.Contains(t, output, "base error")
	})

	t.Run("respects log level for minor errors", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogError) // Warn won't be logged

		minorErr := &errors.Error{
			Code:     3001,
			Message:  "minor error",
			Severity: errors.ErrMinor,
		}

		logger.CustomError(minorErr)

		output := buf.String()
		assert.Empty(t, output, "Minor errors routed to Warn should be filtered by log level")
	})

	t.Run("respects log level for normal errors", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogCritical) // Error won't be logged

		normalErr := &errors.Error{
			Code:     2001,
			Message:  "normal error",
			Severity: errors.ErrError,
		}

		logger.CustomError(normalErr)

		output := buf.String()
		assert.Empty(t, output, "Normal errors should be filtered by log level")
	})

	t.Run("critical errors always logged", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogCritical)

		criticalErr := &errors.Error{
			Code:     1001,
			Message:  "critical error",
			Severity: errors.ErrCritical,
		}

		logger.CustomError(criticalErr)

		output := buf.String()
		assert.Contains(t, output, "CRITICAL:")
		assert.Contains(t, output, "critical error")
	})
}

// TestLogger_OutputRouting tests that regular logs go to Output and errors to ErrorOutput
func TestLogger_OutputRouting(t *testing.T) {
	t.Run("info logs to output", func(t *testing.T) {
		resetLogger()
		var stdout, stderr bytes.Buffer
		cfg := config.LoggerConfig{
			Level:       config.LogInfo,
			Output:      &stdout,
			ErrorOutput: &stderr,
		}
		initialize(cfg)

		instance.Info("test message")

		assert.Contains(t, stdout.String(), "test message")
		assert.Empty(t, stderr.String())
	})

	t.Run("error logs to error output", func(t *testing.T) {
		resetLogger()
		var stdout, stderr bytes.Buffer
		cfg := config.LoggerConfig{
			Level:       config.LogError,
			Output:      &stdout,
			ErrorOutput: &stderr,
		}
		initialize(cfg)

		instance.Error("error message")

		assert.Empty(t, stdout.String())
		assert.Contains(t, stderr.String(), "error message")
	})

	t.Run("critical logs to error output", func(t *testing.T) {
		resetLogger()
		var stdout, stderr bytes.Buffer
		cfg := config.LoggerConfig{
			Level:       config.LogCritical,
			Output:      &stdout,
			ErrorOutput: &stderr,
		}
		initialize(cfg)

		instance.Critical("critical message")

		assert.Empty(t, stdout.String())
		assert.Contains(t, stderr.String(), "critical message")
	})
}

// TestLogger_LogFormat tests the log message format
func TestLogger_LogFormat(t *testing.T) {
	resetLogger()
	var buf bytes.Buffer
	logger := createTestLogger(&buf, config.LogInfo)

	logger.Info("test message")

	output := buf.String()

	// Check for date/time (format: 2006/01/02)
	assert.Regexp(t, `\d{4}/\d{2}/\d{2}`, output, "Should contain date")

	// Check for time (format: 15:04:05)
	assert.Regexp(t, `\d{2}:\d{2}:\d{2}`, output, "Should contain time")

	// Check for file:line (the actual source file is logger.go since that's where the log call originates)
	assert.Contains(t, output, "logger.go:", "Should contain file name")

	// Check for level prefix
	assert.Contains(t, output, "INFO:", "Should contain log level")

	// Check for message
	assert.Contains(t, output, "test message", "Should contain message")
}

// TestLogger_ThreadSafety tests concurrent logging
func TestLogger_ThreadSafety(t *testing.T) {
	resetLogger()
	var buf bytes.Buffer
	logger := createTestLogger(&buf, config.LogInfo)

	const goroutines = 100
	var wg sync.WaitGroup

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			logger.Info("message from goroutine %d", id)
		}(i)
	}

	wg.Wait()

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// All goroutines should have logged
	assert.GreaterOrEqual(t, len(lines), goroutines,
		"Should have at least %d log lines", goroutines)
}

// TestLogger_Integration tests realistic logging scenarios
func TestLogger_Integration(t *testing.T) {
	t.Run("web request logging", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogInfo)

		// Simulate request handling
		logger.Info("incoming request: method=%s path=%s", "GET", "/api/users")
		logger.Debug("query params: limit=10 offset=0") // Won't appear (level is Info)
		logger.Info("request completed: status=%d duration=%dms", 200, 45)

		output := buf.String()
		assert.Contains(t, output, "incoming request")
		assert.Contains(t, output, "GET")
		assert.Contains(t, output, "/api/users")
		assert.NotContains(t, output, "query params") // Debug filtered out
		assert.Contains(t, output, "request completed")
		assert.Contains(t, output, "200")
	})

	t.Run("error handling flow", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogTrace)

		// Simulate error handling
		dbErr := errors.ErrDatabaseRead.Wrap(assert.AnError)
		logger.CustomError(dbErr)

		output := buf.String()
		assert.Contains(t, output, "ERROR:")
		assert.Contains(t, output, "Failed to read from database")
	})

	t.Run("multi-level logging", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		logger := createTestLogger(&buf, config.LogDebug)

		logger.Trace("trace: entering function") // Filtered out
		logger.Debug("debug: processing user data")
		logger.Info("info: user authenticated")
		logger.Warn("warn: rate limit approaching")
		logger.Error("error: validation failed")
		logger.Critical("critical: system shutdown")

		output := buf.String()
		assert.NotContains(t, output, "trace:")
		assert.Contains(t, output, "debug:")
		assert.Contains(t, output, "info:")
		assert.Contains(t, output, "warn:")
		assert.Contains(t, output, "error:")
		assert.Contains(t, output, "critical:")
	})
}

// TestLogger_Initialize tests the initialize function
func TestLogger_Initialize(t *testing.T) {
	t.Run("creates all loggers", func(t *testing.T) {
		resetLogger()
		var buf bytes.Buffer
		cfg := config.LoggerConfig{
			Level:       config.LogTrace,
			Output:      &buf,
			ErrorOutput: &buf,
		}

		initialize(cfg)

		assert.NotNil(t, instance)
		assert.NotNil(t, instance.traceLogger)
		assert.NotNil(t, instance.debugLogger)
		assert.NotNil(t, instance.infoLogger)
		assert.NotNil(t, instance.warnLogger)
		assert.NotNil(t, instance.errorLogger)
		assert.NotNil(t, instance.criticalLogger)
		assert.Equal(t, config.LogTrace, instance.level)
	})
}

// TestLogger_RealOutputs tests with actual stdout/stderr
func TestLogger_RealOutputs(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping stdout/stderr test in CI")
	}

	t.Run("can log to stdout", func(t *testing.T) {
		resetLogger()
		cfg := config.LoggerConfig{
			Level:       config.LogInfo,
			Output:      os.Stdout,
			ErrorOutput: os.Stderr,
		}
		initialize(cfg)

		// This should not panic
		instance.Info("test stdout logging")
	})

	t.Run("can log to stderr", func(t *testing.T) {
		resetLogger()
		cfg := config.LoggerConfig{
			Level:       config.LogError,
			Output:      os.Stdout,
			ErrorOutput: os.Stderr,
		}
		initialize(cfg)

		// This should not panic
		instance.Error("test stderr logging")
	})
}
