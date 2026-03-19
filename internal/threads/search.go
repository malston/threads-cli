package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/malston/threads-cli/internal/api"
)

// Search finds posts matching a keyword query.
func Search(ctx context.Context, client *api.Client, query string, page *api.PageParams) (*PostList, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("fields", strings.Join(defaultFields, ","))
	api.ApplyPaging(params, page)

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
