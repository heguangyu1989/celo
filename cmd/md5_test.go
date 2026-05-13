package cmd

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const emptyMD5 = "d41d8cd98f00b204e9800998ecf8427e"

func TestCalDataInitHashesEmptyString(t *testing.T) {
	data := calData{Name: "", Type: calTypeString}

	err := data.Init()

	require.NoError(t, err)
	assert.Equal(t, emptyMD5, data.MD5)
}

func TestCalDataInitHashesEmptyFile(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "empty.txt")
	touchFile(t, filePath)
	data := calData{Name: filePath, Type: calTypeFile}

	err := data.Init()

	require.NoError(t, err)
	assert.Equal(t, emptyMD5, data.MD5)
}

func TestRunMD5CmdRejectsDirectory(t *testing.T) {
	cmd := GetMD5Cmd()

	err := runMD5Cmd(cmd, []string{t.TempDir()})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "directory")
}

func TestRunMD5CmdRejectsUnsupportedOutput(t *testing.T) {
	cmd := GetMD5Cmd()
	require.NoError(t, cmd.Flags().Set("output", "xml"))

	err := runMD5Cmd(cmd, []string{"hello"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported output format")
}
