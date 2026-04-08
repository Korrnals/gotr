package embed

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func requireSupportedEmbeddedJQPlatform(t *testing.T) {
	t.Helper()

	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		t.Skip("embedded jq binary is unavailable for this platform")
	}
}

func runWithCapturedStdIO(t *testing.T, fn func() error) (error, string, string) {
	t.Helper()

	outFile, err := os.CreateTemp(t.TempDir(), "stdout-*.log")
	if err != nil {
		t.Fatalf("create stdout temp file: %v", err)
	}
	errFile, err := os.CreateTemp(t.TempDir(), "stderr-*.log")
	if err != nil {
		_ = outFile.Close()
		t.Fatalf("create stderr temp file: %v", err)
	}

	oldOut := os.Stdout
	oldErr := os.Stderr
	os.Stdout = outFile
	os.Stderr = errFile

	runErr := fn()

	os.Stdout = oldOut
	os.Stderr = oldErr

	if err := outFile.Close(); err != nil {
		t.Fatalf("close stdout temp file: %v", err)
	}
	if err := errFile.Close(); err != nil {
		t.Fatalf("close stderr temp file: %v", err)
	}

	outBytes, err := os.ReadFile(outFile.Name())
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	errBytes, err := os.ReadFile(errFile.Name())
	if err != nil {
		t.Fatalf("read captured stderr: %v", err)
	}

	return runErr, string(outBytes), string(errBytes)
}

func chdirToTempDir(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}

	return tmpDir
}

func withDevNullStdIO(t *testing.T) func() {
	t.Helper()

	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}

	oldOut := os.Stdout
	oldErr := os.Stderr
	os.Stdout = devNull
	os.Stderr = devNull

	return func() {
		os.Stdout = oldOut
		os.Stderr = oldErr
		_ = devNull.Close()
	}
}

func TestRunEmbeddedJQ_Success(t *testing.T) {
	requireSupportedEmbeddedJQPlatform(t)

	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}

	restore := withDevNullStdIO(t)
	defer restore()

	input := []byte(`{"name":"gotr","count":2}`)
	if err := RunEmbeddedJQ(input, ".name"); err != nil {
		t.Fatalf("RunEmbeddedJQ returned error: %v", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("readdir temp dir: %v", err)
	}

	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "jq-") {
			t.Fatalf("temporary jq binary must be cleaned up, found: %s", filepath.Join(tmpDir, e.Name()))
		}
	}
}

func TestRunEmbeddedJQ_InvalidFilter_ReturnsError(t *testing.T) {
	requireSupportedEmbeddedJQPlatform(t)

	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}

	restore := withDevNullStdIO(t)
	defer restore()

	err = RunEmbeddedJQ([]byte(`{"x":1}`), "(")
	if err == nil {
		t.Fatal("expected error for invalid jq filter")
	}

	if !strings.Contains(err.Error(), "{jq_embed}") {
		t.Fatalf("expected jq_embed context in error, got: %v", err)
	}
}

func TestRunEmbeddedJQ_EmptyFilter_UsesDefault(t *testing.T) {
	requireSupportedEmbeddedJQPlatform(t)

	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}

	restore := withDevNullStdIO(t)
	defer restore()

	if err := RunEmbeddedJQ([]byte(`{"ok":true}`), ""); err != nil {
		t.Fatalf("RunEmbeddedJQ with empty filter returned error: %v", err)
	}
}

func TestRunEmbeddedJQ_InvalidJSONInput_ReturnsError(t *testing.T) {
	requireSupportedEmbeddedJQPlatform(t)

	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}

	restore := withDevNullStdIO(t)
	defer restore()

	err = RunEmbeddedJQ([]byte(`{"x":`), ".")
	if err == nil {
		t.Fatal("expected error for invalid JSON input")
	}
	if !strings.Contains(err.Error(), "{jq_embed}") {
		t.Fatalf("expected jq_embed context in error, got: %v", err)
	}
}

func TestRunEmbeddedJQ_GetwdError(t *testing.T) {
	requireSupportedEmbeddedJQPlatform(t)

	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	if err := os.RemoveAll(tmpDir); err != nil {
		t.Fatalf("remove temp dir: %v", err)
	}

	err = RunEmbeddedJQ([]byte(`{"x":1}`), ".")
	if err == nil {
		t.Fatal("expected error when current working directory is unavailable")
	}
}

