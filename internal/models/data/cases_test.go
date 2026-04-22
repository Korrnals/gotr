package data

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddCaseRequest_MarshalJSON_MergesExtraFields(t *testing.T) {
	autotestOn := int64(1)
	req := AddCaseRequest{
		Title:            "Case A",
		SectionID:        42,
		CustomAutotestOn: &autotestOn,
		ExtraFields: map[string]interface{}{
			"custom_loadtest_on": int64(2),
			"custom_env":         "prod",
		},
	}

	raw, err := json.Marshal(req)
	require.NoError(t, err)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &payload))

	assert.Equal(t, "Case A", payload["title"])
	assert.Equal(t, float64(42), payload["section_id"])
	assert.Equal(t, float64(1), payload["custom_autotest_on"])
	assert.Equal(t, float64(2), payload["custom_loadtest_on"])
	assert.Equal(t, "prod", payload["custom_env"])
	_, hasExtraFields := payload["ExtraFields"]
	assert.False(t, hasExtraFields)
	_, hasExtraFieldsSnake := payload["extra_fields"]
	assert.False(t, hasExtraFieldsSnake)
}