package cmd

import (
	"path/filepath"
	"testing"

	"github.com/heguangyu1989/celo/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildInfoCommand(t *testing.T) {
	assert.NotNil(t, GetBuildInfoCmd())
	require.NoError(t, runBuildInfoCmd(GetBuildInfoCmd(), nil))
}

func TestGenDefaultCommand(t *testing.T) {
	old := config.C
	defer func() { config.C = old }()

	path := filepath.Join(t.TempDir(), "celo.yaml")
	config.C = config.Config{GitlabToken: "token"}
	cmd := GetGenDefaultCmd()
	require.NoError(t, cmd.Flags().Set("dst", path))

	require.NoError(t, runGenDefaultCmd(cmd, nil))
	assert.FileExists(t, path)

	cmd = GetGenDefaultCmd()
	require.NoError(t, cmd.Flags().Set("dst", filepath.Join(t.TempDir(), "celo.txt")))
	require.Error(t, runGenDefaultCmd(cmd, nil))
}

func TestMergeCommandValidation(t *testing.T) {
	cmd := GetMergeCommand()
	require.Error(t, runMergeCommand(cmd, nil))
}

func TestExecuteHelp(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})
	require.NoError(t, Execute())
	rootCmd.SetArgs(nil)
}
