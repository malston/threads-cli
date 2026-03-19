package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
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

// StartCallbackServer starts a local HTTPS server on a random port to receive
// the OAuth callback. Threads requires https redirect URIs. A self-signed
// certificate is generated for localhost. It returns the port, a channel that
// receives the authorization code, a channel for errors, and a shutdown function.
func StartCallbackServer(ctx context.Context) (port int, codeChan <-chan string, errChan <-chan error, shutdown func()) {
	code := make(chan string, 1)
	errs := make(chan error, 1)

	tlsCert, err := generateSelfSignedCert()
	if err != nil {
		errs <- fmt.Errorf("generating TLS certificate: %w", err)
		return 0, code, errs, func() {}
	}

	listener, err := tls.Listen("tcp", ":0", &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	})
	if err != nil {
		errs <- fmt.Errorf("starting TLS listener: %w", err)
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

// generateSelfSignedCert creates a self-signed TLS certificate for localhost.
func generateSelfSignedCert() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:     []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return tls.X509KeyPair(certPEM, keyPEM)
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
