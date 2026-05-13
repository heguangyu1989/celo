package merge

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGitInfoWithFakeGit(t *testing.T) {
	prependFakeGit(t, `#!/bin/sh
echo 'user.name=alice'
echo 'user.email=alice@example.com'
echo 'remote.origin.url=git@gitlab.example.com:group/project.git'
`)

	info, err := GetGitInfo()

	require.NoError(t, err)
	assert.Equal(t, "alice", info.Username)
	assert.Equal(t, "alice@example.com", info.Email)
	assert.Equal(t, "gitlab.example.com", info.GitPathInfo.Host)
}

func TestGetGitInfoMissingRemote(t *testing.T) {
	prependFakeGit(t, `#!/bin/sh
echo 'user.name=alice'
`)

	_, err := GetGitInfo()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "remote.origin.url not found")
}
