package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds user preferences that persist across sessions.
type Config struct {
	Theme  string `json:"theme"`
	Volume *int   `json:"volume,omitempty"`
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	rigDir := filepath.Join(dir, "rig")
	if err := os.MkdirAll(rigDir, 0o750); err != nil {
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

	data, err := os.ReadFile(path) //nolint:gosec // path is derived from os.UserConfigDir, not user input
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
	return os.WriteFile(path, data, 0o600)
}

// SetTheme persists the theme, preserving other config fields.
func SetTheme(theme string) error {
	cfg, _ := Load()
	cfg.Theme = theme
	return Save(cfg)
}

// SetVolume persists the volume (0-100), preserving other config fields.
func SetVolume(vol int) error {
	cfg, _ := Load()
	cfg.Volume = &vol
	return Save(cfg)
}
