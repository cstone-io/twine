package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/cstone-io/twine/internal/scaffold"
	"github.com/spf13/cobra"
)

type ProjectConfig struct {
	ProjectName string
	ModulePath  string
	Port        string
	WithDB      bool
	WithAuth    bool
	NoExamples  bool
}

func NewInitCommand() *cobra.Command {
	var (
		modulePath string
		port       string
		noExamples bool
		withDB     bool
		withAuth   bool
	)

	cmd := &cobra.Command{
		Use:   "init <project-name>",
		Short: "Initialize a new Twine project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			// Default module path
			if modulePath == "" {
				modulePath = fmt.Sprintf("example.com/%s", projectName)
			}

			config := ProjectConfig{
				ProjectName: projectName,
				ModulePath:  modulePath,
				Port:        port,
				WithDB:      withDB,
				WithAuth:    withAuth,
				NoExamples:  noExamples,
			}

			return initProject(config)
		},
	}

	cmd.Flags().StringVarP(&modulePath, "module", "m", "", "Go module path")
	cmd.Flags().StringVarP(&port, "port", "p", "3000", "Server port")
	cmd.Flags().BoolVar(&noExamples, "no-examples", false, "Skip example pages")
	cmd.Flags().BoolVar(&withDB, "with-db", false, "Include database setup")
	cmd.Flags().BoolVar(&withAuth, "with-auth", false, "Include auth setup")

	return cmd
}

func initProject(config ProjectConfig) error {
	// 1. Check Node.js availability
	if err := checkNodeJS(); err != nil {
		return err
	}

	// 2. Check Node.js version
	if err := checkNodeVersion(); err != nil {
		return err
	}

	// 3. Create project directory
	if err := os.Mkdir(config.ProjectName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	projectPath, _ := filepath.Abs(config.ProjectName)
	fmt.Printf("Creating new Twine project: %s\n", projectPath)

	// 4. Generate files from templates
	if err := generateFiles(config, projectPath); err != nil {
		return err
	}

	// 5. Generate Node.js config files
	if err := generateNodeConfig(config, projectPath); err != nil {
		return err
	}

	// 6. Run go mod tidy
	fmt.Println("\n✓ Downloading Go dependencies...")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectPath
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nWarning: Could not download Go dependencies automatically.\n")
		fmt.Printf("This is expected if the Twine framework hasn't been published yet.\n")
		fmt.Printf("You can manually run 'go mod tidy' in the project directory.\n")
	}

	// 7. Install Node.js dependencies
	if err := installNodeDependencies(projectPath); err != nil {
		fmt.Printf("\nWarning: Could not install npm dependencies automatically.\n")
		fmt.Printf("You can manually run 'npm install' in the project directory.\n")
	}

	// 8. Initialize git repository
	if err := initializeGitRepo(projectPath); err != nil {
		fmt.Printf("\nWarning: Could not initialize git repository automatically.\n")
		fmt.Printf("You can manually run 'git init' in the project directory.\n")
	}

	// 9. Print success message
	printSuccessMessage(config)
	return nil
}

func generateFiles(config ProjectConfig, projectPath string) error {
	// Generate from templates
	templates := []struct {
		src  string
		dest string
	}{
		{"main.go.tmpl", "main.go"},
		{"go.mod.tmpl", "go.mod"},
		{"gitignore.tmpl", ".gitignore"},
		{"env.example.tmpl", ".env.example"},
		{"README.md.tmpl", "README.md"},
		{".air.toml.tmpl", ".air.toml"},
	}

	for _, t := range templates {
		if err := generateFromTemplate(config, t.src, filepath.Join(projectPath, t.dest)); err != nil {
			return err
		}
	}

	// Copy HTML templates (no templating needed)
	if err := copyTemplates(config, projectPath); err != nil {
		return err
	}

	// Create app/ directory structure with example routes
	if err := createAppStructure(config, projectPath); err != nil {
		return err
	}

	return nil
}

