package api

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

// Client handles HTTP communication with the Threads API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewClient creates a Client configured for the Threads API.
func NewClient(token string) *Client {
	return &Client{
		baseURL:    "https://graph.threads.net",
		httpClient: &http.Client{},
		token:      token,
	}
}

// NewClientWithHTTP creates a Client with a custom base URL and http.Client.
func NewClientWithHTTP(token string, baseURL string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		token:      token,
	}
}

// Get sends a GET request with optional query parameters.
func (c *Client) Get(ctx context.Context, path string, params url.Values) (*http.Response, error) {
	fullURL := c.baseURL + path
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	return c.httpClient.Do(req)
}

// Post sends a POST request with a form-encoded body.
func (c *Client) Post(ctx context.Context, path string, body url.Values) (*http.Response, error) {
	fullURL := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.httpClient.Do(req)
}

// Delete sends a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) (*http.Response, error) {
	fullURL := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	return c.httpClient.Do(req)
}
