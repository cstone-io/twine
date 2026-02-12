package routing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test fixture helpers

func setupFixture(t *testing.T, structure map[string]string) string {
	t.Helper()
	tmpDir := t.TempDir()

	for path, content := range structure {
		fullPath := filepath.Join(tmpDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}

	return tmpDir
}

func createTestPageHandler(packageName string, methods ...string) string {
	content := "package " + packageName + "\n\n"
	content += "import \"github.com/cstone-io/twine/pkg/kit\"\n\n"

	for _, method := range methods {
		content += "func " + method + "(k *kit.Kit) error {\n"
		content += "\treturn nil\n"
		content += "}\n\n"
	}

	return content
}

func createTestLayout(packageName string) string {
	content := "package " + packageName + "\n\n"
	content += "import \"github.com/cstone-io/twine/pkg/middleware\"\n\n"
	content += "func Layout() middleware.Middleware {\n"
	content += "\treturn func(next kit.HandlerFunc) kit.HandlerFunc {\n"
	content += "\t\treturn next\n"
	content += "\t}\n"
	content += "}\n"
	return content
}

// TestScanRoutes_EmptyDirectory tests scanning empty app directory
func TestScanRoutes_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	appDir := filepath.Join(tmpDir, "app")
	require.NoError(t, os.MkdirAll(appDir, 0755))

	root, err := ScanRoutes(appDir)

	require.NoError(t, err)
	assert.NotNil(t, root)
	assert.Equal(t, appDir, root.Path)
	assert.Empty(t, root.Children)
}

// TestScanRoutes_MissingDirectory tests scanning non-existent directory
func TestScanRoutes_MissingDirectory(t *testing.T) {
	root, err := ScanRoutes("/nonexistent/directory")

	require.NoError(t, err) // Should not error, just return empty tree
	assert.NotNil(t, root)
	assert.Empty(t, root.Children)
}

// TestScanRoutes_SimplePage tests scanning a simple page.go
func TestScanRoutes_SimplePage(t *testing.T) {
	fixture := map[string]string{
		"app/pages/index/page.go": createTestPageHandler("index", "GET"),
	}

	rootDir := setupFixture(t, fixture)
	appDir := filepath.Join(rootDir, "app")

	root, err := ScanRoutes(appDir)

	require.NoError(t, err)
	require.Len(t, root.Children, 1)

	pages := root.Children[0]
	assert.Equal(t, "pages", pages.URLSegment)
	require.Len(t, pages.Children, 1)

	index := pages.Children[0]
	assert.Equal(t, "index", index.URLSegment)
	assert.True(t, index.IsPage)
	assert.False(t, index.IsAPI)
	assert.Equal(t, filepath.Join(rootDir, "app/pages/index/page.go"), index.HandlerFile)
	assert.ElementsMatch(t, []string{"GET"}, index.Methods)
	assert.Equal(t, "index", index.PackageName)
}

// TestScanRoutes_MultiplePages tests scanning multiple pages
func TestScanRoutes_MultiplePages(t *testing.T) {
	fixture := map[string]string{
		"app/pages/users/page.go": createTestPageHandler("users", "GET", "POST"),
		"app/pages/posts/page.go": createTestPageHandler("posts", "GET"),
		"app/pages/about/page.go": createTestPageHandler("about", "GET"),
	}

	rootDir := setupFixture(t, fixture)
	appDir := filepath.Join(rootDir, "app")

	root, err := ScanRoutes(appDir)

	require.NoError(t, err)
	require.Len(t, root.Children, 1)

	pages := root.Children[0]
	assert.Len(t, pages.Children, 3)

	// Find specific pages
	var users, posts, about *RouteNode
	for _, child := range pages.Children {
		switch child.URLSegment {
		case "users":
			users = child
		case "posts":
			posts = child
		case "about":
			about = child
		}
	}

	require.NotNil(t, users)
	require.NotNil(t, posts)
	require.NotNil(t, about)

	assert.ElementsMatch(t, []string{"GET", "POST"}, users.Methods)
	assert.ElementsMatch(t, []string{"GET"}, posts.Methods)
	assert.ElementsMatch(t, []string{"GET"}, about.Methods)
}

