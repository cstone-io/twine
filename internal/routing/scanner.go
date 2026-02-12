package routing

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ScanRoutes walks app/ directory and builds route tree
func ScanRoutes(rootDir string) (*RouteNode, error) {
	root := &RouteNode{
		Path:        rootDir,
		URLSegment:  "",
		IsDirectory: true,
		Children:    make([]*RouteNode, 0),
	}

	// Scan both pages and api directories
	pagesDir := filepath.Join(rootDir, "pages")
	apiDir := filepath.Join(rootDir, "api")

	if _, err := os.Stat(pagesDir); err == nil {
		pagesNode, err := scanDirectoryTree(pagesDir, root, "pages")
		if err != nil {
			return nil, fmt.Errorf("scanning pages: %w", err)
		}
		if pagesNode != nil {
			root.Children = append(root.Children, pagesNode)
		}
	}

	if _, err := os.Stat(apiDir); err == nil {
		apiNode, err := scanDirectoryTree(apiDir, root, "api")
		if err != nil {
			return nil, fmt.Errorf("scanning api: %w", err)
		}
		if apiNode != nil {
			root.Children = append(root.Children, apiNode)
		}
	}

	return root, nil
}

func scanDirectoryTree(dir string, parent *RouteNode, urlSegment string) (*RouteNode, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// Create node for this directory
	node := &RouteNode{
		Path:        dir,
		URLSegment:  urlSegment,
		Parent:      parent,
		IsDirectory: true,
		Children:    make([]*RouteNode, 0),
	}

	// Check for handler and layout files in this directory
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		fullPath := filepath.Join(dir, name)

		switch name {
		case "page.go":
			node.HandlerFile = fullPath
			node.IsPage = true
			methods, err := DetectMethods(fullPath)
			if err != nil {
				return nil, fmt.Errorf("detecting methods in %s: %w", fullPath, err)
			}
			node.Methods = methods
			pkg, err := getPackageName(fullPath)
			if err != nil {
				return nil, fmt.Errorf("getting package name from %s: %w", fullPath, err)
			}
			node.PackageName = pkg

		case "route.go":
			node.HandlerFile = fullPath
			node.IsAPI = true
			methods, err := DetectMethods(fullPath)
			if err != nil {
				return nil, fmt.Errorf("detecting methods in %s: %w", fullPath, err)
			}
			node.Methods = methods
			pkg, err := getPackageName(fullPath)
			if err != nil {
				return nil, fmt.Errorf("getting package name from %s: %w", fullPath, err)
			}
			node.PackageName = pkg

		case "layout.go":
			node.LayoutFile = fullPath
			node.HasLayout = true
			if node.PackageName == "" {
				pkg, err := getPackageName(fullPath)
				if err != nil {
					return nil, fmt.Errorf("getting package name from %s: %w", fullPath, err)
				}
				node.PackageName = pkg
			}
		}
	}

	// Recursively scan subdirectories
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		subPath := filepath.Join(dir, dirName)

		// Determine URL segment for this directory
		segment := dirName
		isDynamic := false
		isCatchAll := false
		paramName := ""

		if strings.HasPrefix(dirName, "[") && strings.HasSuffix(dirName, "]") {
			isDynamic = true
			paramName = strings.TrimSuffix(strings.TrimPrefix(dirName, "["), "]")

			if strings.HasPrefix(paramName, "...") {
				isCatchAll = true
				paramName = strings.TrimPrefix(paramName, "...")
				segment = fmt.Sprintf("{%s...}", paramName)
			} else {
				segment = fmt.Sprintf("{%s}", paramName)
			}
		}

		// Recursively scan subdirectory
		childNode, err := scanDirectoryTree(subPath, node, segment)
		if err != nil {
			return nil, err
		}

		// Add child node if it or its descendants have content
		if childNode != nil && (childNode.HandlerFile != "" || childNode.HasLayout || len(childNode.Children) > 0) {
			childNode.IsDynamic = isDynamic
			childNode.IsCatchAll = isCatchAll
			childNode.ParamName = paramName
			node.Children = append(node.Children, childNode)
		}
	}

	return node, nil
}

// DetectMethods parses a handler file and returns exported HTTP method functions
func DetectMethods(filePath string) ([]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return nil, err
	}

	methods := make([]string, 0)
	validMethods := map[string]bool{
		"GET":    true,
		"POST":   true,
		"PUT":    true,
		"DELETE": true,
		"PATCH":  true,
	}

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		// Check if function is exported and is a valid HTTP method
		if funcDecl.Name.IsExported() && validMethods[funcDecl.Name.Name] {
			methods = append(methods, funcDecl.Name.Name)
		}
	}

	return methods, nil
}

// getPackageName extracts the package name from a Go file
func getPackageName(filePath string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}
	return file.Name.Name, nil
}
