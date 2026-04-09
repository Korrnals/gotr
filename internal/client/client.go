package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Korrnals/gotr/internal/debug"
)

const apiPrefix = "index.php?/api/v2/"

// HTTPClient wraps HTTP transport and base URL handling for TestRail API calls.
type HTTPClient struct {
	client  *http.Client
	baseURL *url.URL
}

// options holds internal client configuration (unexported).
type options struct {
	insecure            bool
	timeout             time.Duration
	tlsHandshakeTimeout time.Duration
}

// authTransport automatically injects Basic Auth into every outgoing request.
type authTransport struct {
	username string
	apiKey   string
	base     http.RoundTripper
}

// RoundTrip injects authentication and required default headers into each request.
func (t authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.username, t.apiKey)
	// Set Content-Type only if not already set
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	// User-Agent is required by some TestRail installations —
	// without a browser-like header the server may return 403/401.
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; gotr/2.7; +https://github.com/Korrnals/gotr)")
	}
	return t.base.RoundTrip(req)
}

// defaultOptions holds the default client configuration values.
var defaultOptions = options{
	insecure:            false,
	timeout:             30 * time.Second,
	tlsHandshakeTimeout: 10 * time.Second,
}

// ClientOption is a functional option for configuring NewClient.
type ClientOption func(*options)

// WithSkipTlsVerify enables or disables TLS certificate verification.
func WithSkipTlsVerify(insecure bool) ClientOption {
	return func(o *options) {
		o.insecure = insecure
	}
}

// WithTimeout sets the HTTP client request timeout.
func WithTimeout(duration time.Duration) ClientOption {
	return func(o *options) {
		o.timeout = duration
	}
}

// NewClient creates a new HTTP client for TestRail API calls with the given options.
func NewClient(baseURLStr, username, apiKey string, debugMode bool, opts ...ClientOption) (*HTTPClient, error) {
	// Parse URL; we rebuild with scheme+host only
	parsed, err := url.Parse(strings.TrimSpace(baseURLStr))
	if err != nil || parsed.Host == "" {
		return nil, fmt.Errorf("invalid or empty base URL: %s", baseURLStr)
	}

	// Build a clean URL with scheme and host only
	cleanURL := &url.URL{
		Scheme: parsed.Scheme,
		Host:   parsed.Host, // includes port if present
	}

	if debugMode {
		debug.DebugPrint("{client} - Original baseURL: %s", baseURLStr)
		debug.DebugPrint("{client} - Normalized baseURL: %s", cleanURL.String())
	}
	// Apply default options, then override with provided ones
	cfg := defaultOptions
	for _, o := range opts {
		o(&cfg)
	}

	if cfg.insecure {
		fmt.Fprintln(os.Stderr, "WARNING: TLS certificate verification is disabled (--insecure). Connection is vulnerable to MITM attacks.")
	}

	// Configure HTTP transport.
	// MaxConnsPerHost MUST match actual concurrency:
	// 2 projects × 8 suites × 10 pages = 160 concurrent requests.
	// With MaxConnsPerHost=50, 110 requests queue inside Go transport;
	// http.Client.Timeout includes queue wait, causing cascading timeouts
	// and exponential-backoff retries → 3× slower than expected.
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.insecure,
		},
		TLSHandshakeTimeout: cfg.tlsHandshakeTimeout,
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 200,
		MaxConnsPerHost:     0, // unlimited — concurrency governed by parallel settings
		IdleConnTimeout:     90 * time.Second,
	}
	// Wrap transport with Basic Auth injector
	auth := authTransport{
		username: username,
		apiKey:   apiKey,
		base:     transport,
	}

	return &HTTPClient{
		client: &http.Client{
			Transport: auth,
			Timeout:   cfg.timeout,
		},
		baseURL: cleanURL,
	}, nil
}

// DoRequest is the universal method for making HTTP requests to TestRail.
// It builds the URL manually to accommodate TestRail's non-standard query format.
func (c *HTTPClient) DoRequest(ctx context.Context, method, endpoint string, body io.Reader, queryParams map[string]string) (*http.Response, error) {
	// Strip leading slash from endpoint
	cleanEndpoint := strings.TrimPrefix(endpoint, "/")
	debug.DebugPrint("{DoRequest} - cleanEndpoint: %s", cleanEndpoint)

	// Build path manually — TestRail requires literal '?' in the path
	path := apiPrefix + cleanEndpoint
	debug.DebugPrint("{DoRequest} - Path: %s", path)
	// Base URL as string (trim trailing slash)
	base := strings.TrimSuffix(c.baseURL.String(), "/")
	debug.DebugPrint("{DoRequest} - Base URL: %s", base)
	// Full URL as string
	fullURL := base + "/" + path

	// Append query params with '&' (TestRail uses '?' inside the path prefix)
	if len(queryParams) > 0 {
		q := url.Values{}
		for k, v := range queryParams {
			q.Add(k, v)
		}
		fullURL += "&" + q.Encode() // '&' instead of '?'
	}

	debug.DebugPrint("{DoRequest} - Constructed URL: %s", fullURL)
	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}

	// Check for Content-Type override in queryParams (e.g. multipart/form-data)
	contentType := "application/json"
	if ct, ok := queryParams["Content-Type"]; ok {
		contentType = ct
		// Remove Content-Type from params so it's not appended to URL
		delete(queryParams, "Content-Type")
	}

	// Set the Content-Type header
	req.Header.Set("Content-Type", contentType)
	// Execute the request
	return c.client.Do(req)
}

// Get performs a GET request with automatic non-200 error handling.
func (c *HTTPClient) Get(ctx context.Context, endpoint string, queryParams map[string]string) (*http.Response, error) {
	resp, err := c.DoRequest(ctx, "GET", endpoint, nil, queryParams)
	if err != nil {
		return nil, err
	}

	// Non-200 status — return a formatted API error
	if resp.StatusCode != http.StatusOK {
		return nil, c.formatAPIError(resp)
	}

	return resp, nil
}

// Post performs a POST request with automatic non-200 error handling.
func (c *HTTPClient) Post(ctx context.Context, endpoint string, body io.Reader, queryParams map[string]string) (*http.Response, error) {
	resp, err := c.DoRequest(ctx, "POST", endpoint, body, queryParams)
	if err != nil {
		return nil, err
	}

	// Non-200 status — return a formatted API error
	if resp.StatusCode != http.StatusOK {
		return nil, c.formatAPIError(resp)
	}

	return resp, nil
}

// formatAPIError formats a non-200 API response into a descriptive error.
func (c *HTTPClient) formatAPIError(resp *http.Response) error {
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("API returned %s, failed to read error body: %w", resp.Status, err)
	}

	// Try parsing as JSON with an "error" field
	var errStruct struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(bodyBytes, &errStruct) == nil && errStruct.Error != "" {
		return fmt.Errorf("API returned %s: %s", resp.Status, errStruct.Error)
	}

	// Fallback: return raw body as error text
	return fmt.Errorf("API returned %s: %s", resp.Status, string(bodyBytes))
}
