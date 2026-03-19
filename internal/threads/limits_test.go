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

func TestGetLimits(t *testing.T) {
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
					"quota_usage":     230,
					"quota_total":     250,
					"quota_duration":  86400,
					"rate_limit_type": "POSTS",
				},
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	result, err := GetLimits(context.Background(), client, "me")
	if err != nil {
		t.Fatalf("GetLimits returned error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if gotPath != "/me/threads_publishing_limit" {
		t.Errorf("path = %q, want %q", gotPath, "/me/threads_publishing_limit")
	}
	if gotQuery.Get("fields") != "quota_usage,quota_total,quota_duration,rate_limit_type" {
		t.Errorf("fields = %q, want %q", gotQuery.Get("fields"), "quota_usage,quota_total,quota_duration,rate_limit_type")
	}

	if len(result.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(result.Data))
	}
	if result.Data[0].QuotaUsage != 230 {
		t.Errorf("QuotaUsage = %d, want 230", result.Data[0].QuotaUsage)
	}
	if result.Data[0].QuotaTotal != 250 {
		t.Errorf("QuotaTotal = %d, want 250", result.Data[0].QuotaTotal)
	}
	if result.Data[0].QuotaDuration != 86400 {
		t.Errorf("QuotaDuration = %d, want 86400", result.Data[0].QuotaDuration)
	}
	if result.Data[0].RateLimitType != "POSTS" {
		t.Errorf("RateLimitType = %q, want %q", result.Data[0].RateLimitType, "POSTS")
	}
}

func TestGetLimitsError(t *testing.T) {
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
	_, err := GetLimits(context.Background(), client, "me")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
