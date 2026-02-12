package routing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestToURLPattern tests URL pattern generation
func TestToURLPattern(t *testing.T) {
	tests := []struct {
		name     string
		node     *RouteNode
		expected string
	}{
		{
			name: "root path",
			node: &RouteNode{
				URLSegment: "",
				Parent:     nil,
			},
			expected: "/",
		},
		{
			name: "simple static route",
			node: &RouteNode{
				URLSegment: "users",
				Parent: &RouteNode{
					URLSegment: "pages",
					Parent:     nil,
				},
			},
			expected: "/users",
		},
		{
			name: "nested static route",
			node: &RouteNode{
				URLSegment: "edit",
				Parent: &RouteNode{
					URLSegment: "users",
					Parent: &RouteNode{
						URLSegment: "pages",
						Parent:     nil,
					},
				},
			},
			expected: "/users/edit",
		},
		{
			name: "dynamic route",
			node: &RouteNode{
				URLSegment: "{id}",
				IsDynamic:  true,
				ParamName:  "id",
				Parent: &RouteNode{
					URLSegment: "users",
					Parent: &RouteNode{
						URLSegment: "pages",
						Parent:     nil,
					},
				},
			},
			expected: "/users/{id}",
		},
		{
			name: "catch-all route",
			node: &RouteNode{
				URLSegment: "{slug...}",
				IsCatchAll: true,
				IsDynamic:  true,
				ParamName:  "slug",
				Parent: &RouteNode{
					URLSegment: "docs",
					Parent: &RouteNode{
						URLSegment: "pages",
						Parent:     nil,
					},
				},
			},
			expected: "/docs/{slug...}",
		},
		{
			name: "API route",
			node: &RouteNode{
				URLSegment: "users",
				Parent: &RouteNode{
					URLSegment: "api",
					Parent:     nil,
				},
			},
			expected: "/api/users",
		},
		{
			name: "deeply nested route",
			node: &RouteNode{
				URLSegment: "comments",
				Parent: &RouteNode{
					URLSegment: "{id}",
					IsDynamic:  true,
					Parent: &RouteNode{
						URLSegment: "posts",
						Parent: &RouteNode{
							URLSegment: "api",
							Parent:     nil,
						},
					},
				},
			},
			expected: "/api/posts/{id}/comments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := tt.node.ToURLPattern()
			assert.Equal(t, tt.expected, pattern)
		})
	}
}

// TestGetFullPath tests full path generation
func TestGetFullPath(t *testing.T) {
	tests := []struct {
		name     string
		node     *RouteNode
		expected string
	}{
		{
			name: "root node",
			node: &RouteNode{
				URLSegment: "",
				Parent:     nil,
			},
			expected: "",
		},
		{
			name: "pages root (filtered)",
			node: &RouteNode{
				URLSegment: "pages",
				Parent:     nil,
			},
			expected: "",
		},
		{
			name: "single segment after pages",
			node: &RouteNode{
				URLSegment: "users",
				Parent: &RouteNode{
					URLSegment: "pages",
					Parent:     nil,
				},
			},
			expected: "/users",
		},
		{
			name: "API segment (included)",
			node: &RouteNode{
				URLSegment: "users",
				Parent: &RouteNode{
					URLSegment: "api",
					Parent:     nil,
				},
			},
			expected: "/api/users",
		},
		{
			name: "multiple segments",
			node: &RouteNode{
				URLSegment: "edit",
				Parent: &RouteNode{
					URLSegment: "{id}",
					Parent: &RouteNode{
						URLSegment: "users",
						Parent: &RouteNode{
							URLSegment: "pages",
							Parent:     nil,
						},
					},
				},
			},
			expected: "/users/{id}/edit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.node.GetFullPath()
			assert.Equal(t, tt.expected, path)
		})
	}
}

