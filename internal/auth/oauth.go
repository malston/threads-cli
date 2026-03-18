package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// BuildAuthURL constructs the Threads OAuth authorization URL with all
// parameters properly URL-encoded.
func BuildAuthURL(appID, redirectURI, state string, scopes []string) string {
	params := url.Values{}
	params.Set("client_id", appID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(scopes, ","))
	params.Set("state", state)

	return "https://threads.net/oauth/authorize?" + params.Encode()
}

// StartCallbackServer starts a local HTTP server on a random port to receive
// the OAuth callback. It returns the port, a channel that receives the
// authorization code, a channel for errors, and a shutdown function.
func StartCallbackServer(ctx context.Context) (port int, codeChan <-chan string, errChan <-chan error, shutdown func()) {
	code := make(chan string, 1)
	errs := make(chan error, 1)

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		errs <- fmt.Errorf("starting listener: %w", err)
		return 0, code, errs, func() {}
	}

	port = listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		authCode := r.URL.Query().Get("code")
		authCode = strings.TrimSuffix(authCode, "#_")

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<html><body><h1>Authorization successful! You can close this window.</h1></body></html>")

		code <- authCode
	})

	server := &http.Server{Handler: mux}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errs <- fmt.Errorf("serving: %w", err)
		}
	}()

	go func() {
		<-ctx.Done()
		server.Close()
	}()

	shutdown = func() {
		server.Close()
	}

	return port, code, errs, shutdown
}

// DefaultScopes returns the default OAuth scopes for the Threads API.
func DefaultScopes() []string {
	return []string{
		"threads_basic",
		"threads_content_publish",
		"threads_manage_insights",
		"threads_manage_replies",
		"threads_read_replies",
	}
}
