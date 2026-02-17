package output

import (
	"fmt"
	"time"
)

// GenerateTimestamp generates a formatted timestamp for filenames.
// Format: YYYY-MM-DD_HH-MM-SS
func GenerateTimestamp(t time.Time) string {
	return t.Format("2006-01-02_15-04-05")
}

// SanitizeResourceName sanitizes resource name for use in filenames.
// Replaces "all" with "all-resources" as per requirements.
func SanitizeResourceName(resource string) string {
	if resource == "all" {
		return "all-resources"
	}
	return resource
}

// BuildFilename builds a complete filename from components.
// Pattern: {resource}_{timestamp}.{format}
func BuildFilename(resource, timestamp, format string) string {
	sanitizedResource := SanitizeResourceName(resource)
	return fmt.Sprintf("%s_%s.%s", sanitizedResource, timestamp, format)
}
