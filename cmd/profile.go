package cmd

import (
	"fmt"
	"net/http"

	"github.com/malston/threads-cli/internal/api"
	"github.com/malston/threads-cli/internal/output"
	"github.com/malston/threads-cli/internal/threads"
	"github.com/spf13/cobra"
)

// profileBaseURL is the Threads API base URL for profile operations.
// Tests override this to point at an httptest server.
var profileBaseURL = "https://graph.threads.net"

func init() {
	profileCmd := &cobra.Command{
		Use:   "profile [username]",
		Short: "Show profile",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runProfile,
	}
	rootCmd.AddCommand(profileCmd)
}

func runProfile(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	client, err := loadProfileClient(cmd)
	if err != nil {
		return err
	}

	var profile *threads.Profile

	if len(args) == 0 {
		userID, err := loadUserID()
		if err != nil {
			return err
		}
		profile, err = threads.GetProfile(ctx, client, userID)
		if err != nil {
			return fmt.Errorf("getting profile: %w", err)
		}
	} else {
		profile, err = threads.LookupProfile(ctx, client, args[0])
		if err != nil {
			return fmt.Errorf("looking up profile: %w", err)
		}
	}

	return output.Print(cmd.OutOrStdout(), profile, getFormat(cmd))
}

// loadProfileClient creates an API client pointing at profileBaseURL.
func loadProfileClient(cmd *cobra.Command) (*api.Client, error) {
	token, _ := cmd.Flags().GetString("token")
	if token == "" {
		client, err := loadClient(cmd)
		if err != nil {
			return nil, err
		}
		if profileBaseURL != "https://graph.threads.net" {
			return api.NewClientWithHTTP(token, profileBaseURL, &http.Client{}), nil
		}
		return client, nil
	}
	if profileBaseURL != "https://graph.threads.net" {
		return api.NewClientWithHTTP(token, profileBaseURL, &http.Client{}), nil
	}
	return api.NewClient(token), nil
}
