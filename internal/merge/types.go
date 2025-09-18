package merge

type MergeRequestSlim struct {
	MergeID      string `json:"merge_id"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	Title        string `json:"title"`
	Desc         string `json:"description"`
	Url          string `json:"url"`
}

type IssueSlim struct {
	IssueID string   `json:"issue_id"`
	Title   string   `json:"title"`
	Desc    string   `json:"description"`
	Url     string   `json:"url"`
	Labels  []string `json:"labels"`
}
