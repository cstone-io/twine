package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewDevCommand tests dev command creation
func TestNewDevCommand(t *testing.T) {
	cmd := NewDevCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "dev", cmd.Use)
	assert.Equal(t, "Start development server with hot reload", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}

// TestIsWatchedFile tests file extension filtering
func TestIsWatchedFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Go file",
			path:     "/app/pages/users/page.go",
			expected: true,
		},
		{
			name:     "Go test file",
			path:     "/app/pages/users/page_test.go",
			expected: true,
		},
		{
			name:     "layout.go",
			path:     "/app/pages/layout.go",
			expected: true,
		},
		{
			name:     "route.go",
			path:     "/app/api/users/route.go",
			expected: true,
		},
		{
			name:     "HTML file",
			path:     "/templates/index.html",
			expected: false,
		},
		{
			name:     "CSS file",
			path:     "/public/style.css",
			expected: false,
		},
		{
			name:     "JavaScript file",
			path:     "/public/app.js",
			expected: false,
		},
		{
			name:     "JSON file",
			path:     "/config.json",
			expected: false,
		},
		{
			name:     "No extension",
			path:     "/app/README",
			expected: false,
		},
		{
			name:     "Directory",
			path:     "/app/pages/",
			expected: false,
		},
		{
			name:     "Uppercase .GO",
			path:     "/app/FILE.GO",
			expected: false, // Case sensitive
		},
		{
			name:     "routes.gen.go should be excluded",
			path:     "app/routes.gen.go",
			expected: false,
		},
		{
			name:     "routes.gen.go with different path should be excluded",
			path:     "/full/path/to/app/routes.gen.go",
			expected: false,
		},
		{
			name:     "other gen.go files should NOT be excluded",
			path:     "app/custom.gen.go",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isWatchedFile(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGenerateRoutes tests route generation function
func TestGenerateRoutes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod
	goModContent := `module github.com/test/project

go 1.22
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644))

	// Create app directory with a simple route
	appDir := filepath.Join(tmpDir, "app")
	pagesDir := filepath.Join(appDir, "pages", "index")
	require.NoError(t, os.MkdirAll(pagesDir, 0755))

	pageContent := `package index

import "github.com/cstone-io/twine/kit"

func GET(k *kit.Kit) error {
	return nil
}
`
	require.NoError(t, os.WriteFile(filepath.Join(pagesDir, "page.go"), []byte(pageContent), 0644))

	// Generate routes
	err := generateRoutes(tmpDir, appDir)
	assert.NoError(t, err)

	// Verify routes.gen.go was created
	routesFile := filepath.Join(appDir, "routes.gen.go")
	assert.FileExists(t, routesFile)

	// Verify content
	content, err := os.ReadFile(routesFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "package app")
	assert.Contains(t, string(content), "RegisterRoutes")
}

// TestGenerateRoutes_NoGoMod tests error when go.mod is missing
func TestGenerateRoutes_NoGoMod(t *testing.T) {
	tmpDir := t.TempDir()

	appDir := filepath.Join(tmpDir, "app")
	require.NoError(t, os.MkdirAll(appDir, 0755))

	err := generateRoutes(tmpDir, appDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "getting module path")
}

// TestGenerateRoutes_InvalidRoute tests validation error
func TestGenerateRoutes_InvalidRoute(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod
	goModContent := `module github.com/test/project

go 1.22
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644))

	// Create invalid route (handler without methods)
	appDir := filepath.Join(tmpDir, "app")
	pagesDir := filepath.Join(appDir, "pages", "test")
	require.NoError(t, os.MkdirAll(pagesDir, 0755))

	pageContent := `package test

func helper() {}
`
	require.NoError(t, os.WriteFile(filepath.Join(pagesDir, "page.go"), []byte(pageContent), 0644))

	err := generateRoutes(tmpDir, appDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation error")
}

