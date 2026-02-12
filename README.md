# Twine

A full-stack Go web framework for building server-side rendered applications with Alpine.js.

## Features

- **Hierarchical Router**: Composable routers with middleware inheritance
- **Alpine.js-First**: Built-in support for Alpine Ajax requests and responses
- **Template System**: Go's stdlib `html/template` with component-based architecture
- **Database Layer**: GORM integration with migrations and generic CRUD stores
- **Authentication**: JWT token generation and validation middleware
- **Error Handling**: Structured errors with severity levels and stack traces
- **Logging**: Configurable logging with multiple severity levels
- **Static Assets**: Embedded static file serving

## Installation

### Install CLI Tool (Recommended)

```bash
go install github.com/cstone-io/twine/cmd/twine@latest
```

### Or Install Framework Only

```bash
go get github.com/cstone-io/twine
```

## Quick Start

### Using the CLI (Easiest)

Create a new project in seconds:

```bash
# Create a new project
twine init my-app

# Navigate to project directory
cd my-app

# Run the application
go run main.go
```

Visit `http://localhost:3000` to see your app!

#### CLI Options

```bash
# Custom module path
twine init my-app --module github.com/myuser/my-app

# Custom port
twine init my-app --port 8080

# Minimal setup (no example pages)
twine init my-app --no-examples

# View all options
twine init --help
```

### Manual Setup

If you prefer to set up manually:

#### 1. Create a new project

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    "github.com/cstone-io/twine/router"
    "github.com/cstone-io/twine/kit"
    "github.com/cstone-io/twine/server"
    "github.com/cstone-io/twine/template"
)

func main() {
    // Load templates
    template.LoadTemplates("templates/**/*.html")

    // Create router
    r := router.NewRouter("")
    r.Get("/", Index)

    // Initialize server
    mux := r.InitializeAsRoot()
    srv := server.NewServer(":3000", mux)
    srv.Start()

    // Graceful shutdown
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()
    srv.AwaitShutdown(ctx)
}

func Index(k *kit.Kit) error {
    data := map[string]any{
        "Title": "Welcome to Twine",
    }
    return k.RenderTemplate("index", data)
}
```

#### 2. Create templates

```html
<!-- templates/pages/index.html -->
{{define "index"}}
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
</head>
<body>
    <h1>{{.Title}}</h1>
    {{template "button" .}}
</body>
</html>
{{end}}

<!-- templates/components/button.html -->
{{define "button"}}
<button>Click me</button>
{{end}}
```

#### 3. Configure environment

```env
# .env
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=postgres
DB_NAME=myapp
DB_SSLMODE=disable
DB_TIMEZONE=UTC

LOGGER_LEVEL=info
LOGGER_OUTPUT=stdout
LOGGER_ERROR_OUTPUT=stderr

AUTH_SECRET=your-secret-key-here
```

## Core Concepts

### Router

The router provides hierarchical routing with middleware inheritance:

```go
r := router.NewRouter("")

// Add middleware
r.Use(middleware.LoggingMiddleware())

// Create sub-router
api := router.NewRouter("/api")
api.Use(middleware.JWTMiddleware())

// Register routes
api.Get("/users", ListUsers)
api.Post("/users", CreateUser)
api.Get("/users/{id}", GetUser)
api.Put("/users/{id}", UpdateUser)
api.Delete("/users/{id}", DeleteUser)

// Mount sub-router
r.Sub(api)
```

### Kit

The Kit wraps `http.ResponseWriter` and `*http.Request` for convenient access:

```go
func Handler(k *kit.Kit) error {
    // Decode request body
    var payload UserPayload
    if err := k.Decode(&payload); err != nil {
        return err
    }

    // Get path parameters
    id := k.PathValue("id")

    // Get context values
    userID := k.GetContext("user")

    // Return JSON response
    return k.JSON(200, map[string]any{
        "message": "Success",
    })
}
```

### Templates

Templates use Go's stdlib `html/template`:

```go
// Load templates
template.LoadTemplates("templates/**/*.html")

// Render full page
func Index(k *kit.Kit) error {
    return k.RenderTemplate("index", data)
}

// Render partial (for Ajax)
func StatsPartial(k *kit.Kit) error {
    return k.RenderPartial("stats-card", stats)
}

// Auto-detect Ajax requests
func Dashboard(k *kit.Kit) error {
    // Automatically renders partial if X-Alpine-Request header is present
    return k.Render("dashboard", data)
}
```

### Database

GORM integration with migrations and generic CRUD stores:

```go
// Define model
type User struct {
    model.BaseModel `gorm:"embedded"`
    Name  string
    Email string
}

// Register migration
func init() {
    database.RegisterMigration(
        database.NewMigrationBuilder().
            Model(&User{}).
            Name("User").
            Build(),
    )
}

// Use CRUD store
store := store.NewCRUDStore[User](database.GORM())
users, err := store.List()
user, err := store.Get(id)
err = store.Create(user)
err = store.Update(user)
err = store.Delete(id)
```

### Middleware

Create custom middleware:

```go
func CustomMiddleware() middleware.Middleware {
    return func(next kit.HandlerFunc) kit.HandlerFunc {
        return func(k *kit.Kit) error {
            // Do something before
            err := next(k)
            // Do something after
            return err
        }
    }
}
```

Built-in middleware:

- `LoggingMiddleware()`: Request logging
- `TimeoutMiddleware(duration)`: Request timeouts
- `JWTMiddleware()`: JWT validation

### Authentication

JWT token generation and validation:

```go
// Generate token
token, err := auth.NewToken(userID, email)

// Validate token (done automatically by JWTMiddleware)
userID, err := auth.ParseToken(tokenString)

// Hash password
hash, err := auth.HashPassword(password)

// Verify password
creds := auth.Credentials{Email: email, Password: password}
err := creds.Authenticate(hashedPassword)
```

### Error Handling

Structured errors with custom handlers:

```go
// Use predefined errors
return errors.ErrNotFound

// Wrap errors
return errors.ErrDatabaseRead.Wrap(err)

// Add context
return errors.ErrDatabaseRead.Wrap(err).WithValue(user)

// Custom error handler
kit.UseErrorHandler(func(k *kit.Kit, err error) {
    if e, ok := err.(*errors.Error); ok {
        k.RenderTemplate("error", e)
    }
})
```

## Alpine.js Integration

Twine is designed to work seamlessly with Alpine.js and Alpine Ajax:

```html
<!-- Full page request -->
<a href="/dashboard">Dashboard</a>

<!-- Alpine Ajax partial request -->
<button x-target="stats" action="/stats">Refresh Stats</button>

<div id="stats">
    {{template "stats-card" .}}
</div>
```

```go
func Stats(k *kit.Kit) error {
    stats := getStats()
    // Automatically returns partial for Ajax requests
    return k.Render("stats-card", stats)
}
```

## Configuration

Configuration is loaded from environment variables and `.env` files:

```go
cfg := config.Get()

// Database config
dsn := cfg.Database.DSN()

// Logger config
level := cfg.Logger.Level

// Auth config
secret := cfg.Auth.SecretKey
```

## Project Structure

```
myapp/
├── main.go
├── .env
├── templates/
│   ├── pages/
│   │   └── index.html
│   ├── components/
│   │   └── button.html
│   └── layouts/
│       └── base.html
├── models/
│   └── user.go
├── handlers/
│   └── user.go
├── public/
│   └── assets/
│       ├── css/
│       └── js/
└── migrations/
    └── migrations.go
```

## License

MIT

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
