// Package compare tests - tests for cases comparison
package compare

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrintCasesStats(t *testing.T) {
	result := &CompareResult{
		Resource:     "cases",
		Project1ID:   30,
		Project2ID:   34,
		OnlyInFirst:  []ItemInfo{{ID: 1, Name: "Case A"}, {ID: 2, Name: "Case B"}},
		OnlyInSecond: []ItemInfo{{ID: 3, Name: "Case C"}},
		Common:       []CommonItemInfo{{Name: "Case D", ID1: 4, ID2: 5, IDsMatch: true}},
	}
	elapsed := 1500 * time.Millisecond

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printCasesStats(result, elapsed)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains key information
	assert.Contains(t, output, "СТАТИСТИКА ВЫПОЛНЕНИЯ")
	assert.Contains(t, output, "1.5s") // elapsed time
	assert.Contains(t, output, "4")   // total cases (2 + 1 + 1)
	assert.Contains(t, output, "30")  // project 1 ID
	assert.Contains(t, output, "34")  // project 2 ID
	assert.Contains(t, output, "2")   // only in first
	assert.Contains(t, output, "1")   // only in second / common
}

func TestPrintCasesStats_ZeroCases(t *testing.T) {
	result := &CompareResult{
		Resource:   "cases",
		Project1ID: 1,
		Project2ID: 2,
	}
	elapsed := 500 * time.Millisecond

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printCasesStats(result, elapsed)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "0") // total cases
	assert.Contains(t, output, "500ms")
}
