package utils

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func FileExists(path string) bool {
	f, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if f.IsDir() {
		return false
	}
	return true
}

func FindAllEnvFiles(rootPath string) ([]string, error) {
	ret := make([]string, 0)
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if d.Name() == ".env" {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".env") {
			ret = append(ret, d.Name())
		}
		return nil
	})
	return ret, err
}