func TestRunEmbeddedJQ_StdoutAndArgvFilter(t *testing.T) {
	requireSupportedEmbeddedJQPlatform(t)
	_ = chdirToTempDir(t)

	runErr, stdout, stderr := runWithCapturedStdIO(t, func() error {
		return RunEmbeddedJQ([]byte(`{"name":"gotr","count":2}`), `.name + "-" + (.count|tostring)`)
	})

	if runErr != nil {
		t.Fatalf("RunEmbeddedJQ returned error: %v", runErr)
	}
	if !strings.Contains(stdout, `"gotr-2"`) {
		t.Fatalf("expected stdout to contain computed jq output, got: %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr on success, got: %q", stderr)
	}
}

func TestRunEmbeddedJQ_InvalidFilter_WritesStderr(t *testing.T) {
	requireSupportedEmbeddedJQPlatform(t)
	_ = chdirToTempDir(t)

	runErr, _, stderr := runWithCapturedStdIO(t, func() error {
		return RunEmbeddedJQ([]byte(`{"x":1}`), "(")
	})

	if runErr == nil {
		t.Fatal("expected error for invalid filter")
	}
	if strings.TrimSpace(stderr) == "" {
		t.Fatal("expected stderr output for invalid jq filter")
	}
}

func TestRunEmbeddedJQ_EmptyInput_NoOutputNoError(t *testing.T) {
	requireSupportedEmbeddedJQPlatform(t)
	_ = chdirToTempDir(t)

	runErr, stdout, stderr := runWithCapturedStdIO(t, func() error {
		return RunEmbeddedJQ(nil, ".")
	})

	if runErr != nil {
		t.Fatalf("expected no error for empty input, got: %v", runErr)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Fatalf("expected empty stdout for empty input, got: %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr for empty input, got: %q", stderr)
	}
}

func TestRunEmbeddedJQ_CreateTempErrorInReadOnlyCwd(t *testing.T) {
	requireSupportedEmbeddedJQPlatform(t)
	if runtime.GOOS == "windows" {
		t.Skip("permissions model differs on windows")
	}

	tmpDir := chdirToTempDir(t)

	if err := os.Chmod(tmpDir, 0o555); err != nil {
		t.Fatalf("chmod read-only temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(tmpDir, 0o755) })

	err := RunEmbeddedJQ([]byte(`{"x":1}`), ".")
	if err == nil {
		t.Fatal("expected error when temp file cannot be created in read-only cwd")
	}
}

func TestRunEmbeddedJQ_IndependentFromPATH(t *testing.T) {
	requireSupportedEmbeddedJQPlatform(t)
	_ = chdirToTempDir(t)

	oldPath, hadPath := os.LookupEnv("PATH")
	if err := os.Setenv("PATH", ""); err != nil {
		t.Fatalf("clear PATH: %v", err)
	}
	t.Cleanup(func() {
		if hadPath {
			_ = os.Setenv("PATH", oldPath)
		} else {
			_ = os.Unsetenv("PATH")
		}
	})

	runErr, stdout, stderr := runWithCapturedStdIO(t, func() error {
		return RunEmbeddedJQ([]byte(`{"ok":true}`), ".ok")
	})

	if runErr != nil {
		t.Fatalf("RunEmbeddedJQ returned error with empty PATH: %v", runErr)
	}
	if !strings.Contains(stdout, "true") {
		t.Fatalf("expected jq output in stdout, got: %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got: %q", stderr)
	}
}

func TestRunEmbeddedJQ_InvalidJSON_WritesStderr(t *testing.T) {
	requireSupportedEmbeddedJQPlatform(t)
	_ = chdirToTempDir(t)

	runErr, _, stderr := runWithCapturedStdIO(t, func() error {
		return RunEmbeddedJQ([]byte(`{"x":`), ".")
	})

	if runErr == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if strings.TrimSpace(stderr) == "" {
		t.Fatal("expected stderr output for invalid JSON")
	}
}

func TestRunEmbeddedJQ_StdinEOFBehavior(t *testing.T) {
	requireSupportedEmbeddedJQPlatform(t)
	_ = chdirToTempDir(t)

	runErr, stdout, stderr := runWithCapturedStdIO(t, func() error {
		return RunEmbeddedJQ([]byte{}, ".")
	})

	if runErr != nil {
		t.Fatalf("RunEmbeddedJQ returned error for explicit empty input: %v", runErr)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Fatalf("expected no stdout for explicit empty input, got: %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected no stderr for explicit empty input, got: %q", stderr)
	}
}

func TestRunEmbeddedJQ_StdStreamsRemainWritableAfterRun(t *testing.T) {
	requireSupportedEmbeddedJQPlatform(t)
	_ = chdirToTempDir(t)

	origOut := os.Stdout
	origErr := os.Stderr

	if err := RunEmbeddedJQ([]byte(`{"k":1}`), ".k"); err != nil {
		t.Fatalf("RunEmbeddedJQ returned error: %v", err)
	}

	if _, err := io.WriteString(os.Stdout, ""); err != nil {
		t.Fatalf("stdout should remain writable: %v", err)
	}
	if _, err := io.WriteString(os.Stderr, ""); err != nil {
		t.Fatalf("stderr should remain writable: %v", err)
	}

	if os.Stdout != origOut {
		t.Fatal("RunEmbeddedJQ must not replace os.Stdout")
	}
	if os.Stderr != origErr {
		t.Fatal("RunEmbeddedJQ must not replace os.Stderr")
	}
}

func assertNoJQTempArtifacts(t *testing.T, dir string) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir temp dir: %v", err)
	}

	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "jq-") {
			t.Fatalf("temporary jq artifact must be cleaned up, found: %s", filepath.Join(dir, e.Name()))
		}
	}
}

