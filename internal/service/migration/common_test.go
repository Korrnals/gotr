package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/paths"
)

// MockClient — alias for client.MockClient for backward compatibility
type MockClient = client.MockClient

// ImportTestCase — universal structure for import tests (generics)
// `T` here is the *response* type from Client (e.g. data.GetSharedStepsResponse),
// so tests use named response types directly.
type ImportTestCase[T any] struct {
	name         string
	filtered     T
	dryRun       bool
	mockFunc     interface{} // func for mock (type assertion in tests)
	wantCount    int         // for mapping.Count (SharedSteps/Suites)
		wantImported int         // for imported counter (Cases)
}

// getImportTestCases — universal test cases function (generics)
func getImportTestCases[T any]() []ImportTestCase[T] {
	return []ImportTestCase[T]{
		{
			name:         "dry-run — import skipped",
			dryRun:       true,
			wantCount:    0,
			wantImported: 0,
		},
		{
			name:         "successful import",
			dryRun:       false,
			wantCount:    1,
			wantImported: 1,
		},
		{
			name:         "import error",
			dryRun:       false,
			wantCount:    0,
			wantImported: 0,
		},
		{
			name:         "many elements — concurrency works",
			dryRun:       false,
			wantCount:    3,
			wantImported: 3,
		},
	}
}

// logDir — directory for test logs (created if it doesn't exist).
// Uses `paths.EnsureLogsDirPath()` and creates a `test_runs` subfolder to isolate test logs.
func logDir() string {
	base, err := paths.EnsureLogsDirPath()
	if err != nil {
		panic(err)
	}
	testPath := filepath.Join(base, "test_runs")
	if err := os.MkdirAll(testPath, 0o755); err != nil {
		panic(err)
	}
	return testPath
}

// setupTestMigration uses the original constructor from types.go
func setupTestMigration(t *testing.T, mock *MockClient) *Migration {
	m, err := NewMigration(mock, 1, 10, 2, 20, "Title", logDir())
	if err != nil {
		t.Fatalf("Failed to setup migration: %v", err)
	}

	// Optionally add automatic Close call on test completion
	t.Cleanup(func() {
		m.Close()
	})

	return m
}
