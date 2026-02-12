package routing

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetModulePath tests extracting module path from go.mod
func TestGetModulePath(t *testing.T) {
	tests := []struct {
		name       string
		goModContent string
		expected   string
		wantError  bool
	}{
		{
			name: "simple module",
			goModContent: `module github.com/user/project

go 1.22
`,
			expected: "github.com/user/project",
			wantError: false,
		},
		{
			name: "module with comments",
			goModContent: `// Project go.mod
module github.com/cstone-io/twine

go 1.22

require (
    github.com/some/dep v1.0.0
)
`,
			expected: "github.com/cstone-io/twine",
			wantError: false,
		},
		{
			name: "module with extra whitespace",
			goModContent: `module     github.com/user/project

go 1.22
`,
			expected: "github.com/user/project",
			wantError: false,
		},
		{
			name: "module at end of file",
			goModContent: `go 1.22

require (
    github.com/some/dep v1.0.0
)

module github.com/user/project
`,
			expected: "github.com/user/project",
			wantError: false,
		},
		{
			name: "missing module declaration",
			goModContent: `go 1.22

require (
    github.com/some/dep v1.0.0
)
`,
			expected: "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			goModPath := filepath.Join(tmpDir, "go.mod")
			require.NoError(t, os.WriteFile(goModPath, []byte(tt.goModContent), 0644))

			modulePath, err := GetModulePath(tmpDir)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, modulePath)
			}
		})
	}
}

// TestGetModulePath_MissingFile tests error when go.mod doesn't exist
func TestGetModulePath_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := GetModulePath(tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading go.mod")
}

