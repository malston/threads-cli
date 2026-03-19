package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppConfigFromEnvVars(t *testing.T) {
	t.Setenv("THREADS_APP_ID", "env-app-id")
	t.Setenv("THREADS_APP_SECRET", "env-app-secret")

	dir := t.TempDir()
	cfg, err := LoadAppConfig(dir)
	if err != nil {
		t.Fatalf("LoadAppConfig: %v", err)
	}

	if cfg.AppID != "env-app-id" {
		t.Errorf("AppID = %q, want %q", cfg.AppID, "env-app-id")
	}
	if cfg.AppSecret != "env-app-secret" {
		t.Errorf("AppSecret = %q, want %q", cfg.AppSecret, "env-app-secret")
	}
	if cfg.RedirectURI != "http://localhost:8888/callback" {
		t.Errorf("RedirectURI = %q, want %q", cfg.RedirectURI, "http://localhost:8888/callback")
	}
}

func TestLoadAppConfigFromEnvVarsOnlyOneSet(t *testing.T) {
	t.Setenv("THREADS_APP_ID", "env-app-id")
	// THREADS_APP_SECRET intentionally not set

	dir := t.TempDir()
	_, err := LoadAppConfig(dir)
	if err == nil {
		t.Fatal("expected error when only THREADS_APP_ID is set")
	}
}

func TestLoadAppConfigFromFile(t *testing.T) {
	dir := t.TempDir()

	cfg := &AppConfig{
		AppID:       "file-app-id",
		AppSecret:   "file-app-secret",
		RedirectURI: "http://example.com/callback",
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := LoadAppConfig(dir)
	if err != nil {
		t.Fatalf("LoadAppConfig: %v", err)
	}

	if got.AppID != "file-app-id" {
		t.Errorf("AppID = %q, want %q", got.AppID, "file-app-id")
	}
	if got.AppSecret != "file-app-secret" {
		t.Errorf("AppSecret = %q, want %q", got.AppSecret, "file-app-secret")
	}
	if got.RedirectURI != "http://example.com/callback" {
		t.Errorf("RedirectURI = %q, want %q", got.RedirectURI, "http://example.com/callback")
	}
}

func TestLoadAppConfigEnvVarsPreferredOverFile(t *testing.T) {
	t.Setenv("THREADS_APP_ID", "env-app-id")
	t.Setenv("THREADS_APP_SECRET", "env-app-secret")

	dir := t.TempDir()

	cfg := &AppConfig{
		AppID:       "file-app-id",
		AppSecret:   "file-app-secret",
		RedirectURI: "http://example.com/callback",
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := LoadAppConfig(dir)
	if err != nil {
		t.Fatalf("LoadAppConfig: %v", err)
	}

	if got.AppID != "env-app-id" {
		t.Errorf("AppID = %q, want %q (env should take precedence)", got.AppID, "env-app-id")
	}
	if got.AppSecret != "env-app-secret" {
		t.Errorf("AppSecret = %q, want %q (env should take precedence)", got.AppSecret, "env-app-secret")
	}
}

func TestLoadAppConfigNothingAvailable(t *testing.T) {
	dir := t.TempDir()

	_, err := LoadAppConfig(dir)
	if err == nil {
		t.Fatal("expected error when no config source is available")
	}
}

func TestSaveAndLoadAppConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()

	want := &AppConfig{
		AppID:       "round-trip-id",
		AppSecret:   "round-trip-secret",
		RedirectURI: "http://localhost:9999/callback",
	}

	if err := SaveAppConfig(dir, want); err != nil {
		t.Fatalf("SaveAppConfig: %v", err)
	}

	got, err := LoadAppConfig(dir)
	if err != nil {
		t.Fatalf("LoadAppConfig: %v", err)
	}

	if got.AppID != want.AppID {
		t.Errorf("AppID = %q, want %q", got.AppID, want.AppID)
	}
	if got.AppSecret != want.AppSecret {
		t.Errorf("AppSecret = %q, want %q", got.AppSecret, want.AppSecret)
	}
	if got.RedirectURI != want.RedirectURI {
		t.Errorf("RedirectURI = %q, want %q", got.RedirectURI, want.RedirectURI)
	}
}

func TestSaveAppConfigCreatesDirectory(t *testing.T) {
	base := t.TempDir()
	nested := filepath.Join(base, "subdir", "config")

	cfg := &AppConfig{
		AppID:       "test-id",
		AppSecret:   "test-secret",
		RedirectURI: "http://localhost:8888/callback",
	}

	if err := SaveAppConfig(nested, cfg); err != nil {
		t.Fatalf("SaveAppConfig: %v", err)
	}

	info, err := os.Stat(nested)
	if err != nil {
		t.Fatalf("Stat directory: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected a directory")
	}

	fileInfo, err := os.Stat(filepath.Join(nested, "config.json"))
	if err != nil {
		t.Fatalf("Stat file: %v", err)
	}
	perm := fileInfo.Mode().Perm()
	if perm != 0600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}
