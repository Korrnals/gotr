// internal/paths/paths_testing.go
//
// Test-mode sandbox guard.
//
// When the current process appears to be a `go test` binary (detected by
// inspecting os.Args[0] for the conventional ".test" suffix), this init()
// redirects HOME to a per-process temporary directory under os.TempDir()
// so that no test can mutate the real ~/.gotr/ on the developer's machine.
//
// Rationale: every code path that touches gotr-managed state
// (snap manifest, reports, exports, cache, config) flows through
// paths.BaseDir(); by installing a process-wide HOME sandbox we get
// complete isolation without requiring
// every package to ship its own TestMain.
//
// This file is compiled into every build of internal/paths, including
// production binaries — but the body is a no-op in non-test mode because
// the suffix check fails. We deliberately avoid importing the "testing"
// package here so that production binaries do not carry test-runtime
// linkage.

package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	if !isGoTestBinary() {
		return
	}
	if v := os.Getenv(HomeEnvVar); v != "" {
		// Honor an explicit GOTR_HOME override from the test process.
		return
	}
	homeDir, err := os.MkdirTemp("", fmt.Sprintf("gotr-test-home-%d-", os.Getpid()))
	if err != nil {
		// Fail closed: better to skip the override than crash all tests
		// on a broken /tmp; production code paths are unaffected.
		return
	}
	// Use HOME instead of GOTR_HOME so tests that intentionally call
	// t.Setenv("HOME", ... ) continue to work without per-package TestMain.
	_ = os.Setenv("HOME", homeDir)
}

// isGoTestBinary reports whether the current process is a "go test"
// compiled binary. Go writes test executables with names that end in
// ".test" (e.g. "paths.test", "snap.test"). When tests run via
// `go test ./...` Go also runs them via temp paths like
// "/tmp/go-build123/b001/foo.test"; the suffix convention still holds.
func isGoTestBinary() bool {
	if len(os.Args) == 0 {
		return false
	}
	exe := strings.ToLower(filepath.Base(os.Args[0]))
	return strings.HasSuffix(exe, ".test") || strings.HasSuffix(exe, ".test.exe")
}