// TestScanRoutes_APIRoute tests scanning API route.go
func TestScanRoutes_APIRoute(t *testing.T) {
	fixture := map[string]string{
		"app/api/users/route.go": createTestPageHandler("users", "GET", "POST", "PUT", "DELETE"),
	}

	rootDir := setupFixture(t, fixture)
	appDir := filepath.Join(rootDir, "app")

	root, err := ScanRoutes(appDir)

	require.NoError(t, err)
	require.Len(t, root.Children, 1)

	api := root.Children[0]
	assert.Equal(t, "api", api.URLSegment)
	require.Len(t, api.Children, 1)

	users := api.Children[0]
	assert.Equal(t, "users", users.URLSegment)
	assert.True(t, users.IsAPI)
	assert.False(t, users.IsPage)
	assert.Equal(t, filepath.Join(rootDir, "app/api/users/route.go"), users.HandlerFile)
	assert.ElementsMatch(t, []string{"GET", "POST", "PUT", "DELETE"}, users.Methods)
}

// TestScanRoutes_BothPagesAndAPI tests scanning both pages and API
func TestScanRoutes_BothPagesAndAPI(t *testing.T) {
	fixture := map[string]string{
		"app/pages/users/page.go": createTestPageHandler("users", "GET"),
		"app/api/users/route.go":  createTestPageHandler("users", "GET", "POST"),
	}

	rootDir := setupFixture(t, fixture)
	appDir := filepath.Join(rootDir, "app")

	root, err := ScanRoutes(appDir)

	require.NoError(t, err)
	assert.Len(t, root.Children, 2)

	// Check both pages and api exist
	var pages, api *RouteNode
	for _, child := range root.Children {
		if child.URLSegment == "pages" {
			pages = child
		} else if child.URLSegment == "api" {
			api = child
		}
	}

	require.NotNil(t, pages)
	require.NotNil(t, api)

	assert.Len(t, pages.Children, 1)
	assert.Len(t, api.Children, 1)
}

// TestScanRoutes_DynamicRoute tests scanning [id] dynamic routes
func TestScanRoutes_DynamicRoute(t *testing.T) {
	fixture := map[string]string{
		"app/pages/users/[id]/page.go": createTestPageHandler("user_id", "GET", "PUT", "DELETE"),
	}

	rootDir := setupFixture(t, fixture)
	appDir := filepath.Join(rootDir, "app")

	root, err := ScanRoutes(appDir)

	require.NoError(t, err)
	require.Len(t, root.Children, 1)

	pages := root.Children[0]
	require.Len(t, pages.Children, 1)

	users := pages.Children[0]
	assert.Equal(t, "users", users.URLSegment)
	require.Len(t, users.Children, 1)

	userID := users.Children[0]
	assert.Equal(t, "{id}", userID.URLSegment)
	assert.True(t, userID.IsDynamic)
	assert.False(t, userID.IsCatchAll)
	assert.Equal(t, "id", userID.ParamName)
	assert.ElementsMatch(t, []string{"GET", "PUT", "DELETE"}, userID.Methods)
}

// TestScanRoutes_CatchAllRoute tests scanning [...slug] catch-all routes
func TestScanRoutes_CatchAllRoute(t *testing.T) {
	fixture := map[string]string{
		"app/pages/docs/[...slug]/page.go": createTestPageHandler("slug_catchall", "GET"),
	}

	rootDir := setupFixture(t, fixture)
	appDir := filepath.Join(rootDir, "app")

	root, err := ScanRoutes(appDir)

	require.NoError(t, err)
	require.Len(t, root.Children, 1)

	pages := root.Children[0]
	require.Len(t, pages.Children, 1)

	docs := pages.Children[0]
	require.Len(t, docs.Children, 1)

	slug := docs.Children[0]
	assert.Equal(t, "{slug...}", slug.URLSegment)
	assert.True(t, slug.IsCatchAll)
	assert.True(t, slug.IsDynamic)
	assert.Equal(t, "slug", slug.ParamName)
}

