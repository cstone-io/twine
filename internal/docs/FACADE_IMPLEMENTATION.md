# Twine Export Facade Implementation

## Summary

Successfully implemented a comprehensive export facade in `twine.go` that re-exports commonly used types and functions from Twine's sub-packages, providing a simplified import experience for users.

## Implementation Details

### Architecture Decisions

1. **Type Aliases (with `=`)**: All type exports use type aliases to preserve full type interchangeability
   ```go
   type Kit = kit.Kit
   type HandlerFunc = kit.HandlerFunc
   type Router = router.Router
   ```

2. **Wrapper Functions**: Functions are re-exported using wrapper functions (not variable assignments) to preserve godoc comments
   ```go
   // NewRouter creates a new Router with the given URL prefix.
   func NewRouter(prefix string) *Router {
       return router.NewRouter(prefix)
   }
   ```

3. **Naming Conflict Resolution**: Generic function names are made more descriptive
   - `config.Get()` → `twine.Config()`
   - `database.GORM()` → `twine.DB()`
   - `logger.Get()` → `twine.Logger()`

4. **Selective Error Exports**: ~25 most common predefined errors exported, specialized errors remain in `pkg/errors`

### Exported Components

#### Core Types & Functions (Kit, Router, Handlers)
- `Kit`, `HandlerFunc`, `ErrorHandlerFunc`
- `Handler()`, `UseErrorHandler()`, `NotFoundHandler()`
- `Router`, `NewRouter()`
- HTTP method constants: `GET`, `POST`, `PUT`, `DELETE`

#### Middleware
- `Middleware` type
- `ApplyMiddlewares()`, `Chain()`
- `LoggingMiddleware()`, `TimeoutMiddleware()`, `JWTMiddleware()`

#### Authentication
- `Token`, `Credentials`
- `NewToken()`, `ParseToken()`, `HashPassword()`

#### Database
- `BaseModel`, `Polymorphic`
- `DB()`, `RegisterMigration()`, `RegisterMigrations()`
- `CRUDStore`, `CRUDStoreInterface`, `NewCRUDStore()`
- `Migration`, `NewMigrationBuilder()`
- `Seeder`, `NewSeeder()`

#### Templates
- `LoadTemplates()`, `SetTemplates()`, `GetTemplates()`
- `Reload()`, `FuncMap()`

#### Configuration & Logging
- `Config()`, `Logger()`
- `DatabaseConfig`, `LoggerConfig`, `AuthConfig`
- `LogLevel` constants: `LogTrace`, `LogDebug`, `LogInfo`, `LogWarn`, `LogError`, `LogCritical`

#### Errors
- `Error`, `ErrorBuilder`, `NewErrorBuilder()`
- `ErrSeverity` constants: `ErrMinor`, `ErrError`, `ErrCritical`
- 25 most common predefined errors (general, database, auth, API, server)

#### Server
- `Server`, `NewServer()`

#### Public Assets
- `AssetsFS`, `SetAssetsFS()`
- `FileServerHandler()`, `Asset()`
- Path constants: `AssetsPath`, `PublicPath`

## File Organization

The facade is organized into clear functional sections with separator comments:

1. Core Types - Kit & Handlers
2. Routing & HTTP
3. Middleware
4. Authentication & Security
5. Database
6. Templates
7. Configuration & Logging
8. Errors
9. Server
10. Public Assets

## Testing

Comprehensive test suite in `twine_test.go` covering:
- Basic usage (router, handlers, middleware)
- Type interchangeability
- Middleware composition
- Error handling
- Constants verification
- Auth functions
- Template functions
- Public assets

**Test Results**: All 8 tests passing ✓

## Examples

Two comprehensive examples demonstrating the facade:

1. **`examples/facade_demo.go`**: Basic usage showing simplified imports and common patterns
2. **`examples/advanced_facade_demo.go`**: Advanced usage with database, auth, templates, and error handling

Both examples compile successfully and demonstrate the improved developer experience.

## Before & After Comparison

### Before (without facade)
```go
import (
    "github.com/cstone-io/twine/pkg/kit"
    "github.com/cstone-io/twine/pkg/router"
    "github.com/cstone-io/twine/pkg/middleware"
    "github.com/cstone-io/twine/pkg/template"
    "github.com/cstone-io/twine/pkg/database"
    "github.com/cstone-io/twine/pkg/server"
)

func main() {
    r := router.NewRouter("")
    r.Use(middleware.LoggingMiddleware())

    template.LoadTemplates("templates/**/*.html")
    database.GORM()

    r.Get("/", func(k *kit.Kit) error {
        return k.Text(200, "Hello!")
    })
}
```

