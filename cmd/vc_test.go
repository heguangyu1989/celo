package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformCleanupUpdatesLRUAfterRemovingCLIVersion(t *testing.T) {
	root := t.TempDir()
	cliDir := filepath.Join(root, "cli", "servers")
	require.NoError(t, os.MkdirAll(cliDir, 0755))

	keepVersion := "Stable-keep"
	removeVersion := "Stable-remove"
	removePath := filepath.Join(cliDir, removeVersion)
	otherPath := filepath.Join(root, "cache.tmp")
	require.NoError(t, os.MkdirAll(removePath, 0755))
	touchFile(t, filepath.Join(removePath, "server"))
	touchFile(t, otherPath)

	lruPath := filepath.Join(cliDir, "lru.json")
	writeLRU(t, lruPath, []string{keepVersion, removeVersion})
	analysis := &analysisResult{
		LRUPath: lruPath,
		Items: []cleanupItem{
			{Path: removePath, Size: 1, Description: "CLI版本: " + removeVersion},
			{Path: otherPath, Size: 1, Description: "临时文件: cache.tmp"},
		},
	}

	cleaned, err := performCleanup(analysis, 1)

	require.NoError(t, err)
	assert.EqualValues(t, 2, cleaned)
	assert.Equal(t, []string{keepVersion}, readLRU(t, lruPath))
	assert.FileExists(t, lruPath+".backup")
}

func TestPerformCleanupRejectsNegativeKeepCount(t *testing.T) {
	_, err := performCleanup(&analysisResult{}, -1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "greater than or equal to 0")
}

func writeLRU(t *testing.T, path string, versions []string) {
	t.Helper()

	data, err := json.Marshal(versions)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0644))
}

func readLRU(t *testing.T, path string) []string {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var versions []string
	require.NoError(t, json.Unmarshal(data, &versions))
	return versions
}
