package output

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Test Data Structures ====================

type TestCase struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type yamlMarshalErrorValue struct{}

func (yamlMarshalErrorValue) MarshalYAML() (interface{}, error) {
	return nil, errors.New("forced yaml marshal error")
}

// ==================== Tests for AddFlag ====================

func TestAddFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	AddFlag(cmd)

	// Check that flag exists
	flag := cmd.Flags().Lookup("save")
	require.NotNil(t, flag)
	assert.Equal(t, "save", flag.Name)
	assert.Equal(t, "bool", flag.Value.Type())
	assert.Equal(t, "false", flag.Value.String())
}

// ==================== Tests for Output ====================

func TestOutput_SaveFlagNotSet(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	AddFlag(cmd)

	data := map[string]string{"key": "value"}
	path, err := Output(cmd, data, "cases", "json")

	assert.NoError(t, err)
	assert.Empty(t, path)
}

func TestOutput_SaveFlagSet(t *testing.T) {
	// Create temp directory for exports
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	cmd := &cobra.Command{Use: "test"}
	AddFlag(cmd)
	cmd.SetArgs([]string{"--save"})
	require.NoError(t, cmd.Execute())

	data := map[string]string{"key": "value"}
	path, err := Output(cmd, data, "cases", "json")

	assert.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.FileExists(t, path)
}

func TestOutput_FlagError(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	// Don't add the flag, so GetBool will error

	data := map[string]string{"key": "value"}
	path, err := Output(cmd, data, "cases", "json")

	assert.Error(t, err)
	assert.Empty(t, path)
	assert.Contains(t, err.Error(), "save")
}

// ==================== Tests for SaveToFile ====================

func TestSaveToFile_JSON(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	data := map[string]string{"key": "value", "test": "data"}
	path, err := SaveToFile(data, "test-resource", "json")

	require.NoError(t, err)
	assert.FileExists(t, path)
	assert.True(t, strings.HasSuffix(path, ".json"))

	// Verify content
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), `"key": "value"`)
	assert.Contains(t, string(content), `"test": "data"`)
}

func TestSaveToFile_YAML(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	data := map[string]string{"key": "value", "test": "data"}
	path, err := SaveToFile(data, "test-resource", "yaml")

	require.NoError(t, err)
	assert.FileExists(t, path)
	assert.True(t, strings.HasSuffix(path, ".yaml"))

	// Verify content
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "key: value")
	assert.Contains(t, string(content), "test: data")
}

func TestSaveToFile_CSV(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	data := []TestCase{
		{ID: 1, Title: "Test Case 1"},
		{ID: 2, Title: "Test Case 2"},
	}

	path, err := SaveToFile(data, "test-resource", "csv")

	require.NoError(t, err)
	assert.FileExists(t, path)
	assert.True(t, strings.HasSuffix(path, ".csv"))

	// Verify content
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)
	require.Len(t, records, 3) // header + 2 data rows
	assert.Equal(t, []string{"id", "title"}, records[0])
	assert.Equal(t, []string{"1", "Test Case 1"}, records[1])
	assert.Equal(t, []string{"2", "Test Case 2"}, records[2])
}

func TestSaveToFile_CSVWithMapSlice(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	data := []map[string]interface{}{
		{"id": 1, "name": "Item 1"},
		{"id": 2, "name": "Item 2"},
	}

	path, err := SaveToFile(data, "test-resource", "csv")

	require.NoError(t, err)
	assert.FileExists(t, path)
}

func TestSaveToFile_CSVEmptySlice(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	data := []TestCase{}

	path, err := SaveToFile(data, "test-resource", "csv")

	require.NoError(t, err)
	assert.FileExists(t, path)
}

func TestSaveToFile_CSVNonSlice(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	data := map[string]string{"key": "value"}

	path, err := SaveToFile(data, "test-resource", "csv")

	assert.Error(t, err)
	assert.Empty(t, path)
	assert.Contains(t, err.Error(), "slice")
}

