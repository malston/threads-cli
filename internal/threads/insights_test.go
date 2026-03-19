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

func TestPostInsights(t *testing.T) {
	var gotMethod, gotPath string
	var gotQuery url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"name":        "views",
					"title":       "Views",
					"description": "Total views",
					"period":      "lifetime",
					"values":      []map[string]any{{"value": 42}},
				},
				{
					"name":        "likes",
					"title":       "Likes",
					"description": "Total likes",
					"period":      "lifetime",
					"values":      []map[string]any{{"value": 10}},
				},
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	result, err := PostInsights(context.Background(), client, "post-123", []string{"views", "likes"})
	if err != nil {
		t.Fatalf("PostInsights returned error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if gotPath != "/post-123/insights" {
		t.Errorf("path = %q, want %q", gotPath, "/post-123/insights")
	}
	if gotQuery.Get("metric") != "views,likes" {
		t.Errorf("metric = %q, want %q", gotQuery.Get("metric"), "views,likes")
	}

	if len(result.Data) != 2 {
		t.Fatalf("len(Data) = %d, want 2", len(result.Data))
	}
	if result.Data[0].Name != "views" {
		t.Errorf("Data[0].Name = %q, want %q", result.Data[0].Name, "views")
	}
	if result.Data[0].Title != "Views" {
		t.Errorf("Data[0].Title = %q, want %q", result.Data[0].Title, "Views")
	}
	if result.Data[0].Description != "Total views" {
		t.Errorf("Data[0].Description = %q, want %q", result.Data[0].Description, "Total views")
	}
	if result.Data[0].Period != "lifetime" {
		t.Errorf("Data[0].Period = %q, want %q", result.Data[0].Period, "lifetime")
	}
	if len(result.Data[0].Values) != 1 {
		t.Fatalf("len(Data[0].Values) = %d, want 1", len(result.Data[0].Values))
	}
	// The JSON number 42 decodes as float64 via any.
	if v, ok := result.Data[0].Values[0].Value.(float64); !ok || v != 42 {
		t.Errorf("Data[0].Values[0].Value = %v, want 42", result.Data[0].Values[0].Value)
	}
	if result.Data[1].Name != "likes" {
		t.Errorf("Data[1].Name = %q, want %q", result.Data[1].Name, "likes")
	}
}

func TestPostInsightsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "invalid metric",
				"type":    "OAuthException",
				"code":    100,
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	_, err := PostInsights(context.Background(), client, "post-123", []string{"bad_metric"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUserInsights(t *testing.T) {
	var gotMethod, gotPath string
	var gotQuery url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"name":        "followers_count",
					"title":       "Followers",
					"description": "Total followers",
					"period":      "day",
					"values":      []map[string]any{{"value": 500}},
				},
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	result, err := UserInsights(context.Background(), client, "me", []string{"followers_count"}, 1700000000, 1700100000)
	if err != nil {
		t.Fatalf("UserInsights returned error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if gotPath != "/me/threads_insights" {
		t.Errorf("path = %q, want %q", gotPath, "/me/threads_insights")
	}
	if gotQuery.Get("metric") != "followers_count" {
		t.Errorf("metric = %q, want %q", gotQuery.Get("metric"), "followers_count")
	}
	if gotQuery.Get("since") != "1700000000" {
		t.Errorf("since = %q, want %q", gotQuery.Get("since"), "1700000000")
	}
	if gotQuery.Get("until") != "1700100000" {
		t.Errorf("until = %q, want %q", gotQuery.Get("until"), "1700100000")
	}

	if len(result.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(result.Data))
	}
	if result.Data[0].Name != "followers_count" {
		t.Errorf("Data[0].Name = %q, want %q", result.Data[0].Name, "followers_count")
	}
}

func TestUserInsightsOmitsZeroTimestamps(t *testing.T) {
	var gotQuery url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	_, err := UserInsights(context.Background(), client, "me", []string{"views"}, 0, 0)
	if err != nil {
		t.Fatalf("UserInsights returned error: %v", err)
	}

	if gotQuery.Get("since") != "" {
		t.Errorf("since should be omitted when zero, got %q", gotQuery.Get("since"))
	}
	if gotQuery.Get("until") != "" {
		t.Errorf("until should be omitted when zero, got %q", gotQuery.Get("until"))
	}
}
