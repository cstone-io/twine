package twine

// Package twine provides convenient access to the Twine web framework.
//
// This package re-exports commonly used types and functions from Twine's
// sub-packages for easier imports. For specialized functionality, import
// the specific sub-packages directly.
//
// Example usage:
//
//	r := twine.NewRouter("")
//	r.Use(twine.LoggingMiddleware())
//	r.Get("/", func(k *twine.Kit) error {
//	    return k.Text(200, "Hello!")
//	})

import (
	"embed"
	"html/template"
	"net/http"
	"time"

	"github.com/cstone-io/twine/pkg/auth"
	"github.com/cstone-io/twine/pkg/config"
	"github.com/cstone-io/twine/pkg/database"
	"github.com/cstone-io/twine/pkg/errors"
	"github.com/cstone-io/twine/pkg/kit"
	"github.com/cstone-io/twine/pkg/logger"
	"github.com/cstone-io/twine/pkg/middleware"
	"github.com/cstone-io/twine/pkg/public"
	"github.com/cstone-io/twine/pkg/router"
	"github.com/cstone-io/twine/pkg/server"
	pkgtemplate "github.com/cstone-io/twine/pkg/template"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================================
// Core Types - Kit & Handlers
// ============================================================================

// Kit wraps http.ResponseWriter and *http.Request for convenient access.
type Kit = kit.Kit

// HandlerFunc is the signature for Twine handlers that return errors.
type HandlerFunc = kit.HandlerFunc

// ErrorHandlerFunc is the signature for custom error handlers.
type ErrorHandlerFunc = kit.ErrorHandlerFunc

// Handler converts a Kit.HandlerFunc to an http.HandlerFunc.
func Handler(h HandlerFunc) http.HandlerFunc {
	return kit.Handler(h)
}

// UseErrorHandler sets a custom error handler for all Kit handlers.
func UseErrorHandler(h ErrorHandlerFunc) {
	kit.UseErrorHandler(h)
}

// NotFoundHandler returns a handler for 404 errors.
func NotFoundHandler() http.HandlerFunc {
	return kit.NotFoundHandler()
}

// ============================================================================
// Routing & HTTP
// ============================================================================

// Router provides hierarchical routing with middleware support.
type Router = router.Router

// NewRouter creates a new Router with the given URL prefix.
// The router supports hierarchical structure with middleware inheritance.
func NewRouter(prefix string) *Router {
	return router.NewRouter(prefix)
}

// HTTP method constants for route registration.
const (
	GET    = router.GET
	POST   = router.POST
	PUT    = router.PUT
	DELETE = router.DELETE
)

// ============================================================================
// Middleware
// ============================================================================

// Middleware is the signature for middleware functions.
type Middleware = middleware.Middleware

// ApplyMiddlewares chains multiple middlewares together.
func ApplyMiddlewares(h HandlerFunc, middlewares ...Middleware) HandlerFunc {
	return middleware.ApplyMiddlewares(h, middlewares...)
}

// Chain combines multiple middlewares into a single middleware.
// Useful for composing middlewares in layout files.
func Chain(middlewares ...Middleware) Middleware {
	return middleware.Chain(middlewares...)
}

// LoggingMiddleware logs incoming requests.
func LoggingMiddleware() Middleware {
	return middleware.LoggingMiddleware()
}

// TimeoutMiddleware adds a timeout to request processing.
func TimeoutMiddleware(d time.Duration) Middleware {
	return middleware.TimeoutMiddleware(d)
}

// JWTMiddleware validates JWT tokens and auto-redirects on failure.
func JWTMiddleware() Middleware {
	return middleware.JWTMiddleware()
}

// ============================================================================
// Authentication & Security
// ============================================================================

// Token represents a JWT authentication token.
type Token = auth.Token

// Credentials holds user authentication credentials.
type Credentials = auth.Credentials

// NewToken generates a new JWT token for a user.
func NewToken(userID uuid.UUID, email string) (*Token, error) {
	return auth.NewToken(userID, email)
}

// ParseToken validates and parses a JWT token, returning the user ID.
func ParseToken(tokenString string) (string, error) {
	return auth.ParseToken(tokenString)
}

// HashPassword hashes a password using bcrypt.
func HashPassword(password string) (string, error) {
	return auth.HashPassword(password)
}

// ============================================================================
// Database
// ============================================================================

// BaseModel provides standard fields for database models.
type BaseModel = database.BaseModel

// Polymorphic provides fields for polymorphic relationships.
type Polymorphic = database.Polymorphic

// DB returns the underlying GORM database instance.
func DB() *gorm.DB {
	return database.GORM()
}

// RegisterMigration adds a migration to the database.
func RegisterMigration(m *database.Migration) {
	database.RegisterMigration(m)
}

// RegisterMigrations adds multiple migrations to the database.
func RegisterMigrations(ms ...*database.Migration) {
	database.RegisterMigrations(ms...)
}

// Migration represents a database table migration with dependencies.
type Migration = database.Migration

// NewMigrationBuilder creates a new MigrationBuilder instance.
func NewMigrationBuilder() *database.MigrationBuilder {
	return database.NewMigrationBuilder()
}

// CRUDStore provides generic CRUD operations for any model type.
type CRUDStore[T any] = database.CRUDStore[T]

// CRUDStoreInterface defines the interface for CRUD operations.
type CRUDStoreInterface[T any] = database.CRUDStoreInterface[T]

// NewCRUDStore creates a new CRUD store for type T.
func NewCRUDStore[T any](client *gorm.DB) *CRUDStore[T] {
	return database.NewCRUDStore[T](client)
}

// Seeder provides a framework for seeding test data.
type Seeder = database.Seeder

// NewSeeder creates a new Seeder instance.
func NewSeeder(db *gorm.DB, batchSize int) *Seeder {
	return database.NewSeeder(db, batchSize)
}

// ============================================================================
// Templates
// ============================================================================

// LoadTemplates loads all templates from the given glob patterns.
func LoadTemplates(patterns ...string) error {
	return pkgtemplate.LoadTemplates(patterns...)
}

// SetTemplates allows users to set a custom template instance.
func SetTemplates(tmpl *template.Template) {
	pkgtemplate.SetTemplates(tmpl)
}

// GetTemplates returns the current template instance.
func GetTemplates() *template.Template {
	return pkgtemplate.GetTemplates()
}

// Reload reloads templates from the same patterns (useful in development).
func Reload(patterns ...string) error {
	return pkgtemplate.Reload(patterns...)
}

// FuncMap returns the default template functions.
func FuncMap() template.FuncMap {
	return pkgtemplate.FuncMap()
}

// ============================================================================
// Configuration & Logging
// ============================================================================

// Config returns the singleton configuration instance.
func Config() *config.Config {
	return config.Get()
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig = config.DatabaseConfig

// LoggerConfig holds logging configuration.
type LoggerConfig = config.LoggerConfig

// AuthConfig holds authentication configuration.
type AuthConfig = config.AuthConfig

// LogLevel represents logging verbosity levels.
type LogLevel = config.LogLevel

// Log level constants.
const (
	LogTrace    = config.LogTrace
	LogDebug    = config.LogDebug
	LogInfo     = config.LogInfo
	LogWarn     = config.LogWarn
	LogError    = config.LogError
	LogCritical = config.LogCritical
)

// Logger returns the singleton logger instance.
func Logger() *logger.Logger {
	return logger.Get()
}

// ============================================================================
// Errors
// ============================================================================

// Error represents a structured application error.
type Error = errors.Error

// ErrorBuilder provides a fluent interface for building errors.
type ErrorBuilder = errors.ErrorBuilder

// NewErrorBuilder creates a new error builder for custom errors.
func NewErrorBuilder() *ErrorBuilder {
	return errors.NewErrorBuilder()
}

// ErrSeverity represents the severity level of an error.
type ErrSeverity = errors.ErrSeverity

// Error severity constants.
const (
	ErrMinor    = errors.ErrMinor
	ErrError    = errors.ErrError
	ErrCritical = errors.ErrCritical
)

// Common predefined errors for typical CRUD applications.
// For specialized errors, import pkg/errors directly.
var (
	// General errors
	ErrNotFound   = errors.ErrNotFound
	ErrDecodeJSON = errors.ErrDecodeJSON
	ErrDecodeForm = errors.ErrDecodeForm

	// Database errors (most common)
	ErrDatabaseRead           = errors.ErrDatabaseRead
	ErrDatabaseWrite          = errors.ErrDatabaseWrite
	ErrDatabaseUpdate         = errors.ErrDatabaseUpdate
	ErrDatabaseDelete         = errors.ErrDatabaseDelete
	ErrDatabaseObjectNotFound = errors.ErrDatabaseObjectNotFound
	ErrDatabaseConn           = errors.ErrDatabaseConn
	ErrDatabaseMigration      = errors.ErrDatabaseMigration

	// Auth errors (most common)
	ErrAuthInvalidToken       = errors.ErrAuthInvalidToken
	ErrAuthExpiredToken       = errors.ErrAuthExpiredToken
	ErrAuthInvalidCredentials = errors.ErrAuthInvalidCredentials
	ErrInsufficientPermissions = errors.ErrInsufficientPermissions
	ErrAuthMissingHeader      = errors.ErrAuthMissingHeader
	ErrHashPassword           = errors.ErrHashPassword
	ErrGenerateToken          = errors.ErrGenerateToken

	// API errors (most common)
	ErrAPIRequestPayload     = errors.ErrAPIRequestPayload
	ErrAPIObjectNotFound     = errors.ErrAPIObjectNotFound
	ErrAPIIDMismatch         = errors.ErrAPIIDMismatch
	ErrAPIPathValue          = errors.ErrAPIPathValue
	ErrAPIRequestContentType = errors.ErrAPIRequestContentType

	// Server errors
	ErrListenAndServe = errors.ErrListenAndServe
	ErrShutdownServer = errors.ErrShutdownServer
)

// ============================================================================
// Server
// ============================================================================

// Server wraps http.Server with graceful shutdown support.
type Server = server.Server

// NewServer creates a new Server with the given address and handler.
func NewServer(addr string, handler http.Handler) *Server {
	return server.NewServer(addr, handler)
}

// ============================================================================
// Public Assets
// ============================================================================

// AssetsFS should be set by the user application using //go:embed.
//
// Example in user's code:
//
//	//go:embed assets
//	var AssetsFS embed.FS
//
//	func init() {
//	    twine.AssetsFS = AssetsFS
//	}
var AssetsFS = &public.AssetsFS

// Public asset path constants.
const (
	AssetsPath = public.AssetsPath
	PublicPath = public.PublicPath
)

// FileServerHandler returns an HTTP handler for serving embedded static files.
func FileServerHandler() http.Handler {
	return public.FileServerHandler()
}

// Asset returns the path to a static asset.
func Asset(name string) string {
	return public.Asset(name)
}

// SetAssetsFS sets the embedded filesystem for static assets.
func SetAssetsFS(fs embed.FS) {
	public.AssetsFS = fs
}
