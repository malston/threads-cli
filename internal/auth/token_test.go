package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/malston/threads-cli/internal/auth"
	"github.com/malston/threads-cli/internal/config"
)

func TestExchangeCode(t *testing.T) {
	t.Run("sends correct form params and parses response", func(t *testing.T) {
		var gotMethod string
		var gotPath string
		var gotForm map[string]string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm: %v", err)
			}
			gotForm = map[string]string{
				"client_id":     r.FormValue("client_id"),
				"client_secret": r.FormValue("client_secret"),
				"code":          r.FormValue("code"),
				"grant_type":    r.FormValue("grant_type"),
				"redirect_uri":  r.FormValue("redirect_uri"),
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"access_token": "short-lived-token-abc",
				"user_id":      12345,
			})
		}))
		defer srv.Close()

		creds, err := auth.ExchangeCode(
			context.Background(),
			srv.Client(),
			srv.URL,
			"my-app-id",
			"my-app-secret",
			"http://localhost:8080/callback",
			"auth-code-xyz",
		)
		if err != nil {
			t.Fatalf("ExchangeCode() error = %v", err)
		}

		if gotMethod != http.MethodPost {
			t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
		}
		if gotPath != "/oauth/access_token" {
			t.Errorf("path = %q, want %q", gotPath, "/oauth/access_token")
		}

		wantForm := map[string]string{
			"client_id":     "my-app-id",
			"client_secret": "my-app-secret",
			"code":          "auth-code-xyz",
			"grant_type":    "authorization_code",
			"redirect_uri":  "http://localhost:8080/callback",
		}
		for k, want := range wantForm {
			if gotForm[k] != want {
				t.Errorf("form[%q] = %q, want %q", k, gotForm[k], want)
			}
		}

		if creds.AccessToken != "short-lived-token-abc" {
			t.Errorf("AccessToken = %q, want %q", creds.AccessToken, "short-lived-token-abc")
		}
		if creds.UserID != "12345" {
			t.Errorf("UserID = %q, want %q", creds.UserID, "12345")
		}
		if creds.TokenType != "bearer" {
			t.Errorf("TokenType = %q, want %q", creds.TokenType, "bearer")
		}
	})

	t.Run("returns error on non-200 response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid_code"}`))
		}))
		defer srv.Close()

		_, err := auth.ExchangeCode(
			context.Background(),
			srv.Client(),
			srv.URL,
			"app", "secret", "http://localhost/cb", "bad-code",
		)
		if err == nil {
			t.Fatal("ExchangeCode() expected error for non-200 response, got nil")
		}
	})
}

func TestExchangeLongLived(t *testing.T) {
	t.Run("sends correct query params and computes ExpiresAt", func(t *testing.T) {
		var gotMethod string
		var gotPath string
		var gotQuery map[string]string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			gotQuery = map[string]string{
				"grant_type":    r.URL.Query().Get("grant_type"),
				"client_secret": r.URL.Query().Get("client_secret"),
				"access_token":  r.URL.Query().Get("access_token"),
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"access_token": "long-lived-token-xyz",
				"token_type":   "bearer",
				"expires_in":   5183944,
			})
		}))
		defer srv.Close()

		before := time.Now()
		creds, err := auth.ExchangeLongLived(
			context.Background(),
			srv.Client(),
			srv.URL,
			"my-app-secret",
			"short-token",
		)
		after := time.Now()

		if err != nil {
			t.Fatalf("ExchangeLongLived() error = %v", err)
		}

		if gotMethod != http.MethodGet {
			t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
		}
		if gotPath != "/access_token" {
			t.Errorf("path = %q, want %q", gotPath, "/access_token")
		}

		wantQuery := map[string]string{
			"grant_type":    "th_exchange_token",
			"client_secret": "my-app-secret",
			"access_token":  "short-token",
		}
		for k, want := range wantQuery {
			if gotQuery[k] != want {
				t.Errorf("query[%q] = %q, want %q", k, gotQuery[k], want)
			}
		}

		if creds.AccessToken != "long-lived-token-xyz" {
			t.Errorf("AccessToken = %q, want %q", creds.AccessToken, "long-lived-token-xyz")
		}
		if creds.TokenType != "bearer" {
			t.Errorf("TokenType = %q, want %q", creds.TokenType, "bearer")
		}

		expectedExpiry := 5183944 * time.Second
		earliestExpiry := before.Add(expectedExpiry)
		latestExpiry := after.Add(expectedExpiry)
		if creds.ExpiresAt.Before(earliestExpiry) || creds.ExpiresAt.After(latestExpiry) {
			t.Errorf("ExpiresAt = %v, want between %v and %v", creds.ExpiresAt, earliestExpiry, latestExpiry)
		}
	})

	t.Run("returns error on non-200 response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "invalid_token"}`))
		}))
		defer srv.Close()

		_, err := auth.ExchangeLongLived(
			context.Background(),
			srv.Client(),
			srv.URL,
			"secret",
			"bad-token",
		)
		if err == nil {
			t.Fatal("ExchangeLongLived() expected error for non-200 response, got nil")
		}
	})
}

