package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cstone-io/twine/internal/routing"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helpers

func setupTestProject(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Create go.mod
	goModContent := `module github.com/test/project

go 1.22
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644))

	// Create app directory structure
	appDir := filepath.Join(tmpDir, "app")
	require.NoError(t, os.MkdirAll(appDir, 0755))

	return tmpDir
}

func createTestRoute(t *testing.T, projectDir, routePath, content string) {
	t.Helper()
	fullPath := filepath.Join(projectDir, "app", routePath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
}

// TestNewRoutesCommand tests routes command creation
func TestNewRoutesCommand(t *testing.T) {
	cmd := NewRoutesCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "routes", cmd.Use)
	assert.Equal(t, "Manage file-based routes", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Verify subcommands
	assert.True(t, cmd.HasSubCommands())
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 2)

	// Find generate and list commands
	var generateCmd, listCmd *cobra.Command
	for _, subcmd := range subcommands {
		if subcmd.Use == "generate" {
			generateCmd = subcmd
		} else if subcmd.Use == "list" {
			listCmd = subcmd
		}
	}

	assert.NotNil(t, generateCmd)
	assert.NotNil(t, listCmd)
}

// TestRoutesGenerateCommand_Success tests successful route generation
func TestRoutesGenerateCommand_Success(t *testing.T) {
	projectDir := setupTestProject(t)

	// Create a simple page
	pageContent := `package index

import "github.com/cstone-io/twine/kit"

func GET(k *kit.Kit) error {
	return nil
}
`
	createTestRoute(t, projectDir, "pages/index/page.go", pageContent)

	// Change to project directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(projectDir))

	// Run generate command
	cmd := newRoutesGenerateCommand()

	err := cmd.Execute()
	assert.NoError(t, err)

	// Verify routes.gen.go was created
	routesFile := filepath.Join(projectDir, "app", "routes.gen.go")
	assert.FileExists(t, routesFile)

	// Verify content
	content, err := os.ReadFile(routesFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "package app")
	assert.Contains(t, string(content), "RegisterRoutes")

	// Note: Output goes to stdout, not captured in test
}

// TestRoutesGenerateCommand_NoAppDirectory tests error when app/ doesn't exist
func TestRoutesGenerateCommand_NoAppDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod but no app/ directory
	goModContent := `module github.com/test/project

go 1.22
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644))

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(tmpDir))

	cmd := newRoutesGenerateCommand()
	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "app/ directory not found")
}

// TestRoutesGenerateCommand_InvalidRoute tests validation error
func TestRoutesGenerateCommand_InvalidRoute(t *testing.T) {
	projectDir := setupTestProject(t)

	// Create invalid route (handler without methods)
	pageContent := `package test

import "github.com/cstone-io/twine/kit"

func helper() {}
`
	createTestRoute(t, projectDir, "pages/test/page.go", pageContent)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(projectDir))

	cmd := newRoutesGenerateCommand()
	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation error")
}

// TestRoutesListCommand_Success tests route listing
func TestRoutesListCommand_Success(t *testing.T) {
	projectDir := setupTestProject(t)

	// Create multiple routes
	usersContent := `package users

import "github.com/cstone-io/twine/kit"

func GET(k *kit.Kit) error { return nil }
func POST(k *kit.Kit) error { return nil }
`
	createTestRoute(t, projectDir, "pages/users/page.go", usersContent)

	postsContent := `package posts

import "github.com/cstone-io/twine/kit"

func GET(k *kit.Kit) error { return nil }
`
	createTestRoute(t, projectDir, "pages/posts/page.go", postsContent)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(projectDir))

	cmd := newRoutesListCommand()

	err := cmd.Execute()
	assert.NoError(t, err)

	// Note: Output goes to stdout via displayRouteTable, not captured in test
	// The command execution succeeding is the main test here
}

