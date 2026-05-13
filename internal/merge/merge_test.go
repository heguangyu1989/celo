package merge

import (
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/heguangyu1989/celo/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeRejectsEmptyTokenBeforeRequest(t *testing.T) {
	old := config.C
	defer func() { config.C = old }()
	config.C = config.Config{}
	prependFakeGit(t, `#!/bin/sh
echo 'remote.origin.url=git@gitlab.example.com:group/project.git'
`)

	err := Merge("feature", "main", "title", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "gitlab_token")
}

func TestMergeReturnsGitInfoError(t *testing.T) {
	prependFakeGit(t, `#!/bin/sh
exit 1
`)

	err := Merge("feature", "main", "title", nil)

	require.Error(t, err)
}

func TestHTTPStatusError(t *testing.T) {
	resp := &resty.Response{
		Request:     &resty.Request{URL: "https://gitlab.example.com/api"},
		RawResponse: &http.Response{StatusCode: http.StatusInternalServerError},
	}
	err := getHttpStatusErr(resp)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "http code")
}
