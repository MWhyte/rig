package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds user preferences that persist across sessions.
type Config struct {
	Theme string `json:"theme"`
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	rigDir := filepath.Join(dir, "rig")
	if err := os.MkdirAll(rigDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(rigDir, "config.json"), nil
}

// Load reads the config file. Returns a default config if the file doesn't exist.
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return &Config{}, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return &Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &Config{}, err
	}
	return &cfg, nil
}

// Save writes the config to disk.
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