// TestSanitizePackageName tests package name sanitization
func TestSanitizePackageName(t *testing.T) {
	tests := []struct {
		name     string
		dirName  string
		expected string
	}{
		{
			name:     "simple dynamic param",
			dirName:  "[id]",
			expected: "id_param",
		},
		{
			name:     "named param",
			dirName:  "[userId]",
			expected: "userId_param",
		},
		{
			name:     "snake_case param",
			dirName:  "[user_id]",
			expected: "user_id_param",
		},
		{
			name:     "catch-all param",
			dirName:  "[...slug]",
			expected: "slug_catchall",
		},
		{
			name:     "named catch-all",
			dirName:  "[...pathSegments]",
			expected: "pathSegments_catchall",
		},
		{
			name:     "static name (no brackets)",
			dirName:  "users",
			expected: "users_param",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizePackageName(tt.dirName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetPackageAlias tests package alias generation
func TestGetPackageAlias(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "root",
			path:     "",
			expected: "root",
		},
		{
			name:     "single segment",
			path:     "pages/users",
			expected: "pages_users",
		},
		{
			name:     "multiple segments",
			path:     "pages/dashboard/reports",
			expected: "pages_dashboard_reports",
		},
		{
			name:     "with dynamic segment",
			path:     "pages/users/[id]",
			expected: "pages_users_id_param",
		},
		{
			name:     "with catch-all segment",
			path:     "pages/docs/[...slug]",
			expected: "pages_docs_slug_catchall",
		},
		{
			name:     "API route",
			path:     "api/users",
			expected: "api_users",
		},
		{
			name:     "nested API route",
			path:     "api/v1/users/[id]",
			expected: "api_v1_users_id_param",
		},
		{
			name:     "with app prefix (filtered)",
			path:     "app/pages/users",
			expected: "pages_users",
		},
		{
			name:     "with leading slash",
			path:     "/app/pages/users",
			expected: "pages_users",
		},
		{
			name:     "with dashes (replaced)",
			path:     "pages/user-profile",
			expected: "pages_user_profile",
		},
		{
			name:     "with dots (replaced)",
			path:     "pages/v1.0",
			expected: "pages_v1_0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &RouteNode{
				Path: tt.path,
			}
			alias := node.GetPackageAlias()
			assert.Equal(t, tt.expected, alias)
		})
	}
}

// TestGetPackagePath tests Go import path generation
func TestGetPackagePath(t *testing.T) {
	tests := []struct {
		name       string
		nodePath   string
		modulePath string
		expected   string
	}{
		{
			name:       "simple page",
			nodePath:   "/app/pages/users",
			modulePath: "github.com/user/project",
			expected:   "github.com/user/project/app/pages/users",
		},
		{
			name:       "API route",
			nodePath:   "/app/api/users",
			modulePath: "github.com/user/project",
			expected:   "github.com/user/project/app/api/users",
		},
		{
			name:       "dynamic route (sanitized)",
			nodePath:   "/app/pages/users/[id]",
			modulePath: "github.com/user/project",
			expected:   "github.com/user/project/app/pages/users/id_param",
		},
		{
			name:       "catch-all route (sanitized)",
			nodePath:   "/app/pages/docs/[...slug]",
			modulePath: "github.com/user/project",
			expected:   "github.com/user/project/app/pages/docs/slug_catchall",
		},
		{
			name:       "nested dynamic route",
			nodePath:   "/app/pages/users/[userId]/posts/[postId]",
			modulePath: "github.com/user/project",
			expected:   "github.com/user/project/app/pages/users/userId_param/posts/postId_param",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &RouteNode{
				Path: tt.nodePath,
			}
			packagePath := node.GetPackagePath(tt.modulePath)
			assert.Equal(t, tt.expected, packagePath)
		})
	}
}

// TestGetPackageAlias_UniqueAliases tests that aliases are unique
func TestGetPackageAlias_UniqueAliases(t *testing.T) {
	nodes := []*RouteNode{
		{Path: "pages/users"},
		{Path: "api/users"},
		{Path: "pages/posts"},
		{Path: "pages/users/[id]"},
		{Path: "api/users/[id]"},
	}

	aliases := make(map[string]bool)
	for _, node := range nodes {
		alias := node.GetPackageAlias()
		assert.False(t, aliases[alias], "duplicate alias: %s", alias)
		aliases[alias] = true
	}
}

// TestToURLPattern_ComplexTree tests patterns in a complex tree
func TestToURLPattern_ComplexTree(t *testing.T) {
	// Build tree:
	// /pages
	//   /users
	//     /[id]
	//       /edit

	root := &RouteNode{
		URLSegment: "",
		Parent:     nil,
	}

	pages := &RouteNode{
		URLSegment: "pages",
		Parent:     root,
	}

	users := &RouteNode{
		URLSegment: "users",
		Parent:     pages,
	}

	userID := &RouteNode{
		URLSegment: "{id}",
		IsDynamic:  true,
		ParamName:  "id",
		Parent:     users,
	}

	edit := &RouteNode{
		URLSegment: "edit",
		Parent:     userID,
	}

	// Test patterns at each level
	assert.Equal(t, "/", root.ToURLPattern())
	assert.Equal(t, "/", pages.ToURLPattern())   // "pages" filtered, returns "/"
	assert.Equal(t, "/users", users.ToURLPattern())
	assert.Equal(t, "/users/{id}", userID.ToURLPattern())
	assert.Equal(t, "/users/{id}/edit", edit.ToURLPattern())
}

