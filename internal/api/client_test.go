package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("test-token")

	if c.baseURL != "https://graph.threads.net" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "https://graph.threads.net")
	}
	if c.token != "test-token" {
		t.Errorf("token = %q, want %q", c.token, "test-token")
	}
	if c.httpClient == nil {
		t.Error("httpClient is nil, want non-nil default")
	}
}

func TestNewClientWithHTTP(t *testing.T) {
	custom := &http.Client{}
	c := NewClientWithHTTP("my-token", "https://example.com", custom)

	if c.baseURL != "https://example.com" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "https://example.com")
	}
	if c.token != "my-token" {
		t.Errorf("token = %q, want %q", c.token, "my-token")
	}
	if c.httpClient != custom {
		t.Error("httpClient not set to provided client")
	}
}

func TestGet(t *testing.T) {
	var gotMethod, gotPath, gotAuth, gotQuery string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := NewClientWithHTTP("tok123", srv.URL, srv.Client())

	params := url.Values{}
	params.Set("fields", "id,text")
	params.Set("limit", "10")

	resp, err := c.Get(context.Background(), "/me/threads", params)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	defer resp.Body.Close()

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if gotPath != "/me/threads" {
		t.Errorf("path = %q, want %q", gotPath, "/me/threads")
	}
	if gotAuth != "Bearer tok123" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer tok123")
	}

	parsed, _ := url.ParseQuery(gotQuery)
	if parsed.Get("fields") != "id,text" {
		t.Errorf("fields param = %q, want %q", parsed.Get("fields"), "id,text")
	}
	if parsed.Get("limit") != "10" {
		t.Errorf("limit param = %q, want %q", parsed.Get("limit"), "10")
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestGetWithNilParams(t *testing.T) {
	var gotQuery string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClientWithHTTP("tok", srv.URL, srv.Client())

	resp, err := c.Get(context.Background(), "/me", nil)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	defer resp.Body.Close()

	if gotQuery != "" {
		t.Errorf("query = %q, want empty", gotQuery)
	}
}

func TestPost(t *testing.T) {
	var gotMethod, gotPath, gotAuth, gotContentType, gotBody string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := NewClientWithHTTP("post-tok", srv.URL, srv.Client())

	body := url.Values{}
	body.Set("text", "hello world")
	body.Set("media_type", "TEXT")

	resp, err := c.Post(context.Background(), "/me/threads", body)
	if err != nil {
		t.Fatalf("Post returned error: %v", err)
	}
	defer resp.Body.Close()

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotPath != "/me/threads" {
		t.Errorf("path = %q, want %q", gotPath, "/me/threads")
	}
	if gotAuth != "Bearer post-tok" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer post-tok")
	}
	if gotContentType != "application/x-www-form-urlencoded" {
		t.Errorf("Content-Type = %q, want %q", gotContentType, "application/x-www-form-urlencoded")
	}

	parsed, _ := url.ParseQuery(gotBody)
	if parsed.Get("text") != "hello world" {
		t.Errorf("text = %q, want %q", parsed.Get("text"), "hello world")
	}
	if parsed.Get("media_type") != "TEXT" {
		t.Errorf("media_type = %q, want %q", parsed.Get("media_type"), "TEXT")
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
}

func TestDelete(t *testing.T) {
	var gotMethod, gotPath, gotAuth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClientWithHTTP("del-tok", srv.URL, srv.Client())

	resp, err := c.Delete(context.Background(), "/12345")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	defer resp.Body.Close()

	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodDelete)
	}
	if gotPath != "/12345" {
		t.Errorf("path = %q, want %q", gotPath, "/12345")
	}
	if gotAuth != "Bearer del-tok" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer del-tok")
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}
