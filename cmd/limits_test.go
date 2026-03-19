package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLimitsCommandRegistered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"limits"})
	if err != nil {
		t.Fatalf("limits command not found: %v", err)
	}
	if cmd.Use != "limits" {
		t.Errorf("command Use = %q, want %q", cmd.Use, "limits")
	}
}

func TestLimits(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"quota_usage":     10,
				"quota_total":     250,
				"quota_duration":  86400,
				"rate_limit_type": "POSTS",
			},
		},
	}

	var requestedPath string
	var requestedFields string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		requestedPath = r.URL.Path
		requestedFields = r.URL.Query().Get("fields")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	dir := t.TempDir()
	origDir := configDir
	configDir = dir
	t.Cleanup(func() { configDir = origDir })

	writeTestCredentials(t, dir, "user-99")

	origBase := limitsBaseURL
	limitsBaseURL = ts.URL
	t.Cleanup(func() { limitsBaseURL = origBase })

	cmd := rootCmd
	cmd.SetArgs([]string{"limits", "--token", "test-token", "--format", "json"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if requestedPath != "/user-99/threads_publishing_limit" {
		t.Errorf("expected path /user-99/threads_publishing_limit, got %s", requestedPath)
	}
	if !strings.Contains(requestedFields, "quota_usage") {
		t.Errorf("expected fields to contain quota_usage, got %s", requestedFields)
	}

	output := out.String()
	if !strings.Contains(output, "quota_usage") || !strings.Contains(output, "10") {
		t.Errorf("output should contain quota data, got: %s", output)
	}
}
