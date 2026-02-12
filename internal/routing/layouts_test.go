package routing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBuildLayoutChain_NoLayouts tests chain with no layouts
func TestBuildLayoutChain_NoLayouts(t *testing.T) {
	// Simple node without layout
	node := &RouteNode{
		Path:        "/app/pages/users",
		URLSegment:  "users",
		HandlerFile: "/app/pages/users/page.go",
		Parent:      nil,
	}

	chain := BuildLayoutChain(node, "github.com/user/project")

	assert.NotNil(t, chain)
	assert.Empty(t, chain.Layouts)
	assert.False(t, chain.HasLayouts())
}

// TestBuildLayoutChain_SingleLayout tests chain with one layout
func TestBuildLayoutChain_SingleLayout(t *testing.T) {
	// Root with layout
	root := &RouteNode{
		Path:       "/app/pages",
		URLSegment: "pages",
		LayoutFile: "/app/pages/layout.go",
		HasLayout:  true,
		Parent:     nil,
	}

	// Child handler
	child := &RouteNode{
		Path:        "/app/pages/users",
		URLSegment:  "users",
		HandlerFile: "/app/pages/users/page.go",
		Parent:      root,
	}

	chain := BuildLayoutChain(child, "github.com/user/project")

	assert.NotNil(t, chain)
	assert.Len(t, chain.Layouts, 1)
	assert.True(t, chain.HasLayouts())

	layout := chain.Layouts[0]
	assert.Equal(t, "/app/pages/layout.go", layout.FilePath)
	assert.Equal(t, "Layout", layout.FuncName)
	assert.Contains(t, layout.PackagePath, "github.com/user/project")
}

// TestBuildLayoutChain_MultipleLayouts tests nested layout inheritance
func TestBuildLayoutChain_MultipleLayouts(t *testing.T) {
	// Build hierarchy:
	// /app
	//   /pages (layout.go)
	//     /dashboard (layout.go)
	//       /admin (layout.go)
	//         /users (page.go)

	app := &RouteNode{
		Path:       "/app",
		URLSegment: "",
		Parent:     nil,
	}

	pages := &RouteNode{
		Path:       "/app/pages",
		URLSegment: "pages",
		LayoutFile: "/app/pages/layout.go",
		HasLayout:  true,
		Parent:     app,
	}

	dashboard := &RouteNode{
		Path:       "/app/pages/dashboard",
		URLSegment: "dashboard",
		LayoutFile: "/app/pages/dashboard/layout.go",
		HasLayout:  true,
		Parent:     pages,
	}

	admin := &RouteNode{
		Path:       "/app/pages/dashboard/admin",
		URLSegment: "admin",
		LayoutFile: "/app/pages/dashboard/admin/layout.go",
		HasLayout:  true,
		Parent:     dashboard,
	}

	users := &RouteNode{
		Path:        "/app/pages/dashboard/admin/users",
		URLSegment:  "users",
		HandlerFile: "/app/pages/dashboard/admin/users/page.go",
		Parent:      admin,
	}

	chain := BuildLayoutChain(users, "github.com/user/project")

	assert.NotNil(t, chain)
	assert.Len(t, chain.Layouts, 3)
	assert.True(t, chain.HasLayouts())

	// Verify order: root to leaf
	assert.Equal(t, "/app/pages/layout.go", chain.Layouts[0].FilePath)
	assert.Equal(t, "/app/pages/dashboard/layout.go", chain.Layouts[1].FilePath)
	assert.Equal(t, "/app/pages/dashboard/admin/layout.go", chain.Layouts[2].FilePath)

	// All should have Layout func name
	for _, layout := range chain.Layouts {
		assert.Equal(t, "Layout", layout.FuncName)
	}
}

// TestBuildLayoutChain_SkipNodesWithoutLayout tests nodes without layout are skipped
func TestBuildLayoutChain_SkipNodesWithoutLayout(t *testing.T) {
	// Build hierarchy:
	// /app
	//   /pages (layout.go)
	//     /dashboard (no layout)
	//       /admin (layout.go)
	//         /users (page.go)

	app := &RouteNode{
		Path:       "/app",
		URLSegment: "",
		Parent:     nil,
	}

	pages := &RouteNode{
		Path:       "/app/pages",
		URLSegment: "pages",
		LayoutFile: "/app/pages/layout.go",
		HasLayout:  true,
		Parent:     app,
	}

	dashboard := &RouteNode{
		Path:       "/app/pages/dashboard",
		URLSegment: "dashboard",
		// No layout
		HasLayout: false,
		Parent:    pages,
	}

	admin := &RouteNode{
		Path:       "/app/pages/dashboard/admin",
		URLSegment: "admin",
		LayoutFile: "/app/pages/dashboard/admin/layout.go",
		HasLayout:  true,
		Parent:     dashboard,
	}

	users := &RouteNode{
		Path:        "/app/pages/dashboard/admin/users",
		URLSegment:  "users",
		HandlerFile: "/app/pages/dashboard/admin/users/page.go",
		Parent:      admin,
	}

	chain := BuildLayoutChain(users, "github.com/user/project")

	assert.NotNil(t, chain)
	assert.Len(t, chain.Layouts, 2)
	assert.True(t, chain.HasLayouts())

	// Should only include pages and admin layouts (dashboard skipped)
	assert.Equal(t, "/app/pages/layout.go", chain.Layouts[0].FilePath)
	assert.Equal(t, "/app/pages/dashboard/admin/layout.go", chain.Layouts[1].FilePath)
}

