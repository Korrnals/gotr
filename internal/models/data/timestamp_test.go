package data

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimestamp_String_zero(t *testing.T) {
	var ts Timestamp
	assert.Equal(t, "", ts.String())
}

func TestTimestamp_String_nonzero(t *testing.T) {
	ts := Timestamp{Time: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)}
	s := ts.String()
	assert.Contains(t, s, "2024-01-15")
}

func TestTimestamp_IsValid_zero(t *testing.T) {
	var ts Timestamp
	assert.False(t, ts.IsValid())
}

func TestTimestamp_IsValid_nonzero(t *testing.T) {
	ts := Timestamp{Time: time.Now()}
	assert.True(t, ts.IsValid())
}

func TestTimestamp_UnmarshalJSON_int(t *testing.T) {
	data := []byte("1705315200")
	var ts Timestamp
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.True(t, ts.IsValid())
}

func TestTimestamp_UnmarshalJSON_null(t *testing.T) {
	// JSON null unmarshal into int64/float64 returns 0 (no error),
	// so the result is time.Unix(0,0) — which is NOT zero in go-time terms.
	// This test simply documents the actual behavior.
	data := []byte("null")
	var ts Timestamp
	require.NoError(t, json.Unmarshal(data, &ts))
	// 0-value unix epoch is not a zero time.Time{}
	assert.False(t, ts.Time.IsZero())
}

func TestTimestamp_UnmarshalJSON_rfc3339(t *testing.T) {
	data := []byte(`"2024-01-15T12:00:00Z"`)
	var ts Timestamp
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.True(t, ts.IsValid())
}

func TestTimestamp_UnmarshalJSON_date(t *testing.T) {
	data := []byte(`"2024-01-15"`)
	var ts Timestamp
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.True(t, ts.IsValid())
}

func TestTimestamp_UnmarshalJSON_float(t *testing.T) {
	data := []byte("1705315200.9")
	var ts Timestamp
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, int64(1705315200), ts.Unix())
}

func TestTimestamp_UnmarshalJSON_emptyString_ReturnsError(t *testing.T) {
	data := []byte(`""`)
	var ts Timestamp
	err := json.Unmarshal(data, &ts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot parse timestamp")
}

func TestTimestamp_UnmarshalJSON_invalidDateString_ReturnsError(t *testing.T) {
	data := []byte(`"2024-13-40"`)
	var ts Timestamp
	err := json.Unmarshal(data, &ts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot parse timestamp")
}

func TestTimestamp_UnmarshalJSON_invalidType_ReturnsError(t *testing.T) {
	data := []byte(`{"value":1}`)
	var ts Timestamp
	err := json.Unmarshal(data, &ts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot unmarshal timestamp from")
}

func TestTimestamp_UnmarshalJSON_bool_ReturnsError(t *testing.T) {
	data := []byte("true")
	var ts Timestamp
	err := json.Unmarshal(data, &ts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot unmarshal timestamp from")
}

func TestTimestamp_UnmarshalJSON_invalidRFC3339Like_ReturnsError(t *testing.T) {
	data := []byte(`"2024-01-15T12:00:00"`)
	var ts Timestamp
	err := json.Unmarshal(data, &ts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot parse timestamp")
}

func TestTimestamp_MarshalJSON_zero(t *testing.T) {
	var ts Timestamp
	out, err := json.Marshal(ts)
	require.NoError(t, err)
	assert.Equal(t, "null", string(out))
}

func TestTimestamp_MarshalJSON_nonzero(t *testing.T) {
	ts := Timestamp{Time: time.Unix(1705315200, 0)}
	out, err := json.Marshal(ts)
	require.NoError(t, err)
	assert.Equal(t, "1705315200", string(out))
}