func generateFromTemplate(config ProjectConfig, templatePath, outputPath string) error {
	// Read template from embed.FS
	content, err := scaffold.FS.ReadFile(templatePath)
	if err != nil {
		return err
	}

	// Parse and execute template
	tmpl, err := template.New("").Parse(string(content))
	if err != nil {
		return err
	}

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, config)
}

func copyTemplates(config ProjectConfig, projectPath string) error {
	// Copy HTML templates as-is
	templateFiles := []string{
		"templates/layouts/base.html",
		"templates/pages/index.html",
		"templates/components/button.html",
	}

	if !config.NoExamples {
		templateFiles = append(templateFiles, "templates/pages/about.html")
	}

	for _, src := range templateFiles {
		content, err := scaffold.FS.ReadFile(src)
		if err != nil {
			return err
		}

		// Determine destination path
		dest := src
		destPath := filepath.Join(projectPath, dest)

		// Create parent directories
		os.MkdirAll(filepath.Dir(destPath), 0755)

		// Write file
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return err
		}
	}

	return nil
}

func createAppStructure(config ProjectConfig, projectPath string) error {
	appPath := filepath.Join(projectPath, "app")

	// Create app/pages directory
	pagesPath := filepath.Join(appPath, "pages")
	if err := os.MkdirAll(pagesPath, 0755); err != nil {
		return err
	}

	// Create app/api/health directory
	healthPath := filepath.Join(appPath, "api", "health")
	if err := os.MkdirAll(healthPath, 0755); err != nil {
		return err
	}

	// Generate app/pages/page.go
	pageContent := `package pages

import "github.com/cstone-io/twine/kit"

func GET(k *kit.Kit) error {
	return k.Render("index", map[string]any{
		"Title":   "Welcome to Twine",
		"Message": "A full-stack Go web framework for server-side rendered applications with HTMX integration.",
	})
}
`
	if err := os.WriteFile(filepath.Join(pagesPath, "page.go"), []byte(pageContent), 0644); err != nil {
		return err
	}

	// Generate app/pages/layout.go
	layoutContent := `package pages

import (
	"github.com/cstone-io/twine/kit"
	"github.com/cstone-io/twine/middleware"
)

func Layout() middleware.Middleware {
	return func(next kit.HandlerFunc) kit.HandlerFunc {
		return func(k *kit.Kit) error {
			// Setup common data available to all pages
			k.SetContext("appName", "My Twine App")
			return next(k)
		}
	}
}
`
	if err := os.WriteFile(filepath.Join(pagesPath, "layout.go"), []byte(layoutContent), 0644); err != nil {
		return err
	}

	// Generate app/api/health/route.go
	healthContent := `package health

import "github.com/cstone-io/twine/kit"

func GET(k *kit.Kit) error {
	return k.JSON(200, map[string]any{
		"status": "healthy",
	})
}
`
	if err := os.WriteFile(filepath.Join(healthPath, "route.go"), []byte(healthContent), 0644); err != nil {
		return err
	}

	return nil
}

func printSuccessMessage(config ProjectConfig) {
	fmt.Println("\n✅ Project created successfully!")
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n\n", config.ProjectName)
	fmt.Printf("For development, run these commands in separate terminals:\n\n")
	fmt.Printf("  Terminal 1:\n")
	fmt.Printf("    npm run watch:css    # Watch and compile CSS\n\n")
	fmt.Printf("  Terminal 2:\n")
	fmt.Printf("    twine dev            # Start dev server with hot reload\n")
	fmt.Printf("    # or\n")
	fmt.Printf("    go run main.go       # Run directly\n")
	fmt.Printf("\nYour application will be running at http://localhost:%s\n", config.Port)
	fmt.Printf("\nFile-based routing is enabled in app/ directory:\n")
	fmt.Printf("  app/pages/           - HTML pages (renders templates)\n")
	fmt.Printf("  app/api/             - JSON API routes\n")
	fmt.Printf("\nFrontend tooling:\n")
	fmt.Printf("  npm run build:css    - Build CSS for production\n")
	fmt.Printf("  npm run watch:css    - Watch CSS during development\n")
}

