package casefields

import (
	"encoding/json"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveRequiredCustomFields_ProjectSpecificDefaults(t *testing.T) {
	payload := `[
		{
			"system_name": "custom_loadtest_on",
			"configs": [{
				"context": {"is_global": false, "project_ids": [34]},
				"id": "cfg-load",
				"options": {"default_value": "2", "is_required": true}
			}]
		},
		{
			"system_name": "custom_env",
			"configs": [{
				"context": {"is_global": false, "project_ids": [34]},
				"id": "cfg-env",
				"options": {"default_value": "prod", "is_required": true}
			}]
		},
		{
			"system_name": "custom_autotest_on",
			"configs": [{
				"context": {"is_global": false, "project_ids": [34]},
				"id": "cfg-auto",
				"options": {"default_value": "1", "is_required": true}
			}]
		},
		{
			"system_name": "custom_optional",
			"configs": [{
				"context": {"is_global": false, "project_ids": [34]},
				"id": "cfg-opt",
				"options": {"default_value": "x", "is_required": false}
			}]
		},
		{
			"system_name": "custom_other_project",
			"configs": [{
				"context": {"is_global": false, "project_ids": [49]},
				"id": "cfg-other",
				"options": {"default_value": "z", "is_required": true}
			}]
		}
	]`

	var fields data.GetCaseFieldsResponse
	require.NoError(t, json.Unmarshal([]byte(payload), &fields))

	resolved := ResolveRequiredCustomFields(fields, 34, []string{"custom_autotest_on"})

	assert.Equal(t, map[string]interface{}{
		"custom_loadtest_on": int64(2),
		"custom_env":         "prod",
	}, resolved)
}