// TestGetFullPath_WithAPIPrefix tests API routes preserve /api prefix
func TestGetFullPath_WithAPIPrefix(t *testing.T) {
	root := &RouteNode{
		URLSegment: "",
		Parent:     nil,
	}

	api := &RouteNode{
		URLSegment: "api",
		Parent:     root,
	}

	users := &RouteNode{
		URLSegment: "users",
		Parent:     api,
	}

	userID := &RouteNode{
		URLSegment: "{id}",
		IsDynamic:  true,
		Parent:     users,
	}

	assert.Equal(t, "/api/users", users.GetFullPath())
	assert.Equal(t, "/api/users/{id}", userID.GetFullPath())
}

// TestGetFullPath_WithPagesPrefix tests pages routes omit /pages prefix
func TestGetFullPath_WithPagesPrefix(t *testing.T) {
	root := &RouteNode{
		URLSegment: "",
		Parent:     nil,
	}

	pages := &RouteNode{
		URLSegment: "pages",
		Parent:     root,
	}

	users := &RouteNode{
		URLSegment: "users",
		Parent:     pages,
	}

	userID := &RouteNode{
		URLSegment: "{id}",
		IsDynamic:  true,
		Parent:     users,
	}

	// pages segment should be filtered out
	assert.Equal(t, "", pages.GetFullPath())
	assert.Equal(t, "/users", users.GetFullPath())
	assert.Equal(t, "/users/{id}", userID.GetFullPath())
}

// TestSanitizePackageName_EdgeCases tests edge cases
func TestSanitizePackageName_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		dirName  string
		expected string
	}{
		{
			name:     "empty brackets",
			dirName:  "[]",
			expected: "_param",
		},
		{
			name:     "just dots",
			dirName:  "[...]",
			expected: "_catchall",
		},
		{
			name:     "number param",
			dirName:  "[123]",
			expected: "123_param",
		},
		{
			name:     "special chars in param",
			dirName:  "[user-id]",
			expected: "user-id_param",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizePackageName(tt.dirName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetPackageAlias_EmptyPath tests empty path handling
func TestGetPackageAlias_EmptyPath(t *testing.T) {
	node := &RouteNode{
		Path: "",
	}
	alias := node.GetPackageAlias()
	assert.Equal(t, "root", alias)
}

// TestGetPackagePath_RelativePathHandling tests various path formats
func TestGetPackagePath_RelativePathHandling(t *testing.T) {
	tests := []struct {
		name       string
		nodePath   string
		modulePath string
		expected   string
	}{
		{
			name:       "path with leading slash",
			nodePath:   "/app/pages/users",
			modulePath: "github.com/user/project",
			expected:   "github.com/user/project/app/pages/users",
		},
		{
			name:       "path without leading slash",
			nodePath:   "app/pages/users",
			modulePath: "github.com/user/project",
			expected:   "github.com/user/project/app/pages/users",
		},
		{
			name:       "empty module path",
			nodePath:   "/app/pages/users",
			modulePath: "",
			expected:   "/app/pages/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &RouteNode{
				Path: tt.nodePath,
			}
			packagePath := node.GetPackagePath(tt.modulePath)
			assert.Equal(t, tt.expected, packagePath)
		})
	}
}

// TestToURLPattern_NilParent tests handling nil parent
func TestToURLPattern_NilParent(t *testing.T) {
	node := &RouteNode{
		URLSegment: "users",
		Parent:     nil,
	}

	pattern := node.ToURLPattern()
	assert.Equal(t, "/users", pattern)
}

// TestGetFullPath_NilParent tests handling nil parent
func TestGetFullPath_NilParent(t *testing.T) {
	node := &RouteNode{
		URLSegment: "users",
		Parent:     nil,
	}

	path := node.GetFullPath()
	assert.Equal(t, "/users", path)
}

// TestGetPackageAlias_ComplexPaths tests complex real-world paths
func TestGetPackageAlias_ComplexPaths(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "dashboard admin users",
			path:     "pages/dashboard/admin/users",
			expected: "pages_dashboard_admin_users",
		},
		{
			name:     "versioned API",
			path:     "api/v1/users/[id]/profile",
			expected: "api_v1_users_id_param_profile",
		},
		{
			name:     "multiple dynamic segments",
			path:     "pages/orgs/[orgId]/projects/[projectId]",
			expected: "pages_orgs_orgId_param_projects_projectId_param",
		},
		{
			name:     "mixed static and dynamic",
			path:     "api/users/[userId]/posts/[postId]/comments",
			expected: "api_users_userId_param_posts_postId_param_comments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &RouteNode{
				Path: tt.path,
			}
			alias := node.GetPackageAlias()
			assert.Equal(t, tt.expected, alias)
		})
	}
}
