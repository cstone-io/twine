package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUpdateCommand(t *testing.T) {
	cmd := NewUpdateCommand()

	assert.Equal(t, "update", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}

func TestUpdateCommandFlags(t *testing.T) {
	cmd := NewUpdateCommand()

	// Check that all required flags are present
	versionFlag := cmd.Flags().Lookup("version")
	assert.NotNil(t, versionFlag)
	assert.Equal(t, "string", versionFlag.Value.Type())

	checkFlag := cmd.Flags().Lookup("check")
	assert.NotNil(t, checkFlag)
	assert.Equal(t, "bool", checkFlag.Value.Type())

	listFlag := cmd.Flags().Lookup("list")
	assert.NotNil(t, listFlag)
	assert.Equal(t, "bool", listFlag.Value.Type())

	yesFlag := cmd.Flags().Lookup("yes")
	assert.NotNil(t, yesFlag)
	assert.Equal(t, "bool", yesFlag.Value.Type())

	// Check shorthand
	yesFlagShorthand := cmd.Flags().ShorthandLookup("y")
	assert.NotNil(t, yesFlagShorthand)
	assert.Equal(t, "yes", yesFlagShorthand.Name)
}

func TestUpdateCommandExamples(t *testing.T) {
	cmd := NewUpdateCommand()

	// Verify the Long description contains examples
	assert.Contains(t, cmd.Long, "twine update")
	assert.Contains(t, cmd.Long, "twine update --check")
	assert.Contains(t, cmd.Long, "twine update --list")
	assert.Contains(t, cmd.Long, "twine update --version")
	assert.Contains(t, cmd.Long, "twine update --yes")
}

func TestUpdateCommandValidation(t *testing.T) {
	cmd := NewUpdateCommand()

	// Verify command is properly configured
	assert.False(t, cmd.Args != nil, "Command should not require args")
	assert.NotNil(t, cmd.RunE, "Command should have RunE handler")
}
