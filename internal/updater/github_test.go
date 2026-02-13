package updater

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLatestRelease(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		release := GitHubRelease{
			TagName:     "v1.0.0",
			Name:        "Version 1.0.0",
			Prerelease:  false,
			PublishedAt: time.Now(),
			Assets: []GitHubAsset{
				{Name: "twine-darwin-arm64", BrowserDownloadURL: "https://example.com/asset", Size: 1024},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/repos/cstone-io/twine/releases/latest", r.URL.Path)
			assert.Equal(t, userAgent, r.Header.Get("User-Agent"))
			assert.Contains(t, r.Header.Get("Accept"), "application/vnd.github")

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(release)
		}))
		defer server.Close()

		client := NewGitHubClient()
		client.baseURL = server.URL

		result, err := client.GetLatestRelease()
		require.NoError(t, err)
		assert.Equal(t, "v1.0.0", result.TagName)
		assert.Equal(t, "Version 1.0.0", result.Name)
		assert.False(t, result.Prerelease)
		assert.Len(t, result.Assets, 1)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := NewGitHubClient()
		client.baseURL = server.URL

		_, err := client.GetLatestRelease()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})

	t.Run("rate limit exceeded", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		client := NewGitHubClient()
		client.baseURL = server.URL

		_, err := client.GetLatestRelease()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := NewGitHubClient()
		client.baseURL = server.URL

		_, err := client.GetLatestRelease()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "decode")
	})
}

func TestGetRelease(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		release := GitHubRelease{
			TagName:     "v0.2.0",
			Name:        "Version 0.2.0",
			Prerelease:  false,
			PublishedAt: time.Now(),
			Assets:      []GitHubAsset{},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/repos/cstone-io/twine/releases/tags/v0.2.0", r.URL.Path)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(release)
		}))
		defer server.Close()

		client := NewGitHubClient()
		client.baseURL = server.URL

		result, err := client.GetRelease("v0.2.0")
		require.NoError(t, err)
		assert.Equal(t, "v0.2.0", result.TagName)
		assert.Equal(t, "Version 0.2.0", result.Name)
	})

	t.Run("version not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := NewGitHubClient()
		client.baseURL = server.URL

		_, err := client.GetRelease("v99.99.99")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestListReleases(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		releases := []GitHubRelease{
			{TagName: "v1.0.0", Name: "Version 1.0.0", Prerelease: false},
			{TagName: "v0.3.0", Name: "Version 0.3.0", Prerelease: false},
			{TagName: "v0.2.0", Name: "Version 0.2.0", Prerelease: false},
			{TagName: "v0.1.0-alpha", Name: "Version 0.1.0 Alpha", Prerelease: true},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/repos/cstone-io/twine/releases", r.URL.Path)
			assert.Equal(t, "100", r.URL.Query().Get("per_page"))

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(releases)
		}))
		defer server.Close()

		client := NewGitHubClient()
		client.baseURL = server.URL

		result, err := client.ListReleases()
		require.NoError(t, err)
		assert.Len(t, result, 4)
		assert.Equal(t, "v1.0.0", result[0].TagName)
		assert.Equal(t, "v0.3.0", result[1].TagName)
		assert.True(t, result[3].Prerelease)
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]GitHubRelease{})
		}))
		defer server.Close()

		client := NewGitHubClient()
		client.baseURL = server.URL

		result, err := client.ListReleases()
		require.NoError(t, err)
		assert.Len(t, result, 0)
	})
}

func TestDownloadAsset(t *testing.T) {
	t.Run("successful download", func(t *testing.T) {
		expectedData := []byte("binary data here")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, userAgent, r.Header.Get("User-Agent"))
			assert.Equal(t, "application/octet-stream", r.Header.Get("Accept"))

			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(expectedData)
		}))
		defer server.Close()

		client := NewGitHubClient()
		data, err := client.DownloadAsset(server.URL + "/asset")
		require.NoError(t, err)
		assert.Equal(t, expectedData, data)
	})

	t.Run("download not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := NewGitHubClient()
		_, err := client.DownloadAsset(server.URL + "/asset")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewGitHubClient()
		_, err := client.DownloadAsset(server.URL + "/asset")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})
}

func TestGitHubClientTimeout(t *testing.T) {
	// Create a server that never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.httpClient.Timeout = 100 * time.Millisecond
	client.baseURL = server.URL

	_, err := client.GetLatestRelease()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request failed")
}
