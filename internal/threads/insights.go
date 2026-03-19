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

// MetricValue holds a single data point for a metric.
type MetricValue struct {
	Value any `json:"value"`
}

// Metric represents one insight metric returned by the Threads API.
type Metric struct {
	Name        string        `json:"name"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Period      string        `json:"period"`
	Values      []MetricValue `json:"values"`
}

// InsightData holds the list of metrics returned by an insights endpoint.
type InsightData struct {
	Data []Metric `json:"data"`
}

// PostInsights fetches insight metrics for a single post.
func PostInsights(ctx context.Context, client *api.Client, postID string, metrics []string) (*InsightData, error) {
	params := url.Values{}
	params.Set("metric", strings.Join(metrics, ","))

	resp, err := client.Get(ctx, fmt.Sprintf("/%s/insights", postID), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, api.ParseError(resp)
	}

	var data InsightData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("parsing insights response: %w", err)
	}
	return &data, nil
}

// UserInsights fetches insight metrics for a user's threads.
// When since or until are zero, they are omitted from the request.
func UserInsights(ctx context.Context, client *api.Client, userID string, metrics []string, since, until int64) (*InsightData, error) {
	params := url.Values{}
	params.Set("metric", strings.Join(metrics, ","))
	if since != 0 {
		params.Set("since", strconv.FormatInt(since, 10))
	}
	if until != 0 {
		params.Set("until", strconv.FormatInt(until, 10))
	}

	resp, err := client.Get(ctx, fmt.Sprintf("/%s/threads_insights", userID), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, api.ParseError(resp)
	}

	var data InsightData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("parsing user insights response: %w", err)
	}
	return &data, nil
}
