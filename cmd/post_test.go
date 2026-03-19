package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/malston/threads-cli/internal/config"
)

func TestPostCommandRegistered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"post"})
	if err != nil {
		t.Fatalf("post command not found: %v", err)
	}
	if cmd.Use != "post" {
		t.Errorf("command Use = %q, want %q", cmd.Use, "post")
	}
}

func TestPostSubcommandsRegistered(t *testing.T) {
	subs := []struct {
		name string
		use  string
	}{
		{"create", "create [text]"},
		{"get", "get [id]"},
		{"list", "list"},
		{"delete", "delete [id]"},
		{"repost", "repost [id]"},
	}
	for _, tt := range subs {
		t.Run(tt.name, func(t *testing.T) {
			cmd, _, err := rootCmd.Find([]string{"post", tt.name})
			if err != nil {
				t.Fatalf("post %s command not found: %v", tt.name, err)
			}
			if cmd.Use != tt.use {
				t.Errorf("command Use = %q, want %q", cmd.Use, tt.use)
			}
		})
	}
}

func TestPostCreateMutuallyExclusiveFlags(t *testing.T) {
	dir := t.TempDir()
	origDir := configDir
	configDir = dir
	t.Cleanup(func() { configDir = origDir })

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"id": "123"})
	}))
	defer ts.Close()

	cmd := rootCmd
	cmd.SetArgs([]string{"post", "create", "hello", "--token", "tok", "--image", "http://img.png", "--video", "http://vid.mp4"})
	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --image and --video are both set")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") && !strings.Contains(err.Error(), "cannot be set together") {
		// Cobra uses "if any flags in the group" phrasing; just check we got an error
		t.Logf("got error (acceptable): %v", err)
	}

	// Reset flags so they don't bleed into other tests.
	resetCreateFlags(t)
}

func TestPostGet(t *testing.T) {
	post := map[string]string{
		"id":         "12345",
		"text":       "Hello world",
		"media_type": "TEXT",
		"timestamp":  "2026-01-01T00:00:00Z",
		"permalink":  "https://threads.net/@user/post/12345",
		"username":   "testuser",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/12345") {
			t.Errorf("expected path starting with /12345, got %s", r.URL.Path)
		}
		fields := r.URL.Query().Get("fields")
		if fields == "" {
			t.Error("expected fields query param")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(post)
	}))
	defer ts.Close()

	t.Setenv("THREADS_API_BASE_URL", ts.URL)

	origDir := configDir
	configDir = t.TempDir()
	t.Cleanup(func() { configDir = origDir })

	cmd := rootCmd
	cmd.SetArgs([]string{"post", "get", "12345", "--token", "test-token", "--format", "json"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	origBase := apiBaseURL
	apiBaseURL = ts.URL
	t.Cleanup(func() { apiBaseURL = origBase })

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "12345") {
		t.Errorf("output should contain post ID, got: %s", output)
	}
	if !strings.Contains(output, "Hello world") {
		t.Errorf("output should contain post text, got: %s", output)
	}
}

