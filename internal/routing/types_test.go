package routing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRouteNode_Construction tests basic RouteNode instantiation
func TestRouteNode_Construction(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *RouteNode
		validate func(*testing.T, *RouteNode)
	}{
		{
			name: "empty route node",
			setup: func() *RouteNode {
				return &RouteNode{}
			},
			validate: func(t *testing.T, n *RouteNode) {
				assert.Empty(t, n.Path)
				assert.Empty(t, n.URLSegment)
				assert.Nil(t, n.Children)
				assert.Nil(t, n.Parent)
				assert.Empty(t, n.HandlerFile)
				assert.Empty(t, n.LayoutFile)
				assert.Empty(t, n.Methods)
				assert.Empty(t, n.PackageName)
				assert.False(t, n.IsDirectory)
				assert.False(t, n.IsPage)
				assert.False(t, n.IsAPI)
				assert.False(t, n.HasLayout)
				assert.False(t, n.IsDynamic)
				assert.False(t, n.IsCatchAll)
				assert.Empty(t, n.ParamName)
			},
		},
		{
			name: "static page route",
			setup: func() *RouteNode {
				return &RouteNode{
					Path:        "/app/pages/users",
					URLSegment:  "users",
					HandlerFile: "/app/pages/users/page.go",
					IsPage:      true,
					Methods:     []string{"GET", "POST"},
					PackageName: "users",
				}
			},
			validate: func(t *testing.T, n *RouteNode) {
				assert.Equal(t, "/app/pages/users", n.Path)
				assert.Equal(t, "users", n.URLSegment)
				assert.Equal(t, "/app/pages/users/page.go", n.HandlerFile)
				assert.True(t, n.IsPage)
				assert.False(t, n.IsAPI)
				assert.ElementsMatch(t, []string{"GET", "POST"}, n.Methods)
				assert.Equal(t, "users", n.PackageName)
			},
		},
		{
			name: "API route",
			setup: func() *RouteNode {
				return &RouteNode{
					Path:        "/app/api/users",
					URLSegment:  "users",
					HandlerFile: "/app/api/users/route.go",
					IsAPI:       true,
					Methods:     []string{"GET", "PUT", "DELETE"},
					PackageName: "users",
				}
			},
			validate: func(t *testing.T, n *RouteNode) {
				assert.Equal(t, "/app/api/users", n.Path)
				assert.True(t, n.IsAPI)
				assert.False(t, n.IsPage)
				assert.Equal(t, "/app/api/users/route.go", n.HandlerFile)
				assert.ElementsMatch(t, []string{"GET", "PUT", "DELETE"}, n.Methods)
			},
		},
		{
			name: "dynamic route",
			setup: func() *RouteNode {
				return &RouteNode{
					Path:        "/app/pages/users/[id]",
					URLSegment:  "{id}",
					HandlerFile: "/app/pages/users/[id]/page.go",
					IsPage:      true,
					IsDynamic:   true,
					ParamName:   "id",
					Methods:     []string{"GET"},
					PackageName: "user_id",
				}
			},
			validate: func(t *testing.T, n *RouteNode) {
				assert.True(t, n.IsDynamic)
				assert.False(t, n.IsCatchAll)
				assert.Equal(t, "id", n.ParamName)
				assert.Equal(t, "{id}", n.URLSegment)
			},
		},
		{
			name: "catch-all route",
			setup: func() *RouteNode {
				return &RouteNode{
					Path:        "/app/pages/[...slug]",
					URLSegment:  "{slug...}",
					HandlerFile: "/app/pages/[...slug]/page.go",
					IsPage:      true,
					IsDynamic:   true,
					IsCatchAll:  true,
					ParamName:   "slug",
					Methods:     []string{"GET"},
					PackageName: "slug_catchall",
				}
			},
			validate: func(t *testing.T, n *RouteNode) {
				assert.True(t, n.IsCatchAll)
				assert.True(t, n.IsDynamic)
				assert.Equal(t, "slug", n.ParamName)
				assert.Equal(t, "{slug...}", n.URLSegment)
			},
		},
		{
			name: "route with layout",
			setup: func() *RouteNode {
				return &RouteNode{
					Path:        "/app/pages/dashboard",
					URLSegment:  "dashboard",
					LayoutFile:  "/app/pages/dashboard/layout.go",
					HasLayout:   true,
					IsDirectory: true,
					PackageName: "dashboard",
				}
			},
			validate: func(t *testing.T, n *RouteNode) {
				assert.True(t, n.HasLayout)
				assert.Equal(t, "/app/pages/dashboard/layout.go", n.LayoutFile)
				assert.Empty(t, n.HandlerFile)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.setup()
			tt.validate(t, node)
		})
	}
}

// TestRouteNode_ParentChild tests parent-child relationships
func TestRouteNode_ParentChild(t *testing.T) {
	parent := &RouteNode{
		Path:       "/app/pages",
		URLSegment: "pages",
		Children:   make([]*RouteNode, 0),
	}

	child := &RouteNode{
		Path:       "/app/pages/users",
		URLSegment: "users",
		Parent:     parent,
	}

	parent.Children = append(parent.Children, child)

	assert.Equal(t, parent, child.Parent)
	assert.Contains(t, parent.Children, child)
	assert.Len(t, parent.Children, 1)
}

// TestRouteNode_MultipleChildren tests nodes with multiple children
func TestRouteNode_MultipleChildren(t *testing.T) {
	parent := &RouteNode{
		Path:       "/app/pages",
		URLSegment: "pages",
		Children:   make([]*RouteNode, 0),
	}

	children := []*RouteNode{
		{Path: "/app/pages/users", URLSegment: "users", Parent: parent},
		{Path: "/app/pages/posts", URLSegment: "posts", Parent: parent},
		{Path: "/app/pages/dashboard", URLSegment: "dashboard", Parent: parent},
	}

	parent.Children = children

	assert.Len(t, parent.Children, 3)
	for _, child := range children {
		assert.Equal(t, parent, child.Parent)
	}
}

