package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Config contains the configuration for this CLI.
// It is used to create a Client for the Pharos API server.
type Config struct {
	BaseURL       string `json:"base_url"`
	AWSProfile    string `json:"aws_profile"`
	AssumeRoleARN string `json:"assume_role_arn"`
	filePath      string
}

const (
	directoryPermissions = 0700
)

// New creates a new Config reference at the given file path.
// Defaults to creating a Config reference at $HOME/.kube/pharos/config.
func New(pharosConfig string) (*Config, error) {
	if pharosConfig == "" {
		pharosConfig = fmt.Sprintf("%s/.kube/pharos/config", os.Getenv("HOME"))
	}
	dir := filepath.Dir(pharosConfig)
	err := os.MkdirAll(dir, directoryPermissions)
	if err != nil {
		return nil, err
	}
	return &Config{filePath: pharosConfig}, nil
}

// Load loads data from the config file into the Config struct.
func (c *Config) Load() error {
	if _, err := os.Stat(c.filePath); os.IsNotExist(err) {
		return errors.New("pharos hasn't been configured yet")
	}
	raw, err := ioutil.ReadFile(c.filePath)
	if err != nil {
		return err
	}
	if string(raw) == "" {
		return errors.New("pharos hasn't been configured yet")
	}

	return json.Unmarshal(raw, c)
}
