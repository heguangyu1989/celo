package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeImageName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"repository only", "nginx", "nginx:latest"},
		{"repository with tag", "nginx:1.25", "nginx:1.25"},
		{"registry port without tag", "registry.local:5000/team/app", "registry.local:5000/team/app:latest"},
		{"registry port with tag", "registry.local:5000/team/app:v1", "registry.local:5000/team/app:v1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeImageName(tt.in))
		})
	}
}

func TestExtractErrorMessage(t *testing.T) {
	tests := []struct {
		output string
		want   string
	}{
		{"manifest unknown: manifest unknown", "Manifest unknown (image not found)"},
		{"no such manifest: repo:tag", "Manifest not found"},
		{"unauthorized: access denied", "Unauthorized (check credentials)"},
		{"connection refused", "Registry unreachable"},
		{"first line\nsecond line", "first line"},
		{"", "Check failed"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, extractErrorMessage(tt.output))
		})
	}
}

func TestDockerCommandsAndTableOutput(t *testing.T) {
	assert.NotNil(t, GetDockerCmd())
	assert.NotNil(t, getDockerCheckCmd())

	output := captureStdout(t, func() {
		printCheckResultsTable([]imageCheckResult{
			{Image: "nginx:latest", Exists: true, Message: "Image exists"},
			{Image: "missing:latest", Exists: false, Message: "Manifest not found"},
		})
	})
	assert.Contains(t, output, "nginx:latest")
	assert.Contains(t, output, "Manifest not found")
}

func TestRunDockerCheckCmdHelpWhenNoArgs(t *testing.T) {
	err := runDockerCheckCmd(getDockerCheckCmd(), nil)
	require.NoError(t, err)
}

func TestDockerCheckWithFakeDocker(t *testing.T) {
	prependFakeCommand(t, "docker", `#!/bin/sh
if [ "$3" = "exists:latest" ]; then
  echo '{"schemaVersion": 2}'
  exit 0
fi
echo 'manifest unknown'
exit 1
`)

	result := checkImageExists("exists")
	assert.True(t, result.Exists)
	assert.Equal(t, "Image exists", result.Message)

	result = checkImageExists("missing")
	assert.False(t, result.Exists)
	assert.Equal(t, "Manifest unknown (image not found)", result.Message)

	cmd := getDockerCheckCmd()
	require.NoError(t, cmd.Flags().Set("output", "json"))
	require.NoError(t, runDockerCheckCmd(cmd, []string{"exists"}))

	cmd = getDockerCheckCmd()
	require.NoError(t, cmd.Flags().Set("output", "yaml"))
	require.NoError(t, runDockerCheckCmd(cmd, []string{"missing"}))

	cmd = getDockerCheckCmd()
	require.NoError(t, cmd.Flags().Set("output", "xml"))
	require.Error(t, runDockerCheckCmd(cmd, []string{"exists"}))

	cmd = getDockerCheckCmd()
	require.NoError(t, cmd.Flags().Set("output", "table"))
	require.NoError(t, runDockerCheckCmd(cmd, []string{"exists", "missing"}))
}
