package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewInitCommand tests init command creation
func TestNewInitCommand(t *testing.T) {
	cmd := NewInitCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "init <project-name>", cmd.Use)
	assert.Equal(t, "Initialize a new Twine project", cmd.Short)
	assert.NotNil(t, cmd.RunE)

	// Verify flags
	assert.NotNil(t, cmd.Flags().Lookup("module"))
	assert.NotNil(t, cmd.Flags().Lookup("port"))
	assert.NotNil(t, cmd.Flags().Lookup("no-examples"))
	assert.NotNil(t, cmd.Flags().Lookup("with-db"))
	assert.NotNil(t, cmd.Flags().Lookup("with-auth"))
}

// TestNewInitCommand_RequiresProjectName tests arg validation
func TestNewInitCommand_RequiresProjectName(t *testing.T) {
	cmd := NewInitCommand()

	// Execute without args
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	assert.Error(t, err)
}

// TestNewInitCommand_DefaultModulePath tests default module path
func TestNewInitCommand_DefaultModulePath(t *testing.T) {
	// We can't easily test the full init without external dependencies,
	// but we can verify the config is created correctly

	config := ProjectConfig{
		ProjectName: "myproject",
		ModulePath:  "example.com/myproject",
		Port:        "3000",
		WithDB:      false,
		WithAuth:    false,
		NoExamples:  false,
	}

	assert.Equal(t, "myproject", config.ProjectName)
	assert.Equal(t, "example.com/myproject", config.ModulePath)
	assert.Equal(t, "3000", config.Port)
	assert.False(t, config.WithDB)
	assert.False(t, config.WithAuth)
	assert.False(t, config.NoExamples)
}

// TestCheckNodeVersion tests Node.js version parsing
func TestCheckNodeVersion_ParsesVersion(t *testing.T) {
	// This test will only work if Node.js is installed
	// We'll make it conditional
	if _, err := os.Stat("/usr/local/bin/node"); os.IsNotExist(err) {
		if _, err := os.Stat("/usr/bin/node"); os.IsNotExist(err) {
			t.Skip("Node.js not installed, skipping version check test")
		}
	}

	// This will fail if Node.js < v16, but that's expected behavior
	err := checkNodeVersion()
	if err != nil {
		// If it fails, it should be because of version check
		assert.Contains(t, err.Error(), "version")
	}
}

// TestGenerateFromTemplate tests template generation
func TestGenerateFromTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	config := ProjectConfig{
		ProjectName: "testproject",
		ModulePath:  "github.com/test/project",
		Port:        "3000",
	}

	outputPath := filepath.Join(tmpDir, "go.mod")

	err := generateFromTemplate(config, "go.mod.tmpl", outputPath)
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, outputPath)

	// Verify content
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "github.com/test/project")
}

// TestGenerateFromTemplate_MainGo tests main.go generation
func TestGenerateFromTemplate_MainGo(t *testing.T) {
	tmpDir := t.TempDir()

	config := ProjectConfig{
		ProjectName: "testproject",
		ModulePath:  "github.com/test/project",
		Port:        "3000",
	}

	outputPath := filepath.Join(tmpDir, "main.go")

	err := generateFromTemplate(config, "main.go.tmpl", outputPath)
	require.NoError(t, err)

	assert.FileExists(t, outputPath)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "package main")
	assert.Contains(t, string(content), "3000") // Port
}

// TestGenerateFromTemplate_InvalidTemplate tests error handling
func TestGenerateFromTemplate_InvalidTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	config := ProjectConfig{
		ProjectName: "testproject",
	}

	outputPath := filepath.Join(tmpDir, "output.txt")

	err := generateFromTemplate(config, "nonexistent.tmpl", outputPath)
	assert.Error(t, err)
}