func TestRefresh(t *testing.T) {
	t.Run("sends correct query params and computes ExpiresAt", func(t *testing.T) {
		var gotMethod string
		var gotPath string
		var gotQuery map[string]string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			gotQuery = map[string]string{
				"grant_type":   r.URL.Query().Get("grant_type"),
				"access_token": r.URL.Query().Get("access_token"),
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"access_token": "refreshed-token-abc",
				"token_type":   "bearer",
				"expires_in":   5183944,
			})
		}))
		defer srv.Close()

		before := time.Now()
		creds, err := auth.Refresh(
			context.Background(),
			srv.Client(),
			srv.URL,
			"old-token",
		)
		after := time.Now()

		if err != nil {
			t.Fatalf("Refresh() error = %v", err)
		}

		if gotMethod != http.MethodGet {
			t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
		}
		if gotPath != "/refresh_access_token" {
			t.Errorf("path = %q, want %q", gotPath, "/refresh_access_token")
		}

		wantQuery := map[string]string{
			"grant_type":   "th_refresh_token",
			"access_token": "old-token",
		}
		for k, want := range wantQuery {
			if gotQuery[k] != want {
				t.Errorf("query[%q] = %q, want %q", k, gotQuery[k], want)
			}
		}

		if creds.AccessToken != "refreshed-token-abc" {
			t.Errorf("AccessToken = %q, want %q", creds.AccessToken, "refreshed-token-abc")
		}
		if creds.TokenType != "bearer" {
			t.Errorf("TokenType = %q, want %q", creds.TokenType, "bearer")
		}

		expectedExpiry := 5183944 * time.Second
		earliestExpiry := before.Add(expectedExpiry)
		latestExpiry := after.Add(expectedExpiry)
		if creds.ExpiresAt.Before(earliestExpiry) || creds.ExpiresAt.After(latestExpiry) {
			t.Errorf("ExpiresAt = %v, want between %v and %v", creds.ExpiresAt, earliestExpiry, latestExpiry)
		}
	})

	t.Run("returns error on non-200 response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		_, err := auth.Refresh(
			context.Background(),
			srv.Client(),
			srv.URL,
			"expired-token",
		)
		if err == nil {
			t.Fatal("Refresh() expected error for non-200 response, got nil")
		}
	})
}

func TestIsExpired(t *testing.T) {
	t.Run("returns true when ExpiresAt is in the past", func(t *testing.T) {
		creds := &config.Credentials{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		if !auth.IsExpired(creds) {
			t.Error("IsExpired() = false, want true for past ExpiresAt")
		}
	})

	t.Run("returns false when ExpiresAt is in the future", func(t *testing.T) {
		creds := &config.Credentials{
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		}
		if auth.IsExpired(creds) {
			t.Error("IsExpired() = true, want false for future ExpiresAt")
		}
	})
}

func TestNeedsRefresh(t *testing.T) {
	t.Run("returns true when ExpiresAt is within 7 days", func(t *testing.T) {
		creds := &config.Credentials{
			ExpiresAt: time.Now().Add(3 * 24 * time.Hour),
		}
		if !auth.NeedsRefresh(creds) {
			t.Error("NeedsRefresh() = false, want true when expiry is within 7 days")
		}
	})

	t.Run("returns false when ExpiresAt is beyond 7 days", func(t *testing.T) {
		creds := &config.Credentials{
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		}
		if auth.NeedsRefresh(creds) {
			t.Error("NeedsRefresh() = true, want false when expiry is beyond 7 days")
		}
	})

	t.Run("returns true when already expired", func(t *testing.T) {
		creds := &config.Credentials{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		if !auth.NeedsRefresh(creds) {
			t.Error("NeedsRefresh() = false, want true when already expired")
		}
	})
}
