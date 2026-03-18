package api

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAPIErrorFormat(t *testing.T) {
	e := &APIError{
		Message: "Invalid access token",
		Type:    "OAuthException",
		Code:    190,
	}

	want := "threads api: Invalid access token (type=OAuthException, code=190)"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestParseErrorValidJSON(t *testing.T) {
	body := `{"error":{"message":"Invalid access token","type":"OAuthException","code":190,"fbtrace_id":"abc123"}}`
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	err := ParseError(resp)
	if err == nil {
		t.Fatal("ParseError returned nil, want error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Message != "Invalid access token" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Invalid access token")
	}
	if apiErr.Type != "OAuthException" {
		t.Errorf("Type = %q, want %q", apiErr.Type, "OAuthException")
	}
	if apiErr.Code != 190 {
		t.Errorf("Code = %d, want %d", apiErr.Code, 190)
	}
	if apiErr.FBTraceID != "abc123" {
		t.Errorf("FBTraceID = %q, want %q", apiErr.FBTraceID, "abc123")
	}
	if apiErr.HTTPStatus != http.StatusBadRequest {
		t.Errorf("HTTPStatus = %d, want %d", apiErr.HTTPStatus, http.StatusBadRequest)
	}
}

func TestParseErrorMissingFields(t *testing.T) {
	body := `{"error":{"message":"Something went wrong","code":2}}`
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	err := ParseError(resp)
	if err == nil {
		t.Fatal("ParseError returned nil, want error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Message != "Something went wrong" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Something went wrong")
	}
	if apiErr.Type != "" {
		t.Errorf("Type = %q, want empty", apiErr.Type)
	}
	if apiErr.Code != 2 {
		t.Errorf("Code = %d, want %d", apiErr.Code, 2)
	}
	if apiErr.FBTraceID != "" {
		t.Errorf("FBTraceID = %q, want empty", apiErr.FBTraceID)
	}
	if apiErr.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("HTTPStatus = %d, want %d", apiErr.HTTPStatus, http.StatusInternalServerError)
	}
}

func TestParseErrorNonJSON(t *testing.T) {
	body := `<html>Service Unavailable</html>`
	resp := &http.Response{
		StatusCode: http.StatusServiceUnavailable,
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	err := ParseError(resp)
	if err == nil {
		t.Fatal("ParseError returned nil, want error")
	}

	if _, ok := err.(*APIError); ok {
		t.Error("expected generic error for non-JSON body, got *APIError")
	}

	wantSubstr := "503"
	if got := err.Error(); !strings.Contains(got, wantSubstr) {
		t.Errorf("error = %q, want it to contain %q", got, wantSubstr)
	}
}

func TestParseErrorEmptyBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusBadGateway,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	err := ParseError(resp)
	if err == nil {
		t.Fatal("ParseError returned nil, want error")
	}

	if _, ok := err.(*APIError); ok {
		t.Error("expected generic error for empty body, got *APIError")
	}

	wantSubstr := "502"
	if got := err.Error(); !strings.Contains(got, wantSubstr) {
		t.Errorf("error = %q, want it to contain %q", got, wantSubstr)
	}
}

func TestIsRateLimited(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "code 4",
			err:  &APIError{Code: 4},
			want: true,
		},
		{
			name: "code 32",
			err:  &APIError{Code: 32},
			want: true,
		},
		{
			name: "HTTP 429",
			err:  &APIError{HTTPStatus: 429},
			want: true,
		},
		{
			name: "not rate limited",
			err:  &APIError{Code: 190, HTTPStatus: 400},
			want: false,
		},
		{
			name: "non-APIError",
			err:  fmt.Errorf("something else"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRateLimited(tt.err); got != tt.want {
				t.Errorf("IsRateLimited() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "OAuthException",
			err:  &APIError{Type: "OAuthException"},
			want: true,
		},
		{
			name: "different type",
			err:  &APIError{Type: "GraphMethodException"},
			want: false,
		},
		{
			name: "non-APIError",
			err:  fmt.Errorf("auth failed"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAuthError(tt.err); got != tt.want {
				t.Errorf("IsAuthError() = %v, want %v", got, tt.want)
			}
		})
	}
}
