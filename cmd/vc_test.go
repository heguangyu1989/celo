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

func TestAnalyzeVSCodeDirAndHelpers(t *testing.T) {
	root := t.TempDir()
	cliDir := filepath.Join(root, "cli", "servers")
	require.NoError(t, os.MkdirAll(filepath.Join(cliDir, "Stable-keep"), 0755))
	removeDir := filepath.Join(cliDir, "Stable-remove")
	require.NoError(t, os.MkdirAll(removeDir, 0755))
	writeFile(t, filepath.Join(removeDir, "server"), "server")
	writeLRU(t, filepath.Join(cliDir, "lru.json"), []string{"Stable-keep", "Stable-remove"})

	cacheDir := filepath.Join(root, "data", "CachedExtensionVSIXs")
	require.NoError(t, os.MkdirAll(cacheDir, 0755))
	writeFile(t, filepath.Join(cacheDir, "ext.vsix"), "vsix")

	tmpFile := filepath.Join(root, "data", "file.tmp")
	writeFile(t, tmpFile, "tmp")

	logsPath := filepath.Join(root, "data", "logs")
	require.NoError(t, os.MkdirAll(logsPath, 0755))
	logFile := filepath.Join(logsPath, "large.log")
	require.NoError(t, os.WriteFile(logFile, []byte("x"), 0644))
	require.NoError(t, os.Truncate(logFile, 11*1024*1024))

	cliLog := filepath.Join(root, ".cli.test.log")
	require.NoError(t, os.WriteFile(cliLog, []byte("x"), 0644))
	require.NoError(t, os.Truncate(cliLog, 101*1024))

	analysis, err := analyzeVSCodeDir(root, 1)

	require.NoError(t, err)
	assert.NotZero(t, analysis.TotalSize)
	assert.NotEmpty(t, analysis.Items)
	require.NoError(t, displayAnalysis(analysis))

	size, err := getDirSize(root)
	require.NoError(t, err)
	assert.Positive(t, size)
	assert.Equal(t, "0 bytes", formatSize(0))
	assert.Equal(t, "1 KB", formatSize(1024))
	assert.Equal(t, "1.0 GB", formatSize(1024*1024*1024))
}

func TestAnalyzeCLIVersionsMissingLRUAndNegativeKeep(t *testing.T) {
	items, err := analyzeCLIVersions(t.TempDir(), filepath.Join(t.TempDir(), "missing-lru.json"), 1)
	require.NoError(t, err)
	assert.Empty(t, items)

	_, err = analyzeCLIVersions(t.TempDir(), filepath.Join(t.TempDir(), "missing-lru.json"), -1)
	require.Error(t, err)
}

func TestUpdateLRUFileNoopAndErrors(t *testing.T) {
	root := t.TempDir()
	lruPath := filepath.Join(root, "lru.json")
	writeLRU(t, lruPath, []string{"Stable-one"})
	require.NoError(t, updateLRUFile(lruPath, 2))

	require.Error(t, updateLRUFile(filepath.Join(root, "missing.json"), 1))

	badPath := filepath.Join(root, "bad.json")
	require.NoError(t, os.WriteFile(badPath, []byte("{"), 0644))
	require.Error(t, updateLRUFile(badPath, 1))
}

func TestVCCleanCommandRunE(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	vscodeDir := filepath.Join(home, ".vscode-server")
	cliDir := filepath.Join(vscodeDir, "cli", "servers")
	require.NoError(t, os.MkdirAll(filepath.Join(cliDir, "Stable-old"), 0755))
	writeFile(t, filepath.Join(cliDir, "Stable-old", "server"), "server")
	writeLRU(t, filepath.Join(cliDir, "lru.json"), []string{"Stable-new", "Stable-old"})

	cmd := getVCCleanCmd()
	require.NoError(t, cmd.Flags().Set("yes", "true"))
	require.NoError(t, cmd.Flags().Set("keep", "1"))
	require.NoError(t, cmd.RunE(cmd, nil))

	cmd = getVCCleanCmd()
	require.NoError(t, cmd.Flags().Set("keep", "-1"))
	require.Error(t, cmd.RunE(cmd, nil))
}

func TestVCSkillAllCommandRunEWithFakeTools(t *testing.T) {
	prependFakeCommand(t, "ps", `#!/bin/sh
echo 'user 123 0.0 0.0 ? ? ? ? ? vscode-server'
`)
	prependFakeCommand(t, "grep", `#!/bin/sh
cat
`)
	prependFakeCommand(t, "kill", `#!/bin/sh
exit 0
`)

	cmd := getVCSkillAllCmd()
	require.NoError(t, cmd.RunE(cmd, nil))
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

func writeFile(t *testing.T, path string, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
}
