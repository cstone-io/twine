package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cstone-io/twine/internal/routing"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

// NewDevCommand creates the dev command
func NewDevCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "dev",
		Short: "Start development server with hot reload",
		Long:  "Start the development server with automatic route generation and hot reload",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get current directory
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting current directory: %w", err)
			}

			// Check if app/ directory exists
			appDir := filepath.Join(cwd, "app")
			if _, err := os.Stat(appDir); err == nil {
				// Generate routes initially
				if err := generateRoutes(cwd, appDir); err != nil {
					fmt.Printf("⚠️  Warning: failed to generate routes: %v\n", err)
				}

				// Start file watcher
				go watchAppDirectory(cwd, appDir)
			} else {
				fmt.Println("ℹ️  No app/ directory found. Skipping route generation.")
				fmt.Println("   Run 'twine init' to create the app/ structure.")
			}

			// Check if Air is installed
			if _, err := exec.LookPath("air"); err != nil {
				return fmt.Errorf("air not found. Install it with: go install github.com/air-verse/air@latest")
			}

			// Start Air
			fmt.Println("🚀 Starting development server with Air...")
			fmt.Println()

			airCmd := exec.Command("air")
			airCmd.Stdout = os.Stdout
			airCmd.Stderr = os.Stderr
			airCmd.Stdin = os.Stdin

			return airCmd.Run()
		},
	}
}

func generateRoutes(cwd, appDir string) error {
	// Scan routes
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

	if err := generator.Generate(); err != nil {
		return fmt.Errorf("generating routes: %w", err)
	}

	return nil
}

func watchAppDirectory(cwd, appDir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("⚠️  Failed to create file watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	// Add app directory and all subdirectories
	if err := addDirectoryRecursive(watcher, appDir); err != nil {
		fmt.Printf("⚠️  Failed to watch app/ directory: %v\n", err)
		return
	}

	// Debounce timer
	var debounceTimer *time.Timer
	debounceDelay := 500 * time.Millisecond

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Only watch .go files and directory changes
			if !isWatchedFile(event.Name) && event.Op != fsnotify.Create {
				continue
			}

			// Reset debounce timer
			if debounceTimer != nil {
				debounceTimer.Stop()
			}

			debounceTimer = time.AfterFunc(debounceDelay, func() {
				fmt.Println("🔄 App directory changed, regenerating routes...")

				// Check if new directory was created
				if event.Op == fsnotify.Create {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						// Add new directory to watcher
						addDirectoryRecursive(watcher, event.Name)
					}
				}

				if err := generateRoutes(cwd, appDir); err != nil {
					fmt.Printf("❌ Failed to regenerate routes: %v\n", err)
				} else {
					fmt.Println("✅ Routes regenerated")
				}
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("⚠️  File watcher error: %v\n", err)
		}
	}
}

func addDirectoryRecursive(watcher *fsnotify.Watcher, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return watcher.Add(path)
		}

		return nil
	})
}

func isWatchedFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".go"
}
