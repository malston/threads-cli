package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/malston/threads-cli/internal/api"
)

// QuotaUsage represents a single rate limit quota entry.
type QuotaUsage struct {
	QuotaUsage    int    `json:"quota_usage"`
	QuotaTotal    int    `json:"quota_total"`
	QuotaDuration int    `json:"quota_duration"`
	RateLimitType string `json:"rate_limit_type"`
}

// RateLimitStatus holds the list of quota entries for a user's publishing limits.
type RateLimitStatus struct {
	Data []QuotaUsage `json:"data"`
}

// GetLimits fetches the current publishing rate limit status for a user.
func GetLimits(ctx context.Context, client *api.Client, userID string) (*RateLimitStatus, error) {
	params := url.Values{}
	params.Set("fields", "quota_usage,quota_total,quota_duration,rate_limit_type")

	resp, err := client.Get(ctx, fmt.Sprintf("/%s/threads_publishing_limit", userID), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, api.ParseError(resp)
	}

	var status RateLimitStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("parsing rate limit response: %w", err)
	}
	return &status, nil
}
