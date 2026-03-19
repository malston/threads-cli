package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchCommandRegistered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"search"})
	if err != nil {
		t.Fatalf("search command not found: %v", err)
	}
	if cmd.Use != "search [query]" {
		t.Errorf("command Use = %q, want %q", cmd.Use, "search [query]")
	}
}

func TestSearchSendsCorrectRequest(t *testing.T) {
	response := map[string]any{
		"data": []map[string]string{
			{"id": "s-1", "text": "found it", "media_type": "TEXT", "username": "alice"},
			{"id": "s-2", "text": "also here", "media_type": "TEXT", "username": "bob"},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/keyword_search") {
			t.Errorf("expected path ending with /keyword_search, got %s", r.URL.Path)
		}
		q := r.URL.Query().Get("q")
		if q != "golang" {
			t.Errorf("expected q=golang, got %s", q)
		}
		fields := r.URL.Query().Get("fields")
		if fields == "" {
			t.Error("expected fields query param")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	origDir := configDir
	configDir = t.TempDir()
	t.Cleanup(func() { configDir = origDir })

	origBase := searchBaseURL
	searchBaseURL = ts.URL
	t.Cleanup(func() { searchBaseURL = origBase })

	cmd := rootCmd
	cmd.SetArgs([]string{"search", "golang", "--token", "test-token", "--format", "json"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "s-1") {
		t.Errorf("output should contain first result ID, got: %s", output)
	}
	if !strings.Contains(output, "s-2") {
		t.Errorf("output should contain second result ID, got: %s", output)
	}
}

func TestSearchAppliesPaginationFlags(t *testing.T) {
	response := map[string]any{
		"data": []map[string]string{},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "5" {
			t.Errorf("expected limit=5, got %s", limit)
		}
		after := r.URL.Query().Get("after")
		if after != "cursor-next" {
			t.Errorf("expected after=cursor-next, got %s", after)
		}
		before := r.URL.Query().Get("before")
		if before != "cursor-prev" {
			t.Errorf("expected before=cursor-prev, got %s", before)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	origDir := configDir
	configDir = t.TempDir()
	t.Cleanup(func() { configDir = origDir })

	origBase := searchBaseURL
	searchBaseURL = ts.URL
	t.Cleanup(func() { searchBaseURL = origBase })

	cmd := rootCmd
	cmd.SetArgs([]string{
		"search", "test-query",
		"--token", "test-token",
		"--limit", "5",
		"--after", "cursor-next",
		"--before", "cursor-prev",
		"--format", "json",
	})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
}