// checkNodeJS verifies that Node.js and npm are installed
func checkNodeJS() error {
	// Check for node
	if _, err := exec.LookPath("node"); err != nil {
		return fmt.Errorf(`Node.js is not installed or not found in PATH.

Twine requires Node.js for frontend tooling (Tailwind CSS, PostCSS).

Please install Node.js v16 or higher:
  - macOS/Linux: https://nodejs.org or use nvm (https://github.com/nvm-sh/nvm)
  - Windows: https://nodejs.org

After installation, restart your terminal and try again.`)
	}

	// Check for npm
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf(`npm is not installed or not found in PATH.

npm should come bundled with Node.js. Please reinstall Node.js from:
  https://nodejs.org

After installation, restart your terminal and try again.`)
	}

	return nil
}

// checkNodeVersion verifies that Node.js version is v16 or higher
func checkNodeVersion() error {
	cmd := exec.Command("node", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check Node.js version: %w", err)
	}

	version := strings.TrimSpace(string(output))
	// Remove 'v' prefix (e.g., "v18.0.0" -> "18.0.0")
	version = strings.TrimPrefix(version, "v")

	// Parse major version
	parts := strings.Split(version, ".")
	if len(parts) < 1 {
		return fmt.Errorf("invalid Node.js version format: %s", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("failed to parse Node.js version: %w", err)
	}

	if major < 16 {
		return fmt.Errorf(`Node.js version v%s is too old. Minimum required: v16.0.0

Please upgrade Node.js:
  - macOS/Linux: https://nodejs.org or use nvm (https://github.com/nvm-sh/nvm)
  - Windows: https://nodejs.org`, version)
	}

	return nil
}

// generateNodeConfig creates Node.js configuration files
func generateNodeConfig(config ProjectConfig, projectPath string) error {
	fmt.Println("✓ Generating Node.js configuration...")

	// Generate package.json
	if err := generateFromTemplate(config, "package.json.tmpl", filepath.Join(projectPath, "package.json")); err != nil {
		return fmt.Errorf("failed to generate package.json: %w", err)
	}

	// Create public/assets/css directory
	cssDir := filepath.Join(projectPath, "public", "assets", "css")
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		return fmt.Errorf("failed to create CSS directory: %w", err)
	}

	// Copy input.css
	inputCSS, err := scaffold.FS.ReadFile("public/assets/css/input.css")
	if err != nil {
		return fmt.Errorf("failed to read input.css: %w", err)
	}

	inputCSSPath := filepath.Join(cssDir, "input.css")
	if err := os.WriteFile(inputCSSPath, inputCSS, 0644); err != nil {
		return fmt.Errorf("failed to write input.css: %w", err)
	}

	return nil
}

// installNodeDependencies runs npm install
func installNodeDependencies(projectPath string) error {
	fmt.Println("\n✓ Installing frontend dependencies...")
	fmt.Println("  This may take a minute...")

	cmd := exec.Command("npm", "install")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// initializeGitRepo initializes a git repository and creates an initial commit
func initializeGitRepo(projectPath string) error {
	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is not installed or not found in PATH")
	}

	fmt.Println("\n✓ Initializing git repository...")

	// Initialize git repository
	initCmd := exec.Command("git", "init")
	initCmd.Dir = projectPath
	if err := initCmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Set default branch to main
	branchCmd := exec.Command("git", "branch", "-M", "main")
	branchCmd.Dir = projectPath
	if err := branchCmd.Run(); err != nil {
		// Non-fatal if this fails
		fmt.Printf("  Warning: Could not set default branch to 'main'\n")
	}

	// Add all files
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = projectPath
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to add files to git: %w", err)
	}

	// Create initial commit
	commitMsg := "chore: initial project setup\n\nGenerated by Twine CLI"
	commitCmd := exec.Command("git", "commit", "-m", commitMsg)
	commitCmd.Dir = projectPath
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	fmt.Println("  Created initial commit on 'main' branch")

	return nil
}