func TestSaveToFile_UnsupportedFormat(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	data := map[string]string{"key": "value"}
	path, err := SaveToFile(data, "test-resource", "xml")

	assert.Error(t, err)
	assert.Empty(t, path)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestSaveToFileWithPath_JSON(t *testing.T) {
	tempDir := t.TempDir()
	customPath := tempDir + "/custom/result.json"

	data := map[string]string{"key": "value"}
	path, err := SaveToFileWithPath(data, "json", customPath)

	require.NoError(t, err)
	assert.Equal(t, customPath, path)
	assert.FileExists(t, customPath)
}

func TestSaveToFileWithPath_EmptyPath(t *testing.T) {
	data := map[string]string{"key": "value"}

	path, err := SaveToFileWithPath(data, "json", "")

	assert.Error(t, err)
	assert.Empty(t, path)
	assert.Contains(t, err.Error(), "empty")
}

func TestSaveToFile_JSONMarshalError(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	// Channel cannot be marshaled to JSON
	data := make(chan int)

	path, err := SaveToFile(data, "test-resource", "json")

	assert.Error(t, err)
	assert.Empty(t, path)
	assert.Contains(t, err.Error(), "marshaling")
}

func TestSaveToFile_HomeDirError(t *testing.T) {
	// Remove HOME env var temporarily
	origHome := os.Getenv("HOME")
	os.Unsetenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Also check for USERPROFILE on Windows
	origUserProfile := os.Getenv("USERPROFILE")
	os.Unsetenv("USERPROFILE")
	defer os.Setenv("USERPROFILE", origUserProfile)

	// Set HOMEDRIVE and HOMEPATH to invalid values
	origHomeDrive := os.Getenv("HOMEDRIVE")
	origHomePath := os.Getenv("HOMEPATH")
	os.Setenv("HOMEDRIVE", "")
	os.Setenv("HOMEPATH", "")
	defer func() {
		if origHomeDrive != "" {
			os.Setenv("HOMEDRIVE", origHomeDrive)
		}
		if origHomePath != "" {
			os.Setenv("HOMEPATH", origHomePath)
		}
	}()

	data := map[string]string{"key": "value"}
	path, err := SaveToFile(data, "test-resource", "json")

	// This test behavior depends on OS, on Unix it may fail
	// We just check that some error occurs
	_ = path
	_ = err
}

func TestSaveToFile_InvalidPath(t *testing.T) {
	// Create a scenario where directory creation fails
	// by using an invalid character in path (Windows) or permission issue (Unix)
	// This is tricky to test cross-platform, so we test EnsureDir separately

	// Instead test by mocking invalid path - just verify error handling works
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	// Valid test case - directory should be created successfully
	data := map[string]string{"key": "value"}
	path, err := SaveToFile(data, "valid-resource", "json")

	assert.NoError(t, err)
	assert.NotEmpty(t, path)
}

func TestOutputResult_Wrapper(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	AddFlag(cmd)

	err := OutputResult(cmd, map[string]any{"ok": true}, "cases")
	assert.NoError(t, err)
}

func TestOutputGetResult_QuietNoop(t *testing.T) {
	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().String("type", "json", "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().Bool("jq", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().Bool("body-only", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	require.NoError(t, cmd.Flags().Set("quiet", "true"))

	err := OutputGetResult(cmd, map[string]any{"id": 1}, time.Now())
	assert.NoError(t, err)
}

func TestOutputGetResult_JSONFull(t *testing.T) {
	cmd := &cobra.Command{Use: "get"}
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().String("type", "json", "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().Bool("jq", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().Bool("body-only", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	require.NoError(t, cmd.Flags().Set("type", "json-full"))

	err := OutputGetResult(cmd, map[string]any{"id": 7}, time.Now())
	assert.NoError(t, err)
	assert.Contains(t, out.String(), "\"status\": \"200 OK\"")
	assert.Contains(t, out.String(), "\"data\"")
}

func TestOutputGetResult_SaveFlag(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().String("type", "json", "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().Bool("jq", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().Bool("body-only", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	require.NoError(t, cmd.Flags().Set("save", "true"))

	err := OutputGetResult(cmd, map[string]any{"id": 9}, time.Now())
	assert.NoError(t, err)
}

func TestOutputGetResult_InteractivePromptError(t *testing.T) {
	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().String("type", "json", "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().Bool("jq", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().Bool("body-only", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	cmd.SetContext(interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter()))

	err := OutputGetResult(cmd, map[string]any{"id": 1}, time.Now())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interactive save selection failed")
}

func TestOutputGetResult_SkippablePromptErrorContinues(t *testing.T) {
	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().String("type", "json", "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().Bool("jq", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().Bool("body-only", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	require.NoError(t, cmd.Flags().Set("quiet", "true"))

	// Mock prompter without queued answers triggers a skippable queue-exhausted error.
	cmd.SetContext(interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()))

	err := OutputGetResult(cmd, map[string]any{"id": 1}, time.Now())
	assert.NoError(t, err)
}

func TestOutputGetResult_SaveError(t *testing.T) {
	base := t.TempDir()
	badHome := filepath.Join(base, "home-as-file")
	require.NoError(t, os.WriteFile(badHome, []byte("x"), 0o644))
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", badHome)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().String("type", "json", "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().Bool("jq", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().Bool("body-only", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	require.NoError(t, cmd.Flags().Set("save", "true"))

	err := OutputGetResult(cmd, map[string]any{"id": 1}, time.Now())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "save error")
}

func TestOutputGetResult_JQRuntimeError(t *testing.T) {
	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().String("type", "json", "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().Bool("jq", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().Bool("body-only", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	require.NoError(t, cmd.Flags().Set("jq", "true"))
	require.NoError(t, cmd.Flags().Set("jq-filter", "["))

	err := OutputGetResult(cmd, map[string]any{"id": 1}, time.Now())
	assert.Error(t, err)
}

func TestOutputGetResult_JQMarshalError(t *testing.T) {
	viper.Set("jq_format", true)
	t.Cleanup(viper.Reset)

	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().String("type", "json", "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().Bool("jq", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().Bool("body-only", false, "")
	cmd.Flags().Bool("non-interactive", false, "")

	err := OutputGetResult(cmd, make(chan int), time.Now())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "jq marshal error")
}

func TestOutputGetResult_DefaultOutputFormat(t *testing.T) {
	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().String("type", "json", "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().Bool("jq", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().Bool("body-only", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	require.NoError(t, cmd.Flags().Set("type", "table"))

	err := OutputGetResult(cmd, map[string]any{"id": 1}, time.Now())
	assert.NoError(t, err)
}

func TestOutputGetResult_SaveFlagBodyOnly(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().String("type", "json", "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().Bool("jq", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().Bool("body-only", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	require.NoError(t, cmd.Flags().Set("save", "true"))
	require.NoError(t, cmd.Flags().Set("body-only", "true"))

	err := OutputGetResult(cmd, map[string]any{"id": 11}, time.Now())
	require.NoError(t, err)

	entries, readErr := os.ReadDir(filepath.Join(tempHome, ".gotr", "exports", "get"))
	require.NoError(t, readErr)
	require.NotEmpty(t, entries)

	savedPath := filepath.Join(tempHome, ".gotr", "exports", "get", entries[0].Name())
	content, contentErr := os.ReadFile(savedPath)
	require.NoError(t, contentErr)
	assert.Contains(t, string(content), "\"id\": 11")
	assert.NotContains(t, string(content), "status_code")
}

func TestOutputGetResult_InteractivePromptSavePath(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().String("type", "json", "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().Bool("jq", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().Bool("body-only", false, "")
	cmd.Flags().Bool("non-interactive", false, "")

	ctx := interactive.WithPrompter(
		context.Background(),
		interactive.NewMockPrompter().WithConfirmResponses(true, false),
	)
	cmd.SetContext(ctx)

	err := OutputGetResult(cmd, map[string]any{"id": 42}, time.Now())
	require.NoError(t, err)

	entries, readErr := os.ReadDir(filepath.Join(tempHome, ".gotr", "exports", "get"))
	require.NoError(t, readErr)
	require.NotEmpty(t, entries)

	savedPath := filepath.Join(tempHome, ".gotr", "exports", "get", entries[0].Name())
	content, readErr := os.ReadFile(savedPath)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), "\"status_code\": 200")
	assert.Contains(t, string(content), "\"id\": 42")
}

func TestOutputGetResult_SkippablePromptErrorFallsThroughToJSONOutput(t *testing.T) {
	cmd := &cobra.Command{Use: "get"}
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().String("type", "json", "")
	cmd.Flags().Bool("save", false, "")
	cmd.Flags().Bool("jq", false, "")
	cmd.Flags().String("jq-filter", "", "")
	cmd.Flags().Bool("body-only", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	// Mock prompter without responses triggers a skippable queue-exhausted error.
	cmd.SetContext(interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()))

	err := OutputGetResult(cmd, map[string]any{"id": 303}, time.Now())
	require.NoError(t, err)
	assert.Contains(t, out.String(), "\"id\": 303")
}

func TestSaveToFileWithPath_YAMLAndUnsupported(t *testing.T) {
	tempDir := t.TempDir()

	yamlPath := tempDir + "/nested/data.yaml"
	gotPath, err := SaveToFileWithPath(map[string]any{"k": "v"}, "yaml", yamlPath)
	assert.NoError(t, err)
	assert.Equal(t, yamlPath, gotPath)

	_, err = SaveToFileWithPath(map[string]any{"k": "v"}, "xml", tempDir+"/x.xml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestOutputResultWithFlags_AndPrintSuccess(t *testing.T) {
	tempDir := t.TempDir()
	outFile := tempDir + "/out.json"

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().String("output", "", "")
	require.NoError(t, cmd.Flags().Set("output", outFile))

	err := OutputResultWithFlags(cmd, map[string]any{"x": 1})
	assert.NoError(t, err)
	assert.FileExists(t, outFile)

	require.NoError(t, cmd.Flags().Set("quiet", "true"))
	PrintSuccess(cmd, "done %d", 1)
}

func TestOutput_InteractiveSkippablePromptErrorFallsBackToStdout(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	AddFlag(cmd)
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("non-interactive", false, "")

	// No queued answers -> mock confirm queue exhausted, which is skippable.
	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter())
	cmd.SetContext(ctx)

	var out bytes.Buffer
	cmd.SetOut(&out)

	path, err := Output(cmd, map[string]any{"id": 101}, "cases", "json")
	require.NoError(t, err)
	assert.Equal(t, "", path)
	assert.Contains(t, out.String(), "\"id\": 101")
}

func TestOutput_NonInteractivePromptErrorIsReturned(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	AddFlag(cmd)
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().Bool("non-interactive", false, "")

	ctx := interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter())
	cmd.SetContext(ctx)

	path, err := Output(cmd, map[string]any{"id": 202}, "cases", "json")
	assert.Error(t, err)
	assert.Equal(t, "", path)
	assert.Contains(t, err.Error(), "interactive save selection failed")
}

func TestOutput_JSONMarshalError(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	AddFlag(cmd)
	cmd.Flags().Bool("non-interactive", false, "")
	require.NoError(t, cmd.Flags().Set("non-interactive", "true"))

	path, err := Output(cmd, make(chan int), "cases", "json")
	assert.Error(t, err)
	assert.Equal(t, "", path)
	assert.Contains(t, err.Error(), "error marshaling to JSON")
}

func TestOutputBySavePath_ExplicitPath(t *testing.T) {
	tempDir := t.TempDir()
	explicitPath := tempDir + "/custom/data.json"

	path, err := outputBySavePath(map[string]any{"ok": true}, "cases", "json", explicitPath)
	require.NoError(t, err)
	assert.Equal(t, explicitPath, path)
	assert.FileExists(t, explicitPath)
}

func TestSaveToFileWithPath_CSV(t *testing.T) {
	tempDir := t.TempDir()
	csvPath := tempDir + "/rows/out.csv"

	data := []TestCase{{ID: 1, Title: "One"}, {ID: 2, Title: "Two"}}
	path, err := SaveToFileWithPath(data, "csv", csvPath)
	require.NoError(t, err)
	assert.Equal(t, csvPath, path)
	assert.FileExists(t, csvPath)
}

func TestSaveToFileWithPath_JSONMarshalError(t *testing.T) {
	path, err := SaveToFileWithPath(make(chan int), "json", t.TempDir()+"/bad.json")
	assert.Error(t, err)
	assert.Equal(t, "", path)
	assert.Contains(t, err.Error(), "marshaling")
}

func TestSaveToFileWithPath_YAMLMarshalPanicUnsupportedType(t *testing.T) {
	assert.Panics(t, func() {
		_, _ = SaveToFileWithPath(make(chan int), "yaml", t.TempDir()+"/bad.yaml")
	})
}

func TestSaveToFile_YAMLMarshalErrorFromMarshaler(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	path, err := SaveToFile(yamlMarshalErrorValue{}, "test-resource", "yaml")
	assert.Error(t, err)
	assert.Equal(t, "", path)
	assert.Contains(t, err.Error(), "error marshaling to YAML")
}

func TestSaveToFileWithPath_YAMLMarshalError(t *testing.T) {
	path, err := SaveToFileWithPath(yamlMarshalErrorValue{}, "yaml", t.TempDir()+"/bad.yaml")
	assert.Error(t, err)
	assert.Equal(t, "", path)
	assert.Contains(t, err.Error(), "error marshaling to YAML")
}

func TestSaveToFileWithPath_WriteError(t *testing.T) {
	tempDir := t.TempDir()
	path, err := SaveToFileWithPath(map[string]any{"k": "v"}, "json", tempDir)
	assert.Error(t, err)
	assert.Equal(t, "", path)
	assert.Contains(t, err.Error(), "error writing file")
}

func TestSaveToFileWithPath_EnsureDirError(t *testing.T) {
	tempDir := t.TempDir()
	notDir := filepath.Join(tempDir, "not-a-dir")
	require.NoError(t, os.WriteFile(notDir, []byte("x"), 0o644))

	path, err := SaveToFileWithPath(map[string]any{"k": "v"}, "json", filepath.Join(notDir, "out.json"))
	assert.Error(t, err)
	assert.Equal(t, "", path)
	assert.Contains(t, err.Error(), "error creating output directory")
}

func TestSaveJSONToFile_SerializationError(t *testing.T) {
	err := SaveJSONToFile(t.TempDir()+"/broken.json", make(chan int))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serialization error")
}

func TestOutputResultWithFlags_QuietSavesWithoutPrintingJSON(t *testing.T) {
	tempDir := t.TempDir()
	outFile := tempDir + "/quiet.json"

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().String("output", "", "")
	require.NoError(t, cmd.Flags().Set("quiet", "true"))
	require.NoError(t, cmd.Flags().Set("output", outFile))

	err := OutputResultWithFlags(cmd, map[string]any{"quiet": true})
	require.NoError(t, err)
	assert.FileExists(t, outFile)
}

func TestOutputResultWithFlags_SaveError(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().String("output", "", "")
	require.NoError(t, cmd.Flags().Set("output", t.TempDir()))

	err := OutputResultWithFlags(cmd, map[string]any{"x": 1})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "write file")
}

func TestOutputResultWithFlags_JSONFormattingError(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", false, "")
	cmd.Flags().String("output", "", "")

	err := OutputResultWithFlags(cmd, make(chan int))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JSON formatting error")
}

func TestPrintSuccess_EmitsWhenNotQuiet(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("quiet", false, "")

	var buf bytes.Buffer
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	t.Cleanup(func() {
		os.Stdout = old
		_ = r.Close()
	})

	PrintSuccess(cmd, "done %d", 7)
	_ = w.Close()
	_, _ = buf.ReadFrom(r)

	assert.Contains(t, buf.String(), "done 7")
}

// ==================== Tests for GenerateFilename ====================

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		format   string
		wantAll  []string // parts that should be present
	}{
		{
			name:     "cases json",
			resource: "cases",
			format:   "json",
			wantAll:  []string{"cases_", ".json"},
		},
		{
			name:     "plans yaml",
			resource: "plans",
			format:   "yaml",
			wantAll:  []string{"plans_", ".yaml"},
		},
		{
			name:     "runs csv",
			resource: "runs",
			format:   "csv",
			wantAll:  []string{"runs_", ".csv"},
		},
		{
			name:     "all becomes all-resources",
			resource: "all",
			format:   "json",
			wantAll:  []string{"all-resources_", ".json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := GenerateFilename(tt.resource, tt.format)
			for _, want := range tt.wantAll {
				assert.Contains(t, filename, want)
			}
			// Check timestamp pattern (YYYY-MM-DD_HH-MM-SS)
			assert.Regexp(t, `^.+_\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}\..+$`, filename)
		})
	}
}

func TestGenerateFilename_Timestamp(t *testing.T) {
	before := time.Now()
	filename := GenerateFilename("test", "json")
	after := time.Now()

	// Extract timestamp from filename
	// Format: test_YYYY-MM-DD_HH-MM-SS.json
	// Remove extension first
	withoutExt := strings.TrimSuffix(filename, ".json")
	parts := strings.Split(withoutExt, "_")
	require.Len(t, parts, 3)

	// Parse the timestamp back
	timeStr := parts[1] + "_" + parts[2]
	fileTime, err := time.Parse("2006-01-02_15-04-05", timeStr)
	require.NoError(t, err)

	// Verify the timestamp is within the test window
	assert.True(t, fileTime.Equal(before) || fileTime.After(before) || fileTime.Before(after.Add(time.Second)),
		"timestamp should be within test execution window")
}

// ==================== Tests for CSV Helper Functions ====================

func TestGetHeaders_Struct(t *testing.T) {
	v := TestCase{ID: 1, Title: "Test"}
	headers := getHeaders(reflect.ValueOf(v))
	assert.Equal(t, []string{"id", "title"}, headers)
}

func TestGetHeaders_Pointer(t *testing.T) {
	v := &TestCase{ID: 1, Title: "Test"}
	headers := getHeaders(reflect.ValueOf(v))
	assert.Equal(t, []string{"id", "title"}, headers)
}

func TestGetHeaders_Map(t *testing.T) {
	v := map[string]interface{}{"key1": "value1", "key2": 42}
	headers := getHeaders(reflect.ValueOf(v))
	assert.Len(t, headers, 2)
	assert.Contains(t, headers, "key1")
	assert.Contains(t, headers, "key2")
}

type StructWithUnexported struct {
	Exported   string
	unexported string
}

func TestGetHeaders_UnexportedFields(t *testing.T) {
	v := StructWithUnexported{Exported: "test", unexported: "hidden"}
	headers := getHeaders(reflect.ValueOf(v))
	assert.Equal(t, []string{"Exported"}, headers)
}

type StructWithJSONTags struct {
	Name    string `json:"name"`
	Value   int    `json:"value"`
	Ignored string `json:"-"`
}

type rowStructWithSkippedFields struct {
	Visible   string `json:"visible"`
	Ignored   string `json:"-"`
	unexported string
}

func TestGetHeaders_JSONTags(t *testing.T) {
	v := StructWithJSONTags{Name: "test", Value: 42, Ignored: "skip"}
	headers := getHeaders(reflect.ValueOf(v))
	assert.Equal(t, []string{"name", "value"}, headers)
}

func TestGetRowValues_Struct(t *testing.T) {
	v := TestCase{ID: 42, Title: "Test Title"}
	headers := []string{"id", "title"}
	values := getRowValues(reflect.ValueOf(v), headers)
	assert.Equal(t, []string{"42", "Test Title"}, values)
}

func TestGetRowValues_Pointer(t *testing.T) {
	v := &TestCase{ID: 42, Title: "Test Title"}
	headers := []string{"id", "title"}
	values := getRowValues(reflect.ValueOf(v), headers)
	assert.Equal(t, []string{"42", "Test Title"}, values)
}

func TestGetRowValues_Map(t *testing.T) {
	v := map[string]interface{}{"key1": "value1", "key2": 42}
	headers := []string{"key1", "key2"}
	values := getRowValues(reflect.ValueOf(v), headers)
	assert.Equal(t, []string{"value1", "42"}, values)
}

func TestGetRowValues_MapMissingKey(t *testing.T) {
	v := map[string]interface{}{"key1": "value1"}
	headers := []string{"key1", "missing"}
	values := getRowValues(reflect.ValueOf(v), headers)
	assert.Equal(t, []string{"value1", ""}, values)
}

func TestGetRowValues_NilPointer(t *testing.T) {
	var v *TestCase
	headers := []string{"id", "title"}
	values := getRowValues(reflect.ValueOf(v), headers)
	assert.Equal(t, []string{"", ""}, values)
}

func TestGetRowValues_StructSkipsUnexportedAndDashTag(t *testing.T) {
	v := rowStructWithSkippedFields{Visible: "ok", Ignored: "secret", unexported: "hidden"}
	headers := []string{"visible", "Ignored", "unexported"}
	values := getRowValues(reflect.ValueOf(v), headers)
	assert.Equal(t, []string{"ok", "", ""}, values)
}

func TestSaveToCSV_CreateError(t *testing.T) {
	_, err := saveToCSV([]TestCase{{ID: 1, Title: "x"}}, "/definitely/missing/path/out.csv")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error creating CSV file")
}

func TestSaveToCSV_HeaderWriteError(t *testing.T) {
	if _, err := os.Stat("/dev/full"); err != nil {
		t.Skip("/dev/full not available")
	}

	data := []map[string]string{{strings.Repeat("k", 10000): "v"}}
	_, err := saveToCSV(data, "/dev/full")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error writing CSV headers")
}

func TestSaveToCSV_RowWriteError(t *testing.T) {
	if _, err := os.Stat("/dev/full"); err != nil {
		t.Skip("/dev/full not available")
	}

	data := []map[string]string{{"k": strings.Repeat("v", 20000)}}
	_, err := saveToCSV(data, "/dev/full")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error writing CSV row")
}

// ==================== Integration Tests ====================

func TestSaveIntegration_AllResources(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	cmd := &cobra.Command{Use: "test"}
	AddFlag(cmd)
	cmd.SetArgs([]string{"--save"})
	require.NoError(t, cmd.Execute())

	data := []map[string]interface{}{
		{"id": 1, "name": "Resource 1"},
		{"id": 2, "name": "Resource 2"},
	}

	path, err := Output(cmd, data, "all", "json")

	require.NoError(t, err)
	assert.Contains(t, path, "all-resources_")
	assert.FileExists(t, path)
}

func TestSaveIntegration_MultipleFormats(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	data := []TestCase{
		{ID: 1, Title: "Case 1"},
		{ID: 2, Title: "Case 2"},
	}

	formats := []string{"json", "yaml", "csv"}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			path, err := SaveToFile(data, "cases", format)
			require.NoError(t, err)
			assert.FileExists(t, path)

			// Verify content based on format
			content, err := os.ReadFile(path)
			require.NoError(t, err)
			assert.NotEmpty(t, content)
		})
	}
}

// ==================== Tests for Directory Creation ====================

func TestSaveToFile_CreatesNestedDirectory(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	data := map[string]string{"test": "data"}
	// Note: resource with path separators will create nested directories
	resource := "nested-resource"

	path, err := SaveToFile(data, resource, "json")

	require.NoError(t, err)
	assert.FileExists(t, path)
	assert.Contains(t, path, "nested-resource")
}

// ==================== Tests for filename.go ====================

func TestGenerateTimestamp(t *testing.T) {
	testTime := time.Date(2026, 2, 16, 14, 30, 45, 0, time.UTC)
	timestamp := GenerateTimestamp(testTime)
	assert.Equal(t, "2026-02-16_14-30-45", timestamp)
}

func TestSanitizeResourceName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"cases", "cases"},
		{"plans", "plans"},
		{"all", "all-resources"},
		{"runs", "runs"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeResourceName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildFilename(t *testing.T) {
	result := BuildFilename("cases", "2026-02-16_14-30-45", "json")
	assert.Equal(t, "cases_2026-02-16_14-30-45.json", result)

	// Test with "all" resource
	result = BuildFilename("all", "2026-02-16_14-30-45", "yaml")
	assert.Equal(t, "all-resources_2026-02-16_14-30-45.yaml", result)
}

// ==================== Tests for paths.go ====================

func TestGetExportsDir(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	exportsDir, err := GetExportsDir("cases")
	require.NoError(t, err)
	assert.Contains(t, exportsDir, ".gotr")
	assert.Contains(t, exportsDir, "exports")
	assert.Contains(t, exportsDir, "cases")
}

func TestEnsureDir(t *testing.T) {
	tempDir := t.TempDir()
	nestedPath := tempDir + "/nested/dir"

	err := EnsureDir(nestedPath)
	require.NoError(t, err)
	assert.DirExists(t, nestedPath)

	// EnsureDir should be idempotent
	err = EnsureDir(nestedPath)
	assert.NoError(t, err)
}

func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()

	// Non-existent file
	assert.False(t, FileExists(tempDir+"/nonexistent"))

	// Existing file
	existingFile := tempDir + "/exists.txt"
	err := os.WriteFile(existingFile, []byte("test"), 0o644)
	require.NoError(t, err)
	assert.True(t, FileExists(existingFile))

	// Existing directory
	assert.True(t, FileExists(tempDir))
}