// TestGenerateRoutes_EmptyApp tests empty app directory
func TestGenerateRoutes_EmptyApp(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod
	goModContent := `module github.com/test/project

go 1.22
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644))

	// Create empty app directory
	appDir := filepath.Join(tmpDir, "app")
	require.NoError(t, os.MkdirAll(appDir, 0755))

	err := generateRoutes(tmpDir, appDir)
	assert.NoError(t, err) // Should succeed with empty routes

	// Verify routes.gen.go exists
	routesFile := filepath.Join(appDir, "routes.gen.go")
	assert.FileExists(t, routesFile)
}

// TestAddDirectoryRecursive tests directory watcher setup
func TestAddDirectoryRecursive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir1", "dir2")
	dir3 := filepath.Join(tmpDir, "dir1", "dir2", "dir3")
	require.NoError(t, os.MkdirAll(dir3, 0755))

	// Create some files
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "file1.go"), []byte("package test"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "file2.go"), []byte("package test"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir3, "file3.go"), []byte("package test"), 0644))

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	defer watcher.Close()

	// Add directories recursively
	err = addDirectoryRecursive(watcher, tmpDir)
	assert.NoError(t, err)

	// Verify all directories were added (check watcher.WatchList)
	watchList := watcher.WatchList()
	assert.GreaterOrEqual(t, len(watchList), 4) // tmpDir + dir1 + dir2 + dir3
}

// TestAddDirectoryRecursive_SingleDirectory tests single directory
func TestAddDirectoryRecursive_SingleDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create single file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("package test"), 0644))

	watcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	defer watcher.Close()

	err = addDirectoryRecursive(watcher, tmpDir)
	assert.NoError(t, err)

	watchList := watcher.WatchList()
	assert.GreaterOrEqual(t, len(watchList), 1)
}

// TestAddDirectoryRecursive_NonexistentDirectory tests error handling
func TestAddDirectoryRecursive_NonexistentDirectory(t *testing.T) {
	watcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	defer watcher.Close()

	err = addDirectoryRecursive(watcher, "/nonexistent/directory")
	assert.Error(t, err)
}

// TestGenerateRoutes_DynamicRoutes tests dynamic route generation
func TestGenerateRoutes_DynamicRoutes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod
	goModContent := `module github.com/test/project

go 1.22
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644))

	// Create dynamic route
	appDir := filepath.Join(tmpDir, "app")
	userIDDir := filepath.Join(appDir, "pages", "users", "[id]")
	require.NoError(t, os.MkdirAll(userIDDir, 0755))

	pageContent := `package user_id

import "github.com/cstone-io/twine/kit"

func GET(k *kit.Kit) error { return nil }
func PUT(k *kit.Kit) error { return nil }
func DELETE(k *kit.Kit) error { return nil }
`
	require.NoError(t, os.WriteFile(filepath.Join(userIDDir, "page.go"), []byte(pageContent), 0644))

	err := generateRoutes(tmpDir, appDir)
	assert.NoError(t, err)

	// Verify generated code includes dynamic route
	routesFile := filepath.Join(appDir, "routes.gen.go")
	content, err := os.ReadFile(routesFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "{id}")
}

// TestGenerateRoutes_WithLayouts tests layout middleware generation
func TestGenerateRoutes_WithLayouts(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod
	goModContent := `module github.com/test/project

go 1.22
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644))

	// Create layout
	appDir := filepath.Join(tmpDir, "app")
	pagesDir := filepath.Join(appDir, "pages")
	require.NoError(t, os.MkdirAll(pagesDir, 0755))

	layoutContent := `package pages

import "github.com/cstone-io/twine/middleware"

func Layout() middleware.Middleware {
	return nil
}
`
	require.NoError(t, os.WriteFile(filepath.Join(pagesDir, "layout.go"), []byte(layoutContent), 0644))

	// Create page
	dashboardDir := filepath.Join(pagesDir, "dashboard")
	require.NoError(t, os.MkdirAll(dashboardDir, 0755))

	pageContent := `package dashboard

import "github.com/cstone-io/twine/kit"

func GET(k *kit.Kit) error { return nil }
`
	require.NoError(t, os.WriteFile(filepath.Join(dashboardDir, "page.go"), []byte(pageContent), 0644))

	err := generateRoutes(tmpDir, appDir)
	assert.NoError(t, err)

	// Verify generated code includes layout middleware
	routesFile := filepath.Join(appDir, "routes.gen.go")
	content, err := os.ReadFile(routesFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "applyMiddleware")
	assert.Contains(t, string(content), ".Layout()")
}

// TestGenerateRoutes_APIRoutes tests API route generation
func TestGenerateRoutes_APIRoutes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod
	goModContent := `module github.com/test/project

go 1.22
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644))

	// Create API route
	appDir := filepath.Join(tmpDir, "app")
	apiDir := filepath.Join(appDir, "api", "users")
	require.NoError(t, os.MkdirAll(apiDir, 0755))

	routeContent := `package users

import "github.com/cstone-io/twine/kit"

func GET(k *kit.Kit) error { return nil }
func POST(k *kit.Kit) error { return nil }
`
	require.NoError(t, os.WriteFile(filepath.Join(apiDir, "route.go"), []byte(routeContent), 0644))

	err := generateRoutes(tmpDir, appDir)
	assert.NoError(t, err)

	// Verify generated code includes API route
	routesFile := filepath.Join(appDir, "routes.gen.go")
	content, err := os.ReadFile(routesFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "/api/users")
	assert.Contains(t, string(content), "// API routes")
}
