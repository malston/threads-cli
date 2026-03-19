package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/malston/saved-threads/internal/api"
	"github.com/malston/saved-threads/internal/config"
	"github.com/malston/saved-threads/internal/output"
	"github.com/malston/saved-threads/internal/threads"
	"github.com/spf13/cobra"
)

// apiBaseURL is the Threads API base URL for post operations.
// Tests override this to point at an httptest server.
var apiBaseURL = "https://graph.threads.net"

func init() {
	postCmd := &cobra.Command{Use: "post", Short: "Manage posts"}

	createCmd := &cobra.Command{
		Use:   "create [text]",
		Short: "Create a post",
		Args:  cobra.ExactArgs(1),
		RunE:  runPostCreate,
	}
	createCmd.Flags().String("image", "", "image URL")
	createCmd.Flags().String("video", "", "video URL")
	createCmd.Flags().StringSlice("carousel", nil, "carousel image URLs")
	createCmd.MarkFlagsMutuallyExclusive("image", "video", "carousel")

	getCmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get a post",
		Args:  cobra.ExactArgs(1),
		RunE:  runPostGet,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List posts",
		RunE:  runPostList,
	}
	listCmd.Flags().Int("limit", 25, "max results")
	listCmd.Flags().String("before", "", "cursor")
	listCmd.Flags().String("after", "", "cursor")

	deleteCmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a post",
		Args:  cobra.ExactArgs(1),
		RunE:  runPostDelete,
	}

	repostCmd := &cobra.Command{
		Use:   "repost [id]",
		Short: "Repost",
		Args:  cobra.ExactArgs(1),
		RunE:  runPostRepost,
	}

	postCmd.AddCommand(createCmd, getCmd, listCmd, deleteCmd, repostCmd)
	rootCmd.AddCommand(postCmd)
}

func runPostCreate(cmd *cobra.Command, args []string) error {
	text := args[0]
	ctx := cmd.Context()

	client, err := loadPostClient(cmd)
	if err != nil {
		return err
	}

	userID, err := loadUserID()
	if err != nil {
		return err
	}

	imageURL, _ := cmd.Flags().GetString("image")
	videoURL, _ := cmd.Flags().GetString("video")
	carouselURLs, _ := cmd.Flags().GetStringSlice("carousel")
	hasCarousel := cmd.Flags().Lookup("carousel").Changed

	var containerID string
	needsPoll := false

	switch {
	case imageURL != "":
		containerID, err = threads.CreateImageContainer(ctx, client, userID, imageURL, text)
		needsPoll = true
	case videoURL != "":
		containerID, err = threads.CreateVideoContainer(ctx, client, userID, videoURL, text)
		needsPoll = true
	case hasCarousel && len(carouselURLs) > 0:
		containerID, err = createCarouselPost(ctx, client, userID, carouselURLs, text)
		needsPoll = true
	default:
		containerID, err = threads.CreateContainer(ctx, client, userID, text, "TEXT")
	}
	if err != nil {
		return fmt.Errorf("creating container: %w", err)
	}

	if needsPoll {
		if err := threads.WaitForReady(ctx, client, containerID, 2*time.Second, 5*time.Minute); err != nil {
			return fmt.Errorf("waiting for media: %w", err)
		}
	}

	postID, err := threads.Publish(ctx, client, userID, containerID)
	if err != nil {
		return fmt.Errorf("publishing: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Published post %s\n", postID)
	return nil
}

func createCarouselPost(ctx context.Context, client *api.Client, userID string, imageURLs []string, text string) (string, error) {
	itemIDs := make([]string, 0, len(imageURLs))
	for _, u := range imageURLs {
		id, err := threads.CreateImageContainer(ctx, client, userID, u, "")
		if err != nil {
			return "", fmt.Errorf("creating carousel item: %w", err)
		}
		if err := threads.WaitForReady(ctx, client, id, 2*time.Second, 5*time.Minute); err != nil {
			return "", fmt.Errorf("waiting for carousel item: %w", err)
		}
		itemIDs = append(itemIDs, id)
	}
	return threads.CreateCarousel(ctx, client, userID, itemIDs, text)
}

func runPostGet(cmd *cobra.Command, args []string) error {
	postID := args[0]
	ctx := cmd.Context()

	client, err := loadPostClient(cmd)
	if err != nil {
		return err
	}

	post, err := threads.Get(ctx, client, postID, threads.DefaultPostFields())
	if err != nil {
		return fmt.Errorf("getting post: %w", err)
	}

	return output.Print(cmd.OutOrStdout(), post, getFormat(cmd))
}

func runPostList(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	client, err := loadPostClient(cmd)
	if err != nil {
		return err
	}

	userID, err := loadUserID()
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

	list, err := threads.List(ctx, client, userID, page)
	if err != nil {
		return fmt.Errorf("listing posts: %w", err)
	}

	return output.Print(cmd.OutOrStdout(), list.Data, getFormat(cmd))
}

func runPostDelete(cmd *cobra.Command, args []string) error {
	postID := args[0]
	ctx := cmd.Context()

	client, err := loadPostClient(cmd)
	if err != nil {
		return err
	}

	if err := threads.Delete(ctx, client, postID); err != nil {
		return fmt.Errorf("deleting post: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted post %s\n", postID)
	return nil
}

func runPostRepost(cmd *cobra.Command, args []string) error {
	postID := args[0]
	ctx := cmd.Context()

	client, err := loadPostClient(cmd)
	if err != nil {
		return err
	}

	repostID, err := threads.Repost(ctx, client, postID)
	if err != nil {
		return fmt.Errorf("reposting: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Reposted as %s\n", repostID)
	return nil
}

// loadPostClient creates an API client pointing at apiBaseURL.
func loadPostClient(cmd *cobra.Command) (*api.Client, error) {
	token, _ := cmd.Flags().GetString("token")
	if token == "" {
		// Fall through to the standard loadClient which checks env and config.
		client, err := loadClient(cmd)
		if err != nil {
			return nil, err
		}
		// If apiBaseURL has been overridden (tests), rebuild with that URL.
		if apiBaseURL != "https://graph.threads.net" {
			return api.NewClientWithHTTP(token, apiBaseURL, &http.Client{}), nil
		}
		return client, nil
	}
	if apiBaseURL != "https://graph.threads.net" {
		return api.NewClientWithHTTP(token, apiBaseURL, &http.Client{}), nil
	}
	return api.NewClient(token), nil
}

// loadUserID loads the user ID from stored credentials.
func loadUserID() (string, error) {
	store := config.NewStore(configDir)
	creds, err := store.Load()
	if err != nil {
		return "", fmt.Errorf("loading credentials: %w (run 'threads auth login' first)", err)
	}
	if creds.UserID == "" {
		return "", fmt.Errorf("no user ID in stored credentials: run 'threads auth login'")
	}
	return creds.UserID, nil
}
