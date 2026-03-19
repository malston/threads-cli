package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/malston/threads-cli/internal/auth"
	"github.com/malston/threads-cli/internal/config"
	"github.com/spf13/cobra"
)

// authBaseURL is the Threads API base URL for token operations.
// Tests override this to point at an httptest server.
var authBaseURL = "https://graph.threads.net"

func init() {
	authCmd := &cobra.Command{Use: "auth", Short: "Authentication commands"}

	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Threads",
		RunE:  runAuthLogin,
	}

	tokenCmd := &cobra.Command{
		Use:   "token",
		Short: "Show or refresh token status",
		RunE:  runAuthToken,
	}
	tokenCmd.Flags().Bool("refresh", false, "refresh the token if needed")

	authCmd.AddCommand(loginCmd, tokenCmd)
	rootCmd.AddCommand(authCmd)
}

func runAuthLogin(cmd *cobra.Command, _ []string) error {
	appID := os.Getenv("THREADS_APP_ID")
	appSecret := os.Getenv("THREADS_APP_SECRET")
	if appID == "" || appSecret == "" {
		return fmt.Errorf("THREADS_APP_ID and THREADS_APP_SECRET environment variables are required")
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
	defer cancel()

	port, codeChan, errChan, shutdown := auth.StartCallbackServer(ctx)
	defer shutdown()

	redirectURI := fmt.Sprintf("https://localhost:%d/callback", port)
	authURL := auth.BuildAuthURL(appID, redirectURI, "cli-login", auth.DefaultScopes())

	fmt.Fprintln(cmd.OutOrStdout(), "Open this URL in your browser to authenticate:")
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), authURL)
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "Waiting for authorization...")

	var code string
	select {
	case code = <-codeChan:
	case err := <-errChan:
		return fmt.Errorf("callback server error: %w", err)
	case <-ctx.Done():
		return fmt.Errorf("timed out waiting for authorization")
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}

	shortCreds, err := auth.ExchangeCode(ctx, httpClient, authBaseURL, appID, appSecret, redirectURI, code)
	if err != nil {
		return fmt.Errorf("exchanging code for token: %w", err)
	}

	longCreds, err := auth.ExchangeLongLived(ctx, httpClient, authBaseURL, appSecret, shortCreds.AccessToken)
	if err != nil {
		return fmt.Errorf("exchanging for long-lived token: %w", err)
	}
	longCreds.UserID = shortCreds.UserID

	store := config.NewStore(configDir)
	if err := store.Save(longCreds); err != nil {
		return fmt.Errorf("saving credentials: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Authentication successful! Credentials saved.")
	return nil
}

func runAuthToken(cmd *cobra.Command, _ []string) error {
	store := config.NewStore(configDir)
	creds, err := store.Load()
	if err != nil {
		return fmt.Errorf("no stored credentials: run 'threads auth login' first")
	}

	refresh, _ := cmd.Flags().GetBool("refresh")

	if refresh && auth.NeedsRefresh(creds) {
		httpClient := &http.Client{Timeout: 30 * time.Second}
		ctx := cmd.Context()

		refreshed, err := auth.Refresh(ctx, httpClient, authBaseURL, creds.AccessToken)
		if err != nil {
			return fmt.Errorf("refreshing token: %w", err)
		}
		refreshed.UserID = creds.UserID

		if err := store.Save(refreshed); err != nil {
			return fmt.Errorf("saving refreshed credentials: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Token refreshed successfully.\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  Token:   %s\n", maskToken(refreshed.AccessToken))
		fmt.Fprintf(cmd.OutOrStdout(), "  Expires: %s\n", refreshed.ExpiresAt.Format(time.RFC3339))
		return nil
	}

	if refresh && !auth.NeedsRefresh(creds) {
		fmt.Fprintf(cmd.OutOrStdout(), "Token does not need refresh.\n")
	}

	status := "valid"
	if auth.IsExpired(creds) {
		status = "expired"
	} else if auth.NeedsRefresh(creds) {
		status = "valid (expires soon, refresh recommended)"
	}

	fmt.Fprintf(cmd.OutOrStdout(), "  Token:   %s\n", maskToken(creds.AccessToken))
	fmt.Fprintf(cmd.OutOrStdout(), "  Status:  %s\n", status)
	fmt.Fprintf(cmd.OutOrStdout(), "  Expires: %s\n", creds.ExpiresAt.Format(time.RFC3339))
	fmt.Fprintf(cmd.OutOrStdout(), "  User ID: %s\n", creds.UserID)

	return nil
}

// maskToken shows the first few characters of a token followed by "...".
func maskToken(token string) string {
	visible := len(token) / 2
	if visible > 6 {
		visible = 6
	}
	return token[:visible] + "..."
}
