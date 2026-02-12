package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TempDir creates a temporary directory for tests that is automatically
// cleaned up when the test completes.
func TempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "twine-test-*")
	require.NoError(t, err, "failed to create temp directory")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// LoadFixture loads a test fixture file from the internal/testutil/fixtures directory.
// Returns the raw bytes of the fixture file.
func LoadFixture(t *testing.T, path string) []byte {
	t.Helper()

	// Try relative to test file first
	data, err := os.ReadFile(filepath.Join("internal", "testutil", "fixtures", path))
	if err != nil {
		// Try from project root
		data, err = os.ReadFile(filepath.Join("fixtures", path))
		if err != nil {
			require.NoError(t, err, "failed to load fixture: %s", path)
		}
	}
	return data
}

// LoadJSONFixture loads a JSON fixture file and unmarshals it into the
// provided value.
func LoadJSONFixture(t *testing.T, path string, v interface{}) {
	t.Helper()
	data := LoadFixture(t, path)
	err := json.Unmarshal(data, v)
	require.NoError(t, err, "failed to unmarshal JSON fixture: %s", path)
}

// WriteFixture writes data to a fixture file in a temporary directory.
// Useful for creating test fixtures on the fly.
func WriteFixture(t *testing.T, dir, filename string, data []byte) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, data, 0644)
	require.NoError(t, err, "failed to write fixture: %s", path)
	return path
}

// SetupTestEnv sets test environment variables and automatically unsets them
// when the test completes.
func SetupTestEnv(t *testing.T, env map[string]string) {
	t.Helper()
	original := make(map[string]string)

	for k, v := range env {
		original[k] = os.Getenv(k)
		os.Setenv(k, v)
	}

	t.Cleanup(func() {
		for k, v := range original {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	})
}

// MkdirAll creates a directory structure in the given base path.
// Useful for creating complex test directory structures.
func MkdirAll(t *testing.T, basePath string, dirs ...string) {
	t.Helper()
	for _, dir := range dirs {
		path := filepath.Join(basePath, dir)
		err := os.MkdirAll(path, 0755)
		require.NoError(t, err, "failed to create directory: %s", path)
	}
}

// CreateFile creates a file with the given content in the specified path.
// Parent directories are created automatically.
func CreateFile(t *testing.T, path string, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		require.NoError(t, err, "failed to create parent directory: %s", dir)
	}
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err, "failed to create file: %s", path)
}

// FileExists checks if a file exists at the given path.
func FileExists(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists at the given path.
func DirExists(t *testing.T, path string) bool {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