func cleanupJQTempArtifacts(t *testing.T, dir string) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "jq-") {
			_ = os.RemoveAll(filepath.Join(dir, e.Name()))
		}
	}
}

func runWriteOrChmodFailureCase(
	t *testing.T,
	name string,
	prepareSabotage func(dir string) (start func(), stop func()),
	wantErrSubstring string,
) {
	t.Helper()

	const maxAttempts = 30
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		tmpDir := chdirToTempDir(t)
		restore := withDevNullStdIO(t)

		start, stop := prepareSabotage(tmpDir)
		start()
		err := RunEmbeddedJQ([]byte(`{"k":1}`), ".")
		stop()
		restore()
		cleanupJQTempArtifacts(t, tmpDir)
		if err == nil {
			lastErr = nil
			continue
		}

		lastErr = err
		if strings.Contains(err.Error(), wantErrSubstring) {
			return
		}
	}

	if lastErr == nil {
		t.Skipf("%s: could not trigger error after %d attempts (race window too tight on this platform)", name, maxAttempts)
	}
	t.Fatalf("%s: expected error containing %q, got: %v", name, wantErrSubstring, lastErr)
}

func TestRunEmbeddedJQ_WriteAndChmodFailureBranches(t *testing.T) {
	requireSupportedEmbeddedJQPlatform(t)
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("filesystem sabotage scenarios are intended for unix-like platforms")
	}

	tests := []struct {
		name             string
		wantErrSubstring string
		prepareSabotage  func(dir string) (start func(), stop func())
	}{
		{
			name:             "Chmod path removed after binary write",
			wantErrSubstring: "failed to set executable permissions",
			prepareSabotage: func(dir string) (start func(), stop func()) {
				stopCh := make(chan struct{})
				doneCh := make(chan struct{})

				start = func() {
					go func() {
						defer close(doneCh)
						deadline := time.Now().Add(300 * time.Millisecond)
						for time.Now().Before(deadline) {
							select {
							case <-stopCh:
								return
							default:
							}

							entries, err := os.ReadDir(dir)
							if err != nil {
								time.Sleep(time.Millisecond)
								continue
							}

							for _, e := range entries {
								if !strings.HasPrefix(e.Name(), "jq-") {
									continue
								}

								path := filepath.Join(dir, e.Name())
								info, statErr := os.Stat(path)
								if statErr != nil {
									continue
								}
								if info.Size() > 0 {
									_ = os.Remove(path)
									return
								}
							}

							time.Sleep(time.Millisecond)
						}
					}()
				}

				stop = func() {
					close(stopCh)
					<-doneCh
				}
				return start, stop
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWriteOrChmodFailureCase(t, tc.name, tc.prepareSabotage, tc.wantErrSubstring)
		})
	}
}

func TestRunEmbeddedJQ_CmdRunError_CleansUpTempBinary(t *testing.T) {
	requireSupportedEmbeddedJQPlatform(t)
	tmpDir := chdirToTempDir(t)
	restore := withDevNullStdIO(t)
	defer restore()

	err := RunEmbeddedJQ([]byte(`{"x":1}`), "(")
	if err == nil {
		t.Fatal("expected error for invalid jq filter")
	}
	if !strings.Contains(err.Error(), "{jq_embed}") {
		t.Fatalf("expected jq_embed context in error, got: %v", err)
	}

	assertNoJQTempArtifacts(t, tmpDir)
}

