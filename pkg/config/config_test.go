package config

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetConfig resets the singleton for testing
func resetConfig() {
	once = sync.Once{}
	instance = nil
}

// setTestEnv sets environment variables for testing and returns a cleanup function
func setTestEnv(t *testing.T, envVars map[string]string) func() {
	t.Helper()

	// Save original environment
	originalEnv := make(map[string]string)
	for key := range envVars {
		originalEnv[key] = os.Getenv(key)
	}

	// Set test environment variables
	for key, value := range envVars {
		os.Setenv(key, value)
	}

	// Return cleanup function
	return func() {
		for key, originalValue := range originalEnv {
			if originalValue == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, originalValue)
			}
		}
	}
}

// TestConfig_Get_Singleton tests that Get() returns the same instance
func TestConfig_Get_Singleton(t *testing.T) {
	resetConfig()
	defer resetConfig()

	cleanup := setTestEnv(t, map[string]string{
		"DB_HOST": "localhost",
		"DB_PORT": "5432",
	})
	defer cleanup()

	config1 := Get()
	config2 := Get()

	assert.Equal(t, config1, config2, "Get() should return the same instance")
	assert.Same(t, config1, config2, "Get() should return the exact same pointer")
}

// TestConfig_DatabaseConfig_FromEnv tests database configuration from environment variables
func TestConfig_DatabaseConfig_FromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected DatabaseConfig
	}{
		{
			name: "all database environment variables set",
			envVars: map[string]string{
				"DB_HOST":     "testhost",
				"DB_PORT":     "5433",
				"DB_USERNAME": "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
				"DB_SSLMODE":  "require",
				"DB_TIMEZONE": "America/New_York",
			},
			expected: DatabaseConfig{
				Host:     "testhost",
				Port:     5433,
				Username: "testuser",
				Password: "testpass",
				Name:     "testdb",
				SSLMode:  "require",
				TimeZone: "America/New_York",
			},
		},
		{
			name: "database config with defaults",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_USERNAME": "user",
				"DB_PASSWORD": "pass",
				"DB_NAME":     "db",
			},
			expected: DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				Username: "user",
				Password: "pass",
				Name:     "db",
				SSLMode:  "disable", // default
				TimeZone: "UTC",     // default
			},
		},
		{
			name: "database config with empty port",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_PORT":     "",
				"DB_USERNAME": "user",
				"DB_PASSWORD": "pass",
				"DB_NAME":     "db",
			},
			expected: DatabaseConfig{
				Host:     "localhost",
				Port:     0, // mustAtoi returns 0 for empty string
				Username: "user",
				Password: "pass",
				Name:     "db",
				SSLMode:  "disable",
				TimeZone: "UTC",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetConfig()
			defer resetConfig()

			cleanup := setTestEnv(t, tt.envVars)
			defer cleanup()

			cfg := Get()

			assert.Equal(t, tt.expected.Host, cfg.Database.Host)
			assert.Equal(t, tt.expected.Port, cfg.Database.Port)
			assert.Equal(t, tt.expected.Username, cfg.Database.Username)
			assert.Equal(t, tt.expected.Password, cfg.Database.Password)
			assert.Equal(t, tt.expected.Name, cfg.Database.Name)
			assert.Equal(t, tt.expected.SSLMode, cfg.Database.SSLMode)
			assert.Equal(t, tt.expected.TimeZone, cfg.Database.TimeZone)
		})
	}
}

// TestDatabaseConfig_DSN tests DSN string generation
func TestDatabaseConfig_DSN(t *testing.T) {
	tests := []struct {
		name           string
		config         DatabaseConfig
		expectedFields map[string]string
	}{
		{
			name: "complete database config",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				Username: "testuser",
				Password: "testpass",
				Name:     "testdb",
				SSLMode:  "require",
				TimeZone: "UTC",
			},
			expectedFields: map[string]string{
				"host":     "localhost",
				"port":     "5432",
				"user":     "testuser",
				"password": "testpass",
				"dbname":   "testdb",
				"sslmode":  "require",
				"TimeZone": "UTC",
			},
		},
		{
			name: "database config with special characters",
			config: DatabaseConfig{
				Host:     "db.example.com",
				Port:     5433,
				Username: "user@domain",
				Password: "p@ss!w0rd",
				Name:     "my-database",
				SSLMode:  "disable",
				TimeZone: "America/New_York",
			},
			expectedFields: map[string]string{
				"host":     "db.example.com",
				"port":     "5433",
				"user":     "user@domain",
				"password": "p@ss!w0rd",
				"dbname":   "my-database",
				"sslmode":  "disable",
				"TimeZone": "America/New_York",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.config.DSN()

			for key, value := range tt.expectedFields {
				expected := key + "=" + value
				assert.Contains(t, dsn, expected,
					"DSN should contain %s", expected)
			}
		})
	}
}

