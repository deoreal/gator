// Package config provides configuration information
package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (cfg *Config) SetUser(user string) {
	cfg.CurrentUserName = user
}

func ReadConfig() (*Config, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return &Config{}, fmt.Errorf("failed to read HomeDir: %s", err)
	}
	file, err := os.Open(homedir + "/" + configFileName)
	if err != nil {
		return &Config{}, fmt.Errorf("failed to read config file: %s", configFileName)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return &Config{}, fmt.Errorf("failed to read data from config file: %s", configFileName)
	}

	cfg := Config{}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return &Config{}, fmt.Errorf("failed to read config into json object")
	}

	return &cfg, nil
}

func WriteConfig(cfg Config) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to read HomeDir: %s", err)
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal data into json object")
	}

	err = os.WriteFile(homedir+"/"+configFileName, data, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %s - %s", configFileName, err)
	}
	return nil
}