// TestGetRouterMethodName tests HTTP method to router method name conversion
func TestGetRouterMethodName(t *testing.T) {
	tests := []struct {
		httpMethod string
		expected   string
	}{
		{"GET", "Get"},
		{"POST", "Post"},
		{"PUT", "Put"},
		{"DELETE", "Delete"},
		{"PATCH", "Patch"},
		{"UNKNOWN", "UNKNOWN"}, // Fallback to uppercase
	}

	for _, tt := range tests {
		t.Run(tt.httpMethod, func(t *testing.T) {
			result := getRouterMethodName(tt.httpMethod)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCodeGenerator_CollectRoutes tests route collection from tree
func TestCodeGenerator_CollectRoutes(t *testing.T) {
	// Build simple tree
	root := &RouteNode{
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

	gen := &CodeGenerator{
		RouteTree: root,
	}

	routes := gen.collectRoutes(root)

	assert.Len(t, routes, 2)
	assert.Contains(t, routes, root.Children[0].Children[0])
	assert.Contains(t, routes, root.Children[0].Children[1])
}

// TestCodeGenerator_CollectRoutes_Nested tests nested route collection
func TestCodeGenerator_CollectRoutes_Nested(t *testing.T) {
	root := &RouteNode{
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
						Children: []*RouteNode{
							{
								Path:        "/app/pages/users/[id]",
								URLSegment:  "{id}",
								IsDynamic:   true,
								HandlerFile: "/app/pages/users/[id]/page.go",
								Methods:     []string{"GET"},
							},
						},
					},
				},
			},
		},
	}

	gen := &CodeGenerator{
		RouteTree: root,
	}

	routes := gen.collectRoutes(root)

	assert.Len(t, routes, 2)
}

// TestCodeGenerator_CollectRoutes_OnlyWithHandlers tests only nodes with handlers are collected
func TestCodeGenerator_CollectRoutes_OnlyWithHandlers(t *testing.T) {
	root := &RouteNode{
		Path:       "/app",
		URLSegment: "",
		Children: []*RouteNode{
			{
				Path:       "/app/pages",
				URLSegment: "pages",
				// No handler
				Children: []*RouteNode{
					{
						Path:       "/app/pages/users",
						URLSegment: "users",
						// No handler
						Children: []*RouteNode{
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

	gen := &CodeGenerator{
		RouteTree: root,
	}

	routes := gen.collectRoutes(root)

	// Should only collect the one node with handler
	assert.Len(t, routes, 1)
	assert.Equal(t, "/app/pages/users/[id]/page.go", routes[0].HandlerFile)
}

// TestCodeGenerator_CollectImports tests import collection
func TestCodeGenerator_CollectImports(t *testing.T) {
	routes := []*RouteNode{
		{
			Path:        "/app/pages/users",
			URLSegment:  "users",
			HandlerFile: "/app/pages/users/page.go",
			Methods:     []string{"GET"},
			Parent: &RouteNode{
				URLSegment: "pages",
			},
		},
		{
			Path:        "/app/pages/posts",
			URLSegment:  "posts",
			HandlerFile: "/app/pages/posts/page.go",
			Methods:     []string{"GET"},
			Parent: &RouteNode{
				URLSegment: "pages",
			},
		},
	}

	gen := &CodeGenerator{
		ModulePath:  "github.com/user/project",
		ProjectRoot: "/project",
	}

	imports := gen.collectImports(routes)

	assert.NotEmpty(t, imports)
	// Should have imports for both handlers
	assert.Len(t, imports, 2)
}

// TestCodeGenerator_CollectImports_WithLayouts tests import collection with layouts
func TestCodeGenerator_CollectImports_WithLayouts(t *testing.T) {
	parent := &RouteNode{
		Path:       "/app/pages",
		URLSegment: "pages",
		LayoutFile: "/app/pages/layout.go",
		HasLayout:  true,
	}

	routes := []*RouteNode{
		{
			Path:        "/app/pages/users",
			URLSegment:  "users",
			HandlerFile: "/app/pages/users/page.go",
			Methods:     []string{"GET"},
			Parent:      parent,
		},
	}

	gen := &CodeGenerator{
		ModulePath:  "github.com/user/project",
		ProjectRoot: "/app",
	}

	imports := gen.collectImports(routes)

	// Should include both handler and layout imports
	assert.GreaterOrEqual(t, len(imports), 2)
}

// TestCodeGenerator_CollectImports_Deduplication tests import alias deduplication
func TestCodeGenerator_CollectImports_Deduplication(t *testing.T) {
	// Create routes with same package alias potential
	parent := &RouteNode{
		URLSegment: "pages",
	}

	routes := []*RouteNode{
		{
			Path:        "/app/pages/users",
			URLSegment:  "users",
			HandlerFile: "/app/pages/users/page.go",
			Methods:     []string{"GET"},
			Parent:      parent,
		},
		{
			Path:        "/app/pages/users",
			URLSegment:  "users",
			HandlerFile: "/app/pages/users/page.go",
			Methods:     []string{"POST"},
			Parent:      parent,
		},
	}

	gen := &CodeGenerator{
		ModulePath:  "github.com/user/project",
		ProjectRoot: "/app",
	}

	imports := gen.collectImports(routes)

	// Should deduplicate - only one import for same path
	aliases := make(map[string]bool)
	for alias := range imports {
		assert.False(t, aliases[alias], "duplicate alias: %s", alias)
		aliases[alias] = true
	}
}

// TestCodeGenerator_Generate_ValidGoCode tests generated code is valid Go
func TestCodeGenerator_Generate_ValidGoCode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod
	goModContent := `module github.com/user/testproject

go 1.22
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644))

	// Create simple route tree
	root := &RouteNode{
		Path:       filepath.Join(tmpDir, "app"),
		URLSegment: "",
		Children: []*RouteNode{
			{
				Path:       filepath.Join(tmpDir, "app/pages"),
				URLSegment: "pages",
				Children: []*RouteNode{
					{
						Path:        filepath.Join(tmpDir, "app/pages/index"),
						URLSegment:  "index",
						HandlerFile: filepath.Join(tmpDir, "app/pages/index/page.go"),
						Methods:     []string{"GET"},
						PackageName: "index",
						Parent: &RouteNode{
							URLSegment: "pages",
						},
					},
				},
			},
		},
	}

	outputFile := filepath.Join(tmpDir, "routes.gen.go")

	gen := &CodeGenerator{
		RouteTree:   root,
		ModulePath:  "github.com/user/testproject",
		ProjectRoot: tmpDir,
		OutputFile:  outputFile,
	}

	err := gen.Generate()
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, outputFile)

	// Read generated code
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	// Verify it's valid Go code
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, outputFile, content, 0)
	assert.NoError(t, err, "Generated code should be valid Go")

	// Verify content contains expected elements
	code := string(content)
	assert.Contains(t, code, "package app")
	assert.Contains(t, code, "func RegisterRoutes(r *router.Router)")
	assert.Contains(t, code, "github.com/cstone-io/twine/kit")
	assert.Contains(t, code, "github.com/cstone-io/twine/router")
}

// TestCodeGenerator_Generate_WithMultipleRoutes tests generation with multiple routes
func TestCodeGenerator_Generate_WithMultipleRoutes(t *testing.T) {
	tmpDir := t.TempDir()

	goModContent := `module github.com/user/testproject

go 1.22
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644))

	pagesNode := &RouteNode{
		Path:       filepath.Join(tmpDir, "app/pages"),
		URLSegment: "pages",
		Parent:     nil,
	}

	root := &RouteNode{
		Path:       filepath.Join(tmpDir, "app"),
		URLSegment: "",
		Children: []*RouteNode{
			{
				Path:       filepath.Join(tmpDir, "app/pages"),
				URLSegment: "pages",
				Children: []*RouteNode{
					{
						Path:        filepath.Join(tmpDir, "app/pages/users"),
						URLSegment:  "users",
						HandlerFile: filepath.Join(tmpDir, "app/pages/users/page.go"),
						Methods:     []string{"GET", "POST"},
						PackageName: "users",
						Parent:      pagesNode,
					},
					{
						Path:        filepath.Join(tmpDir, "app/pages/posts"),
						URLSegment:  "posts",
						HandlerFile: filepath.Join(tmpDir, "app/pages/posts/page.go"),
						Methods:     []string{"GET"},
						PackageName: "posts",
						Parent:      pagesNode,
					},
				},
			},
		},
	}

	outputFile := filepath.Join(tmpDir, "routes.gen.go")

	gen := &CodeGenerator{
		RouteTree:   root,
		ModulePath:  "github.com/user/testproject",
		ProjectRoot: tmpDir,
		OutputFile:  outputFile,
	}

	err := gen.Generate()
	require.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	code := string(content)

	// Verify multiple route registrations
	assert.Contains(t, code, "r.Get")
	assert.Contains(t, code, "r.Post")
	assert.Contains(t, code, "/users")
	assert.Contains(t, code, "/posts")
}

// TestCodeGenerator_Generate_WithAPIRoutes tests API route generation
func TestCodeGenerator_Generate_WithAPIRoutes(t *testing.T) {
	tmpDir := t.TempDir()

	goModContent := `module github.com/user/testproject

go 1.22
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644))

	apiNode := &RouteNode{
		Path:       filepath.Join(tmpDir, "app/api"),
		URLSegment: "api",
		Parent:     nil,
	}

	root := &RouteNode{
		Path:       filepath.Join(tmpDir, "app"),
		URLSegment: "",
		Children: []*RouteNode{
			{
				Path:       filepath.Join(tmpDir, "app/api"),
				URLSegment: "api",
				Children: []*RouteNode{
					{
						Path:        filepath.Join(tmpDir, "app/api/users"),
						URLSegment:  "users",
						HandlerFile: filepath.Join(tmpDir, "app/api/users/route.go"),
						IsAPI:       true,
						Methods:     []string{"GET", "POST"},
						PackageName: "users",
						Parent:      apiNode,
					},
				},
			},
		},
	}

	outputFile := filepath.Join(tmpDir, "routes.gen.go")

	gen := &CodeGenerator{
		RouteTree:   root,
		ModulePath:  "github.com/user/testproject",
		ProjectRoot: tmpDir,
		OutputFile:  outputFile,
	}

	err := gen.Generate()
	require.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	code := string(content)

	// Verify API route registration
	assert.Contains(t, code, "/api/users")
	assert.Contains(t, code, "// API routes")
}

// TestCodeGenerator_Generate_WithLayouts tests layout middleware generation
func TestCodeGenerator_Generate_WithLayouts(t *testing.T) {
	tmpDir := t.TempDir()

	goModContent := `module github.com/user/testproject

go 1.22
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644))

	pagesNode := &RouteNode{
		Path:       filepath.Join(tmpDir, "app/pages"),
		URLSegment: "pages",
		LayoutFile: filepath.Join(tmpDir, "app/pages/layout.go"),
		HasLayout:  true,
		Parent:     nil,
	}

	root := &RouteNode{
		Path:       filepath.Join(tmpDir, "app"),
		URLSegment: "",
		Children: []*RouteNode{
			{
				Path:       filepath.Join(tmpDir, "app/pages"),
				URLSegment: "pages",
				LayoutFile: filepath.Join(tmpDir, "app/pages/layout.go"),
				HasLayout:  true,
				Children: []*RouteNode{
					{
						Path:        filepath.Join(tmpDir, "app/pages/dashboard"),
						URLSegment:  "dashboard",
						HandlerFile: filepath.Join(tmpDir, "app/pages/dashboard/page.go"),
						Methods:     []string{"GET"},
						PackageName: "dashboard",
						Parent:      pagesNode,
					},
				},
			},
		},
	}

	outputFile := filepath.Join(tmpDir, "routes.gen.go")

	gen := &CodeGenerator{
		RouteTree:   root,
		ModulePath:  "github.com/user/testproject",
		ProjectRoot: tmpDir,
		OutputFile:  outputFile,
	}

	err := gen.Generate()
	require.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	code := string(content)

	// Verify layout middleware is applied
	assert.Contains(t, code, "applyMiddleware")
	assert.Contains(t, code, "middleware.Middleware")
	assert.Contains(t, code, ".Layout()")
}

// TestCodeGenerator_Generate_WithDynamicRoutes tests dynamic route generation
func TestCodeGenerator_Generate_WithDynamicRoutes(t *testing.T) {
	tmpDir := t.TempDir()

	goModContent := `module github.com/user/testproject

go 1.22
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644))

	usersNode := &RouteNode{
		Path:       filepath.Join(tmpDir, "app/pages/users"),
		URLSegment: "users",
		Parent:     nil,
	}

	root := &RouteNode{
		Path:       filepath.Join(tmpDir, "app"),
		URLSegment: "",
		Children: []*RouteNode{
			{
				Path:       filepath.Join(tmpDir, "app/pages"),
				URLSegment: "pages",
				Children: []*RouteNode{
					{
						Path:       filepath.Join(tmpDir, "app/pages/users"),
						URLSegment: "users",
						Children: []*RouteNode{
							{
								Path:        filepath.Join(tmpDir, "app/pages/users/[id]"),
								URLSegment:  "{id}",
								IsDynamic:   true,
								ParamName:   "id",
								HandlerFile: filepath.Join(tmpDir, "app/pages/users/[id]/page.go"),
								Methods:     []string{"GET"},
								PackageName: "user_id",
								Parent:      usersNode,
							},
						},
					},
				},
			},
		},
	}

	outputFile := filepath.Join(tmpDir, "routes.gen.go")

	gen := &CodeGenerator{
		RouteTree:   root,
		ModulePath:  "github.com/user/testproject",
		ProjectRoot: tmpDir,
		OutputFile:  outputFile,
	}

	err := gen.Generate()
	require.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	code := string(content)

	// Verify dynamic route parameter
	assert.Contains(t, code, "{id}")
	assert.Contains(t, code, "/users/{id}")
}

// TestCodeGenerator_GenerateCode_Header tests generated code header
func TestCodeGenerator_GenerateCode_Header(t *testing.T) {
	gen := &CodeGenerator{
		RouteTree: &RouteNode{
			Path:     "/app",
			Children: []*RouteNode{},
		},
		ModulePath: "github.com/user/project",
	}

	routes := []*RouteNode{}
	code := gen.generateCode(routes)

	assert.Contains(t, code, "// Code generated by twine routes generate. DO NOT EDIT.")
	assert.Contains(t, code, "package app")
}

// TestCodeGenerator_GenerateCode_Imports tests import generation
func TestCodeGenerator_GenerateCode_Imports(t *testing.T) {
	pagesNode := &RouteNode{
		Path:       "/app/pages",
		URLSegment: "pages",
	}

	gen := &CodeGenerator{
		RouteTree:   &RouteNode{Path: "/app"},
		ModulePath:  "github.com/user/project",
		ProjectRoot: "/",
	}

	routes := []*RouteNode{
		{
			Path:        "/app/pages/users",
			URLSegment:  "users",
			HandlerFile: "/app/pages/users/page.go",
			Methods:     []string{"GET"},
			Parent:      pagesNode,
		},
	}

	code := gen.generateCode(routes)

	// Verify standard imports
	assert.Contains(t, code, `"github.com/cstone-io/twine/kit"`)
	assert.Contains(t, code, `"github.com/cstone-io/twine/router"`)
	assert.Contains(t, code, `"github.com/cstone-io/twine/middleware"`)
}

// TestCodeGenerator_GenerateCode_ApplyMiddleware tests middleware helper function
func TestCodeGenerator_GenerateCode_ApplyMiddleware(t *testing.T) {
	gen := &CodeGenerator{
		RouteTree:  &RouteNode{Path: "/app"},
		ModulePath: "github.com/user/project",
	}

	routes := []*RouteNode{}
	code := gen.generateCode(routes)

	assert.Contains(t, code, "func applyMiddleware(middlewares []middleware.Middleware, handler kit.HandlerFunc) kit.HandlerFunc")
	assert.Contains(t, code, "middleware.ApplyMiddlewares(handler, middlewares...)")
}

// TestCodeGenerator_GenerateCode_RegisterRoutes tests RegisterRoutes function
func TestCodeGenerator_GenerateCode_RegisterRoutes(t *testing.T) {
	gen := &CodeGenerator{
		RouteTree:  &RouteNode{Path: "/app"},
		ModulePath: "github.com/user/project",
	}

	routes := []*RouteNode{}
	code := gen.generateCode(routes)

	assert.Contains(t, code, "func RegisterRoutes(r *router.Router)")
}

// TestCodeGenerator_Generate_EmptyTree tests generation with no routes
func TestCodeGenerator_Generate_EmptyTree(t *testing.T) {
	tmpDir := t.TempDir()

	goModContent := `module github.com/user/testproject

go 1.22
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644))

	root := &RouteNode{
		Path:       filepath.Join(tmpDir, "app"),
		URLSegment: "",
		Children:   []*RouteNode{},
	}

	outputFile := filepath.Join(tmpDir, "routes.gen.go")

	gen := &CodeGenerator{
		RouteTree:   root,
		ModulePath:  "github.com/user/testproject",
		ProjectRoot: tmpDir,
		OutputFile:  outputFile,
	}

	err := gen.Generate()
	require.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	// Verify valid Go even with no routes
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, outputFile, content, 0)
	assert.NoError(t, err)
}

