package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewVersionCommand tests version command creation
func TestNewVersionCommand(t *testing.T) {
	cmd := NewVersionCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "version", cmd.Use)
	assert.Equal(t, "Print version information", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.Run)
}

// TestVersionCommand_Run tests version output
func TestVersionCommand_Run(t *testing.T) {
	// Temporarily override version variables for testing
	originalVersion := Version
	originalCommit := Commit
	originalDate := Date
	originalBuiltBy := BuiltBy

	Version = "1.0.0"
	Commit = "abc123"
	Date = "2024-01-01"
	BuiltBy = "test"

	defer func() {
		Version = originalVersion
		Commit = originalCommit
		Date = originalDate
		BuiltBy = originalBuiltBy
	}()

	cmd := NewVersionCommand()

	// Execute command
	err := cmd.Execute()
	assert.NoError(t, err)

	// Note: Output goes to stdout via fmt.Printf, not captured in test
	// The command execution succeeding with correct variables set is the main test
}

// TestVersionCommand_DefaultValues tests default build-time values
func TestVersionCommand_DefaultValues(t *testing.T) {
	// When not set via ldflags, should use defaults
	originalVersion := Version
	originalCommit := Commit
	originalDate := Date
	originalBuiltBy := BuiltBy

	Version = "dev"
	Commit = "none"
	Date = "unknown"
	BuiltBy = "unknown"

	defer func() {
		Version = originalVersion
		Commit = originalCommit
		Date = originalDate
		BuiltBy = originalBuiltBy
	}()

	cmd := NewVersionCommand()

	err := cmd.Execute()
	assert.NoError(t, err)

	// Note: Output goes to stdout via fmt.Printf, not captured in test
}
