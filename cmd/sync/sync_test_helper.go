package sync

// testContextKey is an unexported key type for context values in tests.
type testContextKey string

// testHTTPClientKey is the context key for tests (must match cmd.httpClientKey).
const testHTTPClientKey testContextKey = "httpClient"