// TestRunEmbeddedJQ_WriteFileError_ReadOnlyBinary races to chmod the temp binary
// to 0444 before os.WriteFile opens it, triggering the WriteFile error branch.
// The race window is tight (between Close and WriteFile's OpenFile), so the test
// skips gracefully when not won; it never hard-fails.
func TestRunEmbeddedJQ_WriteFileError_ReadOnlyBinary(t *testing.T) {
	requireSupportedEmbeddedJQPlatform(t)
	if runtime.GOOS == "windows" {
		t.Skip("unix chmod semantics required for this test")
	}

	const maxAttempts = 300

	tmpDir := t.TempDir()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	restore := withDevNullStdIO(t)
	defer restore()

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		stopCh := make(chan struct{})
		doneCh := make(chan struct{})

		// Goroutine: tight-loop poll for jq-* files at size 0 (just created,
		// not yet written by os.WriteFile). Chmod them to 0444 so that
		// WriteFile's os.OpenFile(O_WRONLY) fails with EACCES.
		go func() {
			defer close(doneCh)
			for {
				select {
				case <-stopCh:
					return
				default:
				}
				entries, rdErr := os.ReadDir(tmpDir)
				if rdErr != nil {
					continue
				}
				for _, e := range entries {
					if !strings.HasPrefix(e.Name(), "jq-") {
						continue
					}
					p := filepath.Join(tmpDir, e.Name())
					info, statErr := os.Stat(p)
					if statErr != nil || info.Size() != 0 {
						continue
					}
					_ = os.Chmod(p, 0444)
				}
			}
		}()

		runErr := RunEmbeddedJQ([]byte(`{"k":1}`), ".")
		close(stopCh)
		<-doneCh

		// Fix and remove any leftover jq-* artifacts (e.g. left at 0444).
		if ents, rdErr := os.ReadDir(tmpDir); rdErr == nil {
			for _, e := range ents {
				if strings.HasPrefix(e.Name(), "jq-") {
					p := filepath.Join(tmpDir, e.Name())
					_ = os.Chmod(p, 0o644)
					_ = os.Remove(p)
				}
			}
		}

		// WriteFile error returns a raw OS error (no "{jq_embed}" prefix).
		// Getwd and CreateTemp errors also have no prefix, but both succeed
		// here since tmpDir is a valid writable directory.
		if runErr != nil && !strings.Contains(runErr.Error(), "{jq_embed}") {
			return // WriteFile error branch covered
		}
	}

	t.Skipf("WriteFile error branch not triggered in %d attempts (race window too tight on this system)", maxAttempts)
}

func TestSelectEmbeddedJQBinary(t *testing.T) {
	tests := []struct {
		name    string
		goos    string
		wantErr bool
	}{
		{name: "linux", goos: "linux", wantErr: false},
		{name: "darwin", goos: "darwin", wantErr: false},
		{name: "windows", goos: "windows", wantErr: false},
		{name: "unsupported", goos: "plan9", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bin, err := selectEmbeddedJQBinary(tc.goos)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for platform %s", tc.goos)
				}
				if !strings.Contains(err.Error(), tc.goos) {
					t.Fatalf("expected platform name in error, got: %v", err)
				}
				if len(bin) != 0 {
					t.Fatalf("expected empty binary on error, got len=%d", len(bin))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error for platform %s: %v", tc.goos, err)
			}
			if len(bin) == 0 {
				t.Fatalf("expected embedded binary bytes for %s", tc.goos)
			}
		})
	}
}

func TestRunEmbeddedJQ_SelectBinaryError(t *testing.T) {
	origSelect := selectEmbeddedJQBinaryFunc
	selectEmbeddedJQBinaryFunc = func(goos string) ([]byte, error) {
		return nil, errors.New("select failure")
	}
	t.Cleanup(func() {
		selectEmbeddedJQBinaryFunc = origSelect
	})

	err := RunEmbeddedJQ([]byte(`{"x":1}`), ".")
	if err == nil {
		t.Fatal("expected selector error")
	}
	if !strings.Contains(err.Error(), "select failure") {
		t.Fatalf("expected selector error text, got: %v", err)
	}
}

func TestRunEmbeddedJQ_WriteBinaryError_Injected(t *testing.T) {
	requireSupportedEmbeddedJQPlatform(t)
	tmpDir := chdirToTempDir(t)
	restore := withDevNullStdIO(t)
	defer restore()

	origWrite := writeEmbeddedBinaryFile
	writeEmbeddedBinaryFile = func(_ string, _ []byte, _ os.FileMode) error {
		return errors.New("injected write failure")
	}
	t.Cleanup(func() {
		writeEmbeddedBinaryFile = origWrite
	})

	err := RunEmbeddedJQ([]byte(`{"x":1}`), ".")
	if err == nil {
		t.Fatal("expected write error")
	}
	if !strings.Contains(err.Error(), "injected write failure") {
		t.Fatalf("expected injected write error text, got: %v", err)
	}

	assertNoJQTempArtifacts(t, tmpDir)
}
