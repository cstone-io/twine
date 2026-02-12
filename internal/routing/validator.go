package routing

import (
	"fmt"
	"unicode"
)

// Validate checks the route tree for conflicts and invalid configurations
func (n *RouteNode) Validate() error {
	// Validate this node
	if err := n.validateNode(); err != nil {
		return err
	}

	// Recursively validate children
	for _, child := range n.Children {
		if err := child.Validate(); err != nil {
			return err
		}
	}

	// Check for route conflicts among children
	if err := n.checkConflicts(); err != nil {
		return err
	}

	return nil
}

func (n *RouteNode) validateNode() error {
	// Validate dynamic segment names
	if n.IsDynamic {
		if err := validateParamName(n.ParamName); err != nil {
			return fmt.Errorf("%s: %w", n.Path, err)
		}
	}

	// Validate catch-all is last segment
	if n.IsCatchAll {
		if len(n.Children) > 0 {
			// Check if any children have handlers
			for _, child := range n.Children {
				if child.HandlerFile != "" {
					return fmt.Errorf("%s: catch-all segment must be the last segment in the route", n.Path)
				}
			}
		}
	}

	// Validate handler has at least one method
	if n.HandlerFile != "" && len(n.Methods) == 0 {
		return fmt.Errorf("%s: handler file must export at least one HTTP method function (GET, POST, PUT, DELETE, PATCH)", n.HandlerFile)
	}

	return nil
}

func (n *RouteNode) checkConflicts() error {
	// Group children by type
	static := make([]*RouteNode, 0)
	dynamic := make([]*RouteNode, 0)
	catchAll := make([]*RouteNode, 0)

	for _, child := range n.Children {
		if child.HandlerFile == "" && !child.HasLayout {
			continue
		}

		if child.IsCatchAll {
			catchAll = append(catchAll, child)
		} else if child.IsDynamic {
			dynamic = append(dynamic, child)
		} else {
			static = append(static, child)
		}
	}

	// Check for multiple catch-all routes
	if len(catchAll) > 1 {
		return fmt.Errorf("%s: multiple catch-all routes at same level", n.Path)
	}

	// Check for conflicts between static and dynamic routes
	if len(static) > 0 && len(dynamic) > 0 {
		// This is allowed by Go's ServeMux, but warn about precedence
		// Static routes will take precedence over dynamic ones
	}

	// Check for duplicate static routes
	seen := make(map[string]*RouteNode)
	for _, node := range static {
		if existing, exists := seen[node.URLSegment]; exists {
			if node.HandlerFile != "" && existing.HandlerFile != "" {
				return fmt.Errorf("duplicate route: %s and %s both map to /%s", node.HandlerFile, existing.HandlerFile, node.URLSegment)
			}
		}
		seen[node.URLSegment] = node
	}

	return nil
}

func validateParamName(name string) error {
	if name == "" {
		return fmt.Errorf("parameter name cannot be empty")
	}

	// Check first character is letter or underscore
	runes := []rune(name)
	if !unicode.IsLetter(runes[0]) && runes[0] != '_' {
		return fmt.Errorf("parameter name must start with letter or underscore: %s", name)
	}

	// Check remaining characters are letters, digits, or underscores
	for i, r := range runes {
		if i == 0 {
			continue
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return fmt.Errorf("parameter name contains invalid character: %s", name)
		}
	}

	return nil
}
