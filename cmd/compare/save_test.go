// Package compare tests - comprehensive test suite for save functionality
package compare

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Тесты для saveToFile ====================

func TestSaveToFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")
	data := []byte("test data")

	err := saveToFile(data, path)

	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, data, content)
}

func TestSaveToFile_InvalidPath(t *testing.T) {
	err := saveToFile([]byte("test"), "/nonexistent/path/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка записи файла")
}

// ==================== Тесты для saveCompareResult ====================

func TestSaveCompareResult_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "result.json")

	result := CompareResult{
		Resource:   "cases",
		Project1ID: 1,
		Project2ID: 2,
		OnlyInFirst: []ItemInfo{
			{ID: 1, Name: "Case A"},
		},
	}

	err := saveCompareResult(result, "json", path)
	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	var decoded CompareResult
	err = json.Unmarshal(content, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "cases", decoded.Resource)
}

func TestSaveCompareResult_YAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "result.yaml")

	result := CompareResult{
		Resource:     "suites",
		Project1ID:   1,
		Project2ID:   2,
		OnlyInFirst:  []ItemInfo{{ID: 1, Name: "Suite A"}},
		OnlyInSecond: []ItemInfo{{ID: 2, Name: "Suite B"}},
	}

	err := saveCompareResult(result, "yaml", path)
	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "resource: suites")
}

func TestSaveCompareResult_CSV(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "result.csv")

	result := CompareResult{
		Resource:     "plans",
		Project1ID:   1,
		Project2ID:   2,
		OnlyInFirst:  []ItemInfo{{ID: 1, Name: "Plan A"}},
		OnlyInSecond: []ItemInfo{{ID: 2, Name: "Plan B"}},
	}

	err := saveCompareResult(result, "csv", path)
	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Type,Name,ID Project 1,ID Project 2")
}

func TestSaveCompareResult_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "result.txt")

	result := CompareResult{Resource: "cases"}

	err := saveCompareResult(result, "txt", path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не поддерживается")
}

func TestSaveCompareResult_InvalidPath(t *testing.T) {
	result := CompareResult{Resource: "cases"}

	err := saveCompareResult(result, "json", "/nonexistent/path/result.json")
	assert.Error(t, err)
}

// ==================== Тесты для saveCSV ====================

func TestSaveCSV_Success(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "result.csv")

	result := CompareResult{
		Resource:   "cases",
		Project1ID: 1,
		Project2ID: 2,
		OnlyInFirst: []ItemInfo{
			{ID: 1, Name: "Case A"},
		},
		OnlyInSecond: []ItemInfo{
			{ID: 2, Name: "Case B"},
		},
		Common: []CommonItemInfo{
			{Name: "Case C", ID1: 3, ID2: 4},
		},
	}

	err := saveCSV(result, path)
	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Type,Name,ID Project 1,ID Project 2")
	assert.Contains(t, string(content), "Only in Project 1")
	assert.Contains(t, string(content), "Case A")
	assert.Contains(t, string(content), "Case B")
	assert.Contains(t, string(content), "Case C")
}

func TestSaveCSV_OnlyFirst(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "result.csv")

	result := CompareResult{
		Resource:    "cases",
		Project1ID:  1,
		Project2ID:  2,
		OnlyInFirst: []ItemInfo{{ID: 1, Name: "Case A"}},
	}

	err := saveCSV(result, path)
	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Case A")
}

func TestSaveCSV_OnlySecond(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "result.csv")

	result := CompareResult{
		Resource:     "cases",
		Project1ID:   1,
		Project2ID:   2,
		OnlyInSecond: []ItemInfo{{ID: 2, Name: "Case B"}},
	}

	err := saveCSV(result, path)
	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Case B")
}

func TestSaveCSV_Common(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "result.csv")

	result := CompareResult{
		Resource:   "cases",
		Project1ID: 1,
		Project2ID: 2,
		Common: []CommonItemInfo{
			{Name: "Case C", ID1: 3, ID2: 4},
		},
	}

	err := saveCSV(result, path)
	assert.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Case C")
}

func TestSaveCSV_InvalidPath(t *testing.T) {
	result := CompareResult{Resource: "cases"}

	err := saveCSV(result, "/nonexistent/path/result.csv")
	assert.Error(t, err)
}

