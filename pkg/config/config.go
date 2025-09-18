package config

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
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
	if ext == ".json" {
		err = json.Unmarshal(data, &C)
		if err != nil {
			return err
		}
	} else if ext == ".yaml" || ext == ".yml" {
		err = yaml.Unmarshal(data, &C)
	} else {
		return fmt.Errorf("unsupport file ext : %s", ext)
	}
	return nil
}

func SaveConfig(path string) error {
	var err error
	var data []byte
	ext := filepath.Ext(path)
	if ext == ".json" {
		data, err = json.Marshal(C)
		if err != nil {
			return err
		}
	} else if ext == ".yaml" || ext == ".yml" {
		data, err = yaml.Marshal(C)
	} else {
		return fmt.Errorf("unsupport file ext : %s", ext)
	}
	return os.WriteFile(path, data, 0644)
}
