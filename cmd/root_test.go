package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/malston/threads-cli/internal/config"
	"github.com/malston/threads-cli/internal/output"
	"github.com/spf13/cobra"
)

func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("token", "", "")
	cmd.Flags().String("format", "text", "")
	return cmd
}

func TestLoadClientPrefersTokenFlag(t *testing.T) {
	t.Setenv("THREADS_ACCESS_TOKEN", "env-token")
	configDir = t.TempDir()

	cmd := newTestCmd()
	cmd.Flags().Set("token", "flag-token")

	client, err := loadClient(cmd)
	if err != nil {
		t.Fatalf("loadClient returned error: %v", err)
	}
	if client == nil {
		t.Fatal("loadClient returned nil client")
	}
}

func TestLoadClientFallsBackToEnvVar(t *testing.T) {
	t.Setenv("THREADS_ACCESS_TOKEN", "env-token")
	configDir = t.TempDir() // empty dir, no credentials file

	cmd := newTestCmd()
	// --token flag not set, defaults to ""

	client, err := loadClient(cmd)
	if err != nil {
		t.Fatalf("loadClient returned error: %v", err)
	}
	if client == nil {
		t.Fatal("loadClient returned nil client")
	}
}

func TestLoadClientFallsBackToConfigStore(t *testing.T) {
	t.Setenv("THREADS_ACCESS_TOKEN", "")

	dir := t.TempDir()
	configDir = dir

	creds := config.Credentials{
		AccessToken: "stored-token",
		TokenType:   "bearer",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		UserID:      "12345",
	}
	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "credentials.json"), data, 0600); err != nil {
		t.Fatalf("write credentials: %v", err)
	}

	cmd := newTestCmd()

	client, err := loadClient(cmd)
	if err != nil {
		t.Fatalf("loadClient returned error: %v", err)
	}
	if client == nil {
		t.Fatal("loadClient returned nil client")
	}
}

func TestLoadClientReturnsErrorWhenNoToken(t *testing.T) {
	t.Setenv("THREADS_ACCESS_TOKEN", "")
	configDir = t.TempDir() // empty dir, no credentials file

	cmd := newTestCmd()

	_, err := loadClient(cmd)
	if err == nil {
		t.Fatal("loadClient should return error when no token is available")
	}
}

func TestGetFormatValidStrings(t *testing.T) {
	tests := []struct {
		input string
		want  output.Format
	}{
		{"json", output.JSON},
		{"text", output.Text},
		{"table", output.Table},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cmd := newTestCmd()
			cmd.Flags().Set("format", tt.input)

			got := getFormat(cmd)
			if got != tt.want {
				t.Errorf("getFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFormatDefaultsToText(t *testing.T) {
	cmd := newTestCmd()
	// format flag not set, defaults to "text"

	got := getFormat(cmd)
	if got != output.Text {
		t.Errorf("getFormat() = %v, want %v", got, output.Text)
	}
}

func TestVersionAndBuildTimeSettable(t *testing.T) {
	origVersion := Version
	origBuildTime := BuildTime
	t.Cleanup(func() {
		Version = origVersion
		BuildTime = origBuildTime
	})

	Version = "1.2.3"
	BuildTime = "2026-01-15T10:00:00Z"

	if Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", Version, "1.2.3")
	}
	if BuildTime != "2026-01-15T10:00:00Z" {
		t.Errorf("BuildTime = %q, want %q", BuildTime, "2026-01-15T10:00:00Z")
	}
}
