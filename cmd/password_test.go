package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunPasswordCmdJSONAndYAML(t *testing.T) {
	cmd := GetPasswordCmd()
	require.NoError(t, cmd.Flags().Set("length", "8"))
	require.NoError(t, cmd.Flags().Set("count", "2"))
	require.NoError(t, cmd.Flags().Set("output", "json"))
	require.NoError(t, runPasswordCmd(cmd, nil))

	cmd = GetPasswordCmd()
	require.NoError(t, cmd.Flags().Set("length", "8"))
	require.NoError(t, cmd.Flags().Set("output", "yaml"))
	require.NoError(t, runPasswordCmd(cmd, nil))
}

func TestRunPasswordCmdTableAndValidation(t *testing.T) {
	cmd := GetPasswordCmd()
	require.NoError(t, cmd.Flags().Set("length", "6"))
	require.NoError(t, cmd.Flags().Set("custom", "abc123"))
	require.NoError(t, cmd.Flags().Set("output", "table"))
	require.NoError(t, runPasswordCmd(cmd, nil))

	cmd = GetPasswordCmd()
	require.NoError(t, cmd.Flags().Set("length", "0"))
	require.Error(t, runPasswordCmd(cmd, nil))

	cmd = GetPasswordCmd()
	require.NoError(t, cmd.Flags().Set("count", "0"))
	require.Error(t, runPasswordCmd(cmd, nil))

	cmd = GetPasswordCmd()
	require.NoError(t, cmd.Flags().Set("output", "xml"))
	err := runPasswordCmd(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported output format")

	cmd = GetPasswordCmd()
	require.NoError(t, cmd.Flags().Set("count", "2"))
	require.NoError(t, cmd.Flags().Set("output", "table"))
	require.NoError(t, runPasswordCmd(cmd, nil))

	cmd = GetPasswordCmd()
	require.NoError(t, cmd.Flags().Set("upper", "false"))
	require.NoError(t, cmd.Flags().Set("lower", "false"))
	require.NoError(t, cmd.Flags().Set("digits", "false"))
	require.Error(t, runPasswordCmd(cmd, nil))
}
