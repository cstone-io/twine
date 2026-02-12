# Test Utilities

**Location:** `internal/testutil/` - This package is in the `internal/` directory, which means it's only importable by code within this project. External users of the Twine library cannot import these test utilities, as they are for internal testing infrastructure only.

This package provides test utilities and helpers for the Twine project. It includes helpers for:

- Creating temporary directories and files
- Setting up test databases (SQLite in-memory by default)
- HTTP response assertions
- Test fixture management
- Environment variable management

## Usage

### Core Test Utilities

```go
import "github.com/cstone-io/twine/internal/testutil"

func TestMyFeature(t *testing.T) {
    // Create temporary directory
    dir := testutil.TempDir(t)
    // Automatically cleaned up after test

    // Create files
    testutil.CreateFile(t, filepath.Join(dir, "test.txt"), "content")

    // Set environment variables
    testutil.SetupTestEnv(t, map[string]string{
        "TEST_VAR": "test_value",
    })
    // Automatically restored after test
}
```

### Database Testing

```go
import "github.com/cstone-io/twine/internal/testutil"

func TestDatabaseOperations(t *testing.T) {
    // Setup in-memory SQLite database
    db := testutil.SetupTestDB(t)
    // Automatically closed after test

    // Run migrations
    testutil.AutoMigrate(t, db, &User{}, &Post{})

    // Seed test data
    users := []User{
        {Email: "user1@example.com"},
        {Email: "user2@example.com"},
    }
    testutil.SeedTestData(t, db, &users)

    // Assert records exist
    testutil.AssertRecordExists(t, db, &User{}, "email = ?", "user1@example.com")
    testutil.AssertRecordCount(t, db, &User{}, 2, "1 = ?", 1)
}
```

#### Using PostgreSQL for Tests

By default, tests use SQLite in-memory databases for speed and simplicity. To test against PostgreSQL:

```bash
# Set environment variable
export POSTGRES_TEST_DSN="host=localhost user=test password=test dbname=test_db port=5432 sslmode=disable"

# Run tests
make test
```

#### Transaction Rollback Pattern

```go
func TestWithRollback(t *testing.T) {
    db := testutil.SetupTestDB(t)
    testutil.AutoMigrate(t, db, &User{})

    // Changes are rolled back after function completes
    testutil.RunInTransaction(t, db, func(tx *gorm.DB) {
        user := User{Email: "test@example.com"}
        err := tx.Create(&user).Error
        require.NoError(t, err)
        // Do your tests here
    })

    // Database is clean again
}
```

### HTTP Assertions

```go
import (
    httpAssert "github.com/cstone-io/twine/internal/testutil/assert"
    "net/http/httptest"
)

func TestHandler(t *testing.T) {
    w := httptest.NewRecorder()
    r := httptest.NewRequest("GET", "/api/users", nil)

    handler(w, r)

    // Assert JSON response
    httpAssert.AssertJSONResponse(t, w, 200, `{"users": []}`)

    // Assert HTML response contains text
    httpAssert.AssertHTMLResponse(t, w, 200, "Welcome")

    // Assert headers
    httpAssert.AssertHeader(t, w, "Content-Type", "application/json")

    // Assert Ajax partial response (no DOCTYPE, no <html>)
    httpAssert.AssertAjaxResponse(t, w, 200)

    // Assert redirect
    httpAssert.AssertRedirect(t, w, "/login")
}
```

### Test Fixtures

```go
// Load raw fixture data
data := testutil.LoadFixture(t, "users.json")

// Load and unmarshal JSON fixture
var users []User
testutil.LoadJSONFixture(t, "users.json", &users)

// Write fixture for test
testutil.WriteFixture(t, tempDir, "config.json", []byte(`{"key": "value"}`))
```

## Directory Structure

```
internal/testutil/      # Internal test utilities (not importable by external projects)
  testutil.go           # Core test helpers
  database.go           # Database test helpers
  assert/
    http.go            # HTTP assertion helpers
  fixtures/            # Test data fixtures
    routes/            # File-based routing fixtures
    templates/         # Template test files
    config/            # Test .env files
```

## Running Tests

```bash
# Run all tests
make test

# Run only fast unit tests (skip integration)
make test-unit

# Run tests with coverage report
make test-coverage

# Check if coverage meets 90% threshold
make test-coverage-check

# Run specific package tests
make test-pkg
# Enter: pkg/router

# Watch tests (requires entr)
make test-watch
```

## Coverage Goals

- **Phase 1 (Current):** Test infrastructure setup - testutil package has 90%+ coverage ✓
- **Phase 2:** Package-by-package testing - target 90%+ coverage per package
- **Overall target:** 90%+ coverage across the entire codebase

## Writing New Tests

### Test File Naming

- `<package>_test.go` - Unit tests for primary functionality
- `<feature>_test.go` - Unit tests for specific feature
- `<package>_integration_test.go` - Integration tests

### Test Function Naming

```go
// Pattern: Test<FunctionName>_<Scenario>_<ExpectedResult>
func TestNewRouter_EmptyPrefix_CreatesRouter(t *testing.T)
func TestRouter_Sub_InheritsMiddleware(t *testing.T)
func TestParseToken_ExpiredToken_ReturnsError(t *testing.T)
```

### Table-Driven Tests

```go
func TestRouter_PathMatching(t *testing.T) {
    tests := []struct {
        name        string
        pattern     string
        path        string
        shouldMatch bool
    }{
        {
            name:        "exact match",
            pattern:     "/users",
            path:        "/users",
            shouldMatch: true,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Best Practices

1. **Always use helpers**: Use `testutil.TempDir()`, `testutil.SetupTestDB()`, etc. instead of manual setup
2. **Automatic cleanup**: All helpers automatically clean up after tests
3. **Isolated tests**: Each test should be independent and not rely on other tests
4. **Fast by default**: Use in-memory SQLite for speed, PostgreSQL only when needed
5. **Descriptive names**: Test names should clearly describe what they test
6. **Table-driven**: Use table-driven tests for multiple similar scenarios
7. **Assert vs Require**: Use `assert` for non-fatal checks, `require` for fatal checks

## Future Enhancements

Potential additions for future phases:

- Mock filesystem utilities (afero integration)
- HTTP mock server helpers
- Snapshot testing utilities
- Performance benchmarking helpers
- Parallel test execution helpers
