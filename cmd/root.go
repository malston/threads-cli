package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/malston/threads-cli/internal/api"
	"github.com/malston/threads-cli/internal/config"
	"github.com/malston/threads-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	Version   string
	BuildTime string

	// configDir is the directory used to load stored credentials.
	configDir = config.DefaultConfigDir()
)

var rootCmd = &cobra.Command{
	Use:   "threads",
	Short: "CLI for the Meta Threads API",
	Long: `CLI wrapper for the Meta Threads API.

Manage posts, replies, profiles, insights, and rate limits
from the command line. Authenticate via OAuth, then use
subcommands to interact with your Threads account.

Token resolution order: --token flag > THREADS_ACCESS_TOKEN env > stored credentials.`,
}

func init() {
	rootCmd.PersistentFlags().String("format", "text", "output format: json, text, or table")
	rootCmd.PersistentFlags().String("token", "", "Threads API access token")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// resolveToken returns the access token from the first available source:
// --token flag, THREADS_ACCESS_TOKEN env var, or stored credentials.
func resolveToken(cmd *cobra.Command) (string, error) {
	token, _ := cmd.Flags().GetString("token")
	if token != "" {
		return token, nil
	}

	token = os.Getenv("THREADS_ACCESS_TOKEN")
	if token != "" {
		return token, nil
	}

	store := config.NewStore(configDir)
	creds, err := store.Load()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("reading stored credentials: %w", err)
		}
	} else if creds.AccessToken != "" {
		return creds.AccessToken, nil
	}

	return "", fmt.Errorf("no access token: set --token, THREADS_ACCESS_TOKEN, or run 'threads auth login'")
}

// loadClient creates an API client with a token resolved via resolveToken.
func loadClient(cmd *cobra.Command) (*api.Client, error) {
	token, err := resolveToken(cmd)
	if err != nil {
		return nil, err
	}
	return api.NewClient(token), nil
}

// getFormat reads the --format flag and parses it into an output.Format.
func getFormat(cmd *cobra.Command) (output.Format, error) {
	s, _ := cmd.Flags().GetString("format")
	return output.ParseFormat(s)
}
