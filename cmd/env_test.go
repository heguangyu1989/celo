package cmd

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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

func TestEnvListValidationBranches(t *testing.T) {
	require.Error(t, runEnvListCmd(GetEnvListCmd(), []string{filepath.Join(t.TempDir(), "missing")}))

	filePath := filepath.Join(t.TempDir(), "file")
	touchFile(t, filePath)
	require.Error(t, runEnvListCmd(GetEnvListCmd(), []string{filePath}))

	require.NoError(t, runEnvListCmd(GetEnvListCmd(), []string{t.TempDir()}))
}

func TestEnvModelAndDelegate(t *testing.T) {
	it := item(".env.local")
	assert.Equal(t, "", it.FilterValue())

	delegate := itemDelegate{}
	assert.Equal(t, 1, delegate.Height())
	assert.Equal(t, 0, delegate.Spacing())
	assert.Nil(t, delegate.Update(nil, nil))
	delegate.Render(io.Discard, list.Model{}, 0, it)

	m := model{}
	assert.Nil(t, m.Init())
	view := m.View()
	assert.Contains(t, view, "\n")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.NotNil(t, cmd)
	assert.Contains(t, updated.(model).View(), "nothing changed")

	updated, cmd = model{list: list.New([]list.Item{it}, itemDelegate{}, 10, 1)}.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, cmd)
	assert.Contains(t, updated.(model).View(), "create link")

	updated, cmd = model{list: list.New([]list.Item{it}, itemDelegate{}, 10, 1)}.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	assert.Nil(t, cmd)
	assert.IsType(t, model{}, updated)
}
