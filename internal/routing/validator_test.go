package routing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateParamName tests parameter name validation
func TestValidateParamName(t *testing.T) {
	tests := []struct {
		name      string
		paramName string
		wantError bool
		errorMsg  string
	}{
		// Valid names
		{
			name:      "simple letter",
			paramName: "id",
			wantError: false,
		},
		{
			name:      "camelCase",
			paramName: "userId",
			wantError: false,
		},
		{
			name:      "snake_case",
			paramName: "user_id",
			wantError: false,
		},
		{
			name:      "starts with underscore",
			paramName: "_id",
			wantError: false,
		},
		{
			name:      "uppercase",
			paramName: "ID",
			wantError: false,
		},
		{
			name:      "with numbers",
			paramName: "user123",
			wantError: false,
		},
		{
			name:      "multiple underscores",
			paramName: "user__id",
			wantError: false,
		},
		{
			name:      "long name",
			paramName: "veryLongParameterName",
			wantError: false,
		},

		// Invalid names
		{
			name:      "empty string",
			paramName: "",
			wantError: true,
			errorMsg:  "parameter name cannot be empty",
		},
		{
			name:      "starts with number",
			paramName: "123id",
			wantError: true,
			errorMsg:  "parameter name must start with letter or underscore",
		},
		{
			name:      "contains dash",
			paramName: "user-id",
			wantError: true,
			errorMsg:  "parameter name contains invalid character",
		},
		{
			name:      "contains space",
			paramName: "user id",
			wantError: true,
			errorMsg:  "parameter name contains invalid character",
		},
		{
			name:      "contains dot",
			paramName: "user.id",
			wantError: true,
			errorMsg:  "parameter name contains invalid character",
		},
		{
			name:      "contains special char",
			paramName: "user@id",
			wantError: true,
			errorMsg:  "parameter name contains invalid character",
		},
		{
			name:      "starts with special char",
			paramName: "$id",
			wantError: true,
			errorMsg:  "parameter name must start with letter or underscore",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateParamName(tt.paramName)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRouteNode_ValidateNode tests single node validation
func TestRouteNode_ValidateNode(t *testing.T) {
	tests := []struct {
		name      string
		node      *RouteNode
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid static route",
			node: &RouteNode{
				Path:        "/app/pages/users",
				URLSegment:  "users",
				HandlerFile: "/app/pages/users/page.go",
				Methods:     []string{"GET", "POST"},
			},
			wantError: false,
		},
		{
			name: "valid dynamic route",
			node: &RouteNode{
				Path:        "/app/pages/users/[id]",
				URLSegment:  "{id}",
				IsDynamic:   true,
				ParamName:   "id",
				HandlerFile: "/app/pages/users/[id]/page.go",
				Methods:     []string{"GET"},
			},
			wantError: false,
		},
		{
			name: "invalid param name",
			node: &RouteNode{
				Path:       "/app/pages/users/[user-id]",
				URLSegment: "{user-id}",
				IsDynamic:  true,
				ParamName:  "user-id",
			},
			wantError: true,
			errorMsg:  "parameter name contains invalid character",
		},
		{
			name: "handler without methods",
			node: &RouteNode{
				Path:        "/app/pages/users",
				HandlerFile: "/app/pages/users/page.go",
				Methods:     []string{},
			},
			wantError: true,
			errorMsg:  "handler file must export at least one HTTP method function",
		},
		{
			name: "catch-all with handler children",
			node: &RouteNode{
				Path:       "/app/pages/[...slug]",
				URLSegment: "{slug...}",
				IsCatchAll: true,
				IsDynamic:  true,
				ParamName:  "slug",
				Children: []*RouteNode{
					{
						HandlerFile: "/app/pages/[...slug]/child/page.go",
						Methods:     []string{"GET"},
					},
				},
			},
			wantError: true,
			errorMsg:  "catch-all segment must be the last segment in the route",
		},
		{
			name: "catch-all without handler children (allowed)",
			node: &RouteNode{
				Path:       "/app/pages/[...slug]",
				URLSegment: "{slug...}",
				IsCatchAll: true,
				IsDynamic:  true,
				ParamName:  "slug",
				Children: []*RouteNode{
					{
						// No handler file
						IsDirectory: true,
					},
				},
			},
			wantError: false,
		},
		{
			name: "directory without handler (valid)",
			node: &RouteNode{
				Path:        "/app/pages/users",
				URLSegment:  "users",
				IsDirectory: true,
			},
			wantError: false,
		},
		{
			name: "layout without handler (valid)",
			node: &RouteNode{
				Path:       "/app/pages/dashboard",
				URLSegment: "dashboard",
				LayoutFile: "/app/pages/dashboard/layout.go",
				HasLayout:  true,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.node.validateNode()

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRouteNode_CheckConflicts tests conflict detection
func TestRouteNode_CheckConflicts(t *testing.T) {
	tests := []struct {
		name      string
		parent    *RouteNode
		wantError bool
		errorMsg  string
	}{
		{
			name: "no conflicts - different static routes",
			parent: &RouteNode{
				Path: "/app/pages",
				Children: []*RouteNode{
					{
						URLSegment:  "users",
						HandlerFile: "/app/pages/users/page.go",
						Methods:     []string{"GET"},
					},
					{
						URLSegment:  "posts",
						HandlerFile: "/app/pages/posts/page.go",
						Methods:     []string{"GET"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "no conflicts - static and dynamic",
			parent: &RouteNode{
				Path: "/app/pages/users",
				Children: []*RouteNode{
					{
						URLSegment:  "new",
						HandlerFile: "/app/pages/users/new/page.go",
						Methods:     []string{"GET"},
					},
					{
						URLSegment:  "{id}",
						IsDynamic:   true,
						ParamName:   "id",
						HandlerFile: "/app/pages/users/[id]/page.go",
						Methods:     []string{"GET"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "duplicate static routes",
			parent: &RouteNode{
				Path: "/app/pages",
				Children: []*RouteNode{
					{
						URLSegment:  "users",
						HandlerFile: "/app/pages/users/page.go",
						Methods:     []string{"GET"},
					},
					{
						URLSegment:  "users",
						HandlerFile: "/app/pages/users2/page.go",
						Methods:     []string{"POST"},
					},
				},
			},
			wantError: true,
			errorMsg:  "duplicate route",
		},
		{
			name: "multiple catch-all routes",
			parent: &RouteNode{
				Path: "/app/pages",
				Children: []*RouteNode{
					{
						URLSegment:  "{slug...}",
						IsCatchAll:  true,
						IsDynamic:   true,
						ParamName:   "slug",
						HandlerFile: "/app/pages/[...slug]/page.go",
						Methods:     []string{"GET"},
					},
					{
						URLSegment:  "{path...}",
						IsCatchAll:  true,
						IsDynamic:   true,
						ParamName:   "path",
						HandlerFile: "/app/pages/[...path]/page.go",
						Methods:     []string{"GET"},
					},
				},
			},
			wantError: true,
			errorMsg:  "multiple catch-all routes at same level",
		},
		{
			name: "single catch-all (allowed)",
			parent: &RouteNode{
				Path: "/app/pages",
				Children: []*RouteNode{
					{
						URLSegment:  "{slug...}",
						IsCatchAll:  true,
						IsDynamic:   true,
						ParamName:   "slug",
						HandlerFile: "/app/pages/[...slug]/page.go",
						Methods:     []string{"GET"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "directories without handlers (no conflict)",
			parent: &RouteNode{
				Path: "/app/pages",
				Children: []*RouteNode{
					{
						URLSegment:  "users",
						IsDirectory: true,
						// No handler file
					},
					{
						URLSegment:  "posts",
						IsDirectory: true,
						// No handler file
					},
				},
			},
			wantError: false,
		},
		{
			name: "layout without handler (no conflict)",
			parent: &RouteNode{
				Path: "/app/pages",
				Children: []*RouteNode{
					{
						URLSegment: "dashboard",
						LayoutFile: "/app/pages/dashboard/layout.go",
						HasLayout:  true,
					},
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.parent.checkConflicts()

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRouteNode_Validate tests full tree validation
func TestRouteNode_Validate(t *testing.T) {
	tests := []struct {
		name      string
		tree      *RouteNode
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid simple tree",
			tree: &RouteNode{
				Path:       "/app",
				URLSegment: "",
				Children: []*RouteNode{
					{
						Path:       "/app/pages",
						URLSegment: "pages",
						Children: []*RouteNode{
							{
								Path:        "/app/pages/users",
								URLSegment:  "users",
								HandlerFile: "/app/pages/users/page.go",
								Methods:     []string{"GET"},
							},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid complex tree",
			tree: &RouteNode{
				Path:       "/app",
				URLSegment: "",
				Children: []*RouteNode{
					{
						Path:       "/app/pages",
						URLSegment: "pages",
						Children: []*RouteNode{
							{
								Path:        "/app/pages/users",
								URLSegment:  "users",
								HandlerFile: "/app/pages/users/page.go",
								Methods:     []string{"GET", "POST"},
								Children: []*RouteNode{
									{
										Path:        "/app/pages/users/[id]",
										URLSegment:  "{id}",
										IsDynamic:   true,
										ParamName:   "id",
										HandlerFile: "/app/pages/users/[id]/page.go",
										Methods:     []string{"GET"},
									},
								},
							},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "invalid param name in child",
			tree: &RouteNode{
				Path:       "/app",
				URLSegment: "",
				Children: []*RouteNode{
					{
						Path:       "/app/pages",
						URLSegment: "pages",
						Children: []*RouteNode{
							{
								Path:       "/app/pages/users/[user-id]",
								URLSegment: "{user-id}",
								IsDynamic:  true,
								ParamName:  "user-id",
							},
						},
					},
				},
			},
			wantError: true,
			errorMsg:  "parameter name contains invalid character",
		},
		{
			name: "duplicate routes in tree",
			tree: &RouteNode{
				Path:       "/app",
				URLSegment: "",
				Children: []*RouteNode{
					{
						Path:       "/app/pages",
						URLSegment: "pages",
						Children: []*RouteNode{
							{
								URLSegment:  "users",
								HandlerFile: "/app/pages/users/page.go",
								Methods:     []string{"GET"},
							},
							{
								URLSegment:  "users",
								HandlerFile: "/app/pages/users2/page.go",
								Methods:     []string{"POST"},
							},
						},
					},
				},
			},
			wantError: true,
			errorMsg:  "duplicate route",
		},
		{
			name: "catch-all not at end",
			tree: &RouteNode{
				Path:       "/app",
				URLSegment: "",
				Children: []*RouteNode{
					{
						Path:       "/app/pages",
						URLSegment: "pages",
						Children: []*RouteNode{
							{
								Path:       "/app/pages/[...slug]",
								URLSegment: "{slug...}",
								IsCatchAll: true,
								IsDynamic:  true,
								ParamName:  "slug",
								Children: []*RouteNode{
									{
										URLSegment:  "extra",
										HandlerFile: "/app/pages/[...slug]/extra/page.go",
										Methods:     []string{"GET"},
									},
								},
							},
						},
					},
				},
			},
			wantError: true,
			errorMsg:  "catch-all segment must be the last segment",
		},
		{
			name: "handler without methods",
			tree: &RouteNode{
				Path:       "/app",
				URLSegment: "",
				Children: []*RouteNode{
					{
						Path:       "/app/pages",
						URLSegment: "pages",
						Children: []*RouteNode{
							{
								Path:        "/app/pages/users",
								URLSegment:  "users",
								HandlerFile: "/app/pages/users/page.go",
								Methods:     []string{}, // Empty methods
							},
						},
					},
				},
			},
			wantError: true,
			errorMsg:  "handler file must export at least one HTTP method function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tree.Validate()

			if tt.wantError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRouteNode_Validate_EmptyTree tests validation of empty tree
func TestRouteNode_Validate_EmptyTree(t *testing.T) {
	tree := &RouteNode{
		Path:       "/app",
		URLSegment: "",
		Children:   []*RouteNode{},
	}

	err := tree.Validate()
	assert.NoError(t, err)
}

// TestRouteNode_Validate_DeepNesting tests deeply nested validation
func TestRouteNode_Validate_DeepNesting(t *testing.T) {
	// Build deeply nested tree
	root := &RouteNode{
		Path:       "/app",
		URLSegment: "",
	}

	pages := &RouteNode{
		Path:       "/app/pages",
		URLSegment: "pages",
		Parent:     root,
	}
	root.Children = []*RouteNode{pages}

	level1 := &RouteNode{
		Path:        "/app/pages/level1",
		URLSegment:  "level1",
		Parent:      pages,
		HandlerFile: "/app/pages/level1/page.go",
		Methods:     []string{"GET"},
	}
	pages.Children = []*RouteNode{level1}

	level2 := &RouteNode{
		Path:        "/app/pages/level1/level2",
		URLSegment:  "level2",
		Parent:      level1,
		HandlerFile: "/app/pages/level1/level2/page.go",
		Methods:     []string{"GET"},
	}
	level1.Children = []*RouteNode{level2}

	level3 := &RouteNode{
		Path:        "/app/pages/level1/level2/level3",
		URLSegment:  "level3",
		Parent:      level2,
		HandlerFile: "/app/pages/level1/level2/level3/page.go",
		Methods:     []string{"GET"},
	}
	level2.Children = []*RouteNode{level3}

	err := root.Validate()
	assert.NoError(t, err)
}

// TestRouteNode_Validate_MixedValidAndInvalid tests tree with some valid and invalid nodes
func TestRouteNode_Validate_MixedValidAndInvalid(t *testing.T) {
	tree := &RouteNode{
		Path:       "/app",
		URLSegment: "",
		Children: []*RouteNode{
			{
				Path:       "/app/pages",
				URLSegment: "pages",
				Children: []*RouteNode{
					{
						// Valid
						URLSegment:  "users",
						HandlerFile: "/app/pages/users/page.go",
						Methods:     []string{"GET"},
					},
					{
						// Invalid - empty methods
						URLSegment:  "posts",
						HandlerFile: "/app/pages/posts/page.go",
						Methods:     []string{},
					},
				},
			},
		},
	}

	err := tree.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler file must export at least one HTTP method function")
}

// TestValidateParamName_UnicodeCharacters tests unicode in param names
func TestValidateParamName_UnicodeCharacters(t *testing.T) {
	tests := []struct {
		name      string
		paramName string
		wantError bool
	}{
		{
			name:      "ASCII letters",
			paramName: "userId",
			wantError: false,
		},
		{
			name:      "unicode letters (allowed by unicode.IsLetter)",
			paramName: "用户ID", // Chinese characters
			wantError: false,
		},
		{
			name:      "unicode with underscore",
			paramName: "user_名前",
			wantError: false,
		},
		{
			name:      "emoji (not a letter)",
			paramName: "user😀",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateParamName(tt.paramName)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