// TestBuildLayoutChain_NodeWithLayout tests building from node that has layout
func TestBuildLayoutChain_NodeWithLayout(t *testing.T) {
	// Node itself has layout
	parent := &RouteNode{
		Path:       "/app/pages",
		URLSegment: "pages",
		LayoutFile: "/app/pages/layout.go",
		HasLayout:  true,
		Parent:     nil,
	}

	node := &RouteNode{
		Path:        "/app/pages/dashboard",
		URLSegment:  "dashboard",
		LayoutFile:  "/app/pages/dashboard/layout.go",
		HasLayout:   true,
		HandlerFile: "/app/pages/dashboard/page.go",
		Parent:      parent,
	}

	chain := BuildLayoutChain(node, "github.com/user/project")

	assert.NotNil(t, chain)
	assert.Len(t, chain.Layouts, 2)

	// Should include both parent and node layouts
	assert.Equal(t, "/app/pages/layout.go", chain.Layouts[0].FilePath)
	assert.Equal(t, "/app/pages/dashboard/layout.go", chain.Layouts[1].FilePath)
}

// TestBuildLayoutChain_PackagePath tests package path generation
func TestBuildLayoutChain_PackagePath(t *testing.T) {
	parent := &RouteNode{
		Path:       "/app/pages",
		URLSegment: "pages",
		LayoutFile: "/app/pages/layout.go",
		HasLayout:  true,
		Parent:     nil,
	}

	child := &RouteNode{
		Path:        "/app/pages/users",
		URLSegment:  "users",
		HandlerFile: "/app/pages/users/page.go",
		Parent:      parent,
	}

	modulePath := "github.com/user/project"
	chain := BuildLayoutChain(child, modulePath)

	assert.Len(t, chain.Layouts, 1)
	layout := chain.Layouts[0]

	// Package path should include module path
	assert.Contains(t, layout.PackagePath, modulePath)
	assert.Contains(t, layout.PackagePath, "/app/pages")
}

// TestBuildLayoutChain_PackageAlias tests package alias generation
func TestBuildLayoutChain_PackageAlias(t *testing.T) {
	parent := &RouteNode{
		Path:       "app/pages",
		URLSegment: "pages",
		LayoutFile: "app/pages/layout.go",
		HasLayout:  true,
		Parent:     nil,
	}

	child := &RouteNode{
		Path:        "app/pages/dashboard",
		URLSegment:  "dashboard",
		LayoutFile:  "app/pages/dashboard/layout.go",
		HasLayout:   true,
		HandlerFile: "app/pages/dashboard/page.go",
		Parent:      parent,
	}

	chain := BuildLayoutChain(child, "github.com/user/project")

	assert.Len(t, chain.Layouts, 2)

	// Package aliases should be generated from path
	assert.NotEmpty(t, chain.Layouts[0].PackageName)
	assert.NotEmpty(t, chain.Layouts[1].PackageName)

	// Aliases should be different
	assert.NotEqual(t, chain.Layouts[0].PackageName, chain.Layouts[1].PackageName)
}

