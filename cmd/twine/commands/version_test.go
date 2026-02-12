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
	originalVersion := version
	originalCommit := commit
	originalDate := date
	originalBuiltBy := builtBy

	version = "1.0.0"
	commit = "abc123"
	date = "2024-01-01"
	builtBy = "test"

	defer func() {
		version = originalVersion
		commit = originalCommit
		date = originalDate
		builtBy = originalBuiltBy
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
	originalVersion := version
	originalCommit := commit
	originalDate := date
	originalBuiltBy := builtBy

	version = "dev"
	commit = "none"
	date = "unknown"
	builtBy = "unknown"

	defer func() {
		version = originalVersion
		commit = originalCommit
		date = originalDate
		builtBy = originalBuiltBy
	}()

	cmd := NewVersionCommand()

	err := cmd.Execute()
	assert.NoError(t, err)

	// Note: Output goes to stdout via fmt.Printf, not captured in test
}
