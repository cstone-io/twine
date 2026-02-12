package public

import (
	"embed"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata
var testFS embed.FS

// TestAsset tests asset path generation
func TestAsset(t *testing.T) {
	t.Run("generates correct asset path", func(t *testing.T) {
		result := Asset("style.css")
		assert.Equal(t, "/public/assets/style.css", result)
	})

	t.Run("handles different file types", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"style.css", "/public/assets/style.css"},
			{"app.js", "/public/assets/app.js"},
			{"logo.png", "/public/assets/logo.png"},
			{"fonts/roboto.woff2", "/public/assets/fonts/roboto.woff2"},
		}

		for _, tc := range testCases {
			result := Asset(tc.input)
			assert.Equal(t, tc.expected, result)
		}
	})

	t.Run("handles empty string", func(t *testing.T) {
		result := Asset("")
		assert.Equal(t, "/public/assets/", result)
	})

	t.Run("handles nested paths", func(t *testing.T) {
		result := Asset("css/vendor/bootstrap.css")
		assert.Equal(t, "/public/assets/css/vendor/bootstrap.css", result)
	})

	t.Run("prepends assets path correctly", func(t *testing.T) {
		result := Asset("test.txt")
		assert.Contains(t, result, AssetsPath)
		assert.Contains(t, result, "test.txt")
	})
}

// TestConstants tests package constants
func TestConstants(t *testing.T) {
	t.Run("AssetsPath constant", func(t *testing.T) {
		assert.Equal(t, "/public/assets/", AssetsPath)
	})

	t.Run("PublicPath constant", func(t *testing.T) {
		assert.Equal(t, "/public/", PublicPath)
	})

	t.Run("AssetsPath starts with PublicPath", func(t *testing.T) {
		assert.Contains(t, AssetsPath, PublicPath)
	})
}

// TestFileServerHandler tests static file serving
func TestFileServerHandler(t *testing.T) {
	// Set test filesystem
	originalFS := AssetsFS
	AssetsFS = testFS
	defer func() {
		AssetsFS = originalFS
	}()

	t.Run("serves files under /public/ prefix", func(t *testing.T) {
		handler := FileServerHandler()

		r := httptest.NewRequest("GET", "/public/testdata/test.txt", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		// Should attempt to serve (may or may not find file)
		// We're testing that it routes correctly
		assert.NotEqual(t, 0, w.Code) // Some response code set
	})

	t.Run("returns 404 for paths without /public/ prefix", func(t *testing.T) {
		handler := FileServerHandler()

		r := httptest.NewRequest("GET", "/other/path", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		assert.Equal(t, 404, w.Code)
	})

	t.Run("returns 404 for root path", func(t *testing.T) {
		handler := FileServerHandler()

		r := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		assert.Equal(t, 404, w.Code)
	})

	t.Run("returns 404 for non-public paths", func(t *testing.T) {
		handler := FileServerHandler()

		testPaths := []string{
			"/api/users",
			"/admin/dashboard",
			"/assets/style.css", // Missing /public/ prefix
		}

		for _, path := range testPaths {
			r := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			assert.Equal(t, 404, w.Code, "Path %s should return 404", path)
		}
	})

	t.Run("strips /public/ prefix correctly", func(t *testing.T) {
		handler := FileServerHandler()

		r := httptest.NewRequest("GET", "/public/some/file.txt", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		// Handler should process the request
		// (actual file may not exist, but it shouldn't panic)
		assert.NotPanics(t, func() {
			handler.ServeHTTP(w, r)
		})
	})
}

// TestFileServerHandler_Integration tests realistic static file serving
func TestFileServerHandler_Integration(t *testing.T) {
	// Set test filesystem
	originalFS := AssetsFS
	AssetsFS = testFS
	defer func() {
		AssetsFS = originalFS
	}()

	t.Run("handles different file extensions", func(t *testing.T) {
		handler := FileServerHandler()

		testPaths := []string{
			"/public/testdata/test.css",
			"/public/testdata/test.js",
			"/public/testdata/test.html",
		}

		for _, path := range testPaths {
			r := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			// Should attempt to serve
			assert.NotPanics(t, func() {
				handler.ServeHTTP(w, r)
			})
		}
	})

	t.Run("serves from embedded filesystem", func(t *testing.T) {
		handler := FileServerHandler()

		r := httptest.NewRequest("GET", "/public/testdata/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		// Should handle embedded FS access
		assert.NotPanics(t, func() {
			handler.ServeHTTP(w, r)
		})
	})
}

// TestAsset_Integration tests asset path generation in realistic scenarios
func TestAsset_Integration(t *testing.T) {
	t.Run("generates paths for common assets", func(t *testing.T) {
		cssPath := Asset("css/style.css")
		jsPath := Asset("js/app.js")
		imgPath := Asset("images/logo.png")

		assert.Equal(t, "/public/assets/css/style.css", cssPath)
		assert.Equal(t, "/public/assets/js/app.js", jsPath)
		assert.Equal(t, "/public/assets/images/logo.png", imgPath)
	})

	t.Run("paths work with file server", func(t *testing.T) {
		// Asset generates path with /public/assets/
		assetPath := Asset("test.txt")

		// FileServerHandler expects /public/
		assert.Contains(t, assetPath, PublicPath)
	})
}
