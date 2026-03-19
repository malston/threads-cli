package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReplyCommandRegistered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"reply"})
	if err != nil {
		t.Fatalf("reply command not found: %v", err)
	}
	if cmd.Use != "reply" {
		t.Errorf("command Use = %q, want %q", cmd.Use, "reply")
	}
}

func TestReplySubcommandsRegistered(t *testing.T) {
	subs := []struct {
		name string
		use  string
	}{
		{"list", "list [post-id]"},
		{"create", "create [post-id] [text]"},
		{"hide", "hide [reply-id]"},
		{"unhide", "unhide [reply-id]"},
	}
	for _, tt := range subs {
		t.Run(tt.name, func(t *testing.T) {
			cmd, _, err := rootCmd.Find([]string{"reply", tt.name})
			if err != nil {
				t.Fatalf("reply %s command not found: %v", tt.name, err)
			}
			if cmd.Use != tt.use {
				t.Errorf("command Use = %q, want %q", cmd.Use, tt.use)
			}
		})
	}
}

func TestReplyList(t *testing.T) {
	response := map[string]any{
		"data": []map[string]string{
			{"id": "r-1", "text": "Reply one", "media_type": "TEXT"},
			{"id": "r-2", "text": "Reply two", "media_type": "TEXT"},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/post-99") {
			t.Errorf("expected path starting with /post-99, got %s", r.URL.Path)
		}
		if !strings.HasSuffix(r.URL.Path, "/replies") {
			t.Errorf("expected path ending with /replies, got %s", r.URL.Path)
		}
		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("expected limit=10, got %s", limit)
		}
		after := r.URL.Query().Get("after")
		if after != "cursor-xyz" {
			t.Errorf("expected after=cursor-xyz, got %s", after)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	origDir := configDir
	configDir = t.TempDir()
	t.Cleanup(func() { configDir = origDir })

	origBase := replyBaseURL
	replyBaseURL = ts.URL
	t.Cleanup(func() { replyBaseURL = origBase })

	cmd := rootCmd
	cmd.SetArgs([]string{"reply", "list", "post-99", "--token", "test-token", "--limit", "10", "--after", "cursor-xyz", "--format", "json"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "r-1") {
		t.Errorf("output should contain first reply ID, got: %s", output)
	}
	if !strings.Contains(output, "r-2") {
		t.Errorf("output should contain second reply ID, got: %s", output)
	}
}

func TestReplyCreate(t *testing.T) {
	var requestCount int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/threads") && r.Method == http.MethodPost:
			// Create container with reply_to_id
			replyToID := r.FormValue("reply_to_id")
			if replyToID != "post-55" {
				t.Errorf("expected reply_to_id=post-55, got %s", replyToID)
			}
			text := r.FormValue("text")
			if text != "Great post!" {
				t.Errorf("expected text='Great post!', got %s", text)
			}
			mediaType := r.FormValue("media_type")
			if mediaType != "TEXT" {
				t.Errorf("expected media_type=TEXT, got %s", mediaType)
			}
			json.NewEncoder(w).Encode(map[string]string{"id": "reply-container"})
		case strings.HasSuffix(r.URL.Path, "/threads_publish") && r.Method == http.MethodPost:
			// Publish
			creationID := r.FormValue("creation_id")
			if creationID != "reply-container" {
				t.Errorf("expected creation_id=reply-container, got %s", creationID)
			}
			json.NewEncoder(w).Encode(map[string]string{"id": "reply-42"})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	dir := t.TempDir()
	origDir := configDir
	configDir = dir
	t.Cleanup(func() { configDir = origDir })

	writeTestCredentials(t, dir, "me-123")

	origBase := replyBaseURL
	replyBaseURL = ts.URL
	t.Cleanup(func() { replyBaseURL = origBase })

	cmd := rootCmd
	cmd.SetArgs([]string{"reply", "create", "post-55", "Great post!", "--token", "test-token"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "reply-42") {
		t.Errorf("output should contain published reply ID, got: %s", output)
	}
	if requestCount != 2 {
		t.Errorf("expected 2 requests (create + publish), got %d", requestCount)
	}
}

func TestReplyHide(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/manage_reply") {
			t.Errorf("expected path ending with /manage_reply, got %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.URL.Path, "/reply-77") {
			t.Errorf("expected path starting with /reply-77, got %s", r.URL.Path)
		}
		hide := r.FormValue("hide")
		if hide != "true" {
			t.Errorf("expected hide=true, got %s", hide)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer ts.Close()

	origDir := configDir
	configDir = t.TempDir()
	t.Cleanup(func() { configDir = origDir })

	origBase := replyBaseURL
	replyBaseURL = ts.URL
	t.Cleanup(func() { replyBaseURL = origBase })

	cmd := rootCmd
	cmd.SetArgs([]string{"reply", "hide", "reply-77", "--token", "test-token"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Hidden") {
		t.Errorf("output should confirm hiding, got: %s", output)
	}
}

func TestReplyUnhide(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/manage_reply") {
			t.Errorf("expected path ending with /manage_reply, got %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.URL.Path, "/reply-88") {
			t.Errorf("expected path starting with /reply-88, got %s", r.URL.Path)
		}
		hide := r.FormValue("hide")
		if hide != "false" {
			t.Errorf("expected hide=false, got %s", hide)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer ts.Close()

	origDir := configDir
	configDir = t.TempDir()
	t.Cleanup(func() { configDir = origDir })

	origBase := replyBaseURL
	replyBaseURL = ts.URL
	t.Cleanup(func() { replyBaseURL = origBase })

	cmd := rootCmd
	cmd.SetArgs([]string{"reply", "unhide", "reply-88", "--token", "test-token"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Unhidden") {
		t.Errorf("output should confirm unhiding, got: %s", output)
	}
}
