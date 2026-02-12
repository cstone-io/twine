package routing

import (
	"path/filepath"
)

// BuildLayoutChain walks from node to root collecting layout.go files
func BuildLayoutChain(node *RouteNode, modulePath string) *LayoutChain {
	chain := &LayoutChain{
		Layouts: make([]LayoutInfo, 0),
	}

	current := node
	for current != nil {
		if current.HasLayout {
			layout := LayoutInfo{
				FilePath:    current.LayoutFile,
				PackagePath: current.GetPackagePath(modulePath),
				PackageName: current.GetPackageAlias(),
				FuncName:    "Layout",
			}
			// Prepend to maintain order from root to leaf
			chain.Layouts = append([]LayoutInfo{layout}, chain.Layouts...)
		}
		current = current.Parent
	}

	return chain
}

// HasLayouts returns true if the chain contains any layouts
func (c *LayoutChain) HasLayouts() bool {
	return len(c.Layouts) > 0
}

// GetLayoutDir returns the directory containing the layout file
func (l *LayoutInfo) GetLayoutDir() string {
	return filepath.Dir(l.FilePath)
}
