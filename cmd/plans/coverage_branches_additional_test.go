package plans

import (
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestAddCmd_NoArgs_NoPrompter_Error(t *testing.T) {
	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, &client.MockClient{}).Context())
	cmd.SetArgs([]string{"--name", "Plan without project"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project_id is required in non-interactive mode")
}

func TestCloseCmd_NoArgs_NoPrompter_Error(t *testing.T) {
	cmd := newCloseCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, &client.MockClient{}).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan_id is required in non-interactive mode")
}

func TestDeleteCmd_NoArgs_NoPrompter_Error(t *testing.T) {
	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, &client.MockClient{}).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan_id is required in non-interactive mode")
}

func TestGetCmd_NoArgs_NoPrompter_Error(t *testing.T) {
	cmd := newGetCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, &client.MockClient{}).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan_id is required in non-interactive mode")
}

func TestListCmd_NoArgs_NoPrompter_Error(t *testing.T) {
	cmd := newListCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, &client.MockClient{}).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project_id is required in non-interactive mode")
}

func TestUpdateCmd_NoArgs_NoPrompter_Error(t *testing.T) {
	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, &client.MockClient{}).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan_id is required in non-interactive mode")
}