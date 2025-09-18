package merge

import (
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/heguangyu1989/celo/pkg/config"
	"github.com/heguangyu1989/celo/pkg/p"
	"net/http"
	"strings"
)

func Merge(srcBranch string, targetBranch string, title string, tags []string) error {
	gitInfo, err := GetGitInfo()
	if err != nil {
		return err
	}

	data := map[string]string{
		"id":            gitInfo.GitPathInfo.Path2GitLabID(),
		"source_branch": srcBranch,
		"target_branch": targetBranch,
		"title":         title,
		"description":   "",
		"private_token": config.C.GitlabToken,
		"labels":        strings.Join(tags, ","),
	}

	reqUrl := fmt.Sprintf("%s://%s/api/v4/projects/%s/merge_requests", gitInfo.GitPathInfo.Scheme, gitInfo.GitPathInfo.Host, gitInfo.GitPathInfo.Path2GitLabID())
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
	return errors.New(fmt.Sprintf("request %s error with http code %d body %s", resp.Request.URL, resp.StatusCode(), resp.String()))
}