// ==================== Тесты для printJSON ====================

func TestPrintJSON(t *testing.T) {
	result := CompareResult{
		Resource:   "cases",
		Project1ID: 1,
		Project2ID: 2,
		OnlyInFirst: []ItemInfo{
			{ID: 1, Name: "Case A"},
		},
		OnlyInSecond: []ItemInfo{
			{ID: 2, Name: "Case B"},
		},
		Common: []CommonItemInfo{
			{Name: "Case C", ID1: 3, ID2: 4, IDsMatch: false},
		},
	}

	var buf bytes.Buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printJSON(result)

	w.Close()
	os.Stdout = old
	buf.ReadFrom(r)

	assert.NoError(t, err)

	var decoded CompareResult
	err = json.Unmarshal(buf.Bytes(), &decoded)
	assert.NoError(t, err)
	assert.Equal(t, result.Resource, decoded.Resource)
}

// ==================== Тесты для printYAML ====================

func TestPrintYAML(t *testing.T) {
	result := CompareResult{
		Resource:   "cases",
		Project1ID: 1,
		Project2ID: 2,
		OnlyInFirst: []ItemInfo{
			{ID: 1, Name: "Case A"},
		},
	}

	var buf bytes.Buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printYAML(result)

	w.Close()
	os.Stdout = old
	buf.ReadFrom(r)

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "resource: cases")
}

// ==================== Тесты для printCSV ====================

func TestPrintCSV(t *testing.T) {
	result := CompareResult{
		Resource:   "cases",
		Project1ID: 1,
		Project2ID: 2,
		OnlyInFirst: []ItemInfo{
			{ID: 1, Name: "Case A"},
		},
		OnlyInSecond: []ItemInfo{
			{ID: 2, Name: "Case B"},
		},
		Common: []CommonItemInfo{
			{Name: "Case C", ID1: 3, ID2: 4},
		},
	}

	var buf bytes.Buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printCSV(result)

	w.Close()
	os.Stdout = old
	buf.ReadFrom(r)

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Type,Name,ID Project 1,ID Project 2")
}

// ==================== Тесты для printTable ====================

func TestPrintTable_OnlyFirst(t *testing.T) {
	result := CompareResult{
		Resource:    "cases",
		Project1ID:  1,
		Project2ID:  2,
		OnlyInFirst: []ItemInfo{{ID: 1, Name: "Case A"}},
	}

	var buf bytes.Buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printTable(result, "Project One", "Project Two")

	w.Close()
	os.Stdout = old
	buf.ReadFrom(r)

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Project One")
	assert.Contains(t, output, "Project Two")
	assert.Contains(t, output, "Case A")
}

func TestPrintTable_OnlySecond(t *testing.T) {
	result := CompareResult{
		Resource:     "cases",
		Project1ID:   1,
		Project2ID:   2,
		OnlyInSecond: []ItemInfo{{ID: 2, Name: "Case B"}},
	}

	var buf bytes.Buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printTable(result, "P1", "P2")

	w.Close()
	os.Stdout = old
	buf.ReadFrom(r)

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Case B")
}

func TestPrintTable_WithIDsMatch(t *testing.T) {
	result := CompareResult{
		Resource:   "cases",
		Project1ID: 1,
		Project2ID: 2,
		Common: []CommonItemInfo{
			{Name: "Case C", ID1: 3, ID2: 3, IDsMatch: true},
		},
	}

	var buf bytes.Buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printTable(result, "P1", "P2")

	w.Close()
	os.Stdout = old
	buf.ReadFrom(r)

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Case C")
	// IDsMatch true, поэтому не должно быть сообщения о разных ID
	assert.NotContains(t, output, "разные ID")
}

func TestPrintTable_WithIDsMismatch(t *testing.T) {
	result := CompareResult{
		Resource:   "cases",
		Project1ID: 1,
		Project2ID: 2,
		Common: []CommonItemInfo{
			{Name: "Case C", ID1: 3, ID2: 4, IDsMatch: false},
		},
	}

	var buf bytes.Buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printTable(result, "P1", "P2")

	w.Close()
	os.Stdout = old
	buf.ReadFrom(r)

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Case C")
	assert.Contains(t, output, "разные ID")
}