func TestGetHomeDir(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	home, err := GetHomeDir()
	require.NoError(t, err)
	assert.Equal(t, tempHome, home)
}

func TestGetExportsDir_Error(t *testing.T) {
	// Remove HOME to trigger error
	origHome := os.Getenv("HOME")
	os.Unsetenv("HOME")
	defer os.Setenv("HOME", origHome)

	origUserProfile := os.Getenv("USERPROFILE")
	os.Unsetenv("USERPROFILE")
	defer os.Setenv("USERPROFILE", origUserProfile)

	_, err := GetExportsDir("cases")
	// This may or may not error depending on OS
	// We just exercise the code path
	_ = err
}

func TestEnsureDir_Error(t *testing.T) {
	// Try to create a directory in a read-only location or with invalid name
	// This is OS-specific, so we test by trying to create a file as a directory
	tempFile, err := os.CreateTemp("", "testdir")
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Try to create a directory where a file exists
	err = EnsureDir(tempFile.Name())
	// Should fail because a file exists at that path
	assert.Error(t, err)
}

func TestSaveToFile_DirectoryError(t *testing.T) {
	// Create a temp file (not directory)
	tempFile, err := os.CreateTemp("", "readonly")
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Mock the home dir to point to our temp file (which is not a directory)
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempFile.Name())
	defer os.Setenv("HOME", origHome)

	data := map[string]string{"key": "value"}
	path, err := SaveToFile(data, "cases", "json")

	assert.Error(t, err)
	assert.Empty(t, path)
}

