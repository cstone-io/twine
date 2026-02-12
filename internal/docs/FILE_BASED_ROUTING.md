# File-Based Routing

Twine supports file-based routing similar to Next.js, where routes are automatically discovered from your filesystem structure. This eliminates manual route registration boilerplate and provides an intuitive way to organize your application.

## Quick Start

1. **Initialize a new project:**
   ```bash
   twine init my-app
   cd my-app
   ```

2. **Generate routes:**
   ```bash
   twine routes generate
   ```

3. **Start development server:**
   ```bash
   twine dev  # Auto-regenerates routes on file changes
   ```

## Directory Structure

```
my-app/
├── app/
│   ├── pages/          # HTML pages (renders templates)
│   │   ├── page.go     # → /
│   │   ├── layout.go   # Layout for all pages
│   │   ├── about/
│   │   │   └── page.go # → /about
│   │   ├── users/
│   │   │   ├── [id]/
│   │   │   │   └── page.go  # → /users/{id}
│   │   │   └── new/
│   │   │       └── page.go  # → /users/new
│   │   └── docs/
│   │       └── [...slug]/
│   │           └── page.go  # → /docs/{slug...}
│   └── api/            # JSON API routes
│       ├── health/
│       │   └── route.go     # → /api/health
│       └── posts/
│           └── route.go     # → /api/posts
└── main.go
```

## Route Files

### Page Routes (`page.go`)

Page routes render HTML templates. Create a `page.go` file in `app/pages/`:

```go
package pages

import "github.com/cstone-io/twine/kit"

func GET(k *kit.Kit) error {
    return k.Render("pages/index", map[string]any{
        "Title": "Welcome to Twine",
    })
}
```

### API Routes (`route.go`)

API routes return JSON. Create a `route.go` file in `app/api/`:

```go
package health

import "github.com/cstone-io/twine/kit"

func GET(k *kit.Kit) error {
    return k.JSON(200, map[string]any{
        "status": "healthy",
    })
}

func POST(k *kit.Kit) error {
    return k.JSON(201, map[string]any{
        "message": "Created",
    })
}
```

### Supported HTTP Methods

Export functions named after HTTP methods:
- `GET`
- `POST`
- `PUT`
- `DELETE`
- `PATCH`

Each route file can export multiple methods.

## Dynamic Routes

### Single Parameter

Create a directory with brackets `[param]`:

```
app/pages/users/[id]/page.go  → /users/{id}
```

Access the parameter:

```go
package id_param

import "github.com/cstone-io/twine/kit"

func GET(k *kit.Kit) error {
    userID := k.PathValue("id")
    return k.Render("pages/user", map[string]any{
        "UserID": userID,
    })
}
```

**Important:** The package name must be sanitized (e.g., `id_param` for `[id]`).

### Catch-All Routes

Create a directory with `[...param]`:

```
app/pages/docs/[...slug]/page.go  → /docs/{slug...}
```

This matches `/docs/intro`, `/docs/guides/setup`, etc.

```go
package slug_catchall

import "github.com/cstone-io/twine/kit"

func GET(k *kit.Kit) error {
    slug := k.PathValue("slug")
    // slug will be "intro" or "guides/setup"
    return k.Render("pages/docs", map[string]any{
        "Slug": slug,
    })
}
```

## Layouts

Layouts are middleware that wrap routes. They execute from root to leaf.

### Root Layout

Create `app/pages/layout.go`:

```go
package pages

import (
    "github.com/cstone-io/twine/kit"
    "github.com/cstone-io/twine/middleware"
)

func Layout() middleware.Middleware {
    return func(next kit.HandlerFunc) kit.HandlerFunc {
        return func(k *kit.Kit) error {
            // Setup common data
            k.SetContext("appName", "My Twine App")
            return next(k)
        }
    }
}
```

### Nested Layouts

Create section-specific layouts:

```
app/pages/dashboard/layout.go  # Applies to /dashboard/*
app/pages/dashboard/page.go    # Has both root + dashboard layouts
```

```go
package dashboard

import (
    "github.com/cstone-io/twine/kit"
    "github.com/cstone-io/twine/middleware"
)

func Layout() middleware.Middleware {
    return func(next kit.HandlerFunc) kit.HandlerFunc {
        return func(k *kit.Kit) error {
            k.SetContext("section", "dashboard")
            return next(k)
        }
    }
}
```

**Layout execution order:** Root → Dashboard → Handler

### Using Existing Middleware

Layouts can return your existing middleware directly:

```go
package dashboard

import "github.com/cstone-io/twine/pkg/middleware"

func Layout() middleware.Middleware {
    // Return existing JWT middleware
    return middleware.JWTMiddleware()
}
```

Now all `/dashboard/*` routes require authentication!

### Chaining Multiple Middlewares

Use `middleware.Chain()` to compose multiple middlewares:

```go
package dashboard

import (
    "time"
    "github.com/cstone-io/twine/pkg/kit"
    "github.com/cstone-io/twine/pkg/middleware"
)

func Layout() middleware.Middleware {
    return middleware.Chain(
        middleware.JWTMiddleware(),              // Auth required
        middleware.TimeoutMiddleware(30*time.Second), // 30s timeout
        dashboardContextMiddleware(),             // Custom context
    )
}

func dashboardContextMiddleware() middleware.Middleware {
    return func(next kit.HandlerFunc) kit.HandlerFunc {
        return func(k *kit.Kit) error {
            k.SetContext("section", "dashboard")
            k.SetContext("nav", "dashboard")
            return next(k)
        }
    }
}
```