// TestRoutesListCommand_WithLayouts tests listing with layouts
func TestRoutesListCommand_WithLayouts(t *testing.T) {
	projectDir := setupTestProject(t)

	// Create layout
	layoutContent := `package pages

import "github.com/cstone-io/twine/middleware"

func Layout() middleware.Middleware {
	return nil
}
`
	createTestRoute(t, projectDir, "pages/layout.go", layoutContent)

	// Create page
	pageContent := `package dashboard

import "github.com/cstone-io/twine/kit"

func GET(k *kit.Kit) error { return nil }
`
	createTestRoute(t, projectDir, "pages/dashboard/page.go", pageContent)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(projectDir))

	cmd := newRoutesListCommand()

	err := cmd.Execute()
	assert.NoError(t, err)

	// Note: Output goes to stdout via displayRouteTable, not captured in test
}

// TestRoutesListCommand_NoRoutes tests empty route list
func TestRoutesListCommand_NoRoutes(t *testing.T) {
	projectDir := setupTestProject(t)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(projectDir))

	cmd := newRoutesListCommand()

	err := cmd.Execute()
	assert.NoError(t, err)

	// Note: Output goes to stdout via displayRouteTable, not captured in test
}

// TestCollectAllRoutes tests route collection
func TestCollectAllRoutes(t *testing.T) {
	// Build test tree
	root := &routing.RouteNode{
		Path:       "/app",
		URLSegment: "",
		Children: []*routing.RouteNode{
			{
				Path:       "/app/pages",
				URLSegment: "pages",
				Children: []*routing.RouteNode{
					{
						Path:        "/app/pages/users",
						URLSegment:  "users",
						HandlerFile: "/app/pages/users/page.go",
						Methods:     []string{"GET"},
					},
					{
						Path:        "/app/pages/posts",
						URLSegment:  "posts",
						HandlerFile: "/app/pages/posts/page.go",
						Methods:     []string{"GET", "POST"},
					},
				},
			},
		},
	}

	routes := collectAllRoutes(root)

	assert.Len(t, routes, 2)
	assert.Equal(t, "/app/pages/users/page.go", routes[0].HandlerFile)
	assert.Equal(t, "/app/pages/posts/page.go", routes[1].HandlerFile)
}

// TestCollectAllRoutes_Empty tests empty tree
func TestCollectAllRoutes_Empty(t *testing.T) {
	root := &routing.RouteNode{
		Path:     "/app",
		Children: []*routing.RouteNode{},
	}

	routes := collectAllRoutes(root)
	assert.Empty(t, routes)
}

// TestCollectAllRoutes_Nested tests nested routes
func TestCollectAllRoutes_Nested(t *testing.T) {
	root := &routing.RouteNode{
		Path: "/app",
		Children: []*routing.RouteNode{
			{
				Path:       "/app/pages",
				URLSegment: "pages",
				Children: []*routing.RouteNode{
					{
						Path:        "/app/pages/users",
						URLSegment:  "users",
						HandlerFile: "/app/pages/users/page.go",
						Methods:     []string{"GET"},
						Children: []*routing.RouteNode{
							{
								Path:        "/app/pages/users/[id]",
								URLSegment:  "{id}",
								HandlerFile: "/app/pages/users/[id]/page.go",
								Methods:     []string{"GET"},
							},
						},
					},
				},
			},
		},
	}

	routes := collectAllRoutes(root)
	assert.Len(t, routes, 2)
}

// TestCollectAllLayouts tests layout collection
func TestCollectAllLayouts(t *testing.T) {
	root := &routing.RouteNode{
		Path: "/app",
		Children: []*routing.RouteNode{
			{
				Path:       "/app/pages",
				URLSegment: "pages",
				LayoutFile: "/app/pages/layout.go",
				HasLayout:  true,
				Children: []*routing.RouteNode{
					{
						Path:       "/app/pages/dashboard",
						URLSegment: "dashboard",
						LayoutFile: "/app/pages/dashboard/layout.go",
						HasLayout:  true,
					},
				},
			},
		},
	}

	layouts := collectAllLayouts(root)

	assert.Len(t, layouts, 2)
	assert.Equal(t, "/app/pages/layout.go", layouts[0].LayoutFile)
	assert.Equal(t, "/app/pages/dashboard/layout.go", layouts[1].LayoutFile)
}

// TestCollectAllLayouts_Empty tests no layouts
func TestCollectAllLayouts_Empty(t *testing.T) {
	root := &routing.RouteNode{
		Path: "/app",
		Children: []*routing.RouteNode{
			{
				Path:        "/app/pages",
				URLSegment:  "pages",
				HandlerFile: "/app/pages/page.go",
			},
		},
	}

	layouts := collectAllLayouts(root)
	assert.Empty(t, layouts)
}

