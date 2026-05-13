package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwitchEnvReplacesRegularEnvFile(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, ".env.local")
	touchFile(t, filepath.Join(root, ".env"))
	touchFile(t, target)

	err := switchEnv(root, ".env.local")

	require.NoError(t, err)
	linkTarget, err := os.Readlink(filepath.Join(root, ".env"))
	require.NoError(t, err)
	assert.Equal(t, target, linkTarget)
}

func TestSwitchEnvReplacesBrokenSymlink(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, ".env.local")
	touchFile(t, target)
	require.NoError(t, os.Symlink(filepath.Join(root, ".env.missing"), filepath.Join(root, ".env")))

	err := switchEnv(root, ".env.local")

	require.NoError(t, err)
	linkTarget, err := os.Readlink(filepath.Join(root, ".env"))
	require.NoError(t, err)
	assert.Equal(t, target, linkTarget)
}

func TestSwitchEnvKeepsCurrentEnvWhenTargetMissing(t *testing.T) {
	root := t.TempDir()
	current := filepath.Join(root, ".env")
	touchFile(t, current)

	err := switchEnv(root, ".env.missing")

	require.Error(t, err)
	assert.FileExists(t, current)
}
