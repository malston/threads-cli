package threads

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/malston/saved-threads/internal/api"
)

func TestListReplies(t *testing.T) {
	var gotMethod, gotPath string
	var gotQuery url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{
				{"id": "r1", "text": "reply one"},
				{"id": "r2", "text": "reply two"},
			},
			"paging": map[string]any{
				"cursors": map[string]any{
					"before": "cur-before",
					"after":  "cur-after",
				},
				"next": "https://graph.threads.net/post-1/replies?after=cur-after",
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	page := &api.PageParams{Limit: 5, After: "prev"}
	list, err := ListReplies(context.Background(), client, "post-1", page)
	if err != nil {
		t.Fatalf("ListReplies returned error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if gotPath != "/post-1/replies" {
		t.Errorf("path = %q, want %q", gotPath, "/post-1/replies")
	}
	if gotQuery.Get("limit") != "5" {
		t.Errorf("limit = %q, want %q", gotQuery.Get("limit"), "5")
	}
	if gotQuery.Get("after") != "prev" {
		t.Errorf("after = %q, want %q", gotQuery.Get("after"), "prev")
	}
	if gotQuery.Get("fields") == "" {
		t.Error("fields param missing")
	}

	if len(list.Data) != 2 {
		t.Fatalf("len(Data) = %d, want 2", len(list.Data))
	}
	if list.Data[0].ID != "r1" {
		t.Errorf("Data[0].ID = %q, want %q", list.Data[0].ID, "r1")
	}
	if list.Data[0].Text != "reply one" {
		t.Errorf("Data[0].Text = %q, want %q", list.Data[0].Text, "reply one")
	}
	if list.Data[1].ID != "r2" {
		t.Errorf("Data[1].ID = %q, want %q", list.Data[1].ID, "r2")
	}

	if list.Paging == nil {
		t.Fatal("Paging is nil")
	}
	if list.Paging.Before != "cur-before" {
		t.Errorf("Paging.Before = %q, want %q", list.Paging.Before, "cur-before")
	}
	if list.Paging.After != "cur-after" {
		t.Errorf("Paging.After = %q, want %q", list.Paging.After, "cur-after")
	}
	if !list.Paging.HasNext() {
		t.Error("Paging.HasNext() = false, want true")
	}
}

func TestListRepliesError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "invalid post",
				"type":    "OAuthException",
				"code":    100,
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	_, err := ListReplies(context.Background(), client, "bad-id", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetConversation(t *testing.T) {
	var gotMethod, gotPath string
	var gotQuery url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{
				{"id": "c1", "text": "convo one"},
				{"id": "c2", "text": "convo two"},
			},
			"paging": map[string]any{
				"cursors": map[string]any{
					"before": "cb",
					"after":  "ca",
				},
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	page := &api.PageParams{Limit: 10}
	list, err := GetConversation(context.Background(), client, "post-1", page)
	if err != nil {
		t.Fatalf("GetConversation returned error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if gotPath != "/post-1/conversation" {
		t.Errorf("path = %q, want %q", gotPath, "/post-1/conversation")
	}
	if gotQuery.Get("limit") != "10" {
		t.Errorf("limit = %q, want %q", gotQuery.Get("limit"), "10")
	}
	if gotQuery.Get("fields") == "" {
		t.Error("fields param missing")
	}

	if len(list.Data) != 2 {
		t.Fatalf("len(Data) = %d, want 2", len(list.Data))
	}
	if list.Data[0].ID != "c1" {
		t.Errorf("Data[0].ID = %q, want %q", list.Data[0].ID, "c1")
	}
	if list.Data[1].ID != "c2" {
		t.Errorf("Data[1].ID = %q, want %q", list.Data[1].ID, "c2")
	}

	if list.Paging == nil {
		t.Fatal("Paging is nil")
	}
	if list.Paging.Before != "cb" {
		t.Errorf("Paging.Before = %q, want %q", list.Paging.Before, "cb")
	}
	if list.Paging.After != "ca" {
		t.Errorf("Paging.After = %q, want %q", list.Paging.After, "ca")
	}
}

func TestGetConversationError(t *testing.T) {
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
	_, err := GetConversation(context.Background(), client, "bad-id", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateReply(t *testing.T) {
	var mu sync.Mutex
	var calls []struct {
		method string
		path   string
		body   url.Values
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body, _ := url.ParseQuery(string(b))
		mu.Lock()
		calls = append(calls, struct {
			method string
			path   string
			body   url.Values
		}{r.Method, r.URL.Path, body})
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		switch r.URL.Path {
		case "/me/threads":
			json.NewEncoder(w).Encode(map[string]string{"id": "container-99"})
		case "/me/threads_publish":
			json.NewEncoder(w).Encode(map[string]string{"id": "reply-42"})
		}
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	id, err := CreateReply(context.Background(), client, "me", "post-1", "my reply")
	if err != nil {
		t.Fatalf("CreateReply returned error: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}

	// First call: create container
	c1 := calls[0]
	if c1.method != http.MethodPost {
		t.Errorf("call 1 method = %q, want %q", c1.method, http.MethodPost)
	}
	if c1.path != "/me/threads" {
		t.Errorf("call 1 path = %q, want %q", c1.path, "/me/threads")
	}
	if c1.body.Get("media_type") != "TEXT" {
		t.Errorf("call 1 media_type = %q, want %q", c1.body.Get("media_type"), "TEXT")
	}
	if c1.body.Get("text") != "my reply" {
		t.Errorf("call 1 text = %q, want %q", c1.body.Get("text"), "my reply")
	}
	if c1.body.Get("reply_to_id") != "post-1" {
		t.Errorf("call 1 reply_to_id = %q, want %q", c1.body.Get("reply_to_id"), "post-1")
	}

	// Second call: publish
	c2 := calls[1]
	if c2.method != http.MethodPost {
		t.Errorf("call 2 method = %q, want %q", c2.method, http.MethodPost)
	}
	if c2.path != "/me/threads_publish" {
		t.Errorf("call 2 path = %q, want %q", c2.path, "/me/threads_publish")
	}
	if c2.body.Get("creation_id") != "container-99" {
		t.Errorf("call 2 creation_id = %q, want %q", c2.body.Get("creation_id"), "container-99")
	}

	if id != "reply-42" {
		t.Errorf("id = %q, want %q", id, "reply-42")
	}
}

func TestCreateReplyContainerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "invalid reply",
				"type":    "OAuthException",
				"code":    100,
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	_, err := CreateReply(context.Background(), client, "me", "post-1", "bad reply")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestHideReply(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody, _ = url.ParseQuery(string(b))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	err := HideReply(context.Background(), client, "reply-1")
	if err != nil {
		t.Fatalf("HideReply returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotPath != "/reply-1/manage_reply" {
		t.Errorf("path = %q, want %q", gotPath, "/reply-1/manage_reply")
	}
	if gotBody.Get("hide") != "true" {
		t.Errorf("hide = %q, want %q", gotBody.Get("hide"), "true")
	}
}

func TestHideReplyError(t *testing.T) {
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
	err := HideReply(context.Background(), client, "reply-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUnhideReply(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody, _ = url.ParseQuery(string(b))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	err := UnhideReply(context.Background(), client, "reply-1")
	if err != nil {
		t.Fatalf("UnhideReply returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotPath != "/reply-1/manage_reply" {
		t.Errorf("path = %q, want %q", gotPath, "/reply-1/manage_reply")
	}
	if gotBody.Get("hide") != "false" {
		t.Errorf("hide = %q, want %q", gotBody.Get("hide"), "false")
	}
}

func TestUnhideReplyError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "not a reply",
				"type":    "OAuthException",
				"code":    100,
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	err := UnhideReply(context.Background(), client, "not-a-reply")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