// TestConfig_LoggerConfig_FromEnv tests logger configuration from environment variables
func TestConfig_LoggerConfig_FromEnv(t *testing.T) {
	tests := []struct {
		name          string
		envVars       map[string]string
		expectedLevel LogLevel
	}{
		{
			name:          "trace level",
			envVars:       map[string]string{"LOGGER_LEVEL": "trace"},
			expectedLevel: LogTrace,
		},
		{
			name:          "debug level",
			envVars:       map[string]string{"LOGGER_LEVEL": "debug"},
			expectedLevel: LogDebug,
		},
		{
			name:          "info level",
			envVars:       map[string]string{"LOGGER_LEVEL": "info"},
			expectedLevel: LogInfo,
		},
		{
			name:          "warn level",
			envVars:       map[string]string{"LOGGER_LEVEL": "warn"},
			expectedLevel: LogWarn,
		},
		{
			name:          "error level",
			envVars:       map[string]string{"LOGGER_LEVEL": "error"},
			expectedLevel: LogError,
		},
		{
			name:          "critical level",
			envVars:       map[string]string{"LOGGER_LEVEL": "critical"},
			expectedLevel: LogCritical,
		},
		{
			name:          "default level for unknown value",
			envVars:       map[string]string{"LOGGER_LEVEL": "unknown"},
			expectedLevel: LogInfo,
		},
		{
			name:          "default level for empty value",
			envVars:       map[string]string{"LOGGER_LEVEL": ""},
			expectedLevel: LogInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetConfig()
			defer resetConfig()

			cleanup := setTestEnv(t, tt.envVars)
			defer cleanup()

			cfg := Get()

			assert.Equal(t, tt.expectedLevel, cfg.Logger.Level)
		})
	}
}

// TestConfig_LoggerConfig_Output tests logger output configuration
func TestConfig_LoggerConfig_Output(t *testing.T) {
	t.Run("stdout output", func(t *testing.T) {
		resetConfig()
		defer resetConfig()

		cleanup := setTestEnv(t, map[string]string{
			"LOGGER_OUTPUT": "stdout",
		})
		defer cleanup()

		cfg := Get()

		assert.Equal(t, os.Stdout, cfg.Logger.Output)
	})

	t.Run("stderr output", func(t *testing.T) {
		resetConfig()
		defer resetConfig()

		cleanup := setTestEnv(t, map[string]string{
			"LOGGER_OUTPUT": "stderr",
		})
		defer cleanup()

		cfg := Get()

		assert.Equal(t, os.Stderr, cfg.Logger.Output)
	})

	t.Run("default to stdout", func(t *testing.T) {
		resetConfig()
		defer resetConfig()

		cleanup := setTestEnv(t, map[string]string{})
		defer cleanup()

		cfg := Get()

		assert.Equal(t, os.Stdout, cfg.Logger.Output)
	})
}

// TestConfig_LoggerConfig_ErrorOutput tests logger error output configuration
func TestConfig_LoggerConfig_ErrorOutput(t *testing.T) {
	t.Run("stderr error output", func(t *testing.T) {
		resetConfig()
		defer resetConfig()

		cleanup := setTestEnv(t, map[string]string{
			"LOGGER_ERROR_OUTPUT": "stderr",
		})
		defer cleanup()

		cfg := Get()

		assert.Equal(t, os.Stderr, cfg.Logger.ErrorOutput)
	})

	t.Run("default to stderr", func(t *testing.T) {
		resetConfig()
		defer resetConfig()

		cleanup := setTestEnv(t, map[string]string{})
		defer cleanup()

		cfg := Get()

		assert.Equal(t, os.Stderr, cfg.Logger.ErrorOutput)
	})
}

// TestConfig_AuthConfig_FromEnv tests auth configuration from environment variables
func TestConfig_AuthConfig_FromEnv(t *testing.T) {
	tests := []struct {
		name          string
		envVars       map[string]string
		expectedValue string
	}{
		{
			name:          "auth secret set",
			envVars:       map[string]string{"AUTH_SECRET": "my-secret-key"},
			expectedValue: "my-secret-key",
		},
		{
			name:          "auth secret empty",
			envVars:       map[string]string{"AUTH_SECRET": ""},
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetConfig()
			defer resetConfig()

			cleanup := setTestEnv(t, tt.envVars)
			defer cleanup()

			cfg := Get()

			assert.Equal(t, tt.expectedValue, cfg.Auth.SecretKey)
		})
	}
}

