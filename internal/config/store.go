package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// Credentials holds the user's Threads API authentication data.
type Credentials struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
	UserID      string    `json:"user_id"`
}

// Store reads and writes credentials to a JSON file on disk.
type Store struct {
	dir string
}

const credentialsFile = "credentials.json"

// DefaultConfigDir returns the default configuration directory (~/.threads-cli).
func DefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".threads-cli")
	}
	return filepath.Join(home, ".threads-cli")
}

// NewStore creates a Store that manages credentials in the given directory.
func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

func (s *Store) path() string {
	return filepath.Join(s.dir, credentialsFile)
}

// Load reads credentials from the JSON file. Returns an error if the file does not exist.
func (s *Store) Load() (*Credentials, error) {
	data, err := os.ReadFile(s.path())
	if err != nil {
		return nil, err
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

// Save writes credentials to the JSON file, creating the directory if needed.
// The directory is created with 0700 permissions and the file with 0600.
func (s *Store) Save(creds *Credentials) error {
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return err
	}

	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}

	return os.WriteFile(s.path(), data, 0600)
}

// Clear deletes the credentials file. Returns nil if the file does not exist.
func (s *Store) Clear() error {
	err := os.Remove(s.path())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
