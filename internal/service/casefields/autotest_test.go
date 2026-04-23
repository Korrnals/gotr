package casefields

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveCustomAutotestOn_ProjectSpecificDefault(t *testing.T) {
	fields := mustCaseFields(t, 49, "1", true)

	resolved, err := ResolveCustomAutotestOn(fields, 49)
	require.NoError(t, err)
	if assert.NotNil(t, resolved) {
		assert.Equal(t, int64(1), *resolved)
	}
}

func TestResolveCustomAutotestOn_RequiredWithoutDefault(t *testing.T) {
	fields := mustCaseFields(t, 49, "", true)

	resolved, err := ResolveCustomAutotestOn(fields, 49)
	assert.Nil(t, resolved)
	assert.ErrorContains(t, err, "required for project 49")
}

func TestResolveCustomAutotestOn_IgnoresOtherProjects(t *testing.T) {
	fields := mustCaseFields(t, 99, "1", true)

	resolved, err := ResolveCustomAutotestOn(fields, 49)
	require.NoError(t, err)
	assert.Nil(t, resolved)
}

func mustCaseFields(t *testing.T, projectID int64, defaultValue string, required bool) data.GetCaseFieldsResponse {
	t.Helper()

	payload := fmt.Sprintf(`[{"name":"custom_autotest_on","system_name":"custom_autotest_on","configs":[{"context":{"is_global":false,"project_ids":[%d]},"id":"cfg","options":{"default_value":%q,"is_required":%t}}]}]`, projectID, defaultValue, required)
	var fields data.GetCaseFieldsResponse
	require.NoError(t, json.Unmarshal([]byte(payload), &fields))
	return fields
}
