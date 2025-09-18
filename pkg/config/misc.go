package config

import (
	"os"
	"path/filepath"
)

func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(home, ".celo.yaml")
}
