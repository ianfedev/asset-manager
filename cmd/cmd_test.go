package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCmdStructure(t *testing.T) {
	// Verify RootCmd exists
	assert.NotNil(t, RootCmd)
	assert.Equal(t, "asset-manager", RootCmd.Use)

	// Verify subcommands
	commands := RootCmd.Commands()
	cmdMap := make(map[string]bool)
	for _, c := range commands {
		cmdMap[c.Use] = true
	}

	assert.True(t, cmdMap["start"], "start command should be registered")
	assert.True(t, cmdMap["integrity"], "integrity command should be registered")
}

func TestIntegrityCmdStructure(t *testing.T) {
	// Verify integrity subcommands
	assert.NotNil(t, integrityCmd)

	commands := integrityCmd.Commands()
	cmdMap := make(map[string]bool)
	for _, c := range commands {
		cmdMap[c.Use] = true
	}

	assert.True(t, cmdMap["furniture"], "furniture command should be registered")
	assert.True(t, cmdMap["structure"], "structure command should be registered")
	assert.True(t, cmdMap["bundle"], "bundle command should be registered")
	assert.True(t, cmdMap["gamedata"], "gamedata command should be registered")
	assert.True(t, cmdMap["server"], "server command should be registered")
}

func TestFlags(t *testing.T) {
	// Verify flags
	fixFlag := structureCmd.Flags().Lookup("fix")
	assert.NotNil(t, fixFlag)

	jsonFlag := furnitureCmd.Flags().Lookup("json")
	assert.NotNil(t, jsonFlag)
}
