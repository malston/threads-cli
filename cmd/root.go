package cmd

import (
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
	// Tests override this to use a temporary directory.
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

// loadClient creates an API client using the first available token source:
// --token flag, THREADS_ACCESS_TOKEN env var, or stored credentials.
func loadClient(cmd *cobra.Command) (*api.Client, error) {
	token, _ := cmd.Flags().GetString("token")
	if token != "" {
		return api.NewClient(token), nil
	}

	token = os.Getenv("THREADS_ACCESS_TOKEN")
	if token != "" {
		return api.NewClient(token), nil
	}

	store := config.NewStore(configDir)
	creds, err := store.Load()
	if err == nil && creds.AccessToken != "" {
		return api.NewClient(creds.AccessToken), nil
	}

	return nil, fmt.Errorf("no access token: set --token, THREADS_ACCESS_TOKEN, or run 'threads auth login'")
}

// getFormat reads the --format flag and parses it into an output.Format.
// Defaults to output.Text if the flag value is unrecognized.
func getFormat(cmd *cobra.Command) output.Format {
	s, _ := cmd.Flags().GetString("format")
	f, err := output.ParseFormat(s)
	if err != nil {
		return output.Text
	}
	return f
}