**Execution order:** JWT Auth → Timeout → Dashboard Context → Handler

## CLI Commands

### `twine routes generate`

Generates `app/routes.gen.go` from your file structure.

```bash
twine routes generate
```

Output:
```
🔍 Scanning routes in app/...
📝 Generating routes.gen.go...
✅ Routes generated successfully

📁 Routes discovered:
   GET     /                 → app/pages/page.go
   GET     /users/{id}       → app/pages/users/[id]/page.go
```

### `twine routes list`

Lists all discovered routes without generating code:

```bash
twine routes list
```

### `twine dev`

Starts development server with automatic route regeneration:

```bash
twine dev
```

- Watches `app/` directory for changes
- Regenerates routes automatically (500ms debounce)
- Uses Air for hot reload

## Generated Code

The generated `app/routes.gen.go` file looks like:

```go
// Code generated by twine routes generate. DO NOT EDIT.

package app

import (
    "github.com/cstone-io/twine/kit"
    "github.com/cstone-io/twine/router"
    "github.com/cstone-io/twine/middleware"

    pages "yourproject/app/pages"
    users_id_param "yourproject/app/pages/users/[id]"
)

func RegisterRoutes(r *router.Router) {
    // Layout chain for /
    pages_middleware := []middleware.Middleware{
        pages.Layout(),
    }
    r.Get("/", applyMiddleware(pages_middleware, pages.GET))

    // Layout chain for /users/{id}
    users_id_param_middleware := []middleware.Middleware{
        pages.Layout(),
    }
    r.Get("/users/{id}", applyMiddleware(users_id_param_middleware, users_id_param.GET))
}
```

## Integration with main.go

Your `main.go` imports and registers the generated routes:

```go
package main

import (
    "yourproject/app"
    "github.com/cstone-io/twine/router"
    "github.com/cstone-io/twine/server"
    "github.com/cstone-io/twine/template"
)

func main() {
    template.LoadTemplates("templates/**/*.html")

    r := router.NewRouter("")

    // Register file-based routes
    app.RegisterRoutes(r)

    // Can still add manual routes
    // r.Get("/custom", CustomHandler)

    mux := r.InitializeAsRoot()
    srv := server.NewServer(":3000", mux)
    srv.Start()
    srv.AwaitShutdown(context.Background())
}
```

## Route Priority

Go's ServeMux handles route priority:

1. **Static routes** match first: `/users/new`
2. **Dynamic routes** match next: `/users/{id}`
3. **Catch-all routes** match last: `/docs/{slug...}`

Within the same priority, longer paths win.

## Package Naming Rules

Dynamic segments must use valid Go identifiers:

| Directory | Package Name | URL Pattern |
|-----------|--------------|-------------|
| `[id]/` | `id_param` | `/users/{id}` |
| `[slug]/` | `slug_param` | `/posts/{slug}` |
| `[...path]/` | `path_catchall` | `/docs/{path...}` |

**Invalid names:**
- `[user-id]/` ❌ (hyphens not allowed)
- `[123]/` ❌ (can't start with number)

**Valid names:**
- `[userId]/` ✅
- `[user_id]/` ✅

## Validation

The route generator validates:

1. **HTTP methods:** At least one exported GET/POST/PUT/DELETE/PATCH
2. **Catch-all position:** Must be the last segment
3. **Parameter names:** Must be valid Go identifiers
4. **Duplicate routes:** Warns about conflicts

Example error:

```
❌ Validation error:
   app/pages/docs/[...path]/more/page.go

   → Catch-all must be last segment
```

## Hot Reload with Air

The generated `.air.toml` watches for changes:

```toml
[build]
  include_dir = [".", "app"]
  include_ext = ["go", "html"]
```

**Workflow:**
1. Modify `app/pages/users/[id]/page.go`
2. File watcher detects change
3. Routes regenerated (500ms debounce)
4. Air rebuilds and restarts server
5. Browser auto-refreshes

## Backward Compatibility

File-based routes **coexist** with manual registration:

```go
r := router.NewRouter("")

// Manual routes (existing code)
r.Get("/legacy", LegacyHandler)

// File-based routes (new code)
app.RegisterRoutes(r)

// Both work together!
```

No migration required - adopt incrementally.

## Best Practices

1. **Use layouts for shared logic:** Authentication, logging, context setup
2. **Keep handlers thin:** Move business logic to services
3. **Group related routes:** Use subdirectories for logical sections
4. **Name dynamic params clearly:** `[userId]` not `[id]` for clarity
5. **Leverage static routes:** `/users/new` before `/users/{id}`

## Troubleshooting

### Routes not discovered

- Check `app/` directory exists
- Verify file names: `page.go` or `route.go`
- Run `twine routes list` to see what's found

### Package name errors

- Directory `[user-id]/` → Rename to `[userId]/`
- Package name must match sanitized directory name

### Layout not applied

- Check `layout.go` exports `func Layout() middleware.Middleware`
- Verify it's in the correct directory (parent of routes)

### Import path errors

- Run `go mod tidy` in project directory
- Check `go.mod` module path matches imports

## Examples

See the full example in `/tmp/test-twine`:

```bash
cd /tmp/test-twine
twine routes list
```

Routes:
- `/` - Home page
- `/users/{id}` - User detail (dynamic)
- `/users/new` - New user form (static)
- `/dashboard` - Dashboard with nested layout
- `/docs/{slug...}` - Documentation (catch-all)
- `/api/health` - Health check API
- `/api/posts` - Posts API (GET, POST)
