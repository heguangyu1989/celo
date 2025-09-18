package merge

import (
	"fmt"
	"time"
)

type GitLabMergeRequest struct {
	ID             int       `json:"id"`
	Iid            int       `json:"iid"`
	ProjectID      int       `json:"project_id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	State          string    `json:"state"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	MergedBy       any       `json:"merged_by"`
	MergeUser      any       `json:"merge_user"`
	MergedAt       any       `json:"merged_at"`
	ClosedBy       any       `json:"closed_by"`
	ClosedAt       any       `json:"closed_at"`
	TargetBranch   string    `json:"target_branch"`
	SourceBranch   string    `json:"source_branch"`
	UserNotesCount int       `json:"user_notes_count"`
	Upvotes        int       `json:"upvotes"`
	Downvotes      int       `json:"downvotes"`
	Author         struct {
		ID        int    `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		State     string `json:"state"`
		Locked    bool   `json:"locked"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	} `json:"author"`
	Assignees                 []any  `json:"assignees"`
	Assignee                  any    `json:"assignee"`
	Reviewers                 []any  `json:"reviewers"`
	SourceProjectID           int    `json:"source_project_id"`
	TargetProjectID           int    `json:"target_project_id"`
	Labels                    []any  `json:"labels"`
	Draft                     bool   `json:"draft"`
	WorkInProgress            bool   `json:"work_in_progress"`
	Milestone                 any    `json:"milestone"`
	MergeWhenPipelineSucceeds bool   `json:"merge_when_pipeline_succeeds"`
	MergeStatus               string `json:"merge_status"`
	DetailedMergeStatus       string `json:"detailed_merge_status"`
	Sha                       string `json:"sha"`
	MergeCommitSha            any    `json:"merge_commit_sha"`
	SquashCommitSha           any    `json:"squash_commit_sha"`
	DiscussionLocked          any    `json:"discussion_locked"`
	ShouldRemoveSourceBranch  any    `json:"should_remove_source_branch"`
	ForceRemoveSourceBranch   any    `json:"force_remove_source_branch"`
	PreparedAt                any    `json:"prepared_at"`
	Reference                 string `json:"reference"`
	References                struct {
		Short    string `json:"short"`
		Relative string `json:"relative"`
		Full     string `json:"full"`
	} `json:"references"`
	WebURL    string `json:"web_url"`
	TimeStats struct {
		TimeEstimate        int `json:"time_estimate"`
		TotalTimeSpent      int `json:"total_time_spent"`
		HumanTimeEstimate   any `json:"human_time_estimate"`
		HumanTotalTimeSpent any `json:"human_total_time_spent"`
	} `json:"time_stats"`
	Squash               bool `json:"squash"`
	SquashOnMerge        bool `json:"squash_on_merge"`
	TaskCompletionStatus struct {
		Count          int `json:"count"`
		CompletedCount int `json:"completed_count"`
	} `json:"task_completion_status"`
	HasConflicts                bool `json:"has_conflicts"`
	BlockingDiscussionsResolved bool `json:"blocking_discussions_resolved"`
	Subscribed                  bool `json:"subscribed"`
	ChangesCount                any  `json:"changes_count"`
	LatestBuildStartedAt        any  `json:"latest_build_started_at"`
	LatestBuildFinishedAt       any  `json:"latest_build_finished_at"`
	FirstDeployedToProductionAt any  `json:"first_deployed_to_production_at"`
	Pipeline                    any  `json:"pipeline"`
	HeadPipeline                any  `json:"head_pipeline"`
	DiffRefs                    any  `json:"diff_refs"`
	MergeError                  any  `json:"merge_error"`
	User                        struct {
		CanMerge bool `json:"can_merge"`
	} `json:"user"`
}

func (gm GitLabMergeRequest) GetMergeRequestSlim() MergeRequestSlim {
	return MergeRequestSlim{
		MergeID:      fmt.Sprintf("%d", gm.Iid),
		SourceBranch: gm.SourceBranch,
		TargetBranch: gm.TargetBranch,
		Title:        gm.Title,
		Desc:         gm.Description,
		Url:          gm.WebURL,
	}
}

type GitLabIssue struct {
	ID          int       `json:"id"`
	Iid         int       `json:"iid"`
	ProjectID   int       `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ClosedAt    any       `json:"closed_at"`
	ClosedBy    any       `json:"closed_by"`
	Labels      []string  `json:"labels"`
	Milestone   any       `json:"milestone"`
	Assignees   []any     `json:"assignees"`
	Author      struct {
		ID        int    `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		State     string `json:"state"`
		Locked    bool   `json:"locked"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	} `json:"author"`
	Type               string `json:"type"`
	Assignee           any    `json:"assignee"`
	UserNotesCount     int    `json:"user_notes_count"`
	MergeRequestsCount int    `json:"merge_requests_count"`
	Upvotes            int    `json:"upvotes"`
	Downvotes          int    `json:"downvotes"`
	DueDate            any    `json:"due_date"`
	Confidential       bool   `json:"confidential"`
	DiscussionLocked   any    `json:"discussion_locked"`
	IssueType          string `json:"issue_type"`
	WebURL             string `json:"web_url"`
	TimeStats          struct {
		TimeEstimate        int `json:"time_estimate"`
		TotalTimeSpent      int `json:"total_time_spent"`
		HumanTimeEstimate   any `json:"human_time_estimate"`
		HumanTotalTimeSpent any `json:"human_total_time_spent"`
	} `json:"time_stats"`
	TaskCompletionStatus struct {
		Count          int `json:"count"`
		CompletedCount int `json:"completed_count"`
	} `json:"task_completion_status"`
	HasTasks   bool   `json:"has_tasks"`
	TaskStatus string `json:"task_status"`
	Links      struct {
		Self                string `json:"self"`
		Notes               string `json:"notes"`
		AwardEmoji          string `json:"award_emoji"`
		Project             string `json:"project"`
		ClosedAsDuplicateOf any    `json:"closed_as_duplicate_of"`
	} `json:"_links"`
	References struct {
		Short    string `json:"short"`
		Relative string `json:"relative"`
		Full     string `json:"full"`
	} `json:"references"`
	Severity           string `json:"severity"`
	Subscribed         bool   `json:"subscribed"`
	MovedToID          any    `json:"moved_to_id"`
	ServiceDeskReplyTo any    `json:"service_desk_reply_to"`
}

func (gi GitLabIssue) GetIssueSlim() IssueSlim {
	return IssueSlim{
		IssueID: fmt.Sprintf("%d", gi.Iid),
		Title:   gi.Title,
		Desc:    gi.Description,
		Url:     gi.WebURL,
		Labels:  gi.Labels,
	}
}
