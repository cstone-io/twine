package testutil_test

import (
	"testing"

	"github.com/cstone-io/twine/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestModel is a simple test model for database tests
// Note: We don't use default UUID generation here because SQLite
// doesn't support gen_random_uuid(). Tests explicitly set IDs.
type TestModel struct {
	ID    uuid.UUID `gorm:"type:uuid;primary_key"`
	Name  string
	Email string
}

func TestSetupTestDB_CreatesDatabase(t *testing.T) {
	db := testutil.SetupTestDB(t)
	assert.NotNil(t, db)

	// Verify we can interact with the database
	var result int
	err := db.Raw("SELECT 1").Scan(&result).Error
	require.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestSetupTestDB_AutoMigrate(t *testing.T) {
	db := testutil.SetupTestDB(t)

	// Verify we can run migrations
	err := db.AutoMigrate(&TestModel{})
	assert.NoError(t, err)

	// Verify table was created by attempting to query it
	var count int64
	err = db.Model(&TestModel{}).Count(&count).Error
	require.NoError(t, err, "table should exist after migration")
	assert.Equal(t, int64(0), count, "table should be empty")
}

func TestAutoMigrate_CreatesTable(t *testing.T) {
	db := testutil.SetupTestDB(t)

	testutil.AutoMigrate(t, db, &TestModel{})

	// Verify we can create a record
	record := TestModel{
		ID:    uuid.New(),
		Name:  "Test",
		Email: "test@example.com",
	}
	err := db.Create(&record).Error
	assert.NoError(t, err)
}

func TestSeedTestData_InsertsRecords(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.AutoMigrate(t, db, &TestModel{})

	// Seed data
	records := []TestModel{
		{ID: uuid.New(), Name: "User 1", Email: "user1@example.com"},
		{ID: uuid.New(), Name: "User 2", Email: "user2@example.com"},
	}
	testutil.SeedTestData(t, db, &records)

	// Verify records were inserted
	var count int64
	err := db.Model(&TestModel{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestTruncateTable_RemovesRecords(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.AutoMigrate(t, db, &TestModel{})

	// Insert records
	records := []TestModel{
		{ID: uuid.New(), Name: "User 1", Email: "user1@example.com"},
		{ID: uuid.New(), Name: "User 2", Email: "user2@example.com"},
	}
	testutil.SeedTestData(t, db, &records)

	// Truncate
	testutil.TruncateTable(t, db, "test_models")

	// Verify records were removed
	var count int64
	err := db.Model(&TestModel{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRunInTransaction_RollsBackChanges(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.AutoMigrate(t, db, &TestModel{})

	// Run operation in transaction
	testutil.RunInTransaction(t, db, func(tx *gorm.DB) {
		record := TestModel{
			ID:    uuid.New(),
			Name:  "Test User",
			Email: "test@example.com",
		}
		err := tx.Create(&record).Error
		require.NoError(t, err)

		// Verify record exists in transaction
		var count int64
		err = tx.Model(&TestModel{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	// Verify record was rolled back
	var count int64
	err := db.Model(&TestModel{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "record should be rolled back")
}

func TestAssertRecordExists_FindsRecords(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.AutoMigrate(t, db, &TestModel{})

	// Insert record
	record := TestModel{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	}
	testutil.SeedTestData(t, db, &record)

	// Assert record exists
	testutil.AssertRecordExists(t, db, &TestModel{}, "email = ?", "test@example.com")
}

func TestAssertRecordNotExists_VerifiesAbsence(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.AutoMigrate(t, db, &TestModel{})

	// Assert record doesn't exist
	testutil.AssertRecordNotExists(t, db, &TestModel{}, "email = ?", "nonexistent@example.com")
}

func TestAssertRecordCount_CountsRecords(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.AutoMigrate(t, db, &TestModel{})

	// Insert records
	records := []TestModel{
		{ID: uuid.New(), Name: "User 1", Email: "user1@example.com"},
		{ID: uuid.New(), Name: "User 2", Email: "user2@example.com"},
		{ID: uuid.New(), Name: "User 3", Email: "user3@example.com"},
	}
	testutil.SeedTestData(t, db, &records)

	// Assert count
	testutil.AssertRecordCount(t, db, &TestModel{}, 3, "1 = ?", 1)
}
