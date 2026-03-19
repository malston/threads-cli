package threads

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/malston/saved-threads/internal/api"
)

func TestCreateContainer(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody, _ = url.ParseQuery(string(b))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "container-123"})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	id, err := CreateContainer(context.Background(), client, "me", "hello world", "TEXT")
	if err != nil {
		t.Fatalf("CreateContainer returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotPath != "/me/threads" {
		t.Errorf("path = %q, want %q", gotPath, "/me/threads")
	}
	if gotBody.Get("text") != "hello world" {
		t.Errorf("text = %q, want %q", gotBody.Get("text"), "hello world")
	}
	if gotBody.Get("media_type") != "TEXT" {
		t.Errorf("media_type = %q, want %q", gotBody.Get("media_type"), "TEXT")
	}
	if id != "container-123" {
		t.Errorf("id = %q, want %q", id, "container-123")
	}
}

func TestCreateContainerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "invalid media type",
				"type":    "OAuthException",
				"code":    100,
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	_, err := CreateContainer(context.Background(), client, "me", "hello", "INVALID")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPublish(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody, _ = url.ParseQuery(string(b))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "post-456"})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	id, err := Publish(context.Background(), client, "me", "container-123")
	if err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotPath != "/me/threads_publish" {
		t.Errorf("path = %q, want %q", gotPath, "/me/threads_publish")
	}
	if gotBody.Get("creation_id") != "container-123" {
		t.Errorf("creation_id = %q, want %q", gotBody.Get("creation_id"), "container-123")
	}
	if id != "post-456" {
		t.Errorf("id = %q, want %q", id, "post-456")
	}
}

func TestPublishError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "server error",
				"type":    "OAuthException",
				"code":    2,
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	_, err := Publish(context.Background(), client, "me", "bad-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGet(t *testing.T) {
	var gotMethod, gotPath string
	var gotQuery url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"id":         "post-789",
			"text":       "hello world",
			"media_type": "TEXT",
			"timestamp":  "2026-03-18T12:00:00Z",
			"permalink":  "https://threads.net/@user/post/abc",
			"username":   "testuser",
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	post, err := Get(context.Background(), client, "post-789", []string{"id", "text", "media_type", "timestamp", "permalink", "username"})
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if gotPath != "/post-789" {
		t.Errorf("path = %q, want %q", gotPath, "/post-789")
	}
	if gotQuery.Get("fields") != "id,text,media_type,timestamp,permalink,username" {
		t.Errorf("fields = %q, want %q", gotQuery.Get("fields"), "id,text,media_type,timestamp,permalink,username")
	}
	if post.ID != "post-789" {
		t.Errorf("ID = %q, want %q", post.ID, "post-789")
	}
	if post.Text != "hello world" {
		t.Errorf("Text = %q, want %q", post.Text, "hello world")
	}
	if post.MediaType != "TEXT" {
		t.Errorf("MediaType = %q, want %q", post.MediaType, "TEXT")
	}
	if post.Timestamp != "2026-03-18T12:00:00Z" {
		t.Errorf("Timestamp = %q, want %q", post.Timestamp, "2026-03-18T12:00:00Z")
	}
	if post.Permalink != "https://threads.net/@user/post/abc" {
		t.Errorf("Permalink = %q, want %q", post.Permalink, "https://threads.net/@user/post/abc")
	}
	if post.Username != "testuser" {
		t.Errorf("Username = %q, want %q", post.Username, "testuser")
	}
}

func TestGetError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "not found",
				"type":    "OAuthException",
				"code":    100,
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	_, err := Get(context.Background(), client, "bad-id", []string{"id"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestList(t *testing.T) {
	var gotMethod, gotPath string
	var gotQuery url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{
				{"id": "p1", "text": "first post"},
				{"id": "p2", "text": "second post"},
			},
			"paging": map[string]any{
				"cursors": map[string]any{
					"before": "cursor-before",
					"after":  "cursor-after",
				},
				"next": "https://graph.threads.net/me/threads?after=cursor-after",
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	page := &api.PageParams{Limit: 2, After: "prev-cursor"}
	list, err := List(context.Background(), client, "me", page)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if gotPath != "/me/threads" {
		t.Errorf("path = %q, want %q", gotPath, "/me/threads")
	}
	if gotQuery.Get("limit") != "2" {
		t.Errorf("limit = %q, want %q", gotQuery.Get("limit"), "2")
	}
	if gotQuery.Get("after") != "prev-cursor" {
		t.Errorf("after = %q, want %q", gotQuery.Get("after"), "prev-cursor")
	}
	if gotQuery.Get("fields") == "" {
		t.Error("fields param missing")
	}

	if len(list.Data) != 2 {
		t.Fatalf("len(Data) = %d, want 2", len(list.Data))
	}
	if list.Data[0].ID != "p1" {
		t.Errorf("Data[0].ID = %q, want %q", list.Data[0].ID, "p1")
	}
	if list.Data[0].Text != "first post" {
		t.Errorf("Data[0].Text = %q, want %q", list.Data[0].Text, "first post")
	}
	if list.Data[1].ID != "p2" {
		t.Errorf("Data[1].ID = %q, want %q", list.Data[1].ID, "p2")
	}

	if list.Paging == nil {
		t.Fatal("Paging is nil")
	}
	if list.Paging.Before != "cursor-before" {
		t.Errorf("Paging.Before = %q, want %q", list.Paging.Before, "cursor-before")
	}
	if list.Paging.After != "cursor-after" {
		t.Errorf("Paging.After = %q, want %q", list.Paging.After, "cursor-after")
	}
	if !list.Paging.HasNext() {
		t.Error("Paging.HasNext() = false, want true")
	}
}

func TestListNilPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	list, err := List(context.Background(), client, "me", nil)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if list.Data == nil {
		t.Error("Data is nil, want empty slice")
	}
}

func TestListError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "invalid token",
				"type":    "OAuthException",
				"code":    190,
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	_, err := List(context.Background(), client, "me", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDelete(t *testing.T) {
	var gotMethod, gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	err := Delete(context.Background(), client, "post-123")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodDelete)
	}
	if gotPath != "/post-123" {
		t.Errorf("path = %q, want %q", gotPath, "/post-123")
	}
}

func TestDeleteError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "permission denied",
				"type":    "OAuthException",
				"code":    10,
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	err := Delete(context.Background(), client, "post-123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRepost(t *testing.T) {
	var gotMethod, gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "repost-789"})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	id, err := Repost(context.Background(), client, "post-123")
	if err != nil {
		t.Fatalf("Repost returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotPath != "/post-123/repost" {
		t.Errorf("path = %q, want %q", gotPath, "/post-123/repost")
	}
	if id != "repost-789" {
		t.Errorf("id = %q, want %q", id, "repost-789")
	}
}

func TestRepostError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "cannot repost",
				"type":    "OAuthException",
				"code":    100,
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	_, err := Repost(context.Background(), client, "post-123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
