package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInsightsCommandRegistered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"insights"})
	if err != nil {
		t.Fatalf("insights command not found: %v", err)
	}
	if cmd.Use != "insights" {
		t.Errorf("command Use = %q, want %q", cmd.Use, "insights")
	}
}

func TestInsightsSubcommandsRegistered(t *testing.T) {
	subs := []struct {
		name string
		use  string
	}{
		{"post", "post [id]"},
		{"user", "user"},
	}
	for _, tt := range subs {
		t.Run(tt.name, func(t *testing.T) {
			cmd, _, err := rootCmd.Find([]string{"insights", tt.name})
			if err != nil {
				t.Fatalf("insights %s command not found: %v", tt.name, err)
			}
			if cmd.Use != tt.use {
				t.Errorf("command Use = %q, want %q", cmd.Use, tt.use)
			}
		})
	}
}

func TestInsightsPost(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"name":   "views",
				"title":  "Views",
				"period": "lifetime",
				"values": []map[string]any{{"value": 42}},
			},
			{
				"name":   "likes",
				"title":  "Likes",
				"period": "lifetime",
				"values": []map[string]any{{"value": 7}},
			},
		},
	}

	var requestedPath string
	var requestedMetric string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		requestedPath = r.URL.Path
		requestedMetric = r.URL.Query().Get("metric")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	origDir := configDir
	configDir = t.TempDir()
	t.Cleanup(func() { configDir = origDir })

	origBase := insightsBaseURL
	insightsBaseURL = ts.URL
	t.Cleanup(func() { insightsBaseURL = origBase })

	cmd := rootCmd
	cmd.SetArgs([]string{"insights", "post", "post-123", "--token", "test-token", "--metrics", "views,likes", "--format", "json"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if requestedPath != "/post-123/insights" {
		t.Errorf("expected path /post-123/insights, got %s", requestedPath)
	}
	if !strings.Contains(requestedMetric, "views") || !strings.Contains(requestedMetric, "likes") {
		t.Errorf("expected metric param to contain views and likes, got %s", requestedMetric)
	}

	output := out.String()
	if !strings.Contains(output, "views") {
		t.Errorf("output should contain 'views', got: %s", output)
	}
	if !strings.Contains(output, "42") {
		t.Errorf("output should contain value 42, got: %s", output)
	}
}

func TestInsightsUser(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"name":   "followers_count",
				"title":  "Followers",
				"period": "day",
				"values": []map[string]any{{"value": 100}},
			},
		},
	}

	var requestedPath string
	var requestedMetric string
	var requestedSince string
	var requestedUntil string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		requestedPath = r.URL.Path
		requestedMetric = r.URL.Query().Get("metric")
		requestedSince = r.URL.Query().Get("since")
		requestedUntil = r.URL.Query().Get("until")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	dir := t.TempDir()
	origDir := configDir
	configDir = dir
	t.Cleanup(func() { configDir = origDir })

	writeTestCredentials(t, dir, "user-42")

	origBase := insightsBaseURL
	insightsBaseURL = ts.URL
	t.Cleanup(func() { insightsBaseURL = origBase })

	cmd := rootCmd
	cmd.SetArgs([]string{
		"insights", "user",
		"--token", "test-token",
		"--metrics", "followers_count",
		"--since", "1700000000",
		"--until", "1700100000",
		"--format", "json",
	})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if requestedPath != "/user-42/threads_insights" {
		t.Errorf("expected path /user-42/threads_insights, got %s", requestedPath)
	}
	if !strings.Contains(requestedMetric, "followers_count") {
		t.Errorf("expected metric param to contain followers_count, got %s", requestedMetric)
	}
	if requestedSince != "1700000000" {
		t.Errorf("expected since=1700000000, got %s", requestedSince)
	}
	if requestedUntil != "1700100000" {
		t.Errorf("expected until=1700100000, got %s", requestedUntil)
	}

	output := out.String()
	if !strings.Contains(output, "followers_count") {
		t.Errorf("output should contain 'followers_count', got: %s", output)
	}
}
