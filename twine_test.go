package twine_test

import (
	"testing"

	"github.com/cstone-io/twine"
)

// TestFacadeBasicUsage verifies the facade works for common use cases
func TestFacadeBasicUsage(t *testing.T) {
	// Test router creation
	r := twine.NewRouter("")
	if r == nil {
		t.Fatal("NewRouter returned nil")
	}

	// Test middleware
	r.Use(twine.LoggingMiddleware())

	// Test handler registration
	r.Get("/", func(k *twine.Kit) error {
		return k.Text(200, "Hello from Twine!")
	})

	r.Post("/api/data", func(k *twine.Kit) error {
		return k.JSON(200, map[string]string{"status": "ok"})
	})
}

// TestFacadeTypes verifies type aliases work correctly
func TestFacadeTypes(t *testing.T) {
	// Test that twine types are interchangeable with pkg types
	var _ twine.HandlerFunc = func(k *twine.Kit) error {
		return nil
	}

	// Test error types
	err := twine.ErrNotFound
	if err == nil {
		t.Fatal("ErrNotFound should not be nil")
	}

	// Test config types exist
	cfg := twine.Config()
	if cfg == nil {
		t.Fatal("Config returned nil")
	}
}

// TestFacadeMiddleware verifies middleware composition
func TestFacadeMiddleware(t *testing.T) {
	// Test middleware chaining
	combined := twine.Chain(
		twine.LoggingMiddleware(),
	)

	if combined == nil {
		t.Fatal("Chain returned nil")
	}

	// Test applying middlewares
	handler := func(k *twine.Kit) error {
		return k.Text(200, "test")
	}

	wrapped := twine.ApplyMiddlewares(handler, twine.LoggingMiddleware())
	if wrapped == nil {
		t.Fatal("ApplyMiddlewares returned nil")
	}
}

// TestFacadeErrorHandling verifies error handling
func TestFacadeErrorHandling(t *testing.T) {
	// Test error builder
	customErr := twine.NewErrorBuilder().
		Code(5000).
		Message("Custom error").
		Severity(twine.ErrError).
		Build()

	if customErr == nil {
		t.Fatal("NewErrorBuilder returned nil")
	}

	if customErr.Code != 5000 {
		t.Errorf("Expected code 5000, got %d", customErr.Code)
	}
}

// TestFacadeConstants verifies exported constants
func TestFacadeConstants(t *testing.T) {
	// Test HTTP method constants
	if twine.GET == "" {
		t.Fatal("GET constant is empty")
	}
	if twine.POST == "" {
		t.Fatal("POST constant is empty")
	}
	if twine.PUT == "" {
		t.Fatal("PUT constant is empty")
	}
	if twine.DELETE == "" {
		t.Fatal("DELETE constant is empty")
	}

	// Test log level constants
	if twine.LogInfo < 0 {
		t.Fatal("LogInfo constant is invalid")
	}

	// Test error severity constants
	if twine.ErrMinor < 0 {
		t.Fatal("ErrMinor constant is invalid")
	}
}

// TestFacadeAuth verifies auth functions
func TestFacadeAuth(t *testing.T) {
	// Test password hashing
	hash, err := twine.HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty string")
	}

	// Verify hash is different from password
	if hash == "testpassword123" {
		t.Fatal("Password was not hashed")
	}
}

// TestFacadeTemplate verifies template functions
func TestFacadeTemplate(t *testing.T) {
	// Test FuncMap
	funcMap := twine.FuncMap()
	if funcMap == nil {
		t.Fatal("FuncMap returned nil")
	}

	// Verify common template functions exist
	if _, ok := funcMap["formatDate"]; !ok {
		t.Error("FuncMap missing formatDate function")
	}
	if _, ok := funcMap["asset"]; !ok {
		t.Error("FuncMap missing asset function")
	}
}

// TestFacadePublicAssets verifies public asset functions
func TestFacadePublicAssets(t *testing.T) {
	// Test asset path generation
	path := twine.Asset("css/style.css")
	expected := "/public/assets/css/style.css"
	if path != expected {
		t.Errorf("Expected %s, got %s", expected, path)
	}

	// Test constants
	if twine.AssetsPath != "/public/assets/" {
		t.Errorf("Expected AssetsPath to be /public/assets/, got %s", twine.AssetsPath)
	}
	if twine.PublicPath != "/public/" {
		t.Errorf("Expected PublicPath to be /public/, got %s", twine.PublicPath)
	}
}
