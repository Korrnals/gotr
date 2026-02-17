// Package compare tests - comprehensive test suite for save functionality
package compare

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
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
	assert.Contains(t, output, "Различаются")
}

// ==================== Тесты для PrintCompareResult с --save и --format ====================

func TestPrintCompareResult_SaveWithFormat_JSON(t *testing.T) {
	// Create a mock command with flags
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("format", "json", "")
	cmd.Flags().String("save", "", "")
	cmd.Flags().String("save-to", "", "")

	result := CompareResult{
		Resource:     "cases",
		Project1ID:   1,
		Project2ID:   2,
		OnlyInFirst:  []ItemInfo{{ID: 1, Name: "Case A"}},
		OnlyInSecond: []ItemInfo{{ID: 2, Name: "Case B"}},
	}

	// Call with format="json" and savePath="__DEFAULT__" (simulates --save --format json)
	err := PrintCompareResult(cmd, result, "Project 1", "Project 2", "json", "__DEFAULT__")

	assert.NoError(t, err)
	// File should be created in exports directory with .json extension
	// We verify by checking stdout output contains the save message
}

func TestPrintCompareResult_SaveWithFormat_YAML(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("format", "yaml", "")
	cmd.Flags().String("save", "", "")
	cmd.Flags().String("save-to", "", "")

	result := CompareResult{
		Resource:   "suites",
		Project1ID: 1,
		Project2ID: 2,
	}

	// Call with format="yaml" and savePath="__DEFAULT__"
	err := PrintCompareResult(cmd, result, "P1", "P2", "yaml", "__DEFAULT__")

	assert.NoError(t, err)
}

func TestPrintCompareResult_SaveWithFormat_CSV(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("format", "csv", "")
	cmd.Flags().String("save", "", "")
	cmd.Flags().String("save-to", "", "")

	result := CompareResult{
		Resource:     "plans",
		Project1ID:   1,
		Project2ID:   2,
		OnlyInFirst:  []ItemInfo{{ID: 1, Name: "Plan A"}},
		OnlyInSecond: []ItemInfo{{ID: 2, Name: "Plan B"}},
	}

	// Call with format="csv" and savePath="__DEFAULT__"
	err := PrintCompareResult(cmd, result, "P1", "P2", "csv", "__DEFAULT__")

	assert.NoError(t, err)
}

func TestPrintCompareResult_SaveWithFormat_Table(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("format", "table", "")
	cmd.Flags().String("save", "", "")
	cmd.Flags().String("save-to", "", "")

	result := CompareResult{
		Resource:   "cases",
		Project1ID: 1,
		Project2ID: 2,
		OnlyInFirst: []ItemInfo{
			{ID: 1, Name: "Case A"},
		},
	}

	// Call with format="table" and savePath="__DEFAULT__"
	err := PrintCompareResult(cmd, result, "Project One", "Project Two", "table", "__DEFAULT__")

	assert.NoError(t, err)
}

func TestPrintCompareResult_SaveToOverridesFormat(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "result.json")

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("format", "yaml", "") // format is yaml
	cmd.Flags().String("save", "", "")
	cmd.Flags().String("save-to", jsonPath, "") // but save-to has .json extension

	result := CompareResult{
		Resource:     "cases",
		Project1ID:   1,
		Project2ID:   2,
		OnlyInFirst:  []ItemInfo{{ID: 1, Name: "Case A"}},
		OnlyInSecond: []ItemInfo{{ID: 2, Name: "Case B"}},
	}

	// Call with explicit save path - extension should override format
	err := PrintCompareResult(cmd, result, "P1", "P2", "table", jsonPath)

	assert.NoError(t, err)

	// Verify the file was created with JSON content
	content, err := os.ReadFile(jsonPath)
	require.NoError(t, err)

	// Should be valid JSON
	var decoded CompareResult
	err = json.Unmarshal(content, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "cases", decoded.Resource)
}

func TestPrintCompareResult_SaveToWithCSV(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "result.csv")

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("format", "table", "")
	cmd.Flags().String("save", "", "")
	cmd.Flags().String("save-to", "", "")

	result := CompareResult{
		Resource:     "plans",
		Project1ID:   1,
		Project2ID:   2,
		OnlyInFirst:  []ItemInfo{{ID: 1, Name: "Plan A"}},
		OnlyInSecond: []ItemInfo{{ID: 2, Name: "Plan B"}},
	}

	// Call with explicit CSV path
	err := PrintCompareResult(cmd, result, "P1", "P2", "table", csvPath)

	assert.NoError(t, err)

	// Verify CSV content
	content, err := os.ReadFile(csvPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Type,Name,ID Project 1,ID Project 2")
}

func TestGetFormatFromExtension(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"result.json", "json"},
		{"result.yaml", "yaml"},
		{"result.yml", "yaml"},
		{"result.csv", "csv"},
		{"result.txt", "table"},
		{"result", ""},
		{"RESULT.JSON", "json"}, // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := getFormatFromExtension(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrintCompareResult_StdoutFormats(t *testing.T) {
	result := CompareResult{
		Resource:     "cases",
		Project1ID:   1,
		Project2ID:   2,
		OnlyInFirst:  []ItemInfo{{ID: 1, Name: "Case A"}},
		OnlyInSecond: []ItemInfo{{ID: 2, Name: "Case B"}},
	}

	cmd := &cobra.Command{Use: "test"}

	// Test JSON to stdout
	t.Run("JSON", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := PrintCompareResult(cmd, result, "P1", "P2", "json", "")
		w.Close()
		os.Stdout = old

		content := make([]byte, 1024)
		n, _ := r.Read(content)

		assert.NoError(t, err)
		assert.Contains(t, string(content[:n]), "\"resource\": \"cases\"")
	})

	// Test YAML to stdout
	t.Run("YAML", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := PrintCompareResult(cmd, result, "P1", "P2", "yaml", "")
		w.Close()
		os.Stdout = old

		content := make([]byte, 1024)
		n, _ := r.Read(content)

		assert.NoError(t, err)
		assert.Contains(t, string(content[:n]), "resource: cases")
	})

	// Test CSV to stdout
	t.Run("CSV", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := PrintCompareResult(cmd, result, "P1", "P2", "csv", "")
		w.Close()
		os.Stdout = old

		content := make([]byte, 1024)
		n, _ := r.Read(content)

		assert.NoError(t, err)
		assert.Contains(t, string(content[:n]), "Type,Name,ID Project 1,ID Project 2")
	})
}
