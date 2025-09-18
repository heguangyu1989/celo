package merge

import (
	"os/exec"
	"strings"
)

type GitInfo struct {
	Username    string      `json:"username"`
	Email       string      `json:"email"`
	GitPathInfo GitPathInfo `json:"git_path_info"`
}

func GetGitInfo() (GitInfo, error) {
	cmd := exec.Command("git", "config", "-l")
	out, err := cmd.Output()
	if err != nil {
		return GitInfo{}, err
	}
	info := GitInfo{}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "user.name=") {
			info.Username = strings.TrimSpace(strings.Replace(line, "user.name=", "", 1))
		}
		if strings.HasPrefix(line, "user.email=") {
			info.Email = strings.TrimSpace(strings.Replace(line, "user.email=", "", 1))
		}
		if strings.HasPrefix(line, "remote.origin.url") {
			info.GitPathInfo, err = ParsePath(strings.TrimSpace(strings.Replace(line, "remote.origin.url=", "", 1)), GitTypeGitlab)
			if err != nil {
				return GitInfo{}, err
			}
		}
	}
	return info, nil
}
