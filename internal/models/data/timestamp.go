package data

import (
	"encoding/json"
	"fmt"
	"time"
)

// Timestamp is a custom type that can unmarshal Unix timestamps from JSON
// TestRail API returns timestamps as integers (Unix epoch), not as RFC3339 strings
type Timestamp struct {
	time.Time
}

// UnmarshalJSON implements custom JSON unmarshaling for Unix timestamps
// It handles both numeric timestamps (int/float) and RFC3339 strings
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as integer (Unix timestamp)
	var unixTimestamp int64
	if err := json.Unmarshal(data, &unixTimestamp); err == nil {
		t.Time = time.Unix(unixTimestamp, 0)
		return nil
	}

	// Try to unmarshal as float (some APIs return timestamps as floats)
	var unixTimestampFloat float64
	if err := json.Unmarshal(data, &unixTimestampFloat); err == nil {
		t.Time = time.Unix(int64(unixTimestampFloat), 0)
		return nil
	}

	// Try to unmarshal as RFC3339 string
	var timeStr string
	if err := json.Unmarshal(data, &timeStr); err == nil {
		parsed, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			// Try other common formats
			parsed, err = time.Parse("2006-01-02", timeStr)
			if err != nil {
				return fmt.Errorf("cannot parse timestamp: %w", err)
			}
		}
		t.Time = parsed
		return nil
	}

	// Handle null value
	if string(data) == "null" {
		t.Time = time.Time{}
		return nil
	}

	return fmt.Errorf("cannot unmarshal timestamp from: %s", string(data))
}

// MarshalJSON implements custom JSON marshaling
func (t Timestamp) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Unix())
}

// String returns the timestamp as a formatted string
func (t Timestamp) String() string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// IsValid returns true if the timestamp is set (not zero)
func (t Timestamp) IsValid() bool {
	return !t.IsZero()
}
