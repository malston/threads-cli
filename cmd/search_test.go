package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestSearchFilterFlagsWiredToRequest(t *testing.T) {
	var gotQuery url.Values

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": []map[string]string{}})
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
		"search", "query",
		"--token", "test-token",
		"--author", "someuser",
		"--sort", "recent",
		"--mode", "tag",
		"--media-type", "image",
		"--since", "1700000000",
		"--until", "1700086400",
		"--format", "json",
	})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if gotQuery.Get("author_username") != "someuser" {
		t.Errorf("author_username = %q, want %q", gotQuery.Get("author_username"), "someuser")
	}
	if gotQuery.Get("search_type") != "RECENT" {
		t.Errorf("search_type = %q, want %q", gotQuery.Get("search_type"), "RECENT")
	}
	if gotQuery.Get("search_mode") != "TAG" {
		t.Errorf("search_mode = %q, want %q", gotQuery.Get("search_mode"), "TAG")
	}
	if gotQuery.Get("media_type") != "IMAGE" {
		t.Errorf("media_type = %q, want %q", gotQuery.Get("media_type"), "IMAGE")
	}
	if gotQuery.Get("since") != "1700000000" {
		t.Errorf("since = %q, want %q", gotQuery.Get("since"), "1700000000")
	}
	if gotQuery.Get("until") != "1700086400" {
		t.Errorf("until = %q, want %q", gotQuery.Get("until"), "1700086400")
	}
}

func resetSearchFlags() {
	searchCmd, _, _ := rootCmd.Find([]string{"search"})
	if searchCmd != nil {
		searchCmd.Flags().Set("sort", "")
		searchCmd.Flags().Set("mode", "")
		searchCmd.Flags().Set("media-type", "")
		searchCmd.Flags().Set("since", "0")
		searchCmd.Flags().Set("until", "0")
		searchCmd.Flags().Set("author", "")
	}
}

func TestSearchRejectsInvalidSort(t *testing.T) {
	t.Cleanup(resetSearchFlags)
	origDir := configDir
	configDir = t.TempDir()
	t.Cleanup(func() { configDir = origDir })

	cmd := rootCmd
	cmd.SetArgs([]string{"search", "query", "--token", "t", "--sort", "newest"})
	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --sort value")
	}
	if !strings.Contains(err.Error(), "sort") {
		t.Errorf("error should mention sort, got: %v", err)
	}
}

func TestSearchRejectsInvalidMediaType(t *testing.T) {
	t.Cleanup(resetSearchFlags)
	origDir := configDir
	configDir = t.TempDir()
	t.Cleanup(func() { configDir = origDir })

	cmd := rootCmd
	cmd.SetArgs([]string{"search", "query", "--token", "t", "--media-type", "audio"})
	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --media-type value")
	}
	if !strings.Contains(err.Error(), "media-type") {
		t.Errorf("error should mention media-type, got: %v", err)
	}
}

func TestSearchRejectsInvalidMode(t *testing.T) {
	t.Cleanup(resetSearchFlags)
	origDir := configDir
	configDir = t.TempDir()
	t.Cleanup(func() { configDir = origDir })

	cmd := rootCmd
	cmd.SetArgs([]string{"search", "query", "--token", "t", "--mode", "fuzzy"})
	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --mode value")
	}
	if !strings.Contains(err.Error(), "mode") {
		t.Errorf("error should mention mode, got: %v", err)
	}
}

func TestSearchRejectsSinceAfterUntil(t *testing.T) {
	t.Cleanup(resetSearchFlags)
	origDir := configDir
	configDir = t.TempDir()
	t.Cleanup(func() { configDir = origDir })

	cmd := rootCmd
	cmd.SetArgs([]string{"search", "query", "--token", "t", "--since", "2000000000", "--until", "1000000000"})
	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --since > --until")
	}
	if !strings.Contains(err.Error(), "since") {
		t.Errorf("error should mention since/until, got: %v", err)
	}
}
