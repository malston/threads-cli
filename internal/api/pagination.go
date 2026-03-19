package api

import (
	"net/url"
	"strconv"
)

// PageParams holds cursor-based pagination request parameters.
type PageParams struct {
	Limit  int
	Before string
	After  string
}

// PageInfo holds cursor-based pagination response metadata.
type PageInfo struct {
	Before   string
	After    string
	Next     string
	Previous string
}

// HasNext reports whether there are more results after the current page.
func (p *PageInfo) HasNext() bool { return p.Next != "" || p.After != "" }

// ApplyPaging adds pagination query parameters to vals based on p.
// Only non-zero/non-empty fields are added. A nil PageParams is a no-op.
func ApplyPaging(vals url.Values, p *PageParams) {
	if p == nil {
		return
	}
	if p.Limit > 0 {
		vals.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Before != "" {
		vals.Set("before", p.Before)
	}
	if p.After != "" {
		vals.Set("after", p.After)
	}
}

// ParsePaging extracts pagination info from a Threads API JSON response.
// It expects the top-level map to contain a "paging" key with cursors and links.
// Returns an empty PageInfo if the paging key is missing.
func ParsePaging(raw map[string]any) *PageInfo {
	info := &PageInfo{}

	paging, ok := raw["paging"].(map[string]any)
	if !ok {
		return info
	}

	if cursors, ok := paging["cursors"].(map[string]any); ok {
		if v, ok := cursors["before"].(string); ok {
			info.Before = v
		}
		if v, ok := cursors["after"].(string); ok {
			info.After = v
		}
	}

	if v, ok := paging["next"].(string); ok {
		info.Next = v
	}
	if v, ok := paging["previous"].(string); ok {
		info.Previous = v
	}

	return info
}
