package testutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cstone-io/twine/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTempDir_CreatesDirectory(t *testing.T) {
	dir := testutil.TempDir(t)

	// Verify directory exists
	assert.DirExists(t, dir)

	// Verify directory is in temp location
	assert.Contains(t, dir, "twine-test-")
}

func TestTempDir_CleansUpAfterTest(t *testing.T) {
	var tempDir string

	// Run a sub-test to create temp dir
	t.Run("create temp dir", func(t *testing.T) {
		tempDir = testutil.TempDir(t)
		assert.DirExists(t, tempDir)
	})

	// After sub-test completes, directory should be cleaned up
	assert.NoDirExists(t, tempDir, "temp directory should be cleaned up after test")
}

func TestSetupTestEnv_SetsAndRestoresEnv(t *testing.T) {
	// Save original values
	originalValue := os.Getenv("TEST_VAR")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("TEST_VAR")
		} else {
			os.Setenv("TEST_VAR", originalValue)
		}
	}()

	// Set test environment in sub-test
	t.Run("with test env", func(t *testing.T) {
		testutil.SetupTestEnv(t, map[string]string{
			"TEST_VAR": "test_value",
		})

		// Verify value is set
		assert.Equal(t, "test_value", os.Getenv("TEST_VAR"))
	})

	// After sub-test, environment should be restored
	assert.Equal(t, originalValue, os.Getenv("TEST_VAR"))
}

func TestMkdirAll_CreatesNestedDirectories(t *testing.T) {
	tempDir := testutil.TempDir(t)

	testutil.MkdirAll(t, tempDir, "a/b/c", "x/y/z")

	// Verify directories were created
	assert.DirExists(t, filepath.Join(tempDir, "a", "b", "c"))
	assert.DirExists(t, filepath.Join(tempDir, "x", "y", "z"))
}

func TestCreateFile_CreatesFileWithContent(t *testing.T) {
	tempDir := testutil.TempDir(t)
	filePath := filepath.Join(tempDir, "subdir", "test.txt")

	testutil.CreateFile(t, filePath, "test content")

	// Verify file exists
	assert.FileExists(t, filePath)

	// Verify content
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, "test content", string(content))
}

func TestFileExists_DetectsFiles(t *testing.T) {
	tempDir := testutil.TempDir(t)
	existingFile := filepath.Join(tempDir, "exists.txt")
	nonExistingFile := filepath.Join(tempDir, "notexists.txt")

	// Create file
	testutil.CreateFile(t, existingFile, "content")

	// Test
	assert.True(t, testutil.FileExists(t, existingFile))
	assert.False(t, testutil.FileExists(t, nonExistingFile))
}

func TestDirExists_DetectsDirectories(t *testing.T) {
	tempDir := testutil.TempDir(t)
	existingDir := filepath.Join(tempDir, "exists")
	nonExistingDir := filepath.Join(tempDir, "notexists")

	// Create directory
	testutil.MkdirAll(t, tempDir, "exists")

	// Test
	assert.True(t, testutil.DirExists(t, existingDir))
	assert.False(t, testutil.DirExists(t, nonExistingDir))
}

func TestWriteFixture_CreatesFileInTempDir(t *testing.T) {
	tempDir := testutil.TempDir(t)
	data := []byte("fixture data")

	path := testutil.WriteFixture(t, tempDir, "fixture.txt", data)

	// Verify file was created
	assert.FileExists(t, path)

	// Verify content
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, data, content)
}

func TestLoadJSONFixture_UnmarshalsJSON(t *testing.T) {
	tempDir := testutil.TempDir(t)

	// Create a JSON fixture
	jsonData := `{"name": "test", "value": 123}`
	testutil.WriteFixture(t, tempDir, "test.json", []byte(jsonData))

	// Change working directory to temp dir for relative path loading
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create fixtures directory structure
	os.MkdirAll("internal/testutil/fixtures", 0755)
	os.WriteFile("internal/testutil/fixtures/test.json", []byte(jsonData), 0644)

	// Load and unmarshal
	var result struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	testutil.LoadJSONFixture(t, "test.json", &result)

	// Verify
	assert.Equal(t, "test", result.Name)
	assert.Equal(t, 123, result.Value)
}