// TestScanRoutes_NestedRoutes tests deeply nested route structure
func TestScanRoutes_NestedRoutes(t *testing.T) {
	fixture := map[string]string{
		"app/pages/users/page.go":              createTestPageHandler("users", "GET"),
		"app/pages/users/[id]/page.go":         createTestPageHandler("user_id", "GET"),
		"app/pages/users/[id]/edit/page.go":    createTestPageHandler("edit", "GET", "POST"),
		"app/pages/users/[id]/delete/page.go":  createTestPageHandler("delete", "POST"),
		"app/pages/users/[id]/profile/page.go": createTestPageHandler("profile", "GET"),
	}

	rootDir := setupFixture(t, fixture)
	appDir := filepath.Join(rootDir, "app")

	root, err := ScanRoutes(appDir)

	require.NoError(t, err)

	pages := root.Children[0]
	users := pages.Children[0]
	assert.True(t, users.IsPage) // /users has page.go

	require.Len(t, users.Children, 1)
	userID := users.Children[0]
	assert.True(t, userID.IsDynamic)
	assert.True(t, userID.IsPage) // /users/[id] has page.go

	// Should have 3 children: edit, delete, profile
	assert.Len(t, userID.Children, 3)
}

// TestScanRoutes_WithLayout tests scanning layout.go files
func TestScanRoutes_WithLayout(t *testing.T) {
	fixture := map[string]string{
		"app/pages/dashboard/layout.go":       createTestLayout("dashboard"),
		"app/pages/dashboard/index/page.go":   createTestPageHandler("index", "GET"),
		"app/pages/dashboard/reports/page.go": createTestPageHandler("reports", "GET"),
	}

	rootDir := setupFixture(t, fixture)
	appDir := filepath.Join(rootDir, "app")

	root, err := ScanRoutes(appDir)

	require.NoError(t, err)

	pages := root.Children[0]
	dashboard := pages.Children[0]

	assert.True(t, dashboard.HasLayout)
	assert.Equal(t, filepath.Join(rootDir, "app/pages/dashboard/layout.go"), dashboard.LayoutFile)
	assert.Equal(t, "dashboard", dashboard.PackageName)
	assert.Len(t, dashboard.Children, 2)
}

// TestScanRoutes_LayoutWithoutHandler tests layout without handler in same directory
func TestScanRoutes_LayoutWithoutHandler(t *testing.T) {
	fixture := map[string]string{
		"app/pages/admin/layout.go":       createTestLayout("admin"),
		"app/pages/admin/users/page.go":   createTestPageHandler("users", "GET"),
		"app/pages/admin/settings/page.go": createTestPageHandler("settings", "GET"),
	}

	rootDir := setupFixture(t, fixture)
	appDir := filepath.Join(rootDir, "app")

	root, err := ScanRoutes(appDir)

	require.NoError(t, err)

	pages := root.Children[0]
	admin := pages.Children[0]

	assert.True(t, admin.HasLayout)
	assert.Empty(t, admin.HandlerFile) // No page.go in admin directory
	assert.True(t, admin.IsDirectory)
	assert.Len(t, admin.Children, 2)
}

// TestScanRoutes_NestedLayouts tests multiple layout.go files in hierarchy
func TestScanRoutes_NestedLayouts(t *testing.T) {
	fixture := map[string]string{
		"app/pages/layout.go":                createTestLayout("pages"),
		"app/pages/dashboard/layout.go":      createTestLayout("dashboard"),
		"app/pages/dashboard/admin/layout.go": createTestLayout("admin"),
		"app/pages/dashboard/admin/users/page.go": createTestPageHandler("users", "GET"),
	}

	rootDir := setupFixture(t, fixture)
	appDir := filepath.Join(rootDir, "app")

	root, err := ScanRoutes(appDir)

	require.NoError(t, err)

	pages := root.Children[0]
	assert.True(t, pages.HasLayout)

	dashboard := pages.Children[0]
	assert.True(t, dashboard.HasLayout)

	admin := dashboard.Children[0]
	assert.True(t, admin.HasLayout)

	users := admin.Children[0]
	assert.True(t, users.IsPage)

	// Verify parent chain for layout inheritance
	assert.Equal(t, admin, users.Parent)
	assert.Equal(t, dashboard, admin.Parent)
	assert.Equal(t, pages, dashboard.Parent)
}

