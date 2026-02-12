# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Twine is a full-stack Go web framework for server-side rendered applications with Alpine.js integration. It consists of:
- A framework library (`pkg/`) providing router, templates, database, auth, and middleware
- A CLI tool (`cmd/twine`) for project scaffolding and development workflows

## Build & Development Commands

This project uses [just](https://github.com/casey/just) as a command runner. Run `just` or `just --list` to see all available commands.

### Building the CLI
```bash
# Build CLI with version info
just build-cli

# Install CLI to $GOPATH/bin
just install-cli

# Build for all platforms
just build-cli-all

# Show version info that will be injected
just version-info
```

### Installing from GitHub
```bash
# Install the latest version from GitHub
go install github.com/cstone-io/twine/cmd/twine@latest

# Or install a specific version
go install github.com/cstone-io/twine/cmd/twine@v0.1.0
```

### Testing
```bash
# Run all tests
just test

# Run only unit tests (skip integration)
just test-unit

# Run tests with coverage report
just test-coverage

# Or directly:
go test -v ./...
```

### Development with Air
The project uses Air for hot reload during development:
```bash
# Use the CLI tool for automatic route generation and hot reload
twine dev
```

Air configuration (`.air.toml`) runs on proxy port 3001, proxying to app on port 3000.

### CSS Compilation
```bash
# Compile Tailwind CSS once
just css

# Watch and recompile on changes
just css-watch
```

### Working with Routes
```bash
# Generate routes from file structure (if using file-based routing)
twine routes generate

# Display route tree
twine routes show
```

## Architecture

### File-Based Routing System

Twine's most distinctive feature is its file-based routing system in the `app/` directory:

**Directory Structure:**
- `app/pages/` - Server-rendered pages (HTML responses)
- `app/api/` - API endpoints (JSON responses)
- Handler files: `page.go` (for pages) or `route.go` (for APIs)
- Layout files: `layout.go` (middleware/layout inheritance)

**Routing Conventions:**
- Directory names map to URL segments: `app/pages/users/` → `/users`
- Dynamic routes: `[id]/` → `{id}` parameter
- Catch-all routes: `[...slug]/` → `{slug...}` parameter
- HTTP methods detected by exported function names: `GET()`, `POST()`, `PUT()`, `DELETE()`, `PATCH()`

**Example:**
```
app/pages/users/[id]/page.go  →  /users/{id}

package user_id
func GET(k *kit.Kit) error { ... }    // Handles GET /users/{id}
func DELETE(k *kit.Kit) error { ... } // Handles DELETE /users/{id}
```

**Layout Inheritance:**
Place a `layout.go` file in any directory to apply middleware to all child routes:
```go
// app/pages/dashboard/layout.go
package dashboard

func Layout() middleware.Middleware {
    return middleware.JWTMiddleware()
}
```
Layouts are inherited from root to leaf (parent layouts execute before child layouts).

**Code Generation:**
The routing scanner (`internal/routing/`) walks the `app/` directory and generates `app/routes.gen.go` which registers all routes. This happens automatically with `twine dev` via file watching.

### Router System

The hierarchical router (`pkg/router/`) enables composable routing with middleware inheritance:

```go
r := router.NewRouter("")
r.Use(middleware.LoggingMiddleware())

api := router.NewRouter("/api")
api.Use(middleware.JWTMiddleware())
api.Get("/users", handlers.ListUsers)

r.Sub(api) // Mount as sub-router
mux := r.InitializeAsRoot() // Returns http.ServeMux
```

Key implementation details:
- Routes are stored in a tree structure with parent/child relationships
- `InitializeAsRoot()` flattens the tree and registers routes with `http.ServeMux`
- Routes are sorted by path length (longest first) for proper precedence
- Middleware from parent routers is inherited and prepended to child middleware

### Kit (Request/Response Wrapper)

The Kit (`pkg/kit/`) wraps `http.ResponseWriter` and `*http.Request`:

```go
type HandlerFunc func(kit *Kit) error
```

Handlers return errors instead of handling them inline. The Kit's error handler manages error responses.

**Key Methods:**
- Request: `Decode()`, `PathValue()`, `GetContext()`, `IsAjax()`
- Response: `JSON()`, `Text()`, `RenderTemplate()`, `RenderPartial()`, `Render()`
- Error handling: Set custom handler with `kit.UseErrorHandler()`

The `Render()` method auto-detects Ajax requests (via `X-Alpine-Request` header) and chooses between full page or partial rendering.

### Template System

Templates (`pkg/template/`) use Go's `html/template` with glob loading and a three-tier architecture:

```go
template.LoadTemplates("templates/**/*.html")
```

Templates are loaded once at startup. Use `{{define "name"}}` blocks for reusable components.

**Three-Tier Template Architecture:**

1. **Base Layouts** (`templates/layouts/`) - Full HTML structure with guaranteed script inclusion
2. **Pages** (`templates/pages/`) - Extend base layouts, define page-specific content
3. **Components** (`templates/components/`) - Reusable HTML fragments

**Important Distinction:** Go `layout.go` files provide *middleware* (request processing), while HTML template layouts provide *HTML structure composition*. These are two separate concerns.

**Alpine.js Integration:** The base layout guarantees Alpine.js and Alpine Ajax are loaded on every full-page render, enabling reactive components and Ajax functionality throughout the application.

#### Base Layout Pattern

The base layout (`templates/layouts/base.html`) guarantees Alpine.js, Alpine Ajax, and jQuery are included on every full-page render:

```html
{{define "base"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "title" .}}Twine App{{end}}</title>
    <link rel="stylesheet" href="/public/assets/css/output.css">

    {{/* Guaranteed Script Inclusion */}}
    <script defer src="https://cdn.jsdelivr.net/npm/@imacrayon/alpine-ajax@0.12.6/dist/cdn.min.js"></script>
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.14.1/dist/cdn.min.js"></script>
    <script src="https://code.jquery.com/jquery-3.7.1.min.js"
            integrity="sha256-/JqT3SQfawRcv/BIHPThkBvs0OEvtFFmqPF/lYI/Cxo="
            crossorigin="anonymous"></script>

    {{block "head" .}}{{end}}
</head>
<body {{block "body-attrs" .}}class="bg-gray-50"{{end}}>
    {{block "content" .}}
    <div class="max-w-4xl mx-auto px-6 py-16">
        <p class="text-gray-500">No content defined</p>
    </div>
    {{end}}

    {{block "scripts" .}}{{end}}
</body>
</html>
{{end}}
```

**Script Versions:**
- Alpine Ajax: v0.12.6 from jsDelivr CDN
- Alpine.js: v3.14.1 from jsDelivr CDN
- jQuery: v3.7.1 from jQuery CDN with SRI integrity hash

#### Page Templates

Pages extend the base layout using template composition:

```html
{{define "index"}}
{{template "base" .}}
{{end}}

{{define "title"}}{{.Title}}{{end}}

{{define "content"}}
<div class="max-w-4xl mx-auto px-6 py-16">
    <h1>{{.Title}}</h1>
    <!-- Alpine.js, Alpine Ajax, and jQuery are automatically available here -->
    <button x-target="response" action="/api/hello">Click me</button>
    <div id="response"></div>

    {{template "button" .}}  {{/* Include component */}}
</div>
{{end}}
```

**Key Points:**
- Pages invoke base layout with `{{template "base" .}}`
- No HTML document structure in pages (it's in base layout)
- No script tags in pages (Alpine.js, Alpine Ajax, and jQuery are guaranteed by base layout)
- Components can be included with `{{template "component-name" .}}`

#### Available Template Blocks

When extending the base layout, pages can customize:

- `{{define "title"}}` - Page title (browser tab)
- `{{define "head"}}` - Additional `<head>` content (meta tags, page styles)
- `{{define "body-attrs"}}` - Customize `<body>` attributes (default: `class="bg-gray-50"`)
- `{{define "content"}}` - Main page content (required)
- `{{define "scripts"}}` - Page-specific JavaScript at end of body

#### Component Templates

Components are pure HTML fragments without layouts:

```html
{{define "button"}}
<button x-target="response" action="/api/hello">
    Click me (Alpine Ajax)
</button>
<div id="response"></div>
{{end}}
```

**Key Points:**
- Components never include script tags (by design for Ajax compatibility)
- Can be included in pages via `{{template "button" .}}`
- Can be rendered as Ajax partial responses via `k.RenderPartial("button", data)`
- When rendered as partials, return pure HTML without base layout

#### Rendering Behavior

The Kit provides automatic Ajax detection:

```go
// Auto-detect based on X-Alpine-Request header
k.Render("index", data)
// - Full page request → uses base layout with scripts
// - Ajax request (has X-Alpine-Request header) → pure HTML fragment

// Explicit rendering:
k.RenderTemplate("index", data)  // Always use base layout
k.RenderPartial("button", data)  // Never use layout (pure HTML)
```

**Template Flow for Full Page Render:**
1. Handler calls `k.RenderTemplate("index", data)`
2. "index" template invokes `{{template "base" .}}`
3. Base layout renders with Alpine Ajax/Alpine.js/jQuery in `<head>`
4. "index" defines `{{define "content"}}` which slots into base's `{{block "content"}}`
5. Final HTML has guaranteed script inclusion

**Template Flow for Ajax Partial:**
1. Alpine Ajax makes request with `X-Alpine-Request` header
2. Handler calls `k.Render("button", data)` or `k.RenderPartial("button", data)`
3. Only the "button" template renders (no base layout)
4. Response is pure HTML fragment for Ajax swap

#### Script Guarantees

**What's Guaranteed:**
- Alpine Ajax (v0.12.6) loaded before Alpine.js core on all full-page renders
- Alpine.js (v3.14.1) loaded with defer attribute on all full-page renders
- jQuery (v3.7.1) loaded in `<head>` on all full-page renders
- Scripts load before content (in `<head>`)
- Tailwind CSS automatically included

**What's NOT Included:**
- Ajax partials don't include scripts (correct - they swap into pages that already have scripts)
- Components don't include scripts (by design)

#### Advanced Usage

**Page-Specific Scripts:**
```html
{{define "dashboard"}}
{{template "base" .}}
{{end}}

{{define "content"}}
<div id="chart"></div>
{{end}}

{{define "scripts"}}
<script>
// jQuery, Alpine.js, and Alpine Ajax already available
$(document).ready(function() {
    $('#chart').initChart();
});
</script>
{{end}}
```

**Custom Meta Tags:**
```html
{{define "product-page"}}
{{template "base" .}}
{{end}}

{{define "head"}}
<meta name="description" content="{{.Description}}">
<meta property="og:title" content="{{.Title}}">
{{end}}

{{define "content"}}
<h1>{{.Title}}</h1>
{{end}}
```

**Multiple Layouts:**
Projects can have multiple base layouts for different sections:
- `base.html` - Public pages
- `base-admin.html` - Admin dashboard with different nav
- `base-minimal.html` - Landing pages without scripts

Pages choose which layout:
```html
{{define "admin-dashboard"}}
{{template "base-admin" .}}
{{end}}
```

#### Error Pages

Error handlers should also use the base layout pattern for consistent UX and script availability:

```html
{{define "error-404"}}
{{template "base" .}}
{{end}}

{{define "title"}}404 - Page Not Found{{end}}

{{define "content"}}
<div class="max-w-4xl mx-auto px-6 py-16 text-center">
    <h1 class="text-6xl font-bold text-gray-900 mb-4">404</h1>
    <p class="text-xl text-gray-600">Page not found</p>
</div>
{{end}}
```

### Database Layer

GORM integration (`pkg/database/`) with:
- `BaseModel` for common fields (ID, CreatedAt, UpdatedAt, DeletedAt)
- Generic CRUD store: `store.NewCRUDStore[T](db)`
- Migration registration: `database.RegisterMigration()`
- Seeder support for dev data

### Middleware

Middleware signature (`pkg/middleware/`):
```go
type Middleware func(next kit.HandlerFunc) kit.HandlerFunc
```

Built-in middleware:
- `LoggingMiddleware()` - Request logging
- `TimeoutMiddleware(duration)` - Request timeouts
- `JWTMiddleware()` - JWT token validation

### Authentication

JWT-based auth (`pkg/auth/`):
- `auth.NewToken(userID, email)` - Generate JWT
- `auth.ParseToken(tokenString)` - Validate and extract claims
- `auth.HashPassword()` / `auth.Credentials.Authenticate()` - Password hashing with bcrypt

Tokens are signed with `AUTH_SECRET` from environment/config.

### Error Handling

Structured errors (`pkg/errors/`) with severity levels and stack traces:
- Predefined errors: `errors.ErrNotFound`, `errors.ErrDatabaseRead`, etc.
- Wrap errors: `errors.ErrDatabaseRead.Wrap(err)`
- Add context: `errors.ErrDatabaseRead.Wrap(err).WithValue(user)`

Custom error handlers can render error templates or return JSON based on request type.

### Configuration

Environment-based config (`pkg/config/`) loaded from `.env` files:
- Database: DSN constructed from DB_HOST, DB_PORT, etc.
- Logger: LOGGER_LEVEL, LOGGER_OUTPUT, LOGGER_ERROR_OUTPUT
- Auth: AUTH_SECRET for JWT signing

Access via `config.Get()`.

## Testing Strategy

When writing tests:
- Unit tests in `*_test.go` files alongside implementation
- Use Go's standard `testing` package
- Mock database with interfaces or use in-memory SQLite for integration tests
- Test error paths and edge cases in error handling code

## CLI Development

The CLI (`cmd/twine/`) uses Cobra for command structure:
- `commands/init.go` - Project scaffolding
- `commands/dev.go` - Development server with file watching
- `commands/routes.go` - Route generation and inspection
- `commands/version.go` - Version information

Version info is injected at build time via ldflags (see `justfile`).

## Internal Packages

The `internal/` directory contains framework-internal code:
- `routing/` - File-based routing scanner and code generator
  - `scanner.go` - Walks `app/` directory and builds route tree
  - `codegen.go` - Generates `app/routes.gen.go` from route tree
  - `types.go` - Core data structures (RouteNode, LayoutChain)
  - `layouts.go` - Layout middleware chain building
  - `validator.go` - Route validation rules
  - `pattern.go` - URL pattern generation from file paths
- `scaffold/` - Project scaffolding templates (embedded files)

## Important Notes

- This project uses conventional commit message format (see the `conventional-commits` skill)
- This project uses `just` (not `make`) for build automation - run `just` to see available commands
- The module path is `github.com/cstone-io/twine`
- File-based routing is regenerated on save when using `twine dev`
- The `app/routes.gen.go` file is auto-generated and should not be manually edited
- Alpine Ajax integration is automatic - the Kit detects `X-Alpine-Request` headers
- Middleware can be applied globally (router), per-route-group (sub-router), or per-directory (layout.go)
