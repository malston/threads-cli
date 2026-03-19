package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/malston/threads-cli/internal/api"
)

// Post represents a Threads post.
type Post struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	MediaType string `json:"media_type"`
	Timestamp string `json:"timestamp"`
	Permalink string `json:"permalink"`
	Username  string `json:"username"`
}

// PostList holds a page of posts with pagination metadata.
type PostList struct {
	Data   []Post        `json:"data"`
	Paging *api.PageInfo `json:"-"`
}

// defaultFields lists the fields requested when fetching posts.
var defaultFields = []string{"id", "text", "media_type", "timestamp", "permalink", "username"}

// DefaultPostFields returns a copy of the default fields requested when fetching posts.
func DefaultPostFields() []string {
	out := make([]string, len(defaultFields))
	copy(out, defaultFields)
	return out
}

// CreateContainer creates a media container for a post.
// Returns the container ID on success.
func CreateContainer(ctx context.Context, client *api.Client, userID, text, mediaType string) (string, error) {
	body := url.Values{}
	body.Set("text", text)
	body.Set("media_type", mediaType)

	resp, err := client.Post(ctx, fmt.Sprintf("/%s/threads", userID), body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", api.ParseError(resp)
	}

	return parseID(resp)
}

// Publish publishes a previously created media container.
// Returns the published post ID on success.
func Publish(ctx context.Context, client *api.Client, userID, containerID string) (string, error) {
	body := url.Values{}
	body.Set("creation_id", containerID)

	resp, err := client.Post(ctx, fmt.Sprintf("/%s/threads_publish", userID), body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", api.ParseError(resp)
	}

	return parseID(resp)
}

// Get fetches a single post by ID with the specified fields.
func Get(ctx context.Context, client *api.Client, postID string, fields []string) (*Post, error) {
	params := url.Values{}
	if len(fields) > 0 {
		params.Set("fields", strings.Join(fields, ","))
	}

	resp, err := client.Get(ctx, fmt.Sprintf("/%s", postID), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, api.ParseError(resp)
	}

	var post Post
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return nil, fmt.Errorf("parsing post response: %w", err)
	}
	return &post, nil
}

// List fetches a page of posts for a user.
func List(ctx context.Context, client *api.Client, userID string, page *api.PageParams) (*PostList, error) {
	return fetchPostList(ctx, client, fmt.Sprintf("/%s/threads", userID), page)
}

// Delete removes a post by ID.
func Delete(ctx context.Context, client *api.Client, postID string) error {
	resp, err := client.Delete(ctx, fmt.Sprintf("/%s", postID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return api.ParseError(resp)
	}
	return nil
}

// Repost reposts an existing post.
// Returns the repost ID on success.
func Repost(ctx context.Context, client *api.Client, postID string) (string, error) {
	resp, err := client.Post(ctx, fmt.Sprintf("/%s/repost", postID), url.Values{})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", api.ParseError(resp)
	}

	return parseID(resp)
}

// parseID extracts the "id" field from a JSON response body.
func parseID(resp *http.Response) (string, error) {
	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("parsing id response: %w", err)
	}
	if result.ID == "" {
		return "", fmt.Errorf("response missing id field")
	}
	return result.ID, nil
}
