package merge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitLabMergeRequestSlim(t *testing.T) {
	mr := GitLabMergeRequest{
		Iid:          12,
		SourceBranch: "feature",
		TargetBranch: "main",
		Title:        "Add feature",
		Description:  "desc",
		WebURL:       "https://gitlab.example.com/mr/12",
	}

	slim := mr.GetMergeRequestSlim()

	assert.Equal(t, "12", slim.MergeID)
	assert.Equal(t, "feature", slim.SourceBranch)
	assert.Equal(t, "main", slim.TargetBranch)
	assert.Equal(t, "Add feature", slim.Title)
	assert.Equal(t, "desc", slim.Desc)
	assert.Equal(t, "https://gitlab.example.com/mr/12", slim.Url)
}

func TestGitLabIssueSlim(t *testing.T) {
	issue := GitLabIssue{
		Iid:         7,
		Title:       "Bug",
		Description: "Fix bug",
		WebURL:      "https://gitlab.example.com/issues/7",
		Labels:      []string{"bug", "urgent"},
	}

	slim := issue.GetIssueSlim()

	assert.Equal(t, "7", slim.IssueID)
	assert.Equal(t, "Bug", slim.Title)
	assert.Equal(t, "Fix bug", slim.Desc)
	assert.Equal(t, "https://gitlab.example.com/issues/7", slim.Url)
	assert.Equal(t, []string{"bug", "urgent"}, slim.Labels)
}
