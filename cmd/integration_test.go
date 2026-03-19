//go:build integration

package cmd

import (
	"bytes"
	"os"
	"testing"
)

// Integration tests require THREADS_ACCESS_TOKEN env var.
// Run with: go test -tags integration ./cmd/...

func requireToken(t *testing.T) string {
	t.Helper()
	token := os.Getenv("THREADS_ACCESS_TOKEN")
	if token == "" {
		t.Skip("THREADS_ACCESS_TOKEN not set")
	}
	return token
}

func TestIntegrationProfile(t *testing.T) {
	token := requireToken(t)

	cmd := rootCmd
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"profile", "--token", token, "--format", "json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("profile command failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected output, got empty")
	}
}

func TestIntegrationLimits(t *testing.T) {
	token := requireToken(t)

	cmd := rootCmd
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"limits", "--token", token, "--format", "json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("limits command failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected output, got empty")
	}
}

func TestIntegrationPostList(t *testing.T) {
	token := requireToken(t)

	cmd := rootCmd
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"post", "list", "--token", token, "--format", "json", "--limit", "5"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("post list command failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected output, got empty")
	}
}
