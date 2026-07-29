package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/heguangyu1989/celo/pkg/p"
	"github.com/spf13/cobra"
)

func GetSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Switch back to master/main, pull latest, and prune deleted branches",
		Long: `Switch the current git repository back to the master or main branch,
pull the latest changes, and clean up local branches whose remote
tracking branch has been deleted.

Equivalent to:
  git checkout master (or main)
  git pull
  git fetch -p
  git branch -d <branches whose upstream is gone>`,
		Args: cobra.NoArgs,
		RunE: runSyncCmd,
	}
	return cmd
}

func runSyncCmd(cmd *cobra.Command, args []string) error {
	if _, err := gitOutput("rev-parse", "--is-inside-work-tree"); err != nil {
		return fmt.Errorf("not inside a git repository")
	}

	branch, err := detectDefaultBranch()
	if err != nil {
		return err
	}

	p.Info(fmt.Sprintf("→ Checking out %s ...", branch))
	if err := runGit("checkout", branch); err != nil {
		return fmt.Errorf("git checkout %s failed: %w", branch, err)
	}

	p.Info("→ Pulling latest changes ...")
	if err := runGit("pull"); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}

	p.Info("→ Pruning deleted remote branches ...")
	if err := runGit("fetch", "-p"); err != nil {
		return fmt.Errorf("git fetch -p failed: %w", err)
	}

	deleted, err := deleteGoneBranches(branch)
	if err != nil {
		return err
	}

	if len(deleted) > 0 {
		p.Success(fmt.Sprintf("✓ Done. Deleted %d stale local branch(es): %s",
			len(deleted), strings.Join(deleted, ", ")))
	} else {
		p.Success(fmt.Sprintf("✓ Done. Now on %s, no stale local branches.", branch))
	}
	return nil
}

// detectDefaultBranch returns "master" or "main", whichever exists locally
// (checked first) or on the origin remote.
func detectDefaultBranch() (string, error) {
	for _, ref := range []string{"refs/heads/", "refs/remotes/origin/"} {
		for _, branch := range []string{"master", "main"} {
			if err := exec.Command("git", "show-ref", "--verify", "--quiet", ref+branch).Run(); err == nil {
				return branch, nil
			}
		}
	}
	return "", fmt.Errorf("no master or main branch found")
}

// deleteGoneBranches deletes local branches whose upstream tracking branch no
// longer exists on the remote, and returns the names of the deleted branches.
func deleteGoneBranches(currentBranch string) ([]string, error) {
	out, err := gitOutput("branch", "-vv")
	if err != nil {
		return nil, fmt.Errorf("git branch -vv failed: %w", err)
	}

	deleted := make([]string, 0)
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, ": gone]") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		// Lines for the current branch start with "*", branches checked out in
		// another worktree start with "+"; the branch name follows the marker.
		name := fields[0]
		if name == "*" || name == "+" {
			if len(fields) < 2 {
				continue
			}
			name = fields[1]
		}
		if name == currentBranch {
			continue
		}

		p.Info(fmt.Sprintf("→ Deleting stale local branch %s ...", name))
		if err := runGit("branch", "-d", name); err != nil {
			p.Error(fmt.Sprintf("✗ Skip %s: not fully merged, use git branch -D to force", name))
			continue
		}
		deleted = append(deleted, name)
	}
	return deleted, nil
}

// runGit runs a git command with stdout/stderr attached to the terminal.
func runGit(args ...string) error {
	c := exec.Command("git", args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// gitOutput runs a git command and returns its trimmed stdout.
func gitOutput(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	return strings.TrimSpace(string(out)), err
}
