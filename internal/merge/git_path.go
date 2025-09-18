package merge

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const (
	GitTypeGitlab = "gitlab"
)

type GitPathInfo struct {
	RawInput string
	Scheme   string
	Host     string // host or host:port (see Hostname and Port methods)
	Path     string // path (relative paths may omit leading slash)
}

func (p GitPathInfo) Path2GitLabID() string {
	return url.QueryEscape(p.Path[1:])
}
func ParsePath(inputUrl string, srcType string) (GitPathInfo, error) {
	switch srcType {
	case GitTypeGitlab:
		return parseGitlabPath(inputUrl)
	}
	return GitPathInfo{}, errors.New(fmt.Sprintf("can not match src type : %s", srcType))
}

func parseGitlabPath(input string) (GitPathInfo, error) {
	ret := GitPathInfo{
		RawInput: input,
	}

	useInput := input
	useInput = strings.TrimSpace(useInput)

	{
		if strings.HasSuffix(useInput, ".wiki.git") {
			useInput = useInput[:len(useInput)-9]
		} else if strings.HasSuffix(useInput, ".git") {
			useInput = useInput[:len(useInput)-4]
		} else if strings.Contains(useInput, "/-/wiki") {
			ii := strings.Index(useInput, "/-/wiki")
			useInput = useInput[:ii]
		}

		if strings.HasSuffix(useInput, "/") {
			useInput = useInput[0:(len(useInput) - 1)]
		}
	}

	{
		if strings.HasPrefix(useInput, "git@") {
			// git协议的地址
			i1 := strings.Index(useInput, ":")
			if i1 < 0 {
				return ret, getParseErr(useInput, "git can not find :")
			}
			ret.Scheme = "https"
			ret.Host = useInput[4:i1]
			ret.Path = "/" + useInput[(i1+1):]
		} else {
			u, err := url.Parse(useInput)
			if err != nil {
				return ret, getParseErr(input, err.Error())
			}
			ret.Scheme = u.Scheme
			ret.Host = u.Host
			ret.Path = u.Path
		}
	}

	return ret, nil
}

func getParseErr(input string, msg string) error {
	return errors.New(fmt.Sprintf("parse input error %s , %s", input, msg))
}
