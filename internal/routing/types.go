package routing

// RouteNode represents a node in the file-based routing tree
type RouteNode struct {
	Path       string       // Filesystem path (e.g., "app/pages/users")
	URLSegment string       // URL segment (e.g., "users" or "{id}")
	Children   []*RouteNode // Child nodes
	Parent     *RouteNode   // Parent node (for layout chain)

	// File detection
	HandlerFile string // "page.go" or "route.go" (full path)
	LayoutFile  string // "layout.go" (full path)

	// Handler metadata
	Methods     []string // ["GET", "POST"] - detected from exports
	PackageName string   // Go package name for this directory

	// Route type detection
	IsDirectory bool // Just a directory (no handler)
	IsPage      bool // page.go found
	IsAPI       bool // route.go found
	HasLayout   bool // layout.go found

	// Dynamic route handling
	IsDynamic  bool   // [param] style
	IsCatchAll bool   // [...param] style
	ParamName  string // "param" extracted from [param] or [...param]
}

// LayoutChain represents an ordered chain of layout middleware
type LayoutChain struct {
	Layouts []LayoutInfo // Ordered from outermost (root) to innermost (leaf)
}

// LayoutInfo contains information about a single layout in the chain
type LayoutInfo struct {
	FilePath    string // Filesystem path to layout.go
	PackagePath string // Go import path
	PackageName string // Package identifier for imports
	FuncName    string // "Layout" (function name to call)
}
