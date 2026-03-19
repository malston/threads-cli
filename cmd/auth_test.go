package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/malston/saved-threads/internal/config"
)

func TestAuthCommandRegistered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"auth"})
	if err != nil {
		t.Fatalf("auth command not found: %v", err)
	}
	if cmd.Use != "auth" {
		t.Errorf("command Use = %q, want %q", cmd.Use, "auth")
	}
}

func TestLoginCommandRegistered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"auth", "login"})
	if err != nil {
		t.Fatalf("auth login command not found: %v", err)
	}
	if cmd.Use != "login" {
		t.Errorf("command Use = %q, want %q", cmd.Use, "login")
	}
}

func TestTokenCommandRegistered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"auth", "token"})
	if err != nil {
		t.Fatalf("auth token command not found: %v", err)
	}
	if cmd.Use != "token" {
		t.Errorf("command Use = %q, want %q", cmd.Use, "token")
	}
}

func TestTokenRefreshFlagExists(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"auth", "token"})
	if err != nil {
		t.Fatalf("auth token command not found: %v", err)
	}
	flag := cmd.Flags().Lookup("refresh")
	if flag == nil {
		t.Fatal("--refresh flag not found on token command")
	}
	if flag.DefValue != "false" {
		t.Errorf("--refresh default = %q, want %q", flag.DefValue, "false")
	}
}

func TestRunAuthTokenShowsExpiryInfo(t *testing.T) {
	dir := t.TempDir()
	origDir := configDir
	configDir = dir
	t.Cleanup(func() { configDir = origDir })

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	creds := config.Credentials{
		AccessToken: "test-token-abc",
		TokenType:   "bearer",
		ExpiresAt:   expiresAt,
		UserID:      "12345",
	}
	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "credentials.json"), data, 0600); err != nil {
		t.Fatalf("write credentials: %v", err)
	}

	cmd := rootCmd
	cmd.SetArgs([]string{"auth", "token"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "valid") {
		t.Errorf("output should contain 'valid', got: %s", output)
	}
	if !strings.Contains(output, "test-t...") {
		t.Errorf("output should contain masked token 'test-t...', got: %s", output)
	}
}

func TestRunAuthTokenExpiredCredentials(t *testing.T) {
	dir := t.TempDir()
	origDir := configDir
	configDir = dir
	t.Cleanup(func() { configDir = origDir })

	creds := config.Credentials{
		AccessToken: "expired-token",
		TokenType:   "bearer",
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
		UserID:      "12345",
	}
	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "credentials.json"), data, 0600); err != nil {
		t.Fatalf("write credentials: %v", err)
	}

	cmd := rootCmd
	cmd.SetArgs([]string{"auth", "token"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "expired") {
		t.Errorf("output should contain 'expired', got: %s", output)
	}
}

func TestRunAuthTokenNoCredentials(t *testing.T) {
	dir := t.TempDir()
	origDir := configDir
	configDir = dir
	t.Cleanup(func() { configDir = origDir })

	cmd := rootCmd
	cmd.SetArgs([]string{"auth", "token"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no credentials exist")
	}
	if !strings.Contains(err.Error(), "no stored credentials") {
		t.Errorf("error should mention 'no stored credentials', got: %v", err)
	}
}

func TestRunAuthTokenRefreshCallsRefresh(t *testing.T) {
	dir := t.TempDir()
	origDir := configDir
	configDir = dir
	t.Cleanup(func() { configDir = origDir })

	// Token that expires within 7 days (needs refresh)
	creds := config.Credentials{
		AccessToken: "needs-refresh-token",
		TokenType:   "bearer",
		ExpiresAt:   time.Now().Add(3 * 24 * time.Hour),
		UserID:      "12345",
	}
	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "credentials.json"), data, 0600); err != nil {
		t.Fatalf("write credentials: %v", err)
	}

	refreshCalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refreshCalled = true
		resp := map[string]any{
			"access_token": "refreshed-token",
			"token_type":   "bearer",
			"expires_in":   5184000,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	origBaseURL := authBaseURL
	authBaseURL = ts.URL
	t.Cleanup(func() { authBaseURL = origBaseURL })

	cmd := rootCmd
	cmd.SetArgs([]string{"auth", "token", "--refresh"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if !refreshCalled {
		t.Error("refresh endpoint was not called")
	}

	output := out.String()
	if !strings.Contains(output, "refreshed") {
		t.Errorf("output should contain 'refreshed', got: %s", output)
	}

	// Verify the refreshed token was saved
	store := config.NewStore(dir)
	saved, err := store.Load()
	if err != nil {
		t.Fatalf("load saved credentials: %v", err)
	}
	if saved.AccessToken != "refreshed-token" {
		t.Errorf("saved token = %q, want %q", saved.AccessToken, "refreshed-token")
	}
}

func TestRunAuthTokenRefreshNotNeeded(t *testing.T) {
	dir := t.TempDir()
	origDir := configDir
	configDir = dir
	t.Cleanup(func() { configDir = origDir })

	// Token that doesn't need refresh (expires in 30 days)
	creds := config.Credentials{
		AccessToken: "good-token",
		TokenType:   "bearer",
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
		UserID:      "12345",
	}
	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "credentials.json"), data, 0600); err != nil {
		t.Fatalf("write credentials: %v", err)
	}

	cmd := rootCmd
	cmd.SetArgs([]string{"auth", "token", "--refresh"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "does not need refresh") {
		t.Errorf("output should contain 'does not need refresh', got: %s", output)
	}
}

func TestRunAuthLoginMissingEnvVars(t *testing.T) {
	t.Setenv("THREADS_APP_ID", "")
	t.Setenv("THREADS_APP_SECRET", "")

	cmd := rootCmd
	cmd.SetArgs([]string{"auth", "login"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when env vars are missing")
	}
	if !strings.Contains(err.Error(), "THREADS_APP_ID") {
		t.Errorf("error should mention THREADS_APP_ID, got: %v", err)
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abcdefghijklmnop", "abcdef..."},
		{"short", "sh..."},
		{"ab", "a..."},
		{"", "..."},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("len_%d", len(tt.input)), func(t *testing.T) {
			got := maskToken(tt.input)
			if got != tt.want {
				t.Errorf("maskToken(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