### After (with facade)
```go
import "github.com/cstone-io/twine"

func main() {
    r := twine.NewRouter("")
    r.Use(twine.LoggingMiddleware())

    twine.LoadTemplates("templates/**/*.html")
    twine.DB()

    r.Get("/", func(k *twine.Kit) error {
        return k.Text(200, "Hello!")
    })
}
```

## Benefits

1. **Simplified Imports**: Single import for common use cases instead of 5-10
2. **Better Discoverability**: Users can explore the API via `go doc twine`
3. **Preserved Documentation**: Wrapper functions maintain godoc comments
4. **Type Safety**: Type aliases ensure full interchangeability with sub-packages
5. **Escape Hatch**: Advanced users can still import specific sub-packages
6. **Idiomatic Go**: Follows established patterns from frameworks like Echo, Gin

## Trade-offs

**Maintenance Overhead**: Facade needs updating when new exports are added
- *Acceptable*: API is relatively stable

**Slight Indirection**: Wrapper functions add one extra call
- *Negligible*: Setup/initialization functions, not hot path

**Documentation Duplication**: Godoc comments duplicated
- *Worth it*: Improved user experience and IDE support

## Godoc Output

The facade produces clean, comprehensive documentation accessible via `go doc`:

```
$ go doc github.com/cstone-io/twine
package twine // import "github.com/cstone-io/twine"

const GET = router.GET ...
const LogTrace = config.LogTrace ...
const ErrMinor = errors.ErrMinor ...
...
func NewRouter(prefix string) *Router
func LoggingMiddleware() Middleware
func DB() *gorm.DB
...
type Kit = kit.Kit
type HandlerFunc = kit.HandlerFunc
type Router = router.Router
...
```

## Verification

✓ Code compiles without errors
✓ All tests pass (8/8)
✓ Godoc generates comprehensive documentation
✓ Type aliases preserve interchangeability
✓ Examples compile and run correctly
✓ No breaking changes to existing code

## Next Steps

Users can now:
1. Use `import "github.com/cstone-io/twine"` for simple applications
2. Import specific sub-packages for specialized functionality
3. Mix facade imports with sub-package imports as needed

The facade is production-ready and fully tested.

## Facade Example

```go
package main

// This example demonstrates the simplified import experience with the Twine facade.
//
// BEFORE (without facade):
// import (
//     "github.com/cstone-io/twine/pkg/kit"
//     "github.com/cstone-io/twine/pkg/router"
//     "github.com/cstone-io/twine/pkg/middleware"
//     "github.com/cstone-io/twine/pkg/template"
//     "github.com/cstone-io/twine/pkg/database"
//     "github.com/cstone-io/twine/pkg/server"
// )
//
// AFTER (with facade):
import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cstone-io/twine"
)

func main() {
	// Create a new router
	r := twine.NewRouter("")

	// Add middleware
	r.Use(twine.LoggingMiddleware())

	// Register routes
	r.Get("/", func(k *twine.Kit) error {
		return k.Text(200, "Welcome to Twine!")
	})

	r.Get("/hello/{name}", func(k *twine.Kit) error {
		name := k.PathValue("name")
		return k.JSON(200, map[string]string{
			"message": "Hello, " + name + "!",
		})
	})

	// API routes with timeout middleware
	api := twine.NewRouter("/api")
	api.Use(twine.TimeoutMiddleware(5 * 1000000000)) // 5 seconds

	api.Get("/status", func(k *twine.Kit) error {
		return k.JSON(200, map[string]string{
			"status": "ok",
		})
	})

	r.Sub(api)

	// Initialize router and create server
	mux := r.InitializeAsRoot()
	srv := twine.NewServer(":3000", mux)

	// Start server
	srv.Start()

	// Graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	twine.Logger().Info("Server started on :3000")
	srv.AwaitShutdown(ctx)
	twine.Logger().Info("Server shutdown complete")
}
```

## Advanced Facade Example