// TestGetLayoutPattern tests layout pattern generation
func TestGetLayoutPattern(t *testing.T) {
	tests := []struct {
		name     string
		node     *routing.RouteNode
		expected string
	}{
		{
			name: "root layout",
			node: &routing.RouteNode{
				URLSegment: "",
				Parent:     nil,
			},
			expected: "/", // Root path
		},
		{
			name: "pages layout",
			node: &routing.RouteNode{
				URLSegment: "users",
				Parent: &routing.RouteNode{
					URLSegment: "pages",
					Parent:     nil,
				},
			},
			expected: "/users/*",
		},
		{
			name: "nested layout",
			node: &routing.RouteNode{
				URLSegment: "admin",
				Parent: &routing.RouteNode{
					URLSegment: "dashboard",
					Parent: &routing.RouteNode{
						URLSegment: "pages",
						Parent:     nil,
					},
				},
			},
			expected: "/dashboard/admin/*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := getLayoutPattern(tt.node)
			assert.Equal(t, tt.expected, pattern)
		})
	}
}

// TestDisplayRouteTable tests route table display
func TestDisplayRouteTable(t *testing.T) {
	root := &routing.RouteNode{
		Path: "/app",
		Children: []*routing.RouteNode{
			{
				Path:       "/app/pages",
				URLSegment: "pages",
				Children: []*routing.RouteNode{
					{
						Path:        "/app/pages/users",
						URLSegment:  "users",
						HandlerFile: "/app/pages/users/page.go",
						Methods:     []string{"GET", "POST"},
						Parent: &routing.RouteNode{
							URLSegment: "pages",
							Parent:     nil,
						},
					},
				},
			},
		},
	}

	// Capture output - should not panic
	assert.NotPanics(t, func() {
		displayRouteTable(root)
	})
}

// TestDisplayRouteTable_EmptyRoutes tests empty route table
func TestDisplayRouteTable_EmptyRoutes(t *testing.T) {
	root := &routing.RouteNode{
		Path:     "/app",
		Children: []*routing.RouteNode{},
	}

	// Should handle empty routes gracefully
	assert.NotPanics(t, func() {
		displayRouteTable(root)
	})
}

// TestRoutesGenerateCommand_ComplexProject tests complex project structure
func TestRoutesGenerateCommand_ComplexProject(t *testing.T) {
	projectDir := setupTestProject(t)

	// Create multiple routes
	usersContent := `package users

import "github.com/cstone-io/twine/kit"

func GET(k *kit.Kit) error { return nil }
func POST(k *kit.Kit) error { return nil }
`
	createTestRoute(t, projectDir, "pages/users/page.go", usersContent)

	userIDContent := `package user_id

import "github.com/cstone-io/twine/kit"

func GET(k *kit.Kit) error { return nil }
func PUT(k *kit.Kit) error { return nil }
func DELETE(k *kit.Kit) error { return nil }
`
	createTestRoute(t, projectDir, "pages/users/[id]/page.go", userIDContent)

	apiContent := `package users

import "github.com/cstone-io/twine/kit"

func GET(k *kit.Kit) error { return nil }
`
	createTestRoute(t, projectDir, "api/users/route.go", apiContent)

	layoutContent := `package pages

import "github.com/cstone-io/twine/middleware"

func Layout() middleware.Middleware {
	return nil
}
`
	createTestRoute(t, projectDir, "pages/layout.go", layoutContent)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(projectDir))

	cmd := newRoutesGenerateCommand()

	err := cmd.Execute()
	assert.NoError(t, err)

	// Verify routes.gen.go exists and has correct content
	routesFile := filepath.Join(projectDir, "app", "routes.gen.go")
	content, err := os.ReadFile(routesFile)
	require.NoError(t, err)

	generated := string(content)
	assert.Contains(t, generated, "RegisterRoutes")
	assert.Contains(t, generated, "/users")
	assert.Contains(t, generated, "/users/{id}")
	assert.Contains(t, generated, "/api/users")
	assert.Contains(t, generated, "applyMiddleware") // Layout middleware
}