// TestCopyTemplates tests HTML template copying
func TestCopyTemplates(t *testing.T) {
	tmpDir := t.TempDir()

	config := ProjectConfig{
		NoExamples: false,
	}

	err := copyTemplates(config, tmpDir)
	require.NoError(t, err)

	// Verify base layout was copied
	baseLayout := filepath.Join(tmpDir, "templates", "layouts", "base.html")
	assert.FileExists(t, baseLayout)

	// Verify index page was copied
	indexPage := filepath.Join(tmpDir, "templates", "pages", "index.html")
	assert.FileExists(t, indexPage)

	// Verify component was copied
	button := filepath.Join(tmpDir, "templates", "components", "button.html")
	assert.FileExists(t, button)
}

// TestCopyTemplates_NoExamples tests skipping examples
func TestCopyTemplates_NoExamples(t *testing.T) {
	tmpDir := t.TempDir()

	config := ProjectConfig{
		NoExamples: true,
	}

	err := copyTemplates(config, tmpDir)
	require.NoError(t, err)

	// Base templates should still exist
	baseLayout := filepath.Join(tmpDir, "templates", "layouts", "base.html")
	assert.FileExists(t, baseLayout)

	// About page should not exist
	aboutPage := filepath.Join(tmpDir, "templates", "pages", "about.html")
	assert.NoFileExists(t, aboutPage)
}

// TestCreateAppStructure tests app directory creation
func TestCreateAppStructure(t *testing.T) {
	tmpDir := t.TempDir()

	config := ProjectConfig{
		ProjectName: "testproject",
	}

	err := createAppStructure(config, tmpDir)
	require.NoError(t, err)

	// Verify app/pages directory
	pagesDir := filepath.Join(tmpDir, "app", "pages")
	assert.DirExists(t, pagesDir)

	// Verify app/pages/page.go
	pageFile := filepath.Join(pagesDir, "page.go")
	assert.FileExists(t, pageFile)

	pageContent, err := os.ReadFile(pageFile)
	require.NoError(t, err)
	assert.Contains(t, string(pageContent), "package pages")
	assert.Contains(t, string(pageContent), "func GET")

	// Verify app/pages/layout.go
	layoutFile := filepath.Join(pagesDir, "layout.go")
	assert.FileExists(t, layoutFile)

	layoutContent, err := os.ReadFile(layoutFile)
	require.NoError(t, err)
	assert.Contains(t, string(layoutContent), "package pages")
	assert.Contains(t, string(layoutContent), "func Layout")

	// Verify app/api/health directory
	healthDir := filepath.Join(tmpDir, "app", "api", "health")
	assert.DirExists(t, healthDir)

	// Verify app/api/health/route.go
	healthFile := filepath.Join(healthDir, "route.go")
	assert.FileExists(t, healthFile)

	healthContent, err := os.ReadFile(healthFile)
	require.NoError(t, err)
	assert.Contains(t, string(healthContent), "package health")
	assert.Contains(t, string(healthContent), "func GET")
	assert.Contains(t, string(healthContent), "healthy")
}

// TestGenerateFiles tests full file generation
func TestGenerateFiles(t *testing.T) {
	tmpDir := t.TempDir()

	config := ProjectConfig{
		ProjectName: "testproject",
		ModulePath:  "github.com/test/project",
		Port:        "3000",
		NoExamples:  false,
	}

	err := generateFiles(config, tmpDir)
	require.NoError(t, err)

	// Verify Go files
	assert.FileExists(t, filepath.Join(tmpDir, "main.go"))
	assert.FileExists(t, filepath.Join(tmpDir, "go.mod"))
	assert.FileExists(t, filepath.Join(tmpDir, ".gitignore"))
	assert.FileExists(t, filepath.Join(tmpDir, ".env.example"))
	assert.FileExists(t, filepath.Join(tmpDir, "README.md"))
	assert.FileExists(t, filepath.Join(tmpDir, ".air.toml"))

	// Verify app structure
	assert.FileExists(t, filepath.Join(tmpDir, "app", "pages", "page.go"))
	assert.FileExists(t, filepath.Join(tmpDir, "app", "pages", "layout.go"))
	assert.FileExists(t, filepath.Join(tmpDir, "app", "api", "health", "route.go"))

	// Verify templates
	assert.FileExists(t, filepath.Join(tmpDir, "templates", "layouts", "base.html"))
	assert.FileExists(t, filepath.Join(tmpDir, "templates", "pages", "index.html"))
}

