package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/malston/threads-cli/internal/api"
)

// CreateImageContainer creates a media container for an image post.
// Returns the container ID on success.
func CreateImageContainer(ctx context.Context, client *api.Client, userID, imageURL, text string) (string, error) {
	body := url.Values{}
	body.Set("media_type", "IMAGE")
	body.Set("image_url", imageURL)
	body.Set("text", text)

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

// CreateVideoContainer creates a media container for a video post.
// Returns the container ID on success.
func CreateVideoContainer(ctx context.Context, client *api.Client, userID, videoURL, text string) (string, error) {
	body := url.Values{}
	body.Set("media_type", "VIDEO")
	body.Set("video_url", videoURL)
	body.Set("text", text)

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

// CreateCarousel creates a media container for a carousel post.
// itemIDs are the container IDs of the carousel items.
// Returns the container ID on success.
func CreateCarousel(ctx context.Context, client *api.Client, userID string, itemIDs []string, text string) (string, error) {
	body := url.Values{}
	body.Set("media_type", "CAROUSEL")
	body.Set("children", strings.Join(itemIDs, ","))
	body.Set("text", text)

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

// ContainerStatus holds the status and optional error message from a container check.
type ContainerStatus struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message"`
}

// CheckStatus retrieves the current status of a media container.
func CheckStatus(ctx context.Context, client *api.Client, containerID string) (*ContainerStatus, error) {
	params := url.Values{}
	params.Set("fields", "status,error_message")

	resp, err := client.Get(ctx, fmt.Sprintf("/%s", containerID), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, api.ParseError(resp)
	}

	var result ContainerStatus
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parsing status response: %w", err)
	}

	return &result, nil
}

// WaitForReady polls CheckStatus until the container reaches FINISHED status.
// Returns an error if the container reaches ERROR status or the timeout elapses.
func WaitForReady(ctx context.Context, client *api.Client, containerID string, pollInterval, timeout time.Duration) error {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		cs, err := CheckStatus(ctx, client, containerID)
		if err != nil {
			return err
		}

		switch cs.Status {
		case "FINISHED":
			return nil
		case "ERROR":
			if cs.ErrorMessage != "" {
				return fmt.Errorf("container error: %s", cs.ErrorMessage)
			}
			return fmt.Errorf("container error: unknown")
		}

		select {
		case <-deadline.C:
			return fmt.Errorf("timed out waiting for container %s", containerID)
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