// TestCodeGenerator_BuildLayoutChain tests layout chain building
func TestCodeGenerator_BuildLayoutChain(t *testing.T) {
	parent := &RouteNode{
		Path:       "/app/pages",
		URLSegment: "pages",
		LayoutFile: "/app/pages/layout.go",
		HasLayout:  true,
		Parent:     nil,
	}

	node := &RouteNode{
		Path:        "/app/pages/users",
		URLSegment:  "users",
		HandlerFile: "/app/pages/users/page.go",
		Parent:      parent,
	}

	gen := &CodeGenerator{
		ModulePath:  "github.com/user/project",
		ProjectRoot: "/app",
	}

	chain := gen.buildLayoutChain(node)

	assert.NotNil(t, chain)
	assert.Len(t, chain.Layouts, 1)
	assert.Equal(t, "/app/pages/layout.go", chain.Layouts[0].FilePath)
}

// TestCodeGenerator_GetPackagePath tests package path generation
func TestCodeGenerator_GetPackagePath(t *testing.T) {
	tests := []struct {
		name        string
		projectRoot string
		nodePath    string
		modulePath  string
		expected    string
	}{
		{
			name:        "simple path",
			projectRoot: "/project",
			nodePath:    "/project/app/pages/users",
			modulePath:  "github.com/user/project",
			expected:    "github.com/user/project/app/pages/users",
		},
		{
			name:        "with dynamic segment",
			projectRoot: "/project",
			nodePath:    "/project/app/pages/users/[id]",
			modulePath:  "github.com/user/project",
			expected:    "github.com/user/project/app/pages/users/[id]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &CodeGenerator{
				ModulePath:  tt.modulePath,
				ProjectRoot: tt.projectRoot,
			}

			node := &RouteNode{
				Path: tt.nodePath,
			}

			packagePath := gen.getPackagePath(node)
			assert.Equal(t, tt.expected, packagePath)
		})
	}
}

