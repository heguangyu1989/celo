package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runGitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, out)
}

// setupSyncTestRepo creates a git repo with a bare origin remote on the main
// branch, plus a local branch "feature" that has been deleted on the remote.
// It returns the repo path (remote "feature" deletion is not yet fetched).
func setupSyncTestRepo(t *testing.T, defaultBranch string) string {
	t.Helper()
	tmp := t.TempDir()

	remote := filepath.Join(tmp, "remote.git")
	runGitIn(t, tmp, "init", "--bare", "-b", defaultBranch, remote)

	repo := filepath.Join(tmp, "repo")
	runGitIn(t, tmp, "init", "-b", defaultBranch, repo)
	runGitIn(t, repo, "config", "user.email", "test@example.com")
	runGitIn(t, repo, "config", "user.name", "test")
	runGitIn(t, repo, "config", "commit.gpgsign", "false")

	require.NoError(t, os.WriteFile(filepath.Join(repo, "README.md"), []byte("hi"), 0o644))
	runGitIn(t, repo, "add", ".")
	runGitIn(t, repo, "commit", "-m", "init")
	runGitIn(t, repo, "remote", "add", "origin", remote)
	runGitIn(t, repo, "push", "-u", "origin", defaultBranch)

	// Push a feature branch, then delete it on the remote.
	runGitIn(t, repo, "checkout", "-b", "feature")
	runGitIn(t, repo, "push", "-u", "origin", "feature")
	runGitIn(t, repo, "checkout", defaultBranch)
	runGitIn(t, repo, "push", "origin", ":feature")

	return repo
}

func TestSyncCommand(t *testing.T) {
	repo := setupSyncTestRepo(t, "main")
	t.Chdir(repo)

	require.NoError(t, runSyncCmd(GetSyncCmd(), nil))

	// Should be on main.
	out, err := gitOutput("branch", "--show-current")
	require.NoError(t, err)
	assert.Equal(t, "main", out)

	// The stale feature branch should be deleted.
	out, err = gitOutput("branch", "--list", "feature")
	require.NoError(t, err)
	assert.Empty(t, out)

	// Remote tracking ref should be pruned.
	err = exec.Command("git", "show-ref", "--verify", "--quiet", "refs/remotes/origin/feature").Run()
	assert.Error(t, err)
}

func TestSyncCommandMasterBranch(t *testing.T) {
	repo := setupSyncTestRepo(t, "master")
	t.Chdir(repo)

	require.NoError(t, runSyncCmd(GetSyncCmd(), nil))

	out, err := gitOutput("branch", "--show-current")
	require.NoError(t, err)
	assert.Equal(t, "master", out)
}

func TestSyncCommandNotGitRepo(t *testing.T) {
	t.Chdir(t.TempDir())
	require.Error(t, runSyncCmd(GetSyncCmd(), nil))
}

func TestDetectDefaultBranch(t *testing.T) {
	repo := setupSyncTestRepo(t, "main")
	t.Chdir(repo)

	branch, err := detectDefaultBranch()
	require.NoError(t, err)
	assert.Equal(t, "main", branch)
}