// TestLayoutChain_HasLayouts tests HasLayouts method
func TestLayoutChain_HasLayouts(t *testing.T) {
	tests := []struct {
		name     string
		chain    *LayoutChain
		expected bool
	}{
		{
			name: "empty chain",
			chain: &LayoutChain{
				Layouts: []LayoutInfo{},
			},
			expected: false,
		},
		{
			name: "nil chain",
			chain: &LayoutChain{
				Layouts: nil,
			},
			expected: false,
		},
		{
			name: "single layout",
			chain: &LayoutChain{
				Layouts: []LayoutInfo{
					{FilePath: "/app/pages/layout.go"},
				},
			},
			expected: true,
		},
		{
			name: "multiple layouts",
			chain: &LayoutChain{
				Layouts: []LayoutInfo{
					{FilePath: "/app/pages/layout.go"},
					{FilePath: "/app/pages/dashboard/layout.go"},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.chain.HasLayouts()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestLayoutInfo_GetLayoutDir tests GetLayoutDir method
func TestLayoutInfo_GetLayoutDir(t *testing.T) {
	tests := []struct {
		name     string
		layout   LayoutInfo
		expected string
	}{
		{
			name: "simple path",
			layout: LayoutInfo{
				FilePath: "/app/pages/layout.go",
			},
			expected: "/app/pages",
		},
		{
			name: "nested path",
			layout: LayoutInfo{
				FilePath: "/app/pages/dashboard/admin/layout.go",
			},
			expected: "/app/pages/dashboard/admin",
		},
		{
			name: "relative path",
			layout: LayoutInfo{
				FilePath: "pages/layout.go",
			},
			expected: "pages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.layout.GetLayoutDir()
			assert.Equal(t, tt.expected, dir)
		})
	}
}

// TestBuildLayoutChain_DynamicRoutes tests layout chain with dynamic routes
func TestBuildLayoutChain_DynamicRoutes(t *testing.T) {
	// /pages/users (layout.go)
	//   /[id] (page.go)

	users := &RouteNode{
		Path:       "/app/pages/users",
		URLSegment: "users",
		LayoutFile: "/app/pages/users/layout.go",
		HasLayout:  true,
		Parent:     nil,
	}

	userID := &RouteNode{
		Path:        "/app/pages/users/[id]",
		URLSegment:  "{id}",
		IsDynamic:   true,
		ParamName:   "id",
		HandlerFile: "/app/pages/users/[id]/page.go",
		Parent:      users,
	}

	chain := BuildLayoutChain(userID, "github.com/user/project")

	assert.NotNil(t, chain)
	assert.Len(t, chain.Layouts, 1)
	assert.Equal(t, "/app/pages/users/layout.go", chain.Layouts[0].FilePath)
}

// TestBuildLayoutChain_APIRoutes tests layout chain for API routes
func TestBuildLayoutChain_APIRoutes(t *testing.T) {
	// /api (layout.go)
	//   /v1 (layout.go)
	//     /users (route.go)

	api := &RouteNode{
		Path:       "/app/api",
		URLSegment: "api",
		LayoutFile: "/app/api/layout.go",
		HasLayout:  true,
		Parent:     nil,
	}

	v1 := &RouteNode{
		Path:       "/app/api/v1",
		URLSegment: "v1",
		LayoutFile: "/app/api/v1/layout.go",
		HasLayout:  true,
		Parent:     api,
	}

	users := &RouteNode{
		Path:        "/app/api/v1/users",
		URLSegment:  "users",
		HandlerFile: "/app/api/v1/users/route.go",
		IsAPI:       true,
		Parent:      v1,
	}

	chain := BuildLayoutChain(users, "github.com/user/project")

	assert.NotNil(t, chain)
	assert.Len(t, chain.Layouts, 2)
	assert.True(t, chain.HasLayouts())

	// Verify order
	assert.Equal(t, "/app/api/layout.go", chain.Layouts[0].FilePath)
	assert.Equal(t, "/app/api/v1/layout.go", chain.Layouts[1].FilePath)
}

// TestBuildLayoutChain_NilNode tests nil node handling
func TestBuildLayoutChain_NilNode(t *testing.T) {
	// Should handle gracefully
	chain := BuildLayoutChain(nil, "github.com/user/project")

	assert.NotNil(t, chain)
	assert.Empty(t, chain.Layouts)
	assert.False(t, chain.HasLayouts())
}

// TestBuildLayoutChain_Order tests layout order is root to leaf
func TestBuildLayoutChain_Order(t *testing.T) {
	// Build chain and verify order is maintained
	level1 := &RouteNode{
		Path:       "/app/pages",
		LayoutFile: "/app/pages/layout.go",
		HasLayout:  true,
		Parent:     nil,
	}

	level2 := &RouteNode{
		Path:       "/app/pages/dashboard",
		LayoutFile: "/app/pages/dashboard/layout.go",
		HasLayout:  true,
		Parent:     level1,
	}

	level3 := &RouteNode{
		Path:       "/app/pages/dashboard/admin",
		LayoutFile: "/app/pages/dashboard/admin/layout.go",
		HasLayout:  true,
		Parent:     level2,
	}

	level4 := &RouteNode{
		Path:        "/app/pages/dashboard/admin/users",
		HandlerFile: "/app/pages/dashboard/admin/users/page.go",
		Parent:      level3,
	}

	chain := BuildLayoutChain(level4, "github.com/user/project")

	assert.Len(t, chain.Layouts, 3)

	// Verify order from root to leaf
	assert.Equal(t, "/app/pages/layout.go", chain.Layouts[0].FilePath)
	assert.Equal(t, "/app/pages/dashboard/layout.go", chain.Layouts[1].FilePath)
	assert.Equal(t, "/app/pages/dashboard/admin/layout.go", chain.Layouts[2].FilePath)
}

// TestBuildLayoutChain_EmptyModulePath tests with empty module path
func TestBuildLayoutChain_EmptyModulePath(t *testing.T) {
	parent := &RouteNode{
		Path:       "/app/pages",
		LayoutFile: "/app/pages/layout.go",
		HasLayout:  true,
		Parent:     nil,
	}

	child := &RouteNode{
		Path:        "/app/pages/users",
		HandlerFile: "/app/pages/users/page.go",
		Parent:      parent,
	}

	chain := BuildLayoutChain(child, "")

	assert.Len(t, chain.Layouts, 1)
	// Should still generate package path even with empty module
	assert.NotEmpty(t, chain.Layouts[0].PackagePath)
}