// TestCodeGenerator_Generate_SortedRoutes tests routes are sorted in output
func TestCodeGenerator_Generate_SortedRoutes(t *testing.T) {
	tmpDir := t.TempDir()

	goModContent := `module github.com/user/testproject

go 1.22
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644))

	pagesNode := &RouteNode{
		Path:       filepath.Join(tmpDir, "app/pages"),
		URLSegment: "pages",
		Parent:     nil,
	}

	root := &RouteNode{
		Path:       filepath.Join(tmpDir, "app"),
		URLSegment: "",
		Children: []*RouteNode{
			{
				Path:       filepath.Join(tmpDir, "app/pages"),
				URLSegment: "pages",
				Children: []*RouteNode{
					{
						Path:        filepath.Join(tmpDir, "app/pages/zebra"),
						URLSegment:  "zebra",
						HandlerFile: filepath.Join(tmpDir, "app/pages/zebra/page.go"),
						Methods:     []string{"GET"},
						PackageName: "zebra",
						Parent:      pagesNode,
					},
					{
						Path:        filepath.Join(tmpDir, "app/pages/apple"),
						URLSegment:  "apple",
						HandlerFile: filepath.Join(tmpDir, "app/pages/apple/page.go"),
						Methods:     []string{"GET"},
						PackageName: "apple",
						Parent:      pagesNode,
					},
				},
			},
		},
	}

	outputFile := filepath.Join(tmpDir, "routes.gen.go")

	gen := &CodeGenerator{
		RouteTree:   root,
		ModulePath:  "github.com/user/testproject",
		ProjectRoot: tmpDir,
		OutputFile:  outputFile,
	}

	err := gen.Generate()
	require.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	code := string(content)

	// Find positions of routes in generated code
	applePos := strings.Index(code, "/apple")
	zebraPos := strings.Index(code, "/zebra")

	// apple should come before zebra (alphabetically sorted)
	assert.Less(t, applePos, zebraPos, "Routes should be sorted alphabetically")
}