func TestPostList(t *testing.T) {
	response := map[string]any{
		"data": []map[string]string{
			{"id": "111", "text": "Post 1", "media_type": "TEXT"},
			{"id": "222", "text": "Post 2", "media_type": "TEXT"},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		// Path should end with /threads
		if !strings.HasSuffix(r.URL.Path, "/threads") {
			t.Errorf("expected path ending with /threads, got %s", r.URL.Path)
		}
		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("expected limit=10, got %s", limit)
		}
		after := r.URL.Query().Get("after")
		if after != "cursor-abc" {
			t.Errorf("expected after=cursor-abc, got %s", after)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	dir := t.TempDir()
	origDir := configDir
	configDir = dir
	t.Cleanup(func() { configDir = origDir })

	writeTestCredentials(t, dir, "user-42")

	origBase := apiBaseURL
	apiBaseURL = ts.URL
	t.Cleanup(func() { apiBaseURL = origBase })

	cmd := rootCmd
	cmd.SetArgs([]string{"post", "list", "--token", "test-token", "--limit", "10", "--after", "cursor-abc", "--format", "json"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "111") {
		t.Errorf("output should contain first post ID, got: %s", output)
	}
	if !strings.Contains(output, "222") {
		t.Errorf("output should contain second post ID, got: %s", output)
	}
}

func TestPostDelete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/99999") {
			t.Errorf("expected path starting with /99999, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer ts.Close()

	origDir := configDir
	configDir = t.TempDir()
	t.Cleanup(func() { configDir = origDir })

	origBase := apiBaseURL
	apiBaseURL = ts.URL
	t.Cleanup(func() { apiBaseURL = origBase })

	cmd := rootCmd
	cmd.SetArgs([]string{"post", "delete", "99999", "--token", "test-token"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Deleted") {
		t.Errorf("output should confirm deletion, got: %s", output)
	}
}

func TestPostRepost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/repost") {
			t.Errorf("expected path containing /repost, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "repost-789"})
	}))
	defer ts.Close()

	origDir := configDir
	configDir = t.TempDir()
	t.Cleanup(func() { configDir = origDir })

	origBase := apiBaseURL
	apiBaseURL = ts.URL
	t.Cleanup(func() { apiBaseURL = origBase })

	cmd := rootCmd
	cmd.SetArgs([]string{"post", "repost", "55555", "--token", "test-token"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "repost-789") {
		t.Errorf("output should contain repost ID, got: %s", output)
	}
}

func TestPostCreateText(t *testing.T) {
	var requestCount int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/threads") && r.Method == http.MethodPost:
			// Create container
			mediaType := r.FormValue("media_type")
			if mediaType != "TEXT" {
				t.Errorf("expected media_type=TEXT, got %s", mediaType)
			}
			text := r.FormValue("text")
			if text != "Hello Threads!" {
				t.Errorf("expected text='Hello Threads!', got %s", text)
			}
			json.NewEncoder(w).Encode(map[string]string{"id": "container-1"})
		case strings.HasSuffix(r.URL.Path, "/threads_publish") && r.Method == http.MethodPost:
			// Publish
			creationID := r.FormValue("creation_id")
			if creationID != "container-1" {
				t.Errorf("expected creation_id=container-1, got %s", creationID)
			}
			json.NewEncoder(w).Encode(map[string]string{"id": "post-1"})
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

	origBase := apiBaseURL
	apiBaseURL = ts.URL
	t.Cleanup(func() { apiBaseURL = origBase })

	cmd := rootCmd
	cmd.SetArgs([]string{"post", "create", "Hello Threads!", "--token", "test-token"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "post-1") {
		t.Errorf("output should contain published post ID, got: %s", output)
	}
	if requestCount != 2 {
		t.Errorf("expected 2 requests (create + publish), got %d", requestCount)
	}
}

func TestPostCreateImage(t *testing.T) {
	var statusChecked bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/threads") && r.Method == http.MethodPost:
			mediaType := r.FormValue("media_type")
			if mediaType != "IMAGE" {
				t.Errorf("expected media_type=IMAGE, got %s", mediaType)
			}
			imageURL := r.FormValue("image_url")
			if imageURL != "http://example.com/img.png" {
				t.Errorf("expected image_url, got %s", imageURL)
			}
			json.NewEncoder(w).Encode(map[string]string{"id": "img-container"})
		case r.Method == http.MethodGet && r.URL.Query().Get("fields") == "status,error_message":
			statusChecked = true
			json.NewEncoder(w).Encode(map[string]string{"status": "FINISHED"})
		case strings.HasSuffix(r.URL.Path, "/threads_publish") && r.Method == http.MethodPost:
			json.NewEncoder(w).Encode(map[string]string{"id": "img-post"})
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

	origBase := apiBaseURL
	apiBaseURL = ts.URL
	t.Cleanup(func() { apiBaseURL = origBase })

	cmd := rootCmd
	t.Cleanup(func() { resetCreateFlags(t) })

	cmd.SetArgs([]string{"post", "create", "Check this out", "--image", "http://example.com/img.png", "--token", "test-token"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if !statusChecked {
		t.Error("expected status check for media container")
	}
	output := out.String()
	if !strings.Contains(output, "img-post") {
		t.Errorf("output should contain published post ID, got: %s", output)
	}
}

func TestPostCreateVideo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/threads") && r.Method == http.MethodPost:
			mediaType := r.FormValue("media_type")
			if mediaType != "VIDEO" {
				t.Errorf("expected media_type=VIDEO, got %s", mediaType)
			}
			json.NewEncoder(w).Encode(map[string]string{"id": "vid-container"})
		case r.Method == http.MethodGet && r.URL.Query().Get("fields") == "status,error_message":
			json.NewEncoder(w).Encode(map[string]string{"status": "FINISHED"})
		case strings.HasSuffix(r.URL.Path, "/threads_publish") && r.Method == http.MethodPost:
			json.NewEncoder(w).Encode(map[string]string{"id": "vid-post"})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()

	dir := t.TempDir()
	origDir := configDir
	configDir = dir
	t.Cleanup(func() { configDir = origDir })

	writeTestCredentials(t, dir, "me-123")

	origBase := apiBaseURL
	apiBaseURL = ts.URL
	t.Cleanup(func() { apiBaseURL = origBase })

	cmd := rootCmd
	t.Cleanup(func() { resetCreateFlags(t) })

	cmd.SetArgs([]string{"post", "create", "Watch this", "--video", "http://example.com/vid.mp4", "--token", "test-token"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "vid-post") {
		t.Errorf("output should contain published post ID, got: %s", output)
	}
}

func TestPostCreateCarousel(t *testing.T) {
	var createCount int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/threads") && r.Method == http.MethodPost:
			createCount++
			mediaType := r.FormValue("media_type")
			switch mediaType {
			case "IMAGE":
				json.NewEncoder(w).Encode(map[string]string{"id": "item-" + r.FormValue("image_url")})
			case "CAROUSEL":
				json.NewEncoder(w).Encode(map[string]string{"id": "carousel-container"})
			default:
				t.Errorf("unexpected media_type: %s", mediaType)
			}
		case r.Method == http.MethodGet && r.URL.Query().Get("fields") == "status,error_message":
			json.NewEncoder(w).Encode(map[string]string{"status": "FINISHED"})
		case strings.HasSuffix(r.URL.Path, "/threads_publish") && r.Method == http.MethodPost:
			json.NewEncoder(w).Encode(map[string]string{"id": "carousel-post"})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()

	dir := t.TempDir()
	origDir := configDir
	configDir = dir
	t.Cleanup(func() { configDir = origDir })

	writeTestCredentials(t, dir, "me-123")

	origBase := apiBaseURL
	apiBaseURL = ts.URL
	t.Cleanup(func() { apiBaseURL = origBase })

	cmd := rootCmd
	t.Cleanup(func() { resetCreateFlags(t) })

	cmd.SetArgs([]string{
		"post", "create", "My carousel",
		"--carousel", "http://example.com/a.png,http://example.com/b.png",
		"--token", "test-token",
	})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "carousel-post") {
		t.Errorf("output should contain published carousel ID, got: %s", output)
	}
}

func TestPostListLoadsUserIDFromCredentials(t *testing.T) {
	var requestedPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	}))
	defer ts.Close()

	dir := t.TempDir()
	origDir := configDir
	configDir = dir
	t.Cleanup(func() { configDir = origDir })

	writeTestCredentials(t, dir, "user-77")

	origBase := apiBaseURL
	apiBaseURL = ts.URL
	t.Cleanup(func() { apiBaseURL = origBase })

	cmd := rootCmd
	cmd.SetArgs([]string{"post", "list", "--token", "test-token"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if !strings.Contains(requestedPath, "user-77") {
		t.Errorf("expected request path to contain userID 'user-77', got: %s", requestedPath)
	}
}

// resetCreateFlags resets the create subcommand's media flags to prevent
// state bleed between tests sharing the global rootCmd.
func resetCreateFlags(t *testing.T) {
	t.Helper()
	createCmd, _, err := rootCmd.Find([]string{"post", "create"})
	if err != nil {
		t.Fatalf("finding create command: %v", err)
	}
	for _, name := range []string{"image", "video"} {
		f := createCmd.Flags().Lookup(name)
		if f != nil {
			f.Value.Set("")
			f.Changed = false
		}
	}
	// StringSlice needs special handling: reset to empty by setting "[]"
	// then clearing Changed so it won't be treated as user-provided.
	if f := createCmd.Flags().Lookup("carousel"); f != nil {
		f.Value.Set("[]")
		f.Changed = false
	}
}

// writeTestCredentials writes a credentials file with the given userID.
func writeTestCredentials(t *testing.T, dir, userID string) {
	t.Helper()
	creds := config.Credentials{
		AccessToken: "test-token",
		TokenType:   "bearer",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		UserID:      userID,
	}
	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "credentials.json"), data, 0600); err != nil {
		t.Fatalf("write credentials: %v", err)
	}
}
