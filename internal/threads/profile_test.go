package threads

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malston/saved-threads/internal/api"
)

func TestGetProfile(t *testing.T) {
	var gotMethod, gotPath, gotFields string

	want := Profile{
		ID:            "12345",
		Username:      "testuser",
		Name:          "Test User",
		Bio:           "Hello world",
		ProfilePicURL: "https://example.com/pic.jpg",
		IsVerified:    true,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotFields = r.URL.Query().Get("fields")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	got, err := GetProfile(context.Background(), client, "12345")
	if err != nil {
		t.Fatalf("GetProfile returned error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if gotPath != "/12345" {
		t.Errorf("path = %q, want %q", gotPath, "/12345")
	}
	wantFields := "id,username,name,threads_biography,threads_profile_picture_url,is_verified"
	if gotFields != wantFields {
		t.Errorf("fields = %q, want %q", gotFields, wantFields)
	}
	if got.ID != want.ID {
		t.Errorf("ID = %q, want %q", got.ID, want.ID)
	}
	if got.Username != want.Username {
		t.Errorf("Username = %q, want %q", got.Username, want.Username)
	}
	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}
	if got.Bio != want.Bio {
		t.Errorf("Bio = %q, want %q", got.Bio, want.Bio)
	}
	if got.ProfilePicURL != want.ProfilePicURL {
		t.Errorf("ProfilePicURL = %q, want %q", got.ProfilePicURL, want.ProfilePicURL)
	}
	if got.IsVerified != want.IsVerified {
		t.Errorf("IsVerified = %v, want %v", got.IsVerified, want.IsVerified)
	}
}

func TestGetProfileError(t *testing.T) {
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

	client := api.NewClientWithHTTP("bad-tok", srv.URL, srv.Client())
	_, err := GetProfile(context.Background(), client, "12345")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLookupProfile(t *testing.T) {
	var gotMethod, gotPath, gotUsername, gotFields string

	want := Profile{
		Username:      "janedoe",
		Name:          "Jane Doe",
		IsVerified:    false,
		FollowerCount: 42,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotUsername = r.URL.Query().Get("username")
		gotFields = r.URL.Query().Get("fields")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	got, err := LookupProfile(context.Background(), client, "janedoe")
	if err != nil {
		t.Fatalf("LookupProfile returned error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if gotPath != "/profile_lookup" {
		t.Errorf("path = %q, want %q", gotPath, "/profile_lookup")
	}
	if gotUsername != "janedoe" {
		t.Errorf("username param = %q, want %q", gotUsername, "janedoe")
	}
	wantFields := "username,name,biography,profile_picture_url,follower_count,is_verified"
	if gotFields != wantFields {
		t.Errorf("fields = %q, want %q", gotFields, wantFields)
	}
	if got.Username != want.Username {
		t.Errorf("Username = %q, want %q", got.Username, want.Username)
	}
	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}
	if got.FollowerCount != want.FollowerCount {
		t.Errorf("FollowerCount = %d, want %d", got.FollowerCount, want.FollowerCount)
	}
}

func TestLookupProfileError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "user not found",
				"type":    "OAuthException",
				"code":    100,
			},
		})
	}))
	defer srv.Close()

	client := api.NewClientWithHTTP("tok", srv.URL, srv.Client())
	_, err := LookupProfile(context.Background(), client, "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
