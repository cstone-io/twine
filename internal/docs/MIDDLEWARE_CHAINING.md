# Middleware Chaining in Layouts

## Basic Usage

The `middleware.Chain()` function allows you to compose multiple middlewares into a single middleware, making it easy to apply several middlewares in layout files.

## Example: Authenticated Dashboard

**`app/pages/dashboard/layout.go`:**

```go
package dashboard

import (
    "time"
    "github.com/cstone-io/twine/pkg/kit"
    "github.com/cstone-io/twine/pkg/middleware"
)

func Layout() middleware.Middleware {
    return middleware.Chain(
        middleware.JWTMiddleware(),
        middleware.TimeoutMiddleware(30*time.Second),
        dashboardContextMiddleware(),
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

## Execution Order

Middlewares are applied in the order they appear in `Chain()`:

```go
middleware.Chain(
    middleware.LoggingMiddleware(),    // 1. Runs first
    middleware.JWTMiddleware(),        // 2. Runs second (after logging)
    customMiddleware(),                 // 3. Runs third
)
// Then the actual handler runs
```

## Nested Layouts

When layouts are nested, they execute from root to leaf:

```
app/pages/layout.go              → Runs first
app/pages/dashboard/layout.go    → Runs second
app/pages/dashboard/settings/layout.go → Runs third
handler                           → Runs last
```

**Example structure:**

```go
// app/pages/layout.go
func Layout() middleware.Middleware {
    return middleware.LoggingMiddleware()
}

// app/pages/dashboard/layout.go
func Layout() middleware.Middleware {
    return middleware.Chain(
        middleware.JWTMiddleware(),
        dashboardSetup(),
    )
}

// app/pages/dashboard/settings/layout.go
func Layout() middleware.Middleware {
    return adminOnlyMiddleware()
}
```

**For `/dashboard/settings`, execution order is:**
1. Logging (from `pages/layout.go`)
2. JWT auth (from `dashboard/layout.go`)
3. Dashboard setup (from `dashboard/layout.go`)
4. Admin check (from `dashboard/settings/layout.go`)
5. Settings page handler

## Common Patterns

### Global API Middleware

**`app/api/layout.go`:**

```go
func Layout() middleware.Middleware {
    return middleware.Chain(
        middleware.LoggingMiddleware(),
        middleware.CORSMiddleware(),
        middleware.JWTMiddleware(),
    )
}
```

All API routes now require auth and have CORS enabled.

### Public Exception in Authenticated Section

If you need a public route within an authenticated section, create a sub-layout:

```
app/pages/dashboard/
├── layout.go          # JWT required
├── page.go
└── public/
    ├── layout.go      # Override with no auth
    └── page.go
```

**`app/pages/dashboard/public/layout.go`:**

```go
func Layout() middleware.Middleware {
    // Return a no-op middleware to skip parent auth
    return func(next kit.HandlerFunc) kit.HandlerFunc {
        return next
    }
}
```

**Note:** This still inherits parent middlewares. If you need to truly skip parent middleware, use manual route registration as an escape hatch.

### Conditional Middleware

Apply middleware based on conditions:

```go
func Layout() middleware.Middleware {
    middlewares := []middleware.Middleware{
        middleware.LoggingMiddleware(),
    }

    if os.Getenv("ENABLE_AUTH") == "true" {
        middlewares = append(middlewares, middleware.JWTMiddleware())
    }

    return middleware.Chain(middlewares...)
}
```

## API Reference

### `Chain(middlewares ...Middleware) Middleware`

Combines multiple middlewares into a single middleware.

**Parameters:**
- `middlewares` - Variadic list of middlewares to chain

**Returns:**
- A single `Middleware` that applies all middlewares in order

**Example:**
```go
combined := middleware.Chain(
    middleware.LoggingMiddleware(),
    middleware.JWTMiddleware(),
)
```

### `ApplyMiddlewares(h kit.HandlerFunc, middlewares ...Middleware) kit.HandlerFunc`

Applies middlewares to a handler (used internally by the framework).

**Note:** Use `Chain()` in your layout files, not `ApplyMiddlewares()`.
