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

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// AddFlag adds the --save boolean flag to a cobra command
func AddFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("save", false, "Save output to file in ~/.gotr/exports/")
}

// Output checks if --save flag is set and saves data to file if so.
// If --save is not set, outputs data to stdout as JSON.
// Returns the saved file path for user notification (empty string if output to stdout).
func Output(cmd *cobra.Command, data interface{}, resource string, format string) (string, error) {
	saveFlag, err := cmd.Flags().GetBool("save")
	if err != nil {
		return "", fmt.Errorf("error reading --save flag: %w", err)
	}

	if !saveFlag {
		// Output to stdout as JSON
		content, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", fmt.Errorf("error marshaling to JSON: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(content))
		return "", nil
	}

	return SaveToFile(data, resource, format)
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
