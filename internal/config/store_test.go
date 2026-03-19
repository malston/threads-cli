package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	want := &Credentials{
		AccessToken: "test-token-abc123",
		TokenType:   "bearer",
		ExpiresAt:   time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
		UserID:      "12345678",
	}

	if err := store.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.AccessToken != want.AccessToken {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, want.AccessToken)
	}
	if got.TokenType != want.TokenType {
		t.Errorf("TokenType = %q, want %q", got.TokenType, want.TokenType)
	}
	if !got.ExpiresAt.Equal(want.ExpiresAt) {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, want.ExpiresAt)
	}
	if got.UserID != want.UserID {
		t.Errorf("UserID = %q, want %q", got.UserID, want.UserID)
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, err := store.Load()
	if err == nil {
		t.Fatal("Load from nonexistent file should return an error")
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	base := t.TempDir()
	nested := filepath.Join(base, "subdir", "config")
	store := NewStore(nested)

	creds := &Credentials{
		AccessToken: "token",
		TokenType:   "bearer",
		ExpiresAt:   time.Now().UTC(),
		UserID:      "1",
	}

	if err := store.Save(creds); err != nil {
		t.Fatalf("Save: %v", err)
	}

	info, err := os.Stat(nested)
	if err != nil {
		t.Fatalf("Stat directory: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected a directory")
	}
}

func TestClearDeletesFile(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	creds := &Credentials{
		AccessToken: "token",
		TokenType:   "bearer",
		ExpiresAt:   time.Now().UTC(),
		UserID:      "1",
	}

	if err := store.Save(creds); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := store.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	credFile := filepath.Join(dir, "credentials.json")
	if _, err := os.Stat(credFile); !os.IsNotExist(err) {
		t.Fatalf("expected credentials file to be deleted, got err: %v", err)
	}
}

func TestClearNonexistentFileNoError(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	if err := store.Clear(); err != nil {
		t.Fatalf("Clear on nonexistent file should not error, got: %v", err)
	}
}

func TestSaveFilePermissions(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	creds := &Credentials{
		AccessToken: "token",
		TokenType:   "bearer",
		ExpiresAt:   time.Now().UTC(),
		UserID:      "1",
	}

	if err := store.Save(creds); err != nil {
		t.Fatalf("Save: %v", err)
	}

	credFile := filepath.Join(dir, "credentials.json")
	info, err := os.Stat(credFile)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}
