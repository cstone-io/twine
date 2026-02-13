# Twine - Justfile
# Run `just` to see all available recipes
# Run `just <recipe>` to execute a specific recipe

# Configuration
binary_name := "twine"
cli_binary := "twine"

# Build info (computed at recipe execution time)
version := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
commit := `git rev-parse --short HEAD 2>/dev/null || echo "none"`
build_date := `date -u '+%Y-%m-%d_%H:%M:%S'`
built_by := env_var_or_default("USER", "unknown")

# Go linker flags for version injection
ldflags := "-ldflags \"-X github.com/cstone-io/twine/cmd/twine/commands.Version=" + version + " -X github.com/cstone-io/twine/cmd/twine/commands.Commit=" + commit + " -X github.com/cstone-io/twine/cmd/twine/commands.Date=" + build_date + " -X github.com/cstone-io/twine/cmd/twine/commands.BuiltBy=" + built_by + "\""

# -----------------------------------------------------------------------------
# HIGH LEVEL COMMANDS
# -----------------------------------------------------------------------------

# Default recipe - list all available recipes
default:
    @just --list

# Show detailed help for all recipes
help:
    @echo "Twine - Just Command Runner"
    @echo ""
    @echo "Usage: just <recipe>"
    @echo ""
    @echo "Available recipes:"
    @just --list
    @echo ""
    @echo "Build info:"
    @echo "  Version: {{version}}"
    @echo "  Commit:  {{commit}}"

# Count lines of code (requires: brew install cloc)
cloc:
    cloc . --vcs=git

# Show version info that will be injected into builds
version-info:
    @echo "Version: {{version}}"
    @echo "Commit:  {{commit}}"
    @echo "Date:    {{build_date}}"
    @echo "Built by: {{built_by}}"

# -----------------------------------------------------------------------------
# GO MANGEMENT COMMANDS
# -----------------------------------------------------------------------------

# Clean build artifacts
clean:
    rm -rf ./bin ./dist coverage.out coverage.html
    @echo "✅ Cleaned build artifacts"

# Run Go mod tidy to clean up dependencies
tidy:
    go mod tidy
    @echo "✅ Cleaned Go module dependencies"

# Format all Go code
fmt:
    go fmt ./...
    @echo "✅ Formatted Go code"

# Run Go linter (requires: brew install golangci-lint)
lint:
    golangci-lint run ./...

# Run all quality checks (fmt, lint, test)
check: fmt lint test
    @echo "✅ All checks passed!"

# -----------------------------------------------------------------------------
# TEST COMMANDS
# -----------------------------------------------------------------------------

# Run all tests (unit + integration), excluding examples
test:
    go test -v $(go list ./... | grep -v /examples)

# Run only fast unit tests (skip integration tests), excluding examples
test-unit:
    go test -v -short $(go list ./... | grep -v /examples)

# Run tests with coverage report (goal: 90%+), excluding examples
test-coverage:
    @echo "Running tests with coverage..."
    go test -v -coverprofile=coverage.out $(go list ./... | grep -v /examples)
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"
    go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $3}'

# Check if coverage meets 90% threshold
test-coverage-check: test-coverage
    #!/usr/bin/env bash
    set -euo pipefail
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$COVERAGE < 90" | bc -l) )); then
        echo "❌ Coverage ${COVERAGE}% is below 90% threshold"
        exit 1
    fi
    echo "✅ Coverage ${COVERAGE}% meets 90% threshold"

# Run tests for a specific package (interactive)
test-pkg:
    #!/usr/bin/env bash
    read -p "Enter package (e.g., pkg/router): " pkg
    go test -v ./${pkg}/...

# Run tests for a specific package (non-interactive)
test-package package:
    go test -v ./{{package}}/...

# Watch and re-run tests on file changes (requires: brew install entr)
test-watch:
    find . -name "*.go" -not -path "./vendor/*" -not -path "./examples/*" | entr -c go test -v $(go list ./... | grep -v /examples)

# -----------------------------------------------------------------------------
# CLI COMMANDS
# -----------------------------------------------------------------------------

# Build CLI with version info injected
build-cli:
    @echo "Building {{cli_binary}} version {{version}}..."
    mkdir -p dist
    go build {{ldflags}} -o dist/{{cli_binary}} ./cmd/twine

# Install CLI to $GOPATH/bin with version info
install-cli:
    @echo "Installing {{cli_binary}} version {{version}}..."
    go install {{ldflags}} ./cmd/twine

# Build CLI for all major platforms
build-cli-all:
    @echo "Building {{cli_binary}} for multiple platforms..."
    mkdir -p dist
    GOOS=darwin GOARCH=amd64 go build {{ldflags}} -o dist/{{cli_binary}}-darwin-amd64 ./cmd/twine
    GOOS=darwin GOARCH=arm64 go build {{ldflags}} -o dist/{{cli_binary}}-darwin-arm64 ./cmd/twine
    GOOS=linux GOARCH=amd64 go build {{ldflags}} -o dist/{{cli_binary}}-linux-amd64 ./cmd/twine
    GOOS=linux GOARCH=arm64 go build {{ldflags}} -o dist/{{cli_binary}}-linux-arm64 ./cmd/twine
    GOOS=windows GOARCH=amd64 go build {{ldflags}} -o dist/{{cli_binary}}-windows-amd64.exe ./cmd/twine
    @echo "✅ Built binaries for all platforms in dist/"

# -----------------------------------------------------------------------------
# DEV COMMANDS
# -----------------------------------------------------------------------------

# Build the main application (depends on CSS compilation)
build: css
    templ generate
    mkdir -p ./bin
    go build -o ./bin/{{binary_name}} ./cmd/{{binary_name}}

# Build and run the application
run: build
    ./bin/{{binary_name}}

# Development workflow - watch for changes and rebuild
dev:
    @echo "Starting development mode..."
    @echo "Run 'just css-watch' in another terminal for CSS hot reload"
    just run

# Compile Tailwind CSS
css:
    tailwindcss -i ./assets/index.css -o ./public/assets/styles.css

# Watch and recompile CSS on changes
css-watch:
    tailwindcss -i ./assets/index.css -o ./public/assets/styles.css --watch
