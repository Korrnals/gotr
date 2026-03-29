package embed

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

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
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		t.Skip("embedded jq binary is unavailable for this platform")
	}

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
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		t.Skip("embedded jq binary is unavailable for this platform")
	}

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
