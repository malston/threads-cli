package threads

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/malston/threads-cli/internal/api"
)

func TestSearchSendsCorrectRequest(t *testing.T) {
	var gotMethod, gotPath string
	var gotQuery url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	_, err := Search(context.Background(), client, "golang", nil, nil)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if gotPath != "/keyword_search" {
		t.Errorf("path = %q, want %q", gotPath, "/keyword_search")
	}
	if gotQuery.Get("q") != "golang" {
		t.Errorf("q = %q, want %q", gotQuery.Get("q"), "golang")
	}
	if gotQuery.Get("fields") == "" {
		t.Error("fields param missing")
	}
}

func TestSearchAppliesPagination(t *testing.T) {
	var gotQuery url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	page := &api.PageParams{Limit: 25, After: "cursor-abc"}
	_, err := Search(context.Background(), client, "test", page, nil)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if gotQuery.Get("limit") != "25" {
		t.Errorf("limit = %q, want %q", gotQuery.Get("limit"), "25")
	}
	if gotQuery.Get("after") != "cursor-abc" {
		t.Errorf("after = %q, want %q", gotQuery.Get("after"), "cursor-abc")
	}
}

func TestSearchParsesPostListWithPaging(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{
				{"id": "s1", "text": "search result one", "username": "alice"},
				{"id": "s2", "text": "search result two", "username": "bob"},
			},
			"paging": map[string]any{
				"cursors": map[string]any{
					"before": "cur-b",
					"after":  "cur-a",
				},
				"next": "https://graph.threads.net/keyword_search?after=cur-a",
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	list, err := Search(context.Background(), client, "result", nil, nil)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if len(list.Data) != 2 {
		t.Fatalf("len(Data) = %d, want 2", len(list.Data))
	}
	if list.Data[0].ID != "s1" {
		t.Errorf("Data[0].ID = %q, want %q", list.Data[0].ID, "s1")
	}
	if list.Data[0].Text != "search result one" {
		t.Errorf("Data[0].Text = %q, want %q", list.Data[0].Text, "search result one")
	}
	if list.Data[0].Username != "alice" {
		t.Errorf("Data[0].Username = %q, want %q", list.Data[0].Username, "alice")
	}
	if list.Data[1].ID != "s2" {
		t.Errorf("Data[1].ID = %q, want %q", list.Data[1].ID, "s2")
	}

	if list.Paging == nil {
		t.Fatal("Paging is nil")
	}
	if list.Paging.Before != "cur-b" {
		t.Errorf("Paging.Before = %q, want %q", list.Paging.Before, "cur-b")
	}
	if list.Paging.After != "cur-a" {
		t.Errorf("Paging.After = %q, want %q", list.Paging.After, "cur-a")
	}
	if !list.Paging.HasNext() {
		t.Error("Paging.HasNext() = false, want true")
	}
}

func TestSearchAppliesFilters(t *testing.T) {
	var gotQuery url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	filters := &SearchFilters{
		Author:    "someuser",
		SortBy:    "recent",
		MediaType: "image",
		Since:     1700000000,
		Until:     1700086400,
	}
	_, err := Search(context.Background(), client, "topic", nil, filters)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if gotQuery.Get("author_username") != "someuser" {
		t.Errorf("author_username = %q, want %q", gotQuery.Get("author_username"), "someuser")
	}
	if gotQuery.Get("search_type") != "RECENT" {
		t.Errorf("search_type = %q, want %q", gotQuery.Get("search_type"), "RECENT")
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

func TestSearchWithNilFilters(t *testing.T) {
	var gotQuery url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	_, err := Search(context.Background(), client, "test", nil, nil)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if gotQuery.Get("author_username") != "" {
		t.Errorf("author_username should be absent, got %q", gotQuery.Get("author_username"))
	}
	if gotQuery.Get("search_type") != "" {
		t.Errorf("search_type should be absent, got %q", gotQuery.Get("search_type"))
	}
}

func TestSearchWithPartialFilters(t *testing.T) {
	var gotQuery url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	filters := &SearchFilters{
		Author: "alice",
	}
	_, err := Search(context.Background(), client, "hello", nil, filters)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if gotQuery.Get("author_username") != "alice" {
		t.Errorf("author_username = %q, want %q", gotQuery.Get("author_username"), "alice")
	}
	if gotQuery.Get("search_type") != "" {
		t.Errorf("search_type should be absent, got %q", gotQuery.Get("search_type"))
	}
	if gotQuery.Get("media_type") != "" {
		t.Errorf("media_type should be absent, got %q", gotQuery.Get("media_type"))
	}
	if gotQuery.Get("since") != "" {
		t.Errorf("since should be absent, got %q", gotQuery.Get("since"))
	}
}

func TestSearchReturnsErrorOnNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "invalid query",
				"type":    "OAuthException",
				"code":    100,
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	_, err := Search(context.Background(), client, "bad", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSearchWithEmptyQuery(t *testing.T) {
	var gotQuery url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	_, err := Search(context.Background(), client, "", nil, nil)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if gotQuery.Get("q") != "" {
		t.Errorf("q = %q, want empty string", gotQuery.Get("q"))
	}
}