// TestConfig_EnvFile tests loading from .env file
func TestConfig_EnvFile(t *testing.T) {
	// Create a temporary .env file
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	envContent := `DB_HOST=envhost
DB_PORT=5434
DB_USERNAME=envuser
DB_PASSWORD=envpass
DB_NAME=envdb
AUTH_SECRET=envsecret
LOGGER_LEVEL=debug
`
	err := os.WriteFile(envFile, []byte(envContent), 0644)
	require.NoError(t, err)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	resetConfig()
	defer resetConfig()

	cfg := Get()

	assert.Equal(t, "envhost", cfg.Database.Host)
	assert.Equal(t, 5434, cfg.Database.Port)
	assert.Equal(t, "envuser", cfg.Database.Username)
	assert.Equal(t, "envpass", cfg.Database.Password)
	assert.Equal(t, "envdb", cfg.Database.Name)
	assert.Equal(t, "envsecret", cfg.Auth.SecretKey)
	assert.Equal(t, LogDebug, cfg.Logger.Level)
}

// TestLogLevel_Values tests log level constants
func TestLogLevel_Values(t *testing.T) {
	assert.Equal(t, LogLevel(0), LogTrace)
	assert.Equal(t, LogLevel(1), LogDebug)
	assert.Equal(t, LogLevel(2), LogInfo)
	assert.Equal(t, LogLevel(3), LogWarn)
	assert.Equal(t, LogLevel(4), LogError)
	assert.Equal(t, LogLevel(5), LogCritical)

	// Verify ordering
	assert.Less(t, int(LogTrace), int(LogDebug))
	assert.Less(t, int(LogDebug), int(LogInfo))
	assert.Less(t, int(LogInfo), int(LogWarn))
	assert.Less(t, int(LogWarn), int(LogError))
	assert.Less(t, int(LogError), int(LogCritical))
}

// TestParseLogLevel tests the parseLogLevel function
func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"trace", LogTrace},
		{"debug", LogDebug},
		{"info", LogInfo},
		{"warn", LogWarn},
		{"error", LogError},
		{"critical", LogCritical},
		{"TRACE", LogInfo},    // case-sensitive, defaults to info
		{"unknown", LogInfo},  // unknown value, defaults to info
		{"", LogInfo},         // empty string, defaults to info
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParseOutput tests the parseOutput function
func TestParseOutput(t *testing.T) {
	t.Run("stdout", func(t *testing.T) {
		output := parseOutput("stdout")
		assert.Equal(t, os.Stdout, output)
	})

	t.Run("stderr", func(t *testing.T) {
		output := parseOutput("stderr")
		assert.Equal(t, os.Stderr, output)
	})

	t.Run("file path", func(t *testing.T) {
		tempDir := t.TempDir()
		logFile := filepath.Join(tempDir, "test.log")

		output := parseOutput(logFile)
		assert.NotNil(t, output)

		// Verify we can write to it
		file, ok := output.(*os.File)
		require.True(t, ok, "Output should be a file")
		defer file.Close()

		_, err := file.WriteString("test\n")
		assert.NoError(t, err)

		// Verify content was written
		content, err := os.ReadFile(logFile)
		require.NoError(t, err)
		assert.Equal(t, "test\n", string(content))
	})

	t.Run("file append mode", func(t *testing.T) {
		tempDir := t.TempDir()
		logFile := filepath.Join(tempDir, "append.log")

		// Write initial content
		err := os.WriteFile(logFile, []byte("initial\n"), 0644)
		require.NoError(t, err)

		// Open with parseOutput (should append)
		output := parseOutput(logFile)
		file, ok := output.(*os.File)
		require.True(t, ok)
		defer file.Close()

		_, err = file.WriteString("appended\n")
		require.NoError(t, err)

		// Verify both lines are present
		content, err := os.ReadFile(logFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "initial\n")
		assert.Contains(t, string(content), "appended\n")
	})
}

// TestGetEnvOrDefault tests the getEnvOrDefault helper function
func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "environment variable set",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "environment variable not set",
			key:          "UNSET_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "environment variable empty string",
			key:          "EMPTY_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvOrDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMustAtoi tests the mustAtoi helper function
