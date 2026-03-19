package cmd

import (
	"fmt"
	"net/http"

	"github.com/malston/threads-cli/internal/api"
	"github.com/malston/threads-cli/internal/output"
	"github.com/malston/threads-cli/internal/threads"
	"github.com/spf13/cobra"
)

// searchBaseURL is the Threads API base URL for search operations.
// Tests override this to point at an httptest server.
var searchBaseURL = "https://graph.threads.net"

func init() {
	searchCmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search posts",
		Args:  cobra.ExactArgs(1),
		RunE:  runSearch,
	}
	searchCmd.Flags().Int("limit", 25, "max results")
	searchCmd.Flags().String("before", "", "cursor")
	searchCmd.Flags().String("after", "", "cursor")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]
	ctx := cmd.Context()

	client, err := loadSearchClient(cmd)
	if err != nil {
		return err
	}

	limit, _ := cmd.Flags().GetInt("limit")
	before, _ := cmd.Flags().GetString("before")
	after, _ := cmd.Flags().GetString("after")

	page := &api.PageParams{
		Limit:  limit,
		Before: before,
		After:  after,
	}

	list, err := threads.Search(ctx, client, query, page)
	if err != nil {
		return fmt.Errorf("searching posts: %w", err)
	}

	return output.Print(cmd.OutOrStdout(), list.Data, getFormat(cmd))
}

// loadSearchClient creates an API client pointing at searchBaseURL.
func loadSearchClient(cmd *cobra.Command) (*api.Client, error) {
	token, _ := cmd.Flags().GetString("token")
	if token == "" {
		client, err := loadClient(cmd)
		if err != nil {
			return nil, err
		}
		if searchBaseURL != "https://graph.threads.net" {
			return api.NewClientWithHTTP(token, searchBaseURL, &http.Client{}), nil
		}
		return client, nil
	}
	if searchBaseURL != "https://graph.threads.net" {
		return api.NewClientWithHTTP(token, searchBaseURL, &http.Client{}), nil
	}
	return api.NewClient(token), nil
}
