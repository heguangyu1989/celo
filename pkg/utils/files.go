package utils

import "os"

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