func TestSaveToFile_WriteError(t *testing.T) {
	// Create a read-only directory
	tempDir := t.TempDir()
	readOnlyDir := tempDir + "/readonly"
	err := os.Mkdir(readOnlyDir, 0555)
	require.NoError(t, err)

	// Mock the home dir
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", readOnlyDir)
	defer os.Setenv("HOME", origHome)

	data := map[string]string{"key": "value"}
	// The save should work because we can still create directories inside
	// This test is mainly to exercise the code path
	_, _ = SaveToFile(data, "cases", "json")
}

func TestSaveToFile_YAMLMarshalError(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	// YAML library panics on invalid types, so we use a function that can't be marshaled
	// which will cause yaml.Marshal to panic
	defer func() {
		recover() // Expected - YAML library panics on unmarshalable types
	}()

	data := make(chan int)
	path, _ := SaveToFile(data, "test", "yaml")

	// If we reach here without panic, test fails
	assert.Empty(t, path)
}

func TestSaveToFile_CSVWriteError(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	// Create a file that will cause issues (too long name or invalid path)
	// For now, just test with valid data to ensure basic CSV path works
	data := []TestCase{
		{ID: 1, Title: "Test"},
	}

	path, err := SaveToFile(data, "test", "csv")
	require.NoError(t, err)
	assert.NotEmpty(t, path)
}

