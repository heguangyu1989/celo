package merge

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/heguangyu1989/celo/pkg/config"
	"github.com/heguangyu1989/celo/pkg/p"
)

func Merge(srcBranch string, targetBranch string, title string, tags []string) error {
	gitInfo, err := GetGitInfo()
	if err != nil {
		return err
	}
	if config.C.GitlabToken == "" {
		return fmt.Errorf("gitlab_token must be set in config")
	}
	projectID, err := gitInfo.GitPathInfo.Path2GitLabID()
	if err != nil {
		return err
	}

	data := map[string]string{
		"id":            projectID,
		"source_branch": srcBranch,
		"target_branch": targetBranch,
		"title":         title,
		"description":   "",
		"private_token": config.C.GitlabToken,
		"labels":        strings.Join(tags, ","),
	}

	reqUrl := fmt.Sprintf("%s://%s/api/v4/projects/%s/merge_requests", gitInfo.GitPathInfo.Scheme, gitInfo.GitPathInfo.Host, projectID)
	p.Info(fmt.Sprintf("gitlabCreateMR request : %s", reqUrl))
	client := resty.New()
	createResp := GitLabMergeRequest{}
	resp, err := client.R().SetResult(&createResp).
		SetBody(data).
		Post(reqUrl)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusCreated {
		return getHttpStatusErr(resp)
	}
	p.Info(fmt.Sprintf("url : %s", createResp.GetMergeRequestSlim().Url))
	return nil
}

func getHttpStatusErr(resp *resty.Response) error {
	return fmt.Errorf("request %s error with http code %d body %s", resp.Request.URL, resp.StatusCode(), resp.String())
}
