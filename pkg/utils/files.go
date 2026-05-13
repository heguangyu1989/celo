package utils

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func FileExists(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
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
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == ".env" {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".env") {
			relPath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return err
			}
			ret = append(ret, relPath)
		}
		return nil
	})
	return ret, err
}
