package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// APIError represents an error response from the Threads API.
type APIError struct {
	Message    string `json:"message"`
	Type       string `json:"type"`
	Code       int    `json:"code"`
	FBTraceID  string `json:"fbtrace_id"`
	HTTPStatus int    `json:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("threads api: %s (type=%s, code=%d)", e.Message, e.Type, e.Code)
}

// ParseError reads the response body and returns an *APIError if the body
// contains the Threads API error JSON format. Otherwise it returns a generic
// error that includes the HTTP status code.
func ParseError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("threads api: failed to read error response (HTTP %d)", resp.StatusCode)
	}

	var envelope struct {
		Error APIError `json:"error"`
	}
	if len(body) > 0 && json.Unmarshal(body, &envelope) == nil && envelope.Error.Message != "" {
		envelope.Error.HTTPStatus = resp.StatusCode
		return &envelope.Error
	}

	return fmt.Errorf("threads api: unexpected error (HTTP %d)", resp.StatusCode)
}

// IsRateLimited reports whether the error indicates a Threads API rate limit.
func IsRateLimited(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.Code == 4 || apiErr.Code == 32 || apiErr.HTTPStatus == 429
}

// IsAuthError reports whether the error indicates an authentication failure.
func IsAuthError(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.Type == "OAuthException"
}
