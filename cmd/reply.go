package cmd

import (
	"fmt"
	"net/http"

	"github.com/malston/threads-cli/internal/api"
	"github.com/malston/threads-cli/internal/output"
	"github.com/malston/threads-cli/internal/threads"
	"github.com/spf13/cobra"
)

// replyBaseURL is the Threads API base URL for reply operations.
// Tests override this to point at an httptest server.
var replyBaseURL = "https://graph.threads.net"

func init() {
	replyCmd := &cobra.Command{Use: "reply", Short: "Manage replies"}

	listCmd := &cobra.Command{
		Use:   "list [post-id]",
		Short: "List replies",
		Args:  cobra.ExactArgs(1),
		RunE:  runReplyList,
	}
	listCmd.Flags().Int("limit", 25, "max results")
	listCmd.Flags().String("before", "", "cursor")
	listCmd.Flags().String("after", "", "cursor")

	createCmd := &cobra.Command{
		Use:   "create [post-id] [text]",
		Short: "Reply to a post",
		Args:  cobra.ExactArgs(2),
		RunE:  runReplyCreate,
	}

	hideCmd := &cobra.Command{
		Use:   "hide [reply-id]",
		Short: "Hide a reply",
		Args:  cobra.ExactArgs(1),
		RunE:  runReplyHide,
	}

	unhideCmd := &cobra.Command{
		Use:   "unhide [reply-id]",
		Short: "Unhide a reply",
		Args:  cobra.ExactArgs(1),
		RunE:  runReplyUnhide,
	}

	replyCmd.AddCommand(listCmd, createCmd, hideCmd, unhideCmd)
	rootCmd.AddCommand(replyCmd)
}

func runReplyList(cmd *cobra.Command, args []string) error {
	postID := args[0]
	ctx := cmd.Context()

	client, err := loadReplyClient(cmd)
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

	list, err := threads.ListReplies(ctx, client, postID, page)
	if err != nil {
		return fmt.Errorf("listing replies: %w", err)
	}

	return output.Print(cmd.OutOrStdout(), list.Data, getFormat(cmd))
}

func runReplyCreate(cmd *cobra.Command, args []string) error {
	replyToID := args[0]
	text := args[1]
	ctx := cmd.Context()

	client, err := loadReplyClient(cmd)
	if err != nil {
		return err
	}

	userID, err := loadUserID()
	if err != nil {
		return err
	}

	replyID, err := threads.CreateReply(ctx, client, userID, replyToID, text)
	if err != nil {
		return fmt.Errorf("creating reply: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Published reply %s\n", replyID)
	return nil
}

func runReplyHide(cmd *cobra.Command, args []string) error {
	replyID := args[0]
	ctx := cmd.Context()

	client, err := loadReplyClient(cmd)
	if err != nil {
		return err
	}

	if err := threads.HideReply(ctx, client, replyID); err != nil {
		return fmt.Errorf("hiding reply: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Hidden reply %s\n", replyID)
	return nil
}

func runReplyUnhide(cmd *cobra.Command, args []string) error {
	replyID := args[0]
	ctx := cmd.Context()

	client, err := loadReplyClient(cmd)
	if err != nil {
		return err
	}

	if err := threads.UnhideReply(ctx, client, replyID); err != nil {
		return fmt.Errorf("unhiding reply: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Unhidden reply %s\n", replyID)
	return nil
}

// loadReplyClient creates an API client pointing at replyBaseURL.
func loadReplyClient(cmd *cobra.Command) (*api.Client, error) {
	token, _ := cmd.Flags().GetString("token")
	if token == "" {
		client, err := loadClient(cmd)
		if err != nil {
			return nil, err
		}
		if replyBaseURL != "https://graph.threads.net" {
			return api.NewClientWithHTTP(token, replyBaseURL, &http.Client{}), nil
		}
		return client, nil
	}
	if replyBaseURL != "https://graph.threads.net" {
		return api.NewClientWithHTTP(token, replyBaseURL, &http.Client{}), nil
	}
	return api.NewClient(token), nil
}
