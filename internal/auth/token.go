package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/malston/threads-cli/internal/config"
)

// exchangeCodeResponse represents the JSON returned by the short-lived token endpoint.
type exchangeCodeResponse struct {
	AccessToken string `json:"access_token"`
	UserID      int64  `json:"user_id"`
}

// longLivedResponse represents the JSON returned by the long-lived and refresh token endpoints.
type longLivedResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// ExchangeCode exchanges an authorization code for a short-lived access token.
func ExchangeCode(ctx context.Context, httpClient *http.Client, baseURL, appID, appSecret, redirectURI, code string) (*config.Credentials, error) {
	form := url.Values{
		"client_id":     {appID},
		"client_secret": {appSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/oauth/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result exchangeCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &config.Credentials{
		AccessToken: result.AccessToken,
		UserID:      strconv.FormatInt(result.UserID, 10),
		TokenType:   "bearer",
	}, nil
}

// ExchangeLongLived exchanges a short-lived token for a long-lived token.
func ExchangeLongLived(ctx context.Context, httpClient *http.Client, baseURL, appSecret, shortToken string) (*config.Credentials, error) {
	params := url.Values{
		"grant_type":    {"th_exchange_token"},
		"client_secret": {appSecret},
		"access_token":  {shortToken},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/access_token?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	return doLongLivedRequest(httpClient, req)
}

// Refresh exchanges an existing long-lived token for a new one with a fresh expiry.
func Refresh(ctx context.Context, httpClient *http.Client, baseURL, token string) (*config.Credentials, error) {
	params := url.Values{
		"grant_type":   {"th_refresh_token"},
		"access_token": {token},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/refresh_access_token?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	return doLongLivedRequest(httpClient, req)
}

func doLongLivedRequest(httpClient *http.Client, req *http.Request) (*config.Credentials, error) {
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token request failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result longLivedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &config.Credentials{
		AccessToken: result.AccessToken,
		TokenType:   result.TokenType,
		ExpiresAt:   time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
	}, nil
}

// IsExpired reports whether the credentials have expired.
func IsExpired(creds *config.Credentials) bool {
	return creds.ExpiresAt.Before(time.Now())
}

// NeedsRefresh reports whether the credentials expire within 7 days.
func NeedsRefresh(creds *config.Credentials) bool {
	return creds.ExpiresAt.Before(time.Now().Add(7 * 24 * time.Hour))
}
