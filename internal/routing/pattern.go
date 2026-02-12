package routing

import (
	"path/filepath"
	"strings"
)

// ToURLPattern converts RouteNode to Go 1.22+ ServeMux pattern
func (n *RouteNode) ToURLPattern() string {
	path := n.GetFullPath()
	if path == "" {
		return "/"
	}
	return path
}

// GetFullPath returns complete URL path from root
func (n *RouteNode) GetFullPath() string {
	segments := make([]string, 0)
	current := n

	// Walk up to root, collecting segments
	for current != nil && current.URLSegment != "" {
		// Skip the root "pages" or "api" segment for pages
		// Include "api" in the path for API routes
		if current.URLSegment == "pages" {
			current = current.Parent
			continue
		}
		segments = append([]string{current.URLSegment}, segments...)
		current = current.Parent
	}

	if len(segments) == 0 {
		return ""
	}

	path := "/" + strings.Join(segments, "/")
	return path
}

// GetPackagePath returns Go import path for handler package
func (n *RouteNode) GetPackagePath(modulePath string) string {
	// Get relative path from project root
	relPath := strings.TrimPrefix(n.Path, "/")

	// Sanitize dynamic segments in path
	parts := strings.Split(relPath, string(filepath.Separator))
	for i, part := range parts {
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			parts[i] = SanitizePackageName(part)
		}
	}

	sanitizedPath := strings.Join(parts, "/")
	return modulePath + "/" + sanitizedPath
}

// SanitizePackageName converts dynamic directory names to valid Go package names
func SanitizePackageName(dirName string) string {
	// Remove brackets
	name := strings.TrimSuffix(strings.TrimPrefix(dirName, "["), "]")

	// Handle catch-all
	if strings.HasPrefix(name, "...") {
		name = strings.TrimPrefix(name, "...")
		return name + "_catchall"
	}

	// Handle dynamic param
	return name + "_param"
}

// GetPackageAlias returns a unique package alias for imports
func (n *RouteNode) GetPackageAlias() string {
	// Build alias from path segments
	parts := strings.Split(n.Path, string(filepath.Separator))

	// Filter out empty parts and common prefixes
	filtered := make([]string, 0)
	for _, part := range parts {
		if part == "" || part == "app" {
			continue
		}
		// Sanitize dynamic segments
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			part = SanitizePackageName(part)
		}
		// Replace invalid characters with underscores
		part = strings.ReplaceAll(part, "-", "_")
		part = strings.ReplaceAll(part, ".", "_")
		filtered = append(filtered, part)
	}

	if len(filtered) == 0 {
		return "root"
	}

	return strings.Join(filtered, "_")
}
