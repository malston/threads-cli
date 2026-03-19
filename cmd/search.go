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
	searchCmd.Flags().String("author", "", "filter by author username")
	searchCmd.Flags().String("sort", "", "sort order: top or recent")
	searchCmd.Flags().String("mode", "", "search mode: keyword or tag")
	searchCmd.Flags().String("media-type", "", "filter by media type: text, image, or video")
	searchCmd.Flags().Int64("since", 0, "results after this unix timestamp")
	searchCmd.Flags().Int64("until", 0, "results before this unix timestamp")
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

	author, _ := cmd.Flags().GetString("author")
	sort, _ := cmd.Flags().GetString("sort")
	mode, _ := cmd.Flags().GetString("mode")
	mediaType, _ := cmd.Flags().GetString("media-type")
	since, _ := cmd.Flags().GetInt64("since")
	until, _ := cmd.Flags().GetInt64("until")

	if err := validateSearchFlags(sort, mode, mediaType, since, until); err != nil {
		return err
	}

	var filters *threads.SearchFilters
	if author != "" || sort != "" || mode != "" || mediaType != "" || since != 0 || until != 0 {
		filters = &threads.SearchFilters{
			Author:    author,
			SortBy:    sort,
			Mode:      mode,
			MediaType: mediaType,
			Since:     since,
			Until:     until,
		}
	}

	list, err := threads.Search(ctx, client, query, page, filters)
	if err != nil {
		return fmt.Errorf("searching posts: %w", err)
	}

	format, err := getFormat(cmd)
	if err != nil {
		return err
	}
	return output.Print(cmd.OutOrStdout(), list.Data, format)
}

func validateSearchFlags(sort, mode, mediaType string, since, until int64) error {
	if sort != "" && sort != "top" && sort != "recent" {
		return fmt.Errorf("invalid --sort %q: must be top or recent", sort)
	}
	if mode != "" && mode != "keyword" && mode != "tag" {
		return fmt.Errorf("invalid --mode %q: must be keyword or tag", mode)
	}
	if mediaType != "" && mediaType != "text" && mediaType != "image" && mediaType != "video" {
		return fmt.Errorf("invalid --media-type %q: must be text, image, or video", mediaType)
	}
	if since != 0 && until != 0 && since > until {
		return fmt.Errorf("--since (%d) must not be after --until (%d)", since, until)
	}
	return nil
}

// loadSearchClient creates an API client pointing at searchBaseURL.
func loadSearchClient(cmd *cobra.Command) (*api.Client, error) {
	token, err := resolveToken(cmd)
	if err != nil {
		return nil, err
	}
	if searchBaseURL != "https://graph.threads.net" {
		return api.NewClientWithHTTP(token, searchBaseURL, &http.Client{}), nil
	}
	return api.NewClient(token), nil
}
