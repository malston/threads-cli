package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// AppConfig holds Threads app credentials.
type AppConfig struct {
	AppID       string `json:"app_id"`
	AppSecret   string `json:"app_secret"`
	RedirectURI string `json:"redirect_uri"`
}

const configFile = "config.json"

// LoadAppConfig loads app configuration from environment variables or a JSON file.
// Environment variables (THREADS_APP_ID, THREADS_APP_SECRET) take precedence.
// Both must be set for env-based config to be used. When loaded from env,
// RedirectURI defaults to https://localhost:8888/callback.
// Falls back to {dir}/config.json if env vars are not set.
func LoadAppConfig(dir string) (*AppConfig, error) {
	appID := os.Getenv("THREADS_APP_ID")
	appSecret := os.Getenv("THREADS_APP_SECRET")

	if appID != "" || appSecret != "" {
		if appID == "" || appSecret == "" {
			return nil, fmt.Errorf("both THREADS_APP_ID and THREADS_APP_SECRET must be set")
		}
		return &AppConfig{
			AppID:       appID,
			AppSecret:   appSecret,
			RedirectURI: "https://localhost:8888/callback",
		}, nil
	}

	data, err := os.ReadFile(filepath.Join(dir, configFile))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("no app configuration found: set THREADS_APP_ID and THREADS_APP_SECRET, or create %s", filepath.Join(dir, configFile))
		}
		return nil, err
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.AppID == "" {
		return nil, fmt.Errorf("app_id is required in %s", filepath.Join(dir, configFile))
	}

	return &cfg, nil
}

// SaveAppConfig writes app configuration to {dir}/config.json.
// Creates the directory with 0700 permissions and the file with 0600.
func SaveAppConfig(dir string, cfg *AppConfig) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, configFile), data, 0600)
}
