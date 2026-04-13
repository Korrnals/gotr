// Package snap_smoke provides end-to-end smoke tests for gotr snap functionality.
// These tests run against a real TestRail server and verify the full snap/rollback cycle.
//
// Configuration via environment variables:
//
//	GOTR_SMOKE_URL      — TestRail base URL (required)
//	GOTR_SMOKE_USER     — TestRail username/email (required)
//	GOTR_SMOKE_KEY      — TestRail API key (required)
//	GOTR_SMOKE_PROJECT  — Project ID to use (required)
//	GOTR_SMOKE_SUITE    — Suite ID (required for multi-suite projects)
//	GOTR_SMOKE_INSECURE — Skip TLS verification ("true"/"1", default: false)
//
// Run:
//
//	go test -tags smoke ./pkg/snap_smoke/ -v
package snap_smoke