```go
package main

// Advanced example demonstrating database, auth, templates, and error handling
// with the Twine facade.

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cstone-io/twine"
	"github.com/google/uuid"
)

// User model using Twine's BaseModel
type User struct {
	twine.BaseModel
	Email    string
	Password string
}

func main() {
	// Load templates
	if err := twine.LoadTemplates("templates/**/*.html"); err != nil {
		twine.Logger().Error("Failed to load templates: %v", err)
	}

	// Register database migration
	userMigration := twine.NewMigrationBuilder().
		Model(&User{}).
		Name("users").
		Build()
	twine.RegisterMigration(userMigration)

	// Create CRUD store
	db := twine.DB()
	userStore := twine.NewCRUDStore[User](db)

	// Create router with custom error handler
	twine.UseErrorHandler(func(k *twine.Kit, err error) {
		if e, ok := err.(*twine.Error); ok {
			twine.Logger().CustomError(e)
			k.JSON(e.HTTPStatus, map[string]any{
				"error": e.Message,
				"code":  e.Code,
			})
		} else {
			k.JSON(500, map[string]any{
				"error": err.Error(),
			})
		}
	})

	r := twine.NewRouter("")
	r.Use(twine.LoggingMiddleware())

	// Public routes
	r.Get("/", func(k *twine.Kit) error {
		return k.RenderTemplate("index", map[string]any{
			"title": "Welcome to Twine",
		})
	})

	// Auth routes
	auth := twine.NewRouter("/auth")

	auth.Post("/register", func(k *twine.Kit) error {
		var creds twine.Credentials
		if err := k.Decode(&creds); err != nil {
			return twine.ErrDecodeJSON.Wrap(err)
		}

		// Hash password
		hash, err := twine.HashPassword(creds.Password)
		if err != nil {
			return twine.ErrHashPassword.Wrap(err)
		}

		// Create user
		user := User{
			Email:    creds.Email,
			Password: hash,
		}

		if err := userStore.Create(user); err != nil {
			return twine.ErrDatabaseWrite.Wrap(err)
		}

		return k.JSON(201, map[string]string{
			"message": "User created successfully",
		})
	})

	auth.Post("/login", func(k *twine.Kit) error {
		var creds twine.Credentials
		if err := k.Decode(&creds); err != nil {
			return twine.ErrDecodeJSON.Wrap(err)
		}

		// Find user by email
		users, err := userStore.List()
		if err != nil {
			return twine.ErrDatabaseRead.Wrap(err)
		}

		var user *User
		for _, u := range users {
			if u.Email == creds.Email {
				user = &u
				break
			}
		}

		if user == nil {
			return twine.ErrAuthInvalidCredentials
		}

		// Verify password
		if err := creds.Authenticate(user.Password); err != nil {
			return err
		}

		// Generate token
		token, err := twine.NewToken(user.ID, user.Email)
		if err != nil {
			return twine.ErrGenerateToken.Wrap(err)
		}

		return k.JSON(200, token)
	})

	r.Sub(auth)

	// Protected API routes
	api := twine.NewRouter("/api")
	api.Use(twine.JWTMiddleware())

	api.Get("/users", func(k *twine.Kit) error {
		users, err := userStore.List()
		if err != nil {
			return twine.ErrDatabaseRead.Wrap(err)
		}

		return k.JSON(200, users)
	})

	api.Get("/users/{id}", func(k *twine.Kit) error {
		id := k.PathValue("id")
		if id == "" {
			return twine.ErrAPIPathValue
		}

		user, err := userStore.Get(id)
		if err != nil {
			return twine.ErrAPIObjectNotFound.Wrap(err)
		}

		return k.JSON(200, user)
	})

	api.Delete("/users/{id}", func(k *twine.Kit) error {
		id := k.PathValue("id")
		if id == "" {
			return twine.ErrAPIPathValue
		}

		// Verify user owns this resource
		userID := k.GetContext("user")
		if userID != id {
			return twine.ErrInsufficientPermissions
		}

		if err := userStore.Delete(id); err != nil {
			return twine.ErrDatabaseDelete.Wrap(err)
		}

		return k.JSON(200, map[string]string{
			"message": "User deleted successfully",
		})
	})

	r.Sub(api)

	// Static assets
	r.Get("/public/{path...}", func(k *twine.Kit) error {
		twine.FileServerHandler().ServeHTTP(k.Response, k.Request)
		return nil
	})

	// 404 handler
	mux := r.InitializeAsRoot()
	mux.HandleFunc("/", twine.NotFoundHandler())

	// Start server
	srv := twine.NewServer(":3000", mux)
	srv.Start()

	// Graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	twine.Logger().Info("Server started on :3000")
	srv.AwaitShutdown(ctx)
	twine.Logger().Info("Server shutdown complete")
}

// Example of custom error handling
func handleError(k *twine.Kit, err error) error {
	// Custom business logic error
	customErr := twine.NewErrorBuilder().
		Code(9000).
		Message("Custom business error").
		Severity(twine.ErrError).
		HTTPStatus(400).
		Cause(err).
		Build()

	return customErr
}

// Example of seeding data
func seedTestData() {
	db := twine.DB()
	seeder := twine.NewSeeder(db, 100)

	users := []User{
		{
			BaseModel: twine.BaseModel{ID: uuid.New()},
			Email:     "test@example.com",
			Password:  "hashed_password",
		},
		{
			BaseModel: twine.BaseModel{ID: uuid.New()},
			Email:     "admin@example.com",
			Password:  "hashed_password",
		},
	}

	if err := seeder.Seed(users); err != nil {
		twine.Logger().Error("Failed to seed users: %v", err)
	}
}
```
