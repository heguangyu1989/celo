package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigJSONRoundTrip(t *testing.T) {
	old := C
	defer func() { C = old }()

	path := filepath.Join(t.TempDir(), "celo.json")
	C = Config{GitlabToken: "token-json"}

	require.NoError(t, SaveConfig(path))
	C = Config{}
	require.NoError(t, LoadConfig(path))

	assert.Equal(t, "token-json", C.GitlabToken)
}

func TestConfigYAMLRoundTrip(t *testing.T) {
	old := C
	defer func() { C = old }()

	path := filepath.Join(t.TempDir(), "celo.yaml")
	C = Config{GitlabToken: "token-yaml"}

	require.NoError(t, SaveConfig(path))
	C = Config{}
	require.NoError(t, LoadConfig(path))

	assert.Equal(t, "token-yaml", C.GitlabToken)
}

func TestConfigUnsupportedExtension(t *testing.T) {
	old := C
	defer func() { C = old }()

	path := filepath.Join(t.TempDir(), "celo.txt")
	require.NoError(t, os.WriteFile(path, []byte("gitlab_token: token"), 0600))

	assert.Error(t, SaveConfig(path))
	assert.Error(t, LoadConfig(path))
}

func TestNewDefaultConfigAndDefaultPath(t *testing.T) {
	assert.Equal(t, Config{}, NewDefaultConfig())
	assert.Contains(t, DefaultPath(), ".celo.yaml")
}
