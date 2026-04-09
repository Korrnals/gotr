package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ResponseData holds a generic response for requests where the structure is unknown.
type ResponseData struct {
	Status     string              `json:"status"`
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       interface{}         `json:"body"`               // parsed JSON
	RawBody    []byte              `json:"raw_body,omitempty"` // raw bytes (optional in JSON output)
	Timestamp  time.Time           `json:"timestamp"`
	Duration   time.Duration       `json:"duration"`
}

// maxResponseBodySize is the upper limit for reading HTTP response bodies (50 MB).
const maxResponseBodySize = 50 * 1024 * 1024

// ReadResponse reads any HTTP response into a generic ResponseData struct.
func (c *HTTPClient) ReadResponse(ctx context.Context, resp *http.Response, duration time.Duration, outputFormat string) (ResponseData, error) {
	// Read the raw response body
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
	if err != nil {
		return ResponseData{}, err
	}
	// Parse raw bytes as JSON for storage in ResponseData
	var bodyData interface{}
	if err := json.Unmarshal(bodyBytes, &bodyData); err != nil {
		bodyData = string(bodyBytes)
	}
	// Map response data into the ResponseData struct
	data := ResponseData{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       bodyData,
		RawBody:    bodyBytes,
		Timestamp:  time.Now(),
		Duration:   duration,
	}
	// Return the fully populated ResponseData
	return data, nil
}

// ReadJSONResponse decodes an HTTP response body into the given typed target.
func (c *HTTPClient) ReadJSONResponse(ctx context.Context, resp *http.Response, target any) error {
	if resp == nil {
		return fmt.Errorf("nil response")
	}
	if resp.Body == nil {
		return fmt.Errorf("nil response body")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
		if err != nil {
			return fmt.Errorf("API error: %s, failed to read error body: %w", resp.Status, err)
		}
		return fmt.Errorf("API error: %s, body: %s", resp.Status, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode error: %w", err)
	}
	return nil
}

// PrintResponseFromData prints a pre-built ResponseData (without re-reading resp.Body).
func (c *HTTPClient) PrintResponseFromData(ctx context.Context, data ResponseData, outputFormat string) {
	switch outputFormat {
	case "json":
		pretty, err := json.MarshalIndent(data.Body, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal response body: %v\n", err)
			return
		}
		fmt.Println(string(pretty))
	case "json-full":
		pretty, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal response data: %v\n", err)
			return
		}
		fmt.Println(string(pretty))
	default: // table
		printTable(data)
	}
}

// SaveResponseToFile saves a generic ResponseData to a file.
func (c *HTTPClient) SaveResponseToFile(ctx context.Context, data ResponseData, filename, outputFormat string) error {
	var (
		toSave []byte
		err    error
	)
	switch outputFormat {
	case "json":
		toSave, err = json.MarshalIndent(data.Body, "", "  ")
	case "json-full":
		toSave, err = json.MarshalIndent(data, "", "  ")
	default: // table
		// For table format, fall back to json-full
		toSave, err = json.MarshalIndent(data, "", "  ")
	}
	if err != nil {
		return fmt.Errorf("failed to marshal response data: %w", err)
	}

	if err := os.WriteFile(filename, toSave, 0o644); err != nil {
		return err
	}
	fmt.Printf("Response saved to %s (format: %s)\n", filename, outputFormat)

	return nil
}

// Private helpers //
// printTable formats a response as a human-readable table.
func printTable(data ResponseData) {
	fmt.Printf("Status: %s (%d)\n", data.Status, data.StatusCode)
	fmt.Printf("Duration: %v\n", data.Duration)
	fmt.Printf("Timestamp: %s\n", data.Timestamp.Format(time.RFC3339))
	fmt.Printf("\nHeaders:\n")
	for k, v := range data.Headers {
		for _, val := range v {
			fmt.Printf("  %s: %s\n", k, val)
		}
	}
	fmt.Printf("\nBody:\n")
	jsonBody, _ := json.MarshalIndent(data.Body, "", "  ")
	fmt.Println(string(jsonBody))
}