// TestLayoutChain_Construction tests LayoutChain construction
func TestLayoutChain_Construction(t *testing.T) {
	tests := []struct {
		name     string
		chain    *LayoutChain
		validate func(*testing.T, *LayoutChain)
	}{
		{
			name: "empty chain",
			chain: &LayoutChain{
				Layouts: make([]LayoutInfo, 0),
			},
			validate: func(t *testing.T, c *LayoutChain) {
				assert.Empty(t, c.Layouts)
				assert.False(t, c.HasLayouts())
			},
		},
		{
			name: "single layout",
			chain: &LayoutChain{
				Layouts: []LayoutInfo{
					{
						FilePath:    "/app/pages/layout.go",
						PackagePath: "github.com/user/app/pages",
						PackageName: "pages",
						FuncName:    "Layout",
					},
				},
			},
			validate: func(t *testing.T, c *LayoutChain) {
				assert.Len(t, c.Layouts, 1)
				assert.True(t, c.HasLayouts())
				assert.Equal(t, "/app/pages/layout.go", c.Layouts[0].FilePath)
				assert.Equal(t, "Layout", c.Layouts[0].FuncName)
			},
		},
		{
			name: "multiple layouts (root to leaf)",
			chain: &LayoutChain{
				Layouts: []LayoutInfo{
					{
						FilePath:    "/app/pages/layout.go",
						PackagePath: "github.com/user/app/pages",
						PackageName: "pages",
						FuncName:    "Layout",
					},
					{
						FilePath:    "/app/pages/dashboard/layout.go",
						PackagePath: "github.com/user/app/pages/dashboard",
						PackageName: "pages_dashboard",
						FuncName:    "Layout",
					},
				},
			},
			validate: func(t *testing.T, c *LayoutChain) {
				assert.Len(t, c.Layouts, 2)
				assert.True(t, c.HasLayouts())
				// Verify order: root to leaf
				assert.Equal(t, "/app/pages/layout.go", c.Layouts[0].FilePath)
				assert.Equal(t, "/app/pages/dashboard/layout.go", c.Layouts[1].FilePath)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.chain)
		})
	}
}

// TestRouteNode_MethodsCombinations tests various HTTP method combinations
func TestRouteNode_MethodsCombinations(t *testing.T) {
	tests := []struct {
		name    string
		methods []string
	}{
		{"single GET", []string{"GET"}},
		{"single POST", []string{"POST"}},
		{"GET and POST", []string{"GET", "POST"}},
		{"all methods", []string{"GET", "POST", "PUT", "DELETE", "PATCH"}},
		{"CRUD methods", []string{"GET", "POST", "PUT", "DELETE"}},
		{"no methods", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &RouteNode{
				Methods: tt.methods,
			}
			assert.ElementsMatch(t, tt.methods, node.Methods)
		})
	}
}

// TestRouteNode_ComplexTree tests a complex route tree structure
func TestRouteNode_ComplexTree(t *testing.T) {
	// Build tree:
	// /pages
	//   /users (page.go, layout.go)
	//     /[id] (page.go)
	//       /edit (page.go)
	//   /posts (page.go)

	root := &RouteNode{
		Path:       "/app/pages",
		URLSegment: "pages",
		Children:   make([]*RouteNode, 0),
	}

	users := &RouteNode{
		Path:        "/app/pages/users",
		URLSegment:  "users",
		Parent:      root,
		HandlerFile: "/app/pages/users/page.go",
		LayoutFile:  "/app/pages/users/layout.go",
		IsPage:      true,
		HasLayout:   true,
		Methods:     []string{"GET"},
		Children:    make([]*RouteNode, 0),
	}

	userID := &RouteNode{
		Path:        "/app/pages/users/[id]",
		URLSegment:  "{id}",
		Parent:      users,
		HandlerFile: "/app/pages/users/[id]/page.go",
		IsPage:      true,
		IsDynamic:   true,
		ParamName:   "id",
		Methods:     []string{"GET", "PUT", "DELETE"},
		Children:    make([]*RouteNode, 0),
	}

	userEdit := &RouteNode{
		Path:        "/app/pages/users/[id]/edit",
		URLSegment:  "edit",
		Parent:      userID,
		HandlerFile: "/app/pages/users/[id]/edit/page.go",
		IsPage:      true,
		Methods:     []string{"GET", "POST"},
	}

	posts := &RouteNode{
		Path:        "/app/pages/posts",
		URLSegment:  "posts",
		Parent:      root,
		HandlerFile: "/app/pages/posts/page.go",
		IsPage:      true,
		Methods:     []string{"GET", "POST"},
	}

	userID.Children = append(userID.Children, userEdit)
	users.Children = append(users.Children, userID)
	root.Children = append(root.Children, users, posts)

	// Validate structure
	assert.Len(t, root.Children, 2)
	assert.Equal(t, users, root.Children[0])
	assert.Equal(t, posts, root.Children[1])

	assert.Len(t, users.Children, 1)
	assert.Equal(t, userID, users.Children[0])

	assert.Len(t, userID.Children, 1)
	assert.Equal(t, userEdit, userID.Children[0])

	// Validate parent links
	assert.Equal(t, root, users.Parent)
	assert.Equal(t, users, userID.Parent)
	assert.Equal(t, userID, userEdit.Parent)

	// Validate properties
	assert.True(t, users.HasLayout)
	assert.True(t, userID.IsDynamic)
	assert.False(t, userEdit.IsDynamic)
}
