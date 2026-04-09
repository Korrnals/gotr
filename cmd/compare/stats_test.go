package compare

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatCompleteness(t *testing.T) {
	ok := formatCompleteness(100, 100, 2, 2, 0)
	assert.Contains(t, ok, "100/100")

	unknownExpectedNoErrors := formatCompleteness(10, 0, 2, 2, 0)
	assert.Contains(t, unknownExpectedNoErrors, "10")
	assert.Contains(t, unknownExpectedNoErrors, "2/2 suites")

	unknownExpectedWithErrors := formatCompleteness(10, 0, 0, 0, 3)
	assert.Contains(t, unknownExpectedWithErrors, "errors: 3 pages")

	moreThanExpected := formatCompleteness(120, 100, 2, 1, 1)
	assert.Contains(t, moreThanExpected, "possible duplicates")

	partial := formatCompleteness(80, 100, 2, 1, 1)
	assert.Contains(t, partial, "80/100")
}

func TestFormatIntegrityCheck(t *testing.T) {
	assert.Equal(t, "", formatIntegrityCheck(10, 0, 10, 0))

	ok := formatIntegrityCheck(10, 2, 10, 0)
	assert.Contains(t, ok, "10 (2 suites")

	withEmpty := formatIntegrityCheck(10, 2, 10, 1)
	assert.Contains(t, withEmpty, "empty: 1")

	delta := formatIntegrityCheck(12, 2, 10, 0)
	assert.Contains(t, delta, "delta +2")
}

func captureStdoutString(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}

	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = oldStdout

	out, readErr := io.ReadAll(r)
	_ = r.Close()
	if readErr != nil {
		t.Fatalf("failed to read stdout: %v", readErr)
	}

	return string(out)
}

func TestPrintCasesStatsWithErrors_InterruptedBannerPriority(t *testing.T) {
	out := captureStdoutString(t, func() {
		PrintCasesStatsWithErrors(
			10, 20,
			1, 2, 3,
			1500*time.Millisecond,
			casesExecutionStats{
				Interrupted:      true,
				FailedPagesAfter: 9,
			},
		)
	})

	assert.Contains(t, out, "INTERRUPTED")
	assert.NotContains(t, out, "PARTIAL")
}

func TestPrintCasesStatsWithErrors_PartialAndRetrySections(t *testing.T) {
	stats := casesExecutionStats{
		Project1: projectDataStats{
			Suites:            3,
			Sections:          4,
			CasesRaw:          12,
			CasesUnique:       10,
			CasesExpected:     11,
			SuitesWithTotal:   3,
			SuitesVerified:    2,
			SuiteDetailsCount: 3,
			SuiteDetailsSum:   10,
			SuiteDetailsEmpty: 1,
			TotalPages:        6,
			FailedPages:       1,
			EmptyTitles:       2,
			Elapsed:           700 * time.Millisecond,
		},
		Project2: projectDataStats{
			Suites:            2,
			Sections:          3,
			CasesRaw:          8,
			CasesUnique:       8,
			CasesExpected:     8,
			SuitesWithTotal:   2,
			SuitesVerified:    2,
			SuiteDetailsCount: 2,
			SuiteDetailsSum:   8,
			TotalPages:        4,
			Elapsed:           600 * time.Millisecond,
		},
		LoadErrorsP1:      1,
		LoadErrorsP2:      2,
		FailedPagesBefore: 3,
		RetryAttempted:    true,
		RetryStats: retryFailedPagesStats{
			RecoveredPages: 1,
			UniquePages:    3,
			RecoveredCases: 5,
		},
		FailedPagesAfter: 2,
	}

	out := captureStdoutString(t, func() {
		PrintCasesStatsWithErrors(10, 20, 2, 1, 7, 2100*time.Millisecond, stats)
	})

	assert.Contains(t, out, "PARTIAL")
	assert.Contains(t, out, "Recovered pages")
	assert.Contains(t, out, "Cases recovered on retry")
	assert.Contains(t, out, "Failed pages after auto-retry")
	assert.Contains(t, out, "Cases (raw before dedup)")
	assert.Contains(t, out, "Download completeness")
	assert.Contains(t, out, "Cases without title")
}

func TestPrintCasesStatsWithErrors_CompleteNoRetry(t *testing.T) {
	out := captureStdoutString(t, func() {
		PrintCasesStatsWithErrors(
			30, 31,
			0, 0, 10,
			time.Second,
			casesExecutionStats{
				Project1: projectDataStats{Suites: 1, Sections: 1, CasesRaw: 5, CasesUnique: 5, Elapsed: 300 * time.Millisecond},
				Project2: projectDataStats{Suites: 1, Sections: 1, CasesRaw: 5, CasesUnique: 5, Elapsed: 300 * time.Millisecond},
			},
		)
	})

	assert.Contains(t, out, "COMPLETE")
	assert.NotContains(t, out, "Recovered pages")
}

func TestPrintCompareStats_StatusBanners(t *testing.T) {
	tests := []struct {
		name   string
		status []CompareStatus
		want   string
	}{
		{name: "default complete", status: nil, want: "COMPLETE"},
		{name: "explicit empty status falls back to complete", status: []CompareStatus{""}, want: "COMPLETE"},
		{name: "partial", status: []CompareStatus{CompareStatusPartial}, want: "PARTIAL"},
		{name: "interrupted", status: []CompareStatus{CompareStatusInterrupted}, want: "INTERRUPTED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdoutString(t, func() {
				PrintCompareStats("cases", 30, 31, 1, 2, 3, 1500*time.Millisecond, tt.status...)
			})
			assert.Contains(t, out, tt.want)
		})
	}
}