// TestDetectMethods_AllHTTPMethods tests detecting all HTTP methods
func TestDetectMethods_AllHTTPMethods(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "GET only",
			content: `package test

import "github.com/cstone-io/twine/pkg/kit"

func GET(k *kit.Kit) error {
	return nil
}
`,
			expected: []string{"GET"},
		},
		{
			name: "POST only",
			content: `package test

import "github.com/cstone-io/twine/pkg/kit"

func POST(k *kit.Kit) error {
	return nil
}
`,
			expected: []string{"POST"},
		},
		{
			name: "GET and POST",
			content: `package test

import "github.com/cstone-io/twine/pkg/kit"

func GET(k *kit.Kit) error {
	return nil
}

func POST(k *kit.Kit) error {
	return nil
}
`,
			expected: []string{"GET", "POST"},
		},
		{
			name: "all methods",
			content: `package test

import "github.com/cstone-io/twine/pkg/kit"

func GET(k *kit.Kit) error { return nil }
func POST(k *kit.Kit) error { return nil }
func PUT(k *kit.Kit) error { return nil }
func DELETE(k *kit.Kit) error { return nil }
func PATCH(k *kit.Kit) error { return nil }
`,
			expected: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		},
		{
			name: "ignores unexported functions",
			content: `package test

import "github.com/cstone-io/twine/pkg/kit"

func GET(k *kit.Kit) error { return nil }
func internal() { }
func helper(x int) int { return x }
`,
			expected: []string{"GET"},
		},
		{
			name: "ignores invalid method names",
			content: `package test

import "github.com/cstone-io/twine/pkg/kit"

func GET(k *kit.Kit) error { return nil }
func OPTIONS(k *kit.Kit) error { return nil }  // Not in valid methods
func HEAD(k *kit.Kit) error { return nil }     // Not in valid methods
func Custom(k *kit.Kit) error { return nil }   // Not a method
`,
			expected: []string{"GET"},
		},
		{
			name:     "no exported methods",
			content: `package test

func helper() {}
`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.go")
			require.NoError(t, os.WriteFile(testFile, []byte(tt.content), 0644))

			methods, err := DetectMethods(testFile)

			require.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, methods)
		})
	}
}

// TestDetectMethods_InvalidSyntax tests handling invalid Go syntax
func TestDetectMethods_InvalidSyntax(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.go")

	invalidContent := `package test

func GET(k *kit.Kit) error {
	// Missing closing brace
`

	require.NoError(t, os.WriteFile(testFile, []byte(invalidContent), 0644))

	_, err := DetectMethods(testFile)
	assert.Error(t, err)
}

// TestDetectMethods_NonexistentFile tests handling missing file
func TestDetectMethods_NonexistentFile(t *testing.T) {
	_, err := DetectMethods("/nonexistent/file.go")
	assert.Error(t, err)
}

// TestGetPackageName_ValidFiles tests package name extraction
func TestGetPackageName_ValidFiles(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expected    string
	}{
		{
			name:     "simple package",
			content:  "package main\n",
			expected: "main",
		},
		{
			name:     "package with imports",
			content: `package users

import "fmt"
`,
			expected: "users",
		},
		{
			name:     "package with underscore",
			content:  "package user_id\n",
			expected: "user_id",
		},
		{
			name:     "package with comment",
			content: `// Package admin provides admin functionality
package admin
`,
			expected: "admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.go")
			require.NoError(t, os.WriteFile(testFile, []byte(tt.content), 0644))

			pkg, err := getPackageName(testFile)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, pkg)
		})
	}
}

// TestGetPackageName_InvalidFile tests error handling
func TestGetPackageName_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.go")

	invalidContent := "not valid go code"
	require.NoError(t, os.WriteFile(testFile, []byte(invalidContent), 0644))

	_, err := getPackageName(testFile)
	assert.Error(t, err)
}

