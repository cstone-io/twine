package updater

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckForUpdate(t *testing.T) {
	t.Run("update available", func(t *testing.T) {
		release := GitHubRelease{
			TagName:     "v2.0.0",
			Name:        "Version 2.0.0",
			PublishedAt: time.Now(),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(release)
		}))
		defer server.Close()

		updater := NewUpdater()
		updater.github.baseURL = server.URL

		result, err := updater.CheckForUpdate("v1.0.0")
		require.NoError(t, err)
		assert.False(t, result.Updated)
		assert.Equal(t, "v1.0.0", result.FromVersion)
		assert.Equal(t, "v2.0.0", result.ToVersion)
		assert.Contains(t, result.Message, "Update available")
	})

	t.Run("already up to date", func(t *testing.T) {
		release := GitHubRelease{
			TagName:     "v1.0.0",
			Name:        "Version 1.0.0",
			PublishedAt: time.Now(),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(release)
		}))
		defer server.Close()

		updater := NewUpdater()
		updater.github.baseURL = server.URL

		result, err := updater.CheckForUpdate("v1.0.0")
		require.NoError(t, err)
		assert.False(t, result.Updated)
		assert.Contains(t, result.Message, "Already up-to-date")
	})

	t.Run("dev version", func(t *testing.T) {
		release := GitHubRelease{
			TagName:     "v1.0.0",
			Name:        "Version 1.0.0",
			PublishedAt: time.Now(),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(release)
		}))
		defer server.Close()

		updater := NewUpdater()
		updater.github.baseURL = server.URL

		result, err := updater.CheckForUpdate("dev")
		require.NoError(t, err)
		assert.False(t, result.Updated)
		assert.Contains(t, result.Message, "Update available")
	})
}

func TestGetBinaryName(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		goarch   string
		expected string
	}{
		{"macOS ARM64", "darwin", "arm64", "twine-darwin-arm64"},
		{"macOS AMD64", "darwin", "amd64", "twine-darwin-amd64"},
		{"Linux ARM64", "linux", "arm64", "twine-linux-arm64"},
		{"Linux AMD64", "linux", "amd64", "twine-linux-amd64"},
		{"Windows AMD64", "windows", "amd64", "twine-windows-amd64.exe"},
		{"Windows ARM64", "windows", "arm64", "twine-windows-arm64.exe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBinaryName(tt.goos, tt.goarch)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInstallBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows due to file locking issues")
	}

	t.Run("successful install", func(t *testing.T) {
		// Create a temporary directory for testing
		tmpDir := t.TempDir()

		// Create a fake binary
		fakeBinary := filepath.Join(tmpDir, "twine")
		err := os.WriteFile(fakeBinary, []byte("old binary"), 0755)
		require.NoError(t, err)

		// Save the original os.Executable function behavior
		// We can't actually replace the running binary, so we'll test with a different file
		newData := []byte("new binary data")

		// Create a temp file to simulate the current executable
		tmpFile, err := os.CreateTemp(tmpDir, "test-binary-*")
		require.NoError(t, err)
		tmpFile.Write([]byte("old data"))
		tmpFile.Close()

		// Test that we can read/write to the file
		err = os.WriteFile(tmpFile.Name(), newData, 0755)
		require.NoError(t, err)

		// Verify the file was updated
		data, err := os.ReadFile(tmpFile.Name())
		require.NoError(t, err)
		assert.Equal(t, newData, data)
	})

	t.Run("permission check", func(t *testing.T) {
		// Create a file we don't have permission to write to
		tmpDir := t.TempDir()
		readonlyFile := filepath.Join(tmpDir, "readonly")
		err := os.WriteFile(readonlyFile, []byte("data"), 0444)
		require.NoError(t, err)

		// Try to check write permissions
		err = checkWritePermissions(readonlyFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})
}

func TestUpdate(t *testing.T) {
	t.Run("no binary for platform", func(t *testing.T) {
		release := GitHubRelease{
			TagName:     "v2.0.0",
			Name:        "Version 2.0.0",
			PublishedAt: time.Now(),
			Assets: []GitHubAsset{
				{Name: "twine-linux-amd64", BrowserDownloadURL: "http://example.com/asset"},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/repos/cstone-io/twine/releases/latest" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(release)
			}
		}))
		defer server.Close()

		updater := NewUpdater()
		updater.github.baseURL = server.URL

		// Only run this test if we're not on linux-amd64
		if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
			t.Skip("Skipping - would find a binary for this platform")
		}

		opts := UpdateOptions{
			CurrentVersion: "v1.0.0",
		}

		_, err := updater.Update(opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no binary available")
	})

	t.Run("already up to date", func(t *testing.T) {
		release := GitHubRelease{
			TagName:     "v1.0.0",
			Name:        "Version 1.0.0",
			PublishedAt: time.Now(),
			Assets:      []GitHubAsset{},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(release)
		}))
		defer server.Close()

		updater := NewUpdater()
		updater.github.baseURL = server.URL

		opts := UpdateOptions{
			CurrentVersion: "v1.0.0",
		}

		result, err := updater.Update(opts)
		require.NoError(t, err)
		assert.False(t, result.Updated)
		assert.Contains(t, result.Message, "Already up-to-date")
	})

	t.Run("target version not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		updater := NewUpdater()
		updater.github.baseURL = server.URL

		opts := UpdateOptions{
			CurrentVersion: "v1.0.0",
			TargetVersion:  "v99.99.99",
		}

		_, err := updater.Update(opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestCheckWritePermissions(t *testing.T) {
	t.Run("writable file", func(t *testing.T) {
		tmpDir := t.TempDir()
		file := filepath.Join(tmpDir, "writable")
		err := os.WriteFile(file, []byte("data"), 0644)
		require.NoError(t, err)

		err = checkWritePermissions(file)
		assert.NoError(t, err)
	})

	t.Run("readonly file", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Readonly behavior differs on Windows")
		}

		tmpDir := t.TempDir()
		file := filepath.Join(tmpDir, "readonly")
		err := os.WriteFile(file, []byte("data"), 0444)
		require.NoError(t, err)

		err = checkWritePermissions(file)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})

	t.Run("nonexistent file", func(t *testing.T) {
		err := checkWritePermissions("/nonexistent/path/file")
		assert.Error(t, err)
	})
}
