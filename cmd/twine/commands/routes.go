package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/cstone-io/twine/internal/routing"
	"github.com/spf13/cobra"
)

// NewRoutesCommand creates the routes command
func NewRoutesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "routes",
		Short: "Manage file-based routes",
		Long:  "Generate and manage file-based routes from app/ directory",
	}

	cmd.AddCommand(newRoutesGenerateCommand())
	cmd.AddCommand(newRoutesListCommand())

	return cmd
}

func newRoutesGenerateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate routes.gen.go from app/ directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get current directory
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting current directory: %w", err)
			}

			// Check if app/ directory exists
			appDir := filepath.Join(cwd, "app")
			if _, err := os.Stat(appDir); os.IsNotExist(err) {
				return fmt.Errorf("app/ directory not found. Create it first or run 'twine init'")
			}

			// Scan routes
			fmt.Println("🔍 Scanning routes in app/...")
			root, err := routing.ScanRoutes(appDir)
			if err != nil {
				return fmt.Errorf("scanning routes: %w", err)
			}

			// Validate routes
			if err := root.Validate(); err != nil {
				return fmt.Errorf("validation error: %w", err)
			}

			// Get module path
			modulePath, err := routing.GetModulePath(cwd)
			if err != nil {
				return fmt.Errorf("getting module path: %w", err)
			}

			// Generate code
			outputFile := filepath.Join(appDir, "routes.gen.go")
			generator := &routing.CodeGenerator{
				RouteTree:   root,
				ModulePath:  modulePath,
				ProjectRoot: cwd,
				OutputFile:  outputFile,
			}

			fmt.Println("📝 Generating routes.gen.go...")
			if err := generator.Generate(); err != nil {
				return fmt.Errorf("generating routes: %w", err)
			}

			fmt.Printf("✅ Routes generated successfully: %s\n", outputFile)

			// Display route table
			displayRouteTable(root)

			return nil
		},
	}
}

func newRoutesListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all discovered routes",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get current directory
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting current directory: %w", err)
			}

			// Check if app/ directory exists
			appDir := filepath.Join(cwd, "app")
			if _, err := os.Stat(appDir); os.IsNotExist(err) {
				return fmt.Errorf("app/ directory not found")
			}

			// Scan routes
			root, err := routing.ScanRoutes(appDir)
			if err != nil {
				return fmt.Errorf("scanning routes: %w", err)
			}

			// Display route table
			displayRouteTable(root)

			return nil
		},
	}
}

func displayRouteTable(root *routing.RouteNode) {
	// Collect all routes
	routes := collectAllRoutes(root)

	if len(routes) == 0 {
		fmt.Println("\n📭 No routes found")
		return
	}

	fmt.Println("\n📁 Routes discovered:")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	for _, route := range routes {
		urlPattern := route.ToURLPattern()
		relPath := strings.TrimPrefix(route.HandlerFile, filepath.Dir(root.Path)+"/")

		for _, method := range route.Methods {
			fmt.Fprintf(w, "   %s\t%s\t→ %s\n", method, urlPattern, relPath)
		}
	}

	w.Flush()

	// Display layouts
	layouts := collectAllLayouts(root)
	if len(layouts) > 0 {
		fmt.Println("\n🎨 Layouts active:")
		fmt.Println()

		for _, layout := range layouts {
			pathPattern := getLayoutPattern(layout)
			relPath := strings.TrimPrefix(layout.LayoutFile, filepath.Dir(root.Path)+"/")
			fmt.Printf("   %s\t→ %s\n", pathPattern, relPath)
		}
		fmt.Println()
	}
}

func collectAllRoutes(node *routing.RouteNode) []*routing.RouteNode {
	routes := make([]*routing.RouteNode, 0)

	if node.HandlerFile != "" {
		routes = append(routes, node)
	}

	for _, child := range node.Children {
		routes = append(routes, collectAllRoutes(child)...)
	}

	return routes
}

func collectAllLayouts(node *routing.RouteNode) []*routing.RouteNode {
	layouts := make([]*routing.RouteNode, 0)

	if node.HasLayout {
		layouts = append(layouts, node)
	}

	for _, child := range node.Children {
		layouts = append(layouts, collectAllLayouts(child)...)
	}

	return layouts
}

func getLayoutPattern(node *routing.RouteNode) string {
	path := node.GetFullPath()
	if path == "" {
		return "/"
	}
	return path + "/*"
}
