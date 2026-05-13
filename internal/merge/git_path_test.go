package merge

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePathGitLabSSH(t *testing.T) {
	info, err := ParsePath("git@gitlab.example.com:group/project.git", GitTypeGitlab)

	require.NoError(t, err)
	assert.Equal(t, "https", info.Scheme)
	assert.Equal(t, "gitlab.example.com", info.Host)
	assert.Equal(t, "/group/project", info.Path)

	projectID, err := info.Path2GitLabID()
	require.NoError(t, err)
	assert.Equal(t, "group%2Fproject", projectID)
}

func TestParsePathGitLabHTTPS(t *testing.T) {
	info, err := ParsePath("https://gitlab.example.com/group/project.git", GitTypeGitlab)

	require.NoError(t, err)
	assert.Equal(t, "https", info.Scheme)
	assert.Equal(t, "gitlab.example.com", info.Host)
	assert.Equal(t, "/group/project", info.Path)
}

func TestParsePathRejectsMissingProjectPath(t *testing.T) {
	_, err := ParsePath("https://gitlab.example.com", GitTypeGitlab)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing project path")
}
