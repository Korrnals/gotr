package compare

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// printOnlyInProjectTable / printCommonTable / printIDMappingTable
// ---------------------------------------------------------------------------

func TestPrintOnlyInProjectTable_WithItems(t *testing.T) {
	out := captureCompareStdout(t, func() {
		printOnlyInProjectTable([]ItemInfo{
			{ID: 1, Name: "Alpha"},
			{ID: 2, Name: "Beta"},
		}, 10, "TestProject")
	})
	assert.Contains(t, out, "Only in project 10")
	assert.Contains(t, out, "Alpha")
	assert.Contains(t, out, "Beta")
}

func TestPrintOnlyInProjectTable_Empty(t *testing.T) {
	out := captureCompareStdout(t, func() {
		printOnlyInProjectTable(nil, 10, "TestProject")
	})
	assert.Contains(t, out, "(none)")
}

func TestPrintCommonTable_WithItems(t *testing.T) {
	out := captureCompareStdout(t, func() {
		printCommonTable([]CommonItemInfo{
			{Name: "SuiteA", ID1: 1, ID2: 1, IDsMatch: true},
			{Name: "SuiteB", ID1: 2, ID2: 5, IDsMatch: false},
		}, 10, 20)
	})
	assert.Contains(t, out, "Common in both projects")
	assert.Contains(t, out, "SuiteA")
	assert.Contains(t, out, "✓ Match")
	assert.Contains(t, out, "⚠ Differ")
}

func TestPrintCommonTable_Empty(t *testing.T) {
	out := captureCompareStdout(t, func() {
		printCommonTable(nil, 10, 20)
	})
	assert.Contains(t, out, "(none)")
}

func TestPrintIDMappingTable_WithDifferences(t *testing.T) {
	out := captureCompareStdout(t, func() {
		printIDMappingTable([]CommonItemInfo{
			{Name: "SuiteA", ID1: 1, ID2: 1, IDsMatch: true},
			{Name: "SuiteB", ID1: 2, ID2: 5, IDsMatch: false},
		})
	})
	assert.Contains(t, out, "ID mapping")
	assert.Contains(t, out, "SuiteB")
	assert.NotContains(t, out, "SuiteA") // matching IDs filtered out
}

func TestPrintIDMappingTable_AllMatch(t *testing.T) {
	out := captureCompareStdout(t, func() {
		printIDMappingTable([]CommonItemInfo{
			{Name: "SuiteA", ID1: 1, ID2: 1, IDsMatch: true},
		})
	})
	assert.Contains(t, out, "(all IDs match)")
}

// ---------------------------------------------------------------------------
// recommendedSyncIndex
// ---------------------------------------------------------------------------

func TestRecommendedSyncIndex(t *testing.T) {
	assert.Equal(t, 1, recommendedSyncIndex("suites"))
	assert.Equal(t, 2, recommendedSyncIndex("sections"))
	assert.Equal(t, 3, recommendedSyncIndex("shared_steps"))
	assert.Equal(t, 3, recommendedSyncIndex("shared-steps"))
	assert.Equal(t, 0, recommendedSyncIndex("cases"))
	assert.Equal(t, 0, recommendedSyncIndex(""))
	assert.Equal(t, 0, recommendedSyncIndex("unknown"))
}

// ---------------------------------------------------------------------------
// flagInt / flagDuration — missing-key branches
// ---------------------------------------------------------------------------

func TestFlagInt_MissingKey(t *testing.T) {
	m := map[string]any{"other": "str"}
	assert.Equal(t, 42, flagInt(m, "missing", 42))
}

func TestFlagInt_WrongType(t *testing.T) {
	m := map[string]any{"k": "not-int"}
	assert.Equal(t, 99, flagInt(m, "k", 99))
}

func TestFlagDuration_MissingKey(t *testing.T) {
	m := map[string]any{}
	assert.Equal(t, 5*time.Second, flagDuration(m, "missing", 5*time.Second))
}

func TestFlagDuration_WrongType(t *testing.T) {
	m := map[string]any{"k": "bad"}
	assert.Equal(t, 3*time.Second, flagDuration(m, "k", 3*time.Second))
}

// ---------------------------------------------------------------------------
// headerLine — short title branch (padding < 0)
// ---------------------------------------------------------------------------

func TestHeaderLine_ShortWidth(t *testing.T) {
	line := headerLine("very long title here", 5)
	assert.Contains(t, line, "│")
}

// ---------------------------------------------------------------------------
// saveAllResult — unsupported format
// ---------------------------------------------------------------------------

func TestSaveAllResult_UnsupportedFormat(t *testing.T) {
	err := saveAllResult(&allResult{}, "xml", "/dev/null")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestSaveAllResult_JSON_Default(t *testing.T) {
	tmp := t.TempDir() + "/out.json"
	err := saveAllResult(&allResult{}, "json", tmp)
	require.NoError(t, err)
	data, _ := os.ReadFile(tmp)
	assert.Contains(t, string(data), "{")
}

func TestSaveAllResult_YAML_Default(t *testing.T) {
	tmp := t.TempDir() + "/out.yaml"
	err := saveAllResult(&allResult{}, "yaml", tmp)
	require.NoError(t, err)
	data, _ := os.ReadFile(tmp)
	assert.NotEmpty(t, data)
}

// ---------------------------------------------------------------------------
// saveToFileWithPath — default format + csv
// ---------------------------------------------------------------------------

func TestSaveToFileWithPath_DefaultFormat(t *testing.T) {
	tmp := t.TempDir() + "/out.bin"
	err := saveToFileWithPath(CompareResult{Resource: "test"}, "unknown-format", tmp)
	require.NoError(t, err)
	data, _ := os.ReadFile(tmp)
	assert.Contains(t, string(data), "test") // JSON fallback
}

func TestSaveToFileWithPath_CSV(t *testing.T) {
	tmp := t.TempDir() + "/out.csv"
	err := saveToFileWithPath(CompareResult{
		Resource: "test",
		OnlyInFirst: []ItemInfo{{ID: 1, Name: "A"}},
	}, "csv", tmp)
	require.NoError(t, err)
	data, _ := os.ReadFile(tmp)
	assert.Contains(t, string(data), "A")
}

func TestSaveToFileWithPath_YAML(t *testing.T) {
	tmp := t.TempDir() + "/out.yaml"
	err := saveToFileWithPath(CompareResult{Resource: "suites"}, "yaml", tmp)
	require.NoError(t, err)
}
