package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/malston/threads-cli/internal/api"
)

// SearchFilters holds optional filter parameters for keyword search.
type SearchFilters struct {
	Author    string // Filter by author username
	SortBy    string // "top" or "recent" (maps to search_type)
	MediaType string // "text", "image", or "video"
	Since     int64  // Unix timestamp for start of date range
	Until     int64  // Unix timestamp for end of date range
}

// Search finds posts matching a keyword query with optional filters.
func Search(ctx context.Context, client *api.Client, query string, page *api.PageParams, filters *SearchFilters) (*PostList, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("fields", strings.Join(defaultFields, ","))
	api.ApplyPaging(params, page)
	applySearchFilters(params, filters)

	resp, err := client.Get(ctx, "/keyword_search", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, api.ParseError(resp)
	}

	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("parsing search response: %w", err)
	}

	list := &PostList{
		Paging: api.ParsePaging(raw),
	}

	dataBytes, err := json.Marshal(raw["data"])
	if err != nil {
		return nil, fmt.Errorf("marshaling search data: %w", err)
	}
	if err := json.Unmarshal(dataBytes, &list.Data); err != nil {
		return nil, fmt.Errorf("parsing search data: %w", err)
	}

	return list, nil
}

func applySearchFilters(params url.Values, f *SearchFilters) {
	if f == nil {
		return
	}
	if f.Author != "" {
		params.Set("author_username", f.Author)
	}
	if f.SortBy != "" {
		params.Set("search_type", strings.ToUpper(f.SortBy))
	}
	if f.MediaType != "" {
		params.Set("media_type", strings.ToUpper(f.MediaType))
	}
	if f.Since != 0 {
		params.Set("since", strconv.FormatInt(f.Since, 10))
	}
	if f.Until != 0 {
		params.Set("until", strconv.FormatInt(f.Until, 10))
	}
}
