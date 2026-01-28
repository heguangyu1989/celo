package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GitlabToken string `yaml:"gitlab_token" json:"gitlab_token"`
}

func NewDefaultConfig() Config {
	return Config{
		GitlabToken: "",
	}
}

var C = NewDefaultConfig()

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &C)
		if err != nil {
			return err
		}
		case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &C)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupport file ext : %s", ext)
	}
	return nil
}

func SaveConfig(path string) error {
	var err error
	var data []byte
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		data, err = json.Marshal(C)
		if err != nil {
			return err
		}
	case ".yaml", ".yml":
		data, err = yaml.Marshal(C)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupport file ext : %s", ext)
	}
	return os.WriteFile(path, data, 0644)
}
