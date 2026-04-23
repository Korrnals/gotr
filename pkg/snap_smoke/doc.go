// Package snap_smoke provides end-to-end smoke tests for gotr snap functionality.
// Tests verify the full snap/rollback cycle against a TestRail-compatible API.
//
// # Default Mode (no configuration needed)
//
// By default, tests start an in-memory FakeTestRail server (httptest) that
// emulates the TestRail API v2 endpoints used by snap operations.
// Routes are compiled from [pkg/testrailapi.APIPath] definitions, ensuring
// the mock stays in sync with the real API structure. List endpoints return
// paginated wrapper responses (TestRail 6.7+ format).
// No external dependencies, no credentials — just run:
//
//	go test -tags smoke ./pkg/snap_smoke/ -v
//
// # Real Server Mode
//
// To run against a live TestRail instance, set environment variables:
//
//	GOTR_SMOKE_URL      — TestRail base URL (required)
//	GOTR_SMOKE_USER     — TestRail username/email (required)
//	GOTR_SMOKE_KEY      — TestRail API key (required)
//	GOTR_SMOKE_PROJECT  — Project ID to use (required)
//	GOTR_SMOKE_SUITE    — Suite ID (required for multi-suite projects)
//	GOTR_SMOKE_INSECURE — Skip TLS verification ("true"/"1", default: false)
//
// Example:
//
//	GOTR_SMOKE_URL=http://localhost:8080 \
//	GOTR_SMOKE_USER=admin@example.com \
//	GOTR_SMOKE_KEY=yourkey \
//	GOTR_SMOKE_PROJECT=3 \
//	GOTR_SMOKE_SUITE=1 \
//	go test -tags smoke ./pkg/snap_smoke/ -v
package snap_smoke