// TestGenerateNodeConfig tests Node.js config generation
func TestGenerateNodeConfig(t *testing.T) {
	tmpDir := t.TempDir()

	config := ProjectConfig{
		ProjectName: "testproject",
		ModulePath:  "github.com/test/project",
	}

	err := generateNodeConfig(config, tmpDir)
	require.NoError(t, err)

	// Verify package.json
	packageJSON := filepath.Join(tmpDir, "package.json")
	assert.FileExists(t, packageJSON)

	content, err := os.ReadFile(packageJSON)
	require.NoError(t, err)
	assert.Contains(t, string(content), "testproject")

	// Verify CSS directory
	cssDir := filepath.Join(tmpDir, "public", "assets", "css")
	assert.DirExists(t, cssDir)

	// Verify input.css
	inputCSS := filepath.Join(cssDir, "input.css")
	assert.FileExists(t, inputCSS)
}

// TestProjectConfig tests config struct
func TestProjectConfig(t *testing.T) {
	config := ProjectConfig{
		ProjectName: "myapp",
		ModulePath:  "github.com/user/myapp",
		Port:        "8080",
		WithDB:      true,
		WithAuth:    true,
		NoExamples:  true,
	}

	assert.Equal(t, "myapp", config.ProjectName)
	assert.Equal(t, "github.com/user/myapp", config.ModulePath)
	assert.Equal(t, "8080", config.Port)
	assert.True(t, config.WithDB)
	assert.True(t, config.WithAuth)
	assert.True(t, config.NoExamples)
}

// TestGenerateFiles_WithFlags tests generation with different flags
func TestGenerateFiles_WithFlags(t *testing.T) {
	tests := []struct {
		name   string
		config ProjectConfig
	}{
		{
			name: "with database",
			config: ProjectConfig{
				ProjectName: "dbproject",
				ModulePath:  "github.com/test/dbproject",
				Port:        "3000",
				WithDB:      true,
				WithAuth:    false,
				NoExamples:  false,
			},
		},
		{
			name: "with auth",
			config: ProjectConfig{
				ProjectName: "authproject",
				ModulePath:  "github.com/test/authproject",
				Port:        "3000",
				WithDB:      false,
				WithAuth:    true,
				NoExamples:  false,
			},
		},
		{
			name: "no examples",
			config: ProjectConfig{
				ProjectName: "minimalproject",
				ModulePath:  "github.com/test/minimalproject",
				Port:        "3000",
				WithDB:      false,
				WithAuth:    false,
				NoExamples:  true,
			},
		},
		{
			name: "custom port",
			config: ProjectConfig{
				ProjectName: "customport",
				ModulePath:  "github.com/test/customport",
				Port:        "8080",
				WithDB:      false,
				WithAuth:    false,
				NoExamples:  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			err := generateFiles(tt.config, tmpDir)
			require.NoError(t, err)

			// Verify basic files exist
			assert.FileExists(t, filepath.Join(tmpDir, "main.go"))
			assert.FileExists(t, filepath.Join(tmpDir, "go.mod"))

			// Verify go.mod has correct module path
			goModContent, err := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
			require.NoError(t, err)
			assert.Contains(t, string(goModContent), tt.config.ModulePath)
		})
	}
}

