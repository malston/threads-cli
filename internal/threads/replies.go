package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/malston/saved-threads/internal/api"
)

// ListReplies fetches a page of replies for a given post.
func ListReplies(ctx context.Context, client *api.Client, postID string, page *api.PageParams) (*PostList, error) {
	return fetchPostList(ctx, client, fmt.Sprintf("/%s/replies", postID), page)
}

// GetConversation fetches the threaded conversation rooted at a post.
func GetConversation(ctx context.Context, client *api.Client, postID string, page *api.PageParams) (*PostList, error) {
	return fetchPostList(ctx, client, fmt.Sprintf("/%s/conversation", postID), page)
}

// CreateReply creates a text reply to a post. It creates a media container
// with the reply_to_id set, then publishes it. Returns the published post ID.
func CreateReply(ctx context.Context, client *api.Client, userID, replyToID, text string) (string, error) {
	body := url.Values{}
	body.Set("media_type", "TEXT")
	body.Set("text", text)
	body.Set("reply_to_id", replyToID)

	resp, err := client.Post(ctx, fmt.Sprintf("/%s/threads", userID), body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", api.ParseError(resp)
	}

	containerID, err := parseID(resp)
	if err != nil {
		return "", err
	}

	return Publish(ctx, client, userID, containerID)
}

// HideReply hides a reply from the conversation.
func HideReply(ctx context.Context, client *api.Client, replyID string) error {
	return manageReply(ctx, client, replyID, true)
}

// UnhideReply unhides a previously hidden reply.
func UnhideReply(ctx context.Context, client *api.Client, replyID string) error {
	return manageReply(ctx, client, replyID, false)
}

func manageReply(ctx context.Context, client *api.Client, replyID string, hide bool) error {
	body := url.Values{}
	if hide {
		body.Set("hide", "true")
	} else {
		body.Set("hide", "false")
	}

	resp, err := client.Post(ctx, fmt.Sprintf("/%s/manage_reply", replyID), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return api.ParseError(resp)
	}
	return nil
}

// fetchPostList is a shared helper for endpoints that return paginated post lists.
func fetchPostList(ctx context.Context, client *api.Client, path string, page *api.PageParams) (*PostList, error) {
	params := url.Values{}
	params.Set("fields", strings.Join(defaultFields, ","))
	api.ApplyPaging(params, page)

	resp, err := client.Get(ctx, path, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, api.ParseError(resp)
	}

	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("parsing list response: %w", err)
	}

	list := &PostList{
		Paging: api.ParsePaging(raw),
	}

	dataBytes, err := json.Marshal(raw["data"])
	if err != nil {
		return nil, fmt.Errorf("marshaling post data: %w", err)
	}
	if err := json.Unmarshal(dataBytes, &list.Data); err != nil {
		return nil, fmt.Errorf("parsing post data: %w", err)
	}

	return list, nil
}
