// Package output provides utilities for output formatting and saving.
package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	embed "github.com/Korrnals/gotr/embedded"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// AddFlag adds the --save boolean flag to a cobra command
func AddFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("save", false, "Save output to file in ~/.gotr/exports/")
}

// OutputResult is a convenience wrapper for Output with format="json".
// It discards the file path and returns only the error.
func OutputResult(cmd *cobra.Command, data interface{}, resource string) error {
	_, err := Output(cmd, data, resource, "json")
	return err
}

// OutputGetResult handles get-command output, including save/json-full/jq modes.
func OutputGetResult(cmd *cobra.Command, data any, start time.Time) error {
	quiet, _ := cmd.Flags().GetBool("quiet")
	outputFormat, _ := cmd.Flags().GetString("type")
	saveFlag, _ := cmd.Flags().GetBool("save")
	jqEnabled, _ := cmd.Flags().GetBool("jq")
	jqFilter, _ := cmd.Flags().GetString("jq-filter")
	bodyOnly, _ := cmd.Flags().GetBool("body-only")

	if !jqEnabled {
		jqEnabled = viper.GetBool("jq_format")
	}

	savePath := ""
	if saveFlag {
		savePath = defaultSavePathMarker
	} else if ShouldPromptForInteractiveSave(cmd) {
		p := interactive.PrompterFromContext(cmd.Context())
		promptedPath, err := PromptSavePathWithOptions(p, "response", false)
		if err != nil {
			if !isSkippableInteractiveSavePromptError(err) {
				return err
			}
			promptedPath = ""
		}
		savePath = promptedPath
	}

	if savePath != "" {
		toSave := data
		if !bodyOnly {
			toSave = struct {
				Status     string        `json:"status"`
				StatusCode int           `json:"status_code"`
				Duration   time.Duration `json:"duration"`
				Timestamp  time.Time     `json:"timestamp"`
				Data       any           `json:"data"`
			}{
				Status:     "200 OK",
				StatusCode: 200,
				Duration:   time.Since(start),
				Timestamp:  time.Now(),
				Data:       data,
			}
		}

		filepath, err := outputBySavePath(toSave, "get", "json", savePath)
		if err != nil {
			return fmt.Errorf("save error: %w", err)
		}
		if !quiet && filepath != "" {
			ui.Infof(os.Stdout, "Response saved to %s", filepath)
		}
		return nil
	}

	if jqEnabled || jqFilter != "" {
		payload, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("jq marshal error: %w", err)
		}
		if err := embed.RunEmbeddedJQ(payload, jqFilter); err != nil {
			return err
		}
		return nil
	}

	if quiet {
		return nil
	}

	switch outputFormat {
	case "json":
		return ui.JSON(cmd, data)
	case "json-full":
		full := struct {
			Status     string        `json:"status"`
			StatusCode int           `json:"status_code"`
			Duration   time.Duration `json:"duration"`
			Timestamp  time.Time     `json:"timestamp"`
			Data       any           `json:"data"`
		}{
			Status:     "200 OK",
			StatusCode: 200,
			Duration:   time.Since(start),
			Timestamp:  time.Now(),
			Data:       data,
		}
		return ui.JSON(cmd, full)
	default:
		ui.Warning(os.Stdout, "Table output not implemented yet")
		return nil
	}
}

// Output checks if --save flag is set and saves data to file if so.
// If --save is not set, outputs data to stdout as JSON.
// Returns the saved file path for user notification (empty string if output to stdout).
func Output(cmd *cobra.Command, data interface{}, resource string, format string) (string, error) {
	saveFlag, err := cmd.Flags().GetBool("save")
	if err != nil {
		return "", fmt.Errorf("error reading --save flag: %w", err)
	}

	savePath := ""
	if saveFlag {
		savePath = defaultSavePathMarker
	} else if ShouldPromptForInteractiveSave(cmd) {
		p := interactive.PrompterFromContext(cmd.Context())
		promptedPath, err := PromptSavePathWithOptions(p, resource+" result", false)
		if err != nil {
			if !isSkippableInteractiveSavePromptError(err) {
				return "", err
			}
			promptedPath = ""
		}
		savePath = promptedPath
	}

	if savePath != "" {
		return outputBySavePath(data, resource, format, savePath)
	}

	// Output to stdout as JSON
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling to JSON: %w", err)
	}
 	fmt.Fprintln(cmd.OutOrStdout(), string(content))
	return "", nil
}

func outputBySavePath(data interface{}, resource string, format string, savePath string) (string, error) {
	if savePath == defaultSavePathMarker {
		return SaveToFile(data, resource, format)
	}

	return SaveToFileWithPath(data, format, savePath)
}