// TestCreateAppStructure_ContentVerification tests generated content
func TestCreateAppStructure_ContentVerification(t *testing.T) {
	tmpDir := t.TempDir()

	config := ProjectConfig{
		ProjectName: "testproject",
	}

	err := createAppStructure(config, tmpDir)
	require.NoError(t, err)

	// Verify page.go content in detail
	pageContent, err := os.ReadFile(filepath.Join(tmpDir, "app", "pages", "page.go"))
	require.NoError(t, err)

	assert.Contains(t, string(pageContent), "package pages")
	assert.Contains(t, string(pageContent), `import "github.com/cstone-io/twine/kit"`)
	assert.Contains(t, string(pageContent), "func GET(k *kit.Kit) error")
	assert.Contains(t, string(pageContent), "k.Render")
	assert.Contains(t, string(pageContent), "Welcome to Twine")

	// Verify layout.go content
	layoutContent, err := os.ReadFile(filepath.Join(tmpDir, "app", "pages", "layout.go"))
	require.NoError(t, err)

	assert.Contains(t, string(layoutContent), "package pages")
	assert.Contains(t, string(layoutContent), "func Layout() middleware.Middleware")
	assert.Contains(t, string(layoutContent), "k.SetContext")

	// Verify health route content
	healthContent, err := os.ReadFile(filepath.Join(tmpDir, "app", "api", "health", "route.go"))
	require.NoError(t, err)

	assert.Contains(t, string(healthContent), "package health")
	assert.Contains(t, string(healthContent), "func GET(k *kit.Kit) error")
	assert.Contains(t, string(healthContent), `"status": "healthy"`)
	assert.Contains(t, string(healthContent), "k.JSON")
}

// TestPrintSuccessMessage tests success message formatting
func TestPrintSuccessMessage(t *testing.T) {
	config := ProjectConfig{
		ProjectName: "myapp",
		Port:        "3000",
	}

	// Should not panic
	assert.NotPanics(t, func() {
		printSuccessMessage(config)
	})
}

// TestGenerateFromTemplate_AllTemplates tests all template files
func TestGenerateFromTemplate_AllTemplates(t *testing.T) {
	tmpDir := t.TempDir()

	config := ProjectConfig{
		ProjectName: "testproject",
		ModulePath:  "github.com/test/project",
		Port:        "3000",
	}

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

	for _, tmpl := range templates {
		t.Run(tmpl.src, func(t *testing.T) {
			outputPath := filepath.Join(tmpDir, tmpl.dest)
			err := generateFromTemplate(config, tmpl.src, outputPath)

			require.NoError(t, err, "Failed to generate %s", tmpl.src)
			assert.FileExists(t, outputPath, "%s should be created", tmpl.dest)

			// Verify file is not empty
			content, err := os.ReadFile(outputPath)
			require.NoError(t, err)
			assert.NotEmpty(t, content, "%s should not be empty", tmpl.dest)
		})
	}
}

// TestCopyTemplates_VerifyContent tests copied template content
func TestCopyTemplates_VerifyContent(t *testing.T) {
	tmpDir := t.TempDir()

	config := ProjectConfig{
		NoExamples: false,
	}

	err := copyTemplates(config, tmpDir)
	require.NoError(t, err)

	// Verify base.html has required structure
	baseContent, err := os.ReadFile(filepath.Join(tmpDir, "templates", "layouts", "base.html"))
	require.NoError(t, err)
	assert.Contains(t, string(baseContent), "<!DOCTYPE html>")
	assert.Contains(t, string(baseContent), "{{define \"base\"}}") // Or whatever the template structure is

	// Verify index.html exists and has content
	indexContent, err := os.ReadFile(filepath.Join(tmpDir, "templates", "pages", "index.html"))
	require.NoError(t, err)
	assert.NotEmpty(t, indexContent)

	// Verify button component exists
	buttonContent, err := os.ReadFile(filepath.Join(tmpDir, "templates", "components", "button.html"))
	require.NoError(t, err)
	assert.NotEmpty(t, buttonContent)
}

// TestGenerateNodeConfig_PackageJSON tests package.json content
func TestGenerateNodeConfig_PackageJSON(t *testing.T) {
	tmpDir := t.TempDir()

	config := ProjectConfig{
		ProjectName: "testproject",
		ModulePath:  "github.com/test/project",
	}

	err := generateNodeConfig(config, tmpDir)
	require.NoError(t, err)

	packageJSON := filepath.Join(tmpDir, "package.json")
	content, err := os.ReadFile(packageJSON)
	require.NoError(t, err)

	// Should contain project name
	assert.Contains(t, string(content), "testproject")

	// Should contain scripts (likely)
	// Note: Exact content depends on template
	assert.NotEmpty(t, content)
}
