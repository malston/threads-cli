package cmd

import (
	"fmt"
	"net/http"

	"github.com/malston/threads-cli/internal/api"
	"github.com/malston/threads-cli/internal/output"
	"github.com/malston/threads-cli/internal/threads"
	"github.com/spf13/cobra"
)

// insightsBaseURL is the Threads API base URL for insights operations.

var insightsBaseURL = "https://graph.threads.net"

func init() {
	insightsCmd := &cobra.Command{Use: "insights", Short: "View insights"}

	postCmd := &cobra.Command{
		Use:   "post [id]",
		Short: "Post insights",
		Args:  cobra.ExactArgs(1),
		RunE:  runInsightsPost,
	}
	postCmd.Flags().StringSlice("metrics", []string{"views", "likes", "replies", "reposts", "quotes"}, "metrics to fetch")

	userCmd := &cobra.Command{
		Use:   "user",
		Short: "User insights",
		RunE:  runInsightsUser,
	}
	userCmd.Flags().StringSlice("metrics", []string{"views", "likes", "replies", "reposts", "quotes", "followers_count"}, "metrics")
	userCmd.Flags().Int64("since", 0, "start timestamp (unix)")
	userCmd.Flags().Int64("until", 0, "end timestamp (unix)")

	insightsCmd.AddCommand(postCmd, userCmd)
	rootCmd.AddCommand(insightsCmd)
}

func runInsightsPost(cmd *cobra.Command, args []string) error {
	postID := args[0]
	ctx := cmd.Context()

	client, err := loadInsightsClient(cmd)
	if err != nil {
		return err
	}

	metrics, _ := cmd.Flags().GetStringSlice("metrics")

	data, err := threads.PostInsights(ctx, client, postID, metrics)
	if err != nil {
		return fmt.Errorf("fetching post insights: %w", err)
	}

	format, err := getFormat(cmd)
	if err != nil {
		return err
	}
	return output.Print(cmd.OutOrStdout(), data, format)
}

func runInsightsUser(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	client, err := loadInsightsClient(cmd)
	if err != nil {
		return err
	}

	userID, err := loadUserID()
	if err != nil {
		return err
	}

	metrics, _ := cmd.Flags().GetStringSlice("metrics")
	since, _ := cmd.Flags().GetInt64("since")
	until, _ := cmd.Flags().GetInt64("until")

	data, err := threads.UserInsights(ctx, client, userID, metrics, since, until)
	if err != nil {
		return fmt.Errorf("fetching user insights: %w", err)
	}

	format, err := getFormat(cmd)
	if err != nil {
		return err
	}
	return output.Print(cmd.OutOrStdout(), data, format)
}

// loadInsightsClient creates an API client pointing at insightsBaseURL.
func loadInsightsClient(cmd *cobra.Command) (*api.Client, error) {
	token, err := resolveToken(cmd)
	if err != nil {
		return nil, err
	}
	if insightsBaseURL != "https://graph.threads.net" {
		return api.NewClientWithHTTP(token, insightsBaseURL, &http.Client{}), nil
	}
	return api.NewClient(token), nil
}
