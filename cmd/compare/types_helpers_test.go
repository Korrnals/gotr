package compare

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func captureCompareStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe create error: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String()
}

func TestTruncateString(t *testing.T) {
	assert.Equal(t, "abc", truncateString("abcdef", 3))
	assert.Equal(t, "ab...", truncateString("abcdef", 5))
	assert.Equal(t, "abc", truncateString("abc", 5))
	assert.Equal(t, "", truncateString("abc", 0))
}

func TestPadRight(t *testing.T) {
	assert.Equal(t, "abc  ", padRight("abc", 5))
	assert.Equal(t, "abcdef", padRight("abcdef", 3))
}

func TestPrintHorizontalBorder(t *testing.T) {
	out := captureCompareStdout(t, func() {
		printHorizontalBorder("[", "+", "]", []int{2, 3})
	})

	assert.Contains(t, out, "[")
	assert.Contains(t, out, "+")
	assert.Contains(t, out, "]")
}

func TestPrintRow(t *testing.T) {
	out := captureCompareStdout(t, func() {
		printRow([]string{"abcdef", "x"}, []int{4, 2})
	})

	assert.Contains(t, out, "a...")
	assert.Contains(t, out, "x")
	assert.Contains(t, out, "│")
}

func TestPrintHeader(t *testing.T) {
	out := captureCompareStdout(t, func() {
		printHeader("Very long title", 6)
	})

	assert.NotEmpty(t, strings.TrimSpace(out))
}

func TestPrintSeparator(t *testing.T) {
	out := captureCompareStdout(t, func() {
		printSeparator([]int{2, 3})
	})

	assert.Contains(t, out, "┼")
}