func TestMustAtoi(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "valid integer",
			input:    "42",
			expected: 42,
		},
		{
			name:     "zero",
			input:    "0",
			expected: 0,
		},
		{
			name:     "negative integer",
			input:    "-10",
			expected: -10,
		},
		{
			name:     "empty string returns zero",
			input:    "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mustAtoi(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestConfig_Integration tests a complete configuration scenario
func TestConfig_Integration(t *testing.T) {
	resetConfig()
	defer resetConfig()

	envVars := map[string]string{
		"DB_HOST":              "prod-db.example.com",
		"DB_PORT":              "5432",
		"DB_USERNAME":          "app_user",
		"DB_PASSWORD":          "secure_password",
		"DB_NAME":              "app_database",
		"DB_SSLMODE":           "require",
		"DB_TIMEZONE":          "UTC",
		"LOGGER_LEVEL":         "warn",
		"LOGGER_OUTPUT":        "stdout",
		"LOGGER_ERROR_OUTPUT":  "stderr",
		"AUTH_SECRET":          "super-secret-key-12345",
	}

	cleanup := setTestEnv(t, envVars)
	defer cleanup()

	cfg := Get()

	// Verify database config
	assert.Equal(t, "prod-db.example.com", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "app_user", cfg.Database.Username)
	assert.Equal(t, "secure_password", cfg.Database.Password)
	assert.Equal(t, "app_database", cfg.Database.Name)
	assert.Equal(t, "require", cfg.Database.SSLMode)
	assert.Equal(t, "UTC", cfg.Database.TimeZone)

	// Verify DSN
	dsn := cfg.Database.DSN()
	assert.Contains(t, dsn, "host=prod-db.example.com")
	assert.Contains(t, dsn, "port=5432")
	assert.Contains(t, dsn, "user=app_user")
	assert.Contains(t, dsn, "password=secure_password")
	assert.Contains(t, dsn, "dbname=app_database")
	assert.Contains(t, dsn, "sslmode=require")
	assert.Contains(t, dsn, "TimeZone=UTC")

	// Verify logger config
	assert.Equal(t, LogWarn, cfg.Logger.Level)
	assert.Equal(t, os.Stdout, cfg.Logger.Output)
	assert.Equal(t, os.Stderr, cfg.Logger.ErrorOutput)

	// Verify auth config
	assert.Equal(t, "super-secret-key-12345", cfg.Auth.SecretKey)

	// Verify we can write to logger output
	var buf bytes.Buffer
	cfg.Logger.Output = &buf
	cfg.Logger.Output.Write([]byte("test"))
	assert.Equal(t, "test", buf.String())
}

// TestConfig_ThreadSafety tests concurrent access to Get()
func TestConfig_ThreadSafety(t *testing.T) {
	resetConfig()
	defer resetConfig()

	cleanup := setTestEnv(t, map[string]string{
		"DB_HOST": "localhost",
	})
	defer cleanup()

	const goroutines = 100
	configs := make([]*Config, goroutines)
	var wg sync.WaitGroup

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(index int) {
			defer wg.Done()
			configs[index] = Get()
		}(i)
	}

	wg.Wait()

	// All configs should be the same instance
	firstConfig := configs[0]
	for i := 1; i < goroutines; i++ {
		assert.Same(t, firstConfig, configs[i],
			"All goroutines should get the same config instance")
	}
}

// TestConfig_FileOutput_Cleanup tests that file outputs are properly managed
func TestConfig_FileOutput_Cleanup(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "output.log")

	resetConfig()
	defer resetConfig()

	cleanup := setTestEnv(t, map[string]string{
		"LOGGER_OUTPUT": logFile,
	})
	defer cleanup()

	cfg := Get()

	// Verify output is a file
	file, ok := cfg.Logger.Output.(*os.File)
	require.True(t, ok)

	// Write to it
	_, err := file.WriteString("test output\n")
	require.NoError(t, err)

	// Verify file exists and contains content
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "test output")

	// Clean up
	file.Close()
}

// TestConfig_MultipleWriters tests using io.MultiWriter for logger output
func TestConfig_MultipleWriters(t *testing.T) {
	resetConfig()
	defer resetConfig()

	cleanup := setTestEnv(t, map[string]string{
		"LOGGER_OUTPUT": "stdout",
	})
	defer cleanup()

	cfg := Get()

	var buf1, buf2 bytes.Buffer
	multiWriter := io.MultiWriter(&buf1, &buf2)
	cfg.Logger.Output = multiWriter

	message := "test message"
	cfg.Logger.Output.Write([]byte(message))

	assert.Equal(t, message, buf1.String())
	assert.Equal(t, message, buf2.String())
}
