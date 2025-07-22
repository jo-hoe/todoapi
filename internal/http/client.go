// Package http provides HTTP client utilities with enhanced functionality
package http

import (
	"context"
	"net/http"
	"time"
)

// AddHeaderTransport is a RoundTripper that adds default headers to requests
type AddHeaderTransport struct {
	Transport      http.RoundTripper
	DefaultHeaders map[string]string
}

// RoundTrip implements the http.RoundTripper interface
func (t *AddHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	reqClone := req.Clone(req.Context())

	// Add default headers if they don't already exist
	for name, value := range t.DefaultHeaders {
		if reqClone.Header.Get(name) == "" {
			reqClone.Header.Set(name, value)
		}
	}

	return t.Transport.RoundTrip(reqClone)
}

// NewAddHeaderTransport creates a new AddHeaderTransport
func NewAddHeaderTransport(transport http.RoundTripper, headers map[string]string) *AddHeaderTransport {
	if transport == nil {
		transport = http.DefaultTransport
	}

	return &AddHeaderTransport{
		Transport:      transport,
		DefaultHeaders: headers,
	}
}

// ClientConfig holds configuration for HTTP clients
type ClientConfig struct {
	Timeout         time.Duration
	MaxIdleConns    int
	IdleConnTimeout time.Duration
}

// DefaultClientConfig returns a default client configuration
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		Timeout:         30 * time.Second,
		MaxIdleConns:    100,
		IdleConnTimeout: 90 * time.Second,
	}
}

// NewHTTPClientWithHeader creates an HTTP client with a single default header
func NewHTTPClientWithHeader(headerName, headerValue string) *http.Client {
	headers := map[string]string{
		headerName: headerValue,
	}
	return NewHTTPClientWithHeaders(headers)
}

// NewHTTPClientWithHeaders creates an HTTP client with multiple default headers
func NewHTTPClientWithHeaders(headers map[string]string) *http.Client {
	config := DefaultClientConfig()
	return NewHTTPClientWithConfig(headers, config)
}

// NewHTTPClientWithConfig creates an HTTP client with custom configuration
func NewHTTPClientWithConfig(headers map[string]string, config *ClientConfig) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:    config.MaxIdleConns,
		IdleConnTimeout: config.IdleConnTimeout,
	}

	return &http.Client{
		Transport: NewAddHeaderTransport(transport, headers),
		Timeout:   config.Timeout,
	}
}

// DoWithContext performs an HTTP request with context
func DoWithContext(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	return client.Do(req)
}
