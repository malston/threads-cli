package cmd

import (
	"fmt"
	"net/http"

	"github.com/malston/threads-cli/internal/api"
	"github.com/malston/threads-cli/internal/output"
	"github.com/malston/threads-cli/internal/threads"
	"github.com/spf13/cobra"
)

// limitsBaseURL is the Threads API base URL for rate limit operations.
// Tests override this to point at an httptest server.
var limitsBaseURL = "https://graph.threads.net"

func init() {
	limitsCmd := &cobra.Command{
		Use:   "limits",
		Short: "Show rate limit status",
		RunE:  runLimits,
	}
	rootCmd.AddCommand(limitsCmd)
}

func runLimits(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	client, err := loadLimitsClient(cmd)
	if err != nil {
		return err
	}

	userID, err := loadUserID()
	if err != nil {
		return err
	}

	status, err := threads.GetLimits(ctx, client, userID)
	if err != nil {
		return fmt.Errorf("fetching rate limits: %w", err)
	}

	return output.Print(cmd.OutOrStdout(), status, getFormat(cmd))
}

// loadLimitsClient creates an API client pointing at limitsBaseURL.
func loadLimitsClient(cmd *cobra.Command) (*api.Client, error) {
	token, err := resolveToken(cmd)
	if err != nil {
		return nil, err
	}
	if limitsBaseURL != "https://graph.threads.net" {
		return api.NewClientWithHTTP(token, limitsBaseURL, &http.Client{}), nil
	}
	return api.NewClient(token), nil
}