// SaveToFile saves data to a file in the exports directory.
// Returns the full path of the saved file.
func SaveToFile(data interface{}, resource string, format string) (string, error) {
	// Generate filename
	filename := GenerateFilename(resource, format)

	// Get exports directory for this resource
	exportsDir, err := GetExportsDir(resource)
	if err != nil {
		return "", fmt.Errorf("error getting exports directory: %w", err)
	}

	// Ensure directory exists
	if err := EnsureDir(exportsDir); err != nil {
		return "", fmt.Errorf("error creating exports directory: %w", err)
	}

	// Full file path
	filePath := filepath.Join(exportsDir, filename)

	// Marshal data based on format
	var content []byte
	switch format {
	case "json":
		content, err = json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", fmt.Errorf("error marshaling to JSON: %w", err)
		}
	case "yaml":
		content, err = yaml.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("error marshaling to YAML: %w", err)
		}
	case "csv":
		return saveToCSV(data, filePath)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}

	// Write file
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return "", fmt.Errorf("error writing file: %w", err)
	}

	return filePath, nil
}

// SaveToFileWithPath saves data to an explicit file path.
// Returns the full path of the saved file.
func SaveToFileWithPath(data interface{}, format, filePath string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("file path is empty")
	}

	if err := EnsureDir(filepath.Dir(filePath)); err != nil {
		return "", fmt.Errorf("error creating output directory: %w", err)
	}

	var content []byte
	var err error
	switch format {
	case "json":
		content, err = json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", fmt.Errorf("error marshaling to JSON: %w", err)
		}
	case "yaml":
		content, err = yaml.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("error marshaling to YAML: %w", err)
		}
	case "csv":
		return saveToCSV(data, filePath)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return "", fmt.Errorf("error writing file: %w", err)
	}

	return filePath, nil
}

// saveToCSV saves data to a CSV file
func saveToCSV(data interface{}, filePath string) (string, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("error creating CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Handle slice of structs/maps
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		return "", fmt.Errorf("CSV export requires a slice, got %s", v.Kind())
	}

	if v.Len() == 0 {
		return filePath, nil
	}

	// Get headers from first element
	firstElem := v.Index(0)
	headers := getHeaders(firstElem)

	// Write headers
	if err := writer.Write(headers); err != nil {
		return "", fmt.Errorf("error writing CSV headers: %w", err)
	}

	// Write data rows
	for i := 0; i < v.Len(); i++ {
		row := getRowValues(v.Index(i), headers)
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("error writing CSV row: %w", err)
		}
	}

	return filePath, nil
}

// getHeaders extracts header names from a struct or map
func getHeaders(v reflect.Value) []string {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var headers []string

	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			// Skip unexported fields
			if field.PkgPath != "" {
				continue
			}
			// Use json tag if available, skip if "-"
			name := field.Name
			if tag := field.Tag.Get("json"); tag != "" {
				if tag == "-" {
					continue
				}
				name = tag
			}
			headers = append(headers, name)
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			headers = append(headers, fmt.Sprintf("%v", key.Interface()))
		}
	}

	return headers
}

// getRowValues extracts values from a struct or map matching the headers
func getRowValues(v reflect.Value, headers []string) []string {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	values := make([]string, len(headers))

	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		fieldMap := make(map[string]string)
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" {
				continue
			}
			name := field.Name
			if tag := field.Tag.Get("json"); tag != "" {
				if tag == "-" {
					continue
				}
				name = tag
			}
			fieldMap[name] = fmt.Sprintf("%v", v.Field(i).Interface())
		}
		for i, h := range headers {
			values[i] = fieldMap[h]
		}
	case reflect.Map:
		for i, h := range headers {
			val := v.MapIndex(reflect.ValueOf(h))
			if val.IsValid() {
				values[i] = fmt.Sprintf("%v", val.Interface())
			}
		}
	}

	return values
}

// GenerateFilename generates a filename with pattern: {resource}_YYYY-MM-DD_HH-MM-SS.{format}
// For resource "all", uses "all-resources" as prefix
func GenerateFilename(resource string, format string) string {
	// Handle special case for "all" resource
	if resource == "all" {
		resource = "all-resources"
	}

	// Format timestamp: YYYY-MM-DD_HH-MM-SS
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	return fmt.Sprintf("%s_%s.%s", resource, timestamp, format)
}

// SaveJSONToFile writes JSON with indentation to a specific file path.
func SaveJSONToFile(filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("serialization error: %w", err)
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// OutputResultWithFlags prints JSON and supports root-level --output/--quiet flags.
func OutputResultWithFlags(cmd *cobra.Command, data interface{}) error {
	quiet, _ := cmd.Flags().GetBool("quiet")
	outputPath, _ := cmd.Flags().GetString("output")

	if outputPath != "" {
		if err := SaveJSONToFile(outputPath, data); err != nil {
			return err
		}
		if !quiet {
			ui.Infof(os.Stdout, "Response saved to %s", outputPath)
		}
	}

	if !quiet {
		pretty, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("JSON formatting error: %w", err)
		}
		fmt.Println(string(pretty))
	}

	return nil
}

// PrintSuccess prints success message unless --quiet is enabled.
func PrintSuccess(cmd *cobra.Command, format string, args ...interface{}) {
	quiet, _ := cmd.Flags().GetBool("quiet")
	if !quiet {
		ui.Successf(os.Stdout, format, args...)
	}
}
