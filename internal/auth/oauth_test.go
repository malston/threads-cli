package auth_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/malston/saved-threads/internal/auth"
)

func TestBuildAuthURL(t *testing.T) {
	t.Run("produces correct URL with all params encoded", func(t *testing.T) {
		got := auth.BuildAuthURL("12345", "http://localhost:8080/callback", "abc123", []string{"threads_basic", "threads_content_publish"})

		parsed, err := url.Parse(got)
		if err != nil {
			t.Fatalf("BuildAuthURL returned invalid URL: %v", err)
		}

		if parsed.Scheme != "https" {
			t.Errorf("scheme = %q, want %q", parsed.Scheme, "https")
		}
		if parsed.Host != "threads.net" {
			t.Errorf("host = %q, want %q", parsed.Host, "threads.net")
		}
		if parsed.Path != "/oauth/authorize" {
			t.Errorf("path = %q, want %q", parsed.Path, "/oauth/authorize")
		}

		q := parsed.Query()
		if q.Get("client_id") != "12345" {
			t.Errorf("client_id = %q, want %q", q.Get("client_id"), "12345")
		}
		if q.Get("redirect_uri") != "http://localhost:8080/callback" {
			t.Errorf("redirect_uri = %q, want %q", q.Get("redirect_uri"), "http://localhost:8080/callback")
		}
		if q.Get("response_type") != "code" {
			t.Errorf("response_type = %q, want %q", q.Get("response_type"), "code")
		}
		if q.Get("scope") != "threads_basic,threads_content_publish" {
			t.Errorf("scope = %q, want %q", q.Get("scope"), "threads_basic,threads_content_publish")
		}
		if q.Get("state") != "abc123" {
			t.Errorf("state = %q, want %q", q.Get("state"), "abc123")
		}
	})

	t.Run("encodes special characters in state param", func(t *testing.T) {
		got := auth.BuildAuthURL("12345", "http://localhost:8080/callback", "state with spaces&special=chars", []string{"threads_basic"})

		parsed, err := url.Parse(got)
		if err != nil {
			t.Fatalf("BuildAuthURL returned invalid URL: %v", err)
		}

		q := parsed.Query()
		if q.Get("state") != "state with spaces&special=chars" {
			t.Errorf("state = %q, want %q", q.Get("state"), "state with spaces&special=chars")
		}
	})
}

func TestStartCallbackServer(t *testing.T) {
	t.Run("starts on a random port and returns it", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		port, _, _, shutdown := auth.StartCallbackServer(ctx)
		defer shutdown()

		if port <= 0 {
			t.Fatalf("port = %d, want > 0", port)
		}
	})

	t.Run("receives code via callback and strips #_ suffix", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		port, codeChan, errChan, shutdown := auth.StartCallbackServer(ctx)
		defer shutdown()

		callbackURL := fmt.Sprintf("http://localhost:%d/callback?code=test_code%%23_", port)
		resp, err := http.Get(callbackURL)
		if err != nil {
			t.Fatalf("GET %s failed: %v", callbackURL, err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("reading response body: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		bodyStr := string(body)
		if bodyStr == "" {
			t.Error("response body is empty")
		}

		select {
		case code := <-codeChan:
			if code != "test_code" {
				t.Errorf("code = %q, want %q", code, "test_code")
			}
		case err := <-errChan:
			t.Fatalf("unexpected error: %v", err)
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for code")
		}
	})

	t.Run("context cancellation shuts down server", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		port, _, _, _ := auth.StartCallbackServer(ctx)

		// Cancel context to trigger shutdown
		cancel()

		// Give server time to shut down
		time.Sleep(100 * time.Millisecond)

		// Attempt a request -- should fail because the server is shut down
		callbackURL := fmt.Sprintf("http://localhost:%d/callback?code=test", port)
		_, err := http.Get(callbackURL)
		if err == nil {
			t.Error("expected error after context cancellation, got nil")
		}
	})
}

func TestDefaultScopes(t *testing.T) {
	expected := []string{
		"threads_basic",
		"threads_content_publish",
		"threads_manage_insights",
		"threads_manage_replies",
		"threads_read_replies",
	}

	got := auth.DefaultScopes()

	if len(got) != len(expected) {
		t.Fatalf("len(DefaultScopes()) = %d, want %d", len(got), len(expected))
	}

	for i, scope := range expected {
		if got[i] != scope {
			t.Errorf("DefaultScopes()[%d] = %q, want %q", i, got[i], scope)
		}
	}
}