// TestScanRoutes_ComplexHierarchy tests realistic complex structure
func TestScanRoutes_ComplexHierarchy(t *testing.T) {
	fixture := map[string]string{
		// Root pages
		"app/pages/index/page.go": createTestPageHandler("index", "GET"),
		"app/pages/about/page.go": createTestPageHandler("about", "GET"),

		// User pages with dynamic routes
		"app/pages/users/page.go":         createTestPageHandler("users", "GET", "POST"),
		"app/pages/users/[id]/page.go":    createTestPageHandler("user_id", "GET", "PUT", "DELETE"),
		"app/pages/users/[id]/edit/page.go": createTestPageHandler("edit", "GET", "POST"),

		// Dashboard with layout
		"app/pages/dashboard/layout.go":       createTestLayout("dashboard"),
		"app/pages/dashboard/index/page.go":   createTestPageHandler("index", "GET"),
		"app/pages/dashboard/reports/page.go": createTestPageHandler("reports", "GET"),

		// API routes
		"app/api/users/route.go":          createTestPageHandler("users", "GET", "POST"),
		"app/api/users/[id]/route.go":     createTestPageHandler("user_id", "GET", "PUT", "DELETE"),
		"app/api/posts/route.go":          createTestPageHandler("posts", "GET", "POST"),
		"app/api/posts/[id]/route.go":     createTestPageHandler("post_id", "GET"),
		"app/api/auth/login/route.go":     createTestPageHandler("login", "POST"),
		"app/api/auth/logout/route.go":    createTestPageHandler("logout", "POST"),

		// Catch-all route
		"app/pages/docs/[...slug]/page.go": createTestPageHandler("slug_catchall", "GET"),
	}

	rootDir := setupFixture(t, fixture)
	appDir := filepath.Join(rootDir, "app")

	root, err := ScanRoutes(appDir)

	require.NoError(t, err)
	assert.Len(t, root.Children, 2) // pages and api

	// Verify pages structure
	var pages, api *RouteNode
	for _, child := range root.Children {
		if child.URLSegment == "pages" {
			pages = child
		} else if child.URLSegment == "api" {
			api = child
		}
	}

	require.NotNil(t, pages)
	require.NotNil(t, api)

	// Verify pages has: index, about, users, dashboard, docs
	assert.Len(t, pages.Children, 5)

	// Verify API has: users, posts, auth
	assert.Len(t, api.Children, 3)
}

// TestScanRoutes_EmptySubdirectories tests directories without handlers
func TestScanRoutes_EmptySubdirectories(t *testing.T) {
	fixture := map[string]string{
		"app/pages/users/active/page.go": createTestPageHandler("active", "GET"),
	}

	rootDir := setupFixture(t, fixture)
	appDir := filepath.Join(rootDir, "app")

	// Create empty intermediate directory
	emptyDir := filepath.Join(rootDir, "app/pages/empty")
	require.NoError(t, os.MkdirAll(emptyDir, 0755))

	root, err := ScanRoutes(appDir)

	require.NoError(t, err)

	pages := root.Children[0]
	// Should only have 'users', not 'empty' (empty dirs are filtered)
	assert.Len(t, pages.Children, 1)
	assert.Equal(t, "users", pages.Children[0].URLSegment)
}

// TestScanRoutes_BothPageAndRoute tests invalid mixed handler types
func TestScanRoutes_BothPageAndRoute(t *testing.T) {
	// This is technically invalid but scanner should handle it
	fixture := map[string]string{
		"app/pages/users/page.go":  createTestPageHandler("users", "GET"),
		"app/pages/users/route.go": createTestPageHandler("users", "POST"),
	}

	rootDir := setupFixture(t, fixture)
	appDir := filepath.Join(rootDir, "app")

	root, err := ScanRoutes(appDir)

	// Should not error - both files will be detected
	// (validation happens later in the validator)
	require.NoError(t, err)

	pages := root.Children[0]
	users := pages.Children[0]

	// Both IsPage and IsAPI will be true (invalid state)
	assert.True(t, users.IsPage)
	assert.True(t, users.IsAPI)
}

// TestScanRoutes_ParentLinks tests parent links are correctly set
func TestScanRoutes_ParentLinks(t *testing.T) {
	fixture := map[string]string{
		"app/pages/users/[id]/edit/page.go": createTestPageHandler("edit", "GET"),
	}

	rootDir := setupFixture(t, fixture)
	appDir := filepath.Join(rootDir, "app")

	root, err := ScanRoutes(appDir)

	require.NoError(t, err)

	// Walk down tree and verify parent links
	pages := root.Children[0]
	assert.Equal(t, root, pages.Parent)

	users := pages.Children[0]
	assert.Equal(t, pages, users.Parent)

	userID := users.Children[0]
	assert.Equal(t, users, userID.Parent)

	edit := userID.Children[0]
	assert.Equal(t, userID, edit.Parent)
}
