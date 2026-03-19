package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/malston/saved-threads/internal/api"
)

// Profile represents a Threads user profile.
type Profile struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Name          string `json:"name"`
	Bio           string `json:"threads_biography"`
	ProfilePicURL string `json:"threads_profile_picture_url"`
	IsVerified    bool   `json:"is_verified"`
	FollowerCount int    `json:"follower_count,omitempty"`
}

// profileFields lists the fields requested when fetching a profile by ID.
var profileFields = []string{"id", "username", "name", "threads_biography", "threads_profile_picture_url", "is_verified"}

// lookupFields lists the fields requested when looking up a profile by username.
var lookupFields = []string{"username", "name", "biography", "profile_picture_url", "follower_count", "is_verified"}

// GetProfile fetches a user profile by ID.
func GetProfile(ctx context.Context, client *api.Client, userID string) (*Profile, error) {
	params := url.Values{}
	params.Set("fields", strings.Join(profileFields, ","))

	resp, err := client.Get(ctx, fmt.Sprintf("/%s", userID), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, api.ParseError(resp)
	}

	var profile Profile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("parsing profile response: %w", err)
	}
	return &profile, nil
}

// LookupProfile finds a user profile by username.
func LookupProfile(ctx context.Context, client *api.Client, username string) (*Profile, error) {
	params := url.Values{}
	params.Set("username", username)
	params.Set("fields", strings.Join(lookupFields, ","))

	resp, err := client.Get(ctx, "/profile_lookup", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, api.ParseError(resp)
	}

	var profile Profile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("parsing profile lookup response: %w", err)
	}
	return &profile, nil
}
