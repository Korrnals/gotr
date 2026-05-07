package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
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
	client      *http.Client
	baseURL     *url.URL
	retryPolicy RetryPolicy
}

// options holds internal client configuration (unexported).
type options struct {
	insecure            bool
	timeout             time.Duration
	tlsHandshakeTimeout time.Duration
	caBundlePath        string
	retryPolicy         RetryPolicy
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
//
// timeout: 180s — heavy legacy endpoints (get_results_for_run on huge runs)
// have a fixed server-side overhead of ~30–60s independent of page size, so
// reducing the limit does not help. p99 measured at ~56s with limit=250;
// 180s leaves a 3× headroom for retry bursts under concurrent load.
var defaultOptions = options{
	insecure:            false,
	timeout:             180 * time.Second,
	tlsHandshakeTimeout: 10 * time.Second,
	retryPolicy:         DefaultRetryPolicy(),
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

// WithRetryPolicy overrides the automatic retry policy for transient GET
// failures (network timeouts, HTTP 5xx, 429). Pass RetryPolicy{MaxAttempts: 1}
// to disable retries entirely.
func WithRetryPolicy(p RetryPolicy) ClientOption {
	return func(o *options) {
		o.retryPolicy = p
	}
}

// WithCABundle loads a PEM-encoded CA bundle from the given path and
// installs it as the TLS RootCAs pool. This is the preferred alternative to
// WithSkipTlsVerify for corporate environments that use private CAs.
// An empty path is a no-op.
func WithCABundle(path string) ClientOption {
	return func(o *options) {
		o.caBundlePath = path
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

	// NOTE: The "TLS verification disabled" banner is emitted by the caller
	// (cmd/root.go) via internal/warnings so it honors ui.suppress_warnings
	// and the --show-warnings CLI override.
	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.insecure, //nolint:gosec // toggled only via explicit --insecure flag
	}
	if cfg.caBundlePath != "" {
		pool, err := loadCAPool(cfg.caBundlePath)
		if err != nil {
			return nil, fmt.Errorf("load CA bundle %q: %w", cfg.caBundlePath, err)
		}
		tlsCfg.RootCAs = pool
	}

	// Configure HTTP transport.
	// MaxConnsPerHost MUST match actual concurrency:
	// 2 projects × 8 suites × 10 pages = 160 concurrent requests.
	// With MaxConnsPerHost=50, 110 requests queue inside Go transport;
	// http.Client.Timeout includes queue wait, causing cascading timeouts
	// and exponential-backoff retries → 3× slower than expected.
	transport := &http.Transport{
		TLSClientConfig:     tlsCfg,
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
		baseURL:     cleanURL,
		retryPolicy: cfg.retryPolicy,
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

	// Extract Content-Type override BEFORE building the query string —
	// otherwise multipart upload headers leak into TestRail's GET
	// parameters and the server rejects the request with
	// "Invalid characters in GET: [Content-Type] [multipart/form-data; ...]".
	contentType := "application/json"
	if ct, ok := queryParams["Content-Type"]; ok {
		contentType = ct
		delete(queryParams, "Content-Type")
	}

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

	// Set the Content-Type header
	req.Header.Set("Content-Type", contentType)
	// Execute the request
	return c.client.Do(req)
}

// Get performs a GET request with automatic non-200 error handling and
// transparent retry on transient failures (network timeouts, HTTP 5xx, 429).
//
// Retries respect the parent context: cancellation aborts the loop
// immediately. Non-retryable errors (4xx other than 429, caller cancellation)
// short-circuit on the first attempt.
func (c *HTTPClient) Get(ctx context.Context, endpoint string, queryParams map[string]string) (*http.Response, error) {
	p := c.retryPolicy
	if p.MaxAttempts < 1 {
		p.MaxAttempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < p.MaxAttempts; attempt++ {
		resp, err := c.DoRequest(ctx, "GET", endpoint, nil, queryParams)

		// Success path.
		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		isLast := attempt == p.MaxAttempts-1
		switch {
		case err != nil:
			if !isRetryableErr(err) || isLast {
				return nil, err
			}
			lastErr = err
		case isRetryableStatus(resp.StatusCode):
			// Drain & close body so the connection can be reused.
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if isLast {
				return nil, fmt.Errorf("API returned %s after %d attempts", resp.Status, p.MaxAttempts)
			}
			lastErr = fmt.Errorf("transient %s", resp.Status)
		default:
			// Non-retryable HTTP status — hand off to the existing formatter
			// which reads the body for the diagnostic message.
			return nil, c.formatAPIError(resp)
		}

		delay := nextBackoff(attempt, p)
		logRetry(endpoint, attempt, p.MaxAttempts, delay, lastErr.Error())
		if sleepErr := sleepWithCtx(ctx, delay); sleepErr != nil {
			return nil, sleepErr
		}
	}
	return nil, lastErr
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

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
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

// loadCAPool reads a PEM-encoded CA bundle from path and returns an
// x509.CertPool that contains its certificates.
func loadCAPool(path string) (*x509.CertPool, error) {
	pem, err := os.ReadFile(path) //nolint:gosec // path comes from explicit config/flag
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("no certificates parsed from %s", path)
	}
	return pool, nil
}