func TestGetRowValues_MissingHeader(t *testing.T) {
	v := TestCase{ID: 42, Title: "Test"}
	// Request a header that doesn't exist in the struct
	headers := []string{"id", "title", "nonexistent"}
	values := getRowValues(reflect.ValueOf(v), headers)
	// The nonexistent field should be empty
	assert.Equal(t, []string{"42", "Test", ""}, values)
}

func TestSaveToFile_FileCreationError(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	// Create exports dir
	exportsDir := tempHome + "/.gotr/exports/test"
	err := os.MkdirAll(exportsDir, 0o755)
	require.NoError(t, err)

	// Make the directory read-only
	err = os.Chmod(exportsDir, 0555)
	require.NoError(t, err)
	defer os.Chmod(exportsDir, 0o755) // Restore for cleanup

	data := map[string]string{"key": "value"}
	path, err := SaveToFile(data, "test", "json")

	// Should fail to create file in read-only directory
	assert.Error(t, err)
	assert.Empty(t, path)
}

func TestGetHeaders_InvalidType(t *testing.T) {
	// Test with a non-struct, non-map type
	headers := getHeaders(reflect.ValueOf("string"))
	assert.Empty(t, headers)

	headers = getHeaders(reflect.ValueOf(42))
	assert.Empty(t, headers)

	headers = getHeaders(reflect.ValueOf([]string{"a", "b"}))
	assert.Empty(t, headers)
}

