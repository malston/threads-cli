package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProfileCommandRegistered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"profile"})
	if err != nil {
		t.Fatalf("profile command not found: %v", err)
	}
	if cmd.Use != "profile [username]" {
		t.Errorf("command Use = %q, want %q", cmd.Use, "profile [username]")
	}
}

func TestProfileNoArgs(t *testing.T) {
	profile := map[string]any{
		"id":                          "user-42",
		"username":                    "testuser",
		"name":                        "Test User",
		"threads_biography":           "Hello world",
		"threads_profile_picture_url": "https://example.com/pic.jpg",
		"is_verified":                 true,
	}

	var requestedPath string
	var requestedFields string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		requestedFields = r.URL.Query().Get("fields")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(profile)
	}))
	defer ts.Close()

	dir := t.TempDir()
	origDir := configDir
	configDir = dir
	t.Cleanup(func() { configDir = origDir })

	writeTestCredentials(t, dir, "user-42")

	origBase := profileBaseURL
	profileBaseURL = ts.URL
	t.Cleanup(func() { profileBaseURL = origBase })

	cmd := rootCmd
	cmd.SetArgs([]string{"profile", "--token", "test-token", "--format", "json"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if requestedPath != "/user-42" {
		t.Errorf("expected request path /user-42, got %s", requestedPath)
	}
	if requestedFields == "" {
		t.Error("expected fields query param")
	}

	output := out.String()
	if !strings.Contains(output, "user-42") {
		t.Errorf("output should contain user ID, got: %s", output)
	}
	if !strings.Contains(output, "testuser") {
		t.Errorf("output should contain username, got: %s", output)
	}
}

func TestProfileWithUsername(t *testing.T) {
	profile := map[string]any{
		"username":            "someone",
		"name":                "Someone Famous",
		"biography":           "A bio",
		"profile_picture_url": "https://example.com/someone.jpg",
		"follower_count":      1000,
		"is_verified":         false,
	}

	var requestedPath string
	var requestedUsername string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		requestedUsername = r.URL.Query().Get("username")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(profile)
	}))
	defer ts.Close()

	origDir := configDir
	configDir = t.TempDir()
	t.Cleanup(func() { configDir = origDir })

	origBase := profileBaseURL
	profileBaseURL = ts.URL
	t.Cleanup(func() { profileBaseURL = origBase })

	cmd := rootCmd
	cmd.SetArgs([]string{"profile", "someone", "--token", "test-token", "--format", "json"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if requestedPath != "/profile_lookup" {
		t.Errorf("expected request path /profile_lookup, got %s", requestedPath)
	}
	if requestedUsername != "someone" {
		t.Errorf("expected username=someone, got %s", requestedUsername)
	}

	output := out.String()
	if !strings.Contains(output, "someone") {
		t.Errorf("output should contain username, got: %s", output)
	}
	if !strings.Contains(output, "Someone Famous") {
		t.Errorf("output should contain name, got: %s", output)
	}
}
