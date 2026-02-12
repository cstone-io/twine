package testutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates a test database. By default, it creates an in-memory
// SQLite database for fast, isolated testing. If POSTGRES_TEST_DSN environment
// variable is set, it will use a PostgreSQL database instead.
//
// The database is automatically closed when the test completes.
//
// Example usage:
//
//	db := testutil.SetupTestDB(t)
//	err := db.AutoMigrate(&User{})
//	require.NoError(t, err)
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	var db *gorm.DB
	var err error

	// Check if POSTGRES_TEST_DSN is set for manual Postgres testing
	if dsn := os.Getenv("POSTGRES_TEST_DSN"); dsn != "" {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err, "failed to connect to PostgreSQL test database")
	} else {
		// Default: use SQLite in-memory database
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err, "failed to create in-memory SQLite database")
	}

	// Cleanup: close database connection when test completes
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	})

	return db
}

// SeedTestData seeds the test database with the provided data.
// The data parameter should be a pointer to a struct or slice of structs.
//
// Example usage:
//
//	users := []User{
//	    {Email: "user1@example.com"},
//	    {Email: "user2@example.com"},
//	}
//	testutil.SeedTestData(t, db, &users)
func SeedTestData(t *testing.T, db *gorm.DB, data interface{}) {
	t.Helper()
	err := db.Create(data).Error
	require.NoError(t, err, "failed to seed test data")
}

// TruncateTable truncates the specified table in the database.
// Useful for cleaning up between test cases.
//
// Example usage:
//
//	testutil.TruncateTable(t, db, "users")
func TruncateTable(t *testing.T, db *gorm.DB, tableName string) {
	t.Helper()
	err := db.Exec("DELETE FROM " + tableName).Error
	require.NoError(t, err, "failed to truncate table: %s", tableName)
}

// RunInTransaction runs the provided function within a database transaction
// that is automatically rolled back after the function completes. This is
// useful for testing database operations without persisting changes.
//
// Example usage:
//
//	testutil.RunInTransaction(t, db, func(tx *gorm.DB) {
//	    user := User{Email: "test@example.com"}
//	    err := tx.Create(&user).Error
//	    require.NoError(t, err)
//	    // Changes are rolled back after this function returns
//	})
func RunInTransaction(t *testing.T, db *gorm.DB, fn func(tx *gorm.DB)) {
	t.Helper()
	tx := db.Begin()
	require.NoError(t, tx.Error, "failed to begin transaction")

	defer func() {
		tx.Rollback()
	}()

	fn(tx)
}

// AutoMigrate runs database migrations for the provided models.
// This is a convenience wrapper around GORM's AutoMigrate.
//
// Example usage:
//
//	testutil.AutoMigrate(t, db, &User{}, &Post{})
func AutoMigrate(t *testing.T, db *gorm.DB, models ...interface{}) {
	t.Helper()
	err := db.AutoMigrate(models...)
	require.NoError(t, err, "failed to run migrations")
}

// AssertRecordExists asserts that a record exists in the database with the
// given conditions.
//
// Example usage:
//
//	testutil.AssertRecordExists(t, db, &User{}, "email = ?", "test@example.com")
func AssertRecordExists(t *testing.T, db *gorm.DB, model interface{}, where string, args ...interface{}) {
	t.Helper()
	var count int64
	err := db.Model(model).Where(where, args...).Count(&count).Error
	require.NoError(t, err, "failed to query database")
	require.Greater(t, count, int64(0), "expected record to exist but found none")
}

// AssertRecordNotExists asserts that no record exists in the database with
// the given conditions.
//
// Example usage:
//
//	testutil.AssertRecordNotExists(t, db, &User{}, "email = ?", "deleted@example.com")
func AssertRecordNotExists(t *testing.T, db *gorm.DB, model interface{}, where string, args ...interface{}) {
	t.Helper()
	var count int64
	err := db.Model(model).Where(where, args...).Count(&count).Error
	require.NoError(t, err, "failed to query database")
	require.Equal(t, int64(0), count, "expected no records but found %d", count)
}

// AssertRecordCount asserts that the database contains the expected number
// of records matching the given conditions.
//
// Example usage:
//
//	testutil.AssertRecordCount(t, db, &User{}, 5, "active = ?", true)
func AssertRecordCount(t *testing.T, db *gorm.DB, model interface{}, expectedCount int64, where string, args ...interface{}) {
	t.Helper()
	var count int64
	err := db.Model(model).Where(where, args...).Count(&count).Error
	require.NoError(t, err, "failed to query database")
	require.Equal(t, expectedCount, count, "unexpected record count")
}