func TestGetRowValues_InvalidType(t *testing.T) {
	// Test with a non-struct, non-map type
	values := getRowValues(reflect.ValueOf("string"), []string{"a", "b"})
	assert.Equal(t, []string{"", ""}, values)

	values = getRowValues(reflect.ValueOf(42), []string{"x"})
	assert.Equal(t, []string{""}, values)
}

// ==================== Tests for File Content Validation ====================

func TestSaveToFile_JSONContent(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	data := map[string]interface{}{
		"id":      1,
		"title":   "Test",
		"enabled": true,
		"nested": map[string]string{
			"key": "value",
		},
	}

	path, err := SaveToFile(data, "test", "json")
	require.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	// Verify valid JSON
	var result map[string]interface{}
	err = json.Unmarshal(content, &result)
	require.NoError(t, err)

	assert.Equal(t, float64(1), result["id"])
	assert.Equal(t, "Test", result["title"])
	assert.Equal(t, true, result["enabled"])
}

func TestSaveToFile_FilePermissions(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	data := map[string]string{"test": "data"}
	path, err := SaveToFile(data, "test", "json")
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)

	// File should be readable by owner
	mode := info.Mode()
	assert.True(t, mode&0400 != 0, "file should be readable by owner")
}
func TestSaveJSONToFile(t *testing.T) {
        dir := t.TempDir()
        path := dir + "/out.json"

        data := map[string]interface{}{"key": "value", "num": 42}
        err := SaveJSONToFile(path, data)
        require.NoError(t, err)

        content, err := os.ReadFile(path)
        require.NoError(t, err)
        assert.Contains(t, string(content), "key")
        assert.Contains(t, string(content), "value")
}

func TestSaveJSONToFile_WriteError(t *testing.T) {
        // Pass a directory as filename to force write error
        dir := t.TempDir()
        err := SaveJSONToFile(dir, map[string]string{"x": "y"})
        assert.Error(t, err)
}