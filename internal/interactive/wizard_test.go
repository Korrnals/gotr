package interactive

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAskRunWithPrompter_Create_RequiresName(t *testing.T) {
	p := NewMockPrompter().
		WithInputResponses("", "").
		WithConfirmResponses(true)

	answers, err := AskRunWithPrompter(p, false)
	assert.Error(t, err)
	assert.Nil(t, answers)
	assert.Contains(t, err.Error(), "run name is required")
}

func TestAskRunWithPrompter_Update_AllowsEmptyName(t *testing.T) {
	p := NewMockPrompter().
		WithInputResponses("", "update description").
		WithConfirmResponses(true)

	answers, err := AskRunWithPrompter(p, true)
	assert.NoError(t, err)
	assert.NotNil(t, answers)
	assert.Equal(t, "", answers.Name)
	assert.Equal(t, "update description", answers.Description)
	assert.True(t, answers.IncludeAll)
}

func TestAskProjectWithPrompter_Create(t *testing.T) {
	p := NewMockPrompter().
		WithInputResponses("Project A", "Announcement").
		WithConfirmResponses(true)

	answers, err := AskProjectWithPrompter(p, false)
	require.NoError(t, err)
	assert.Equal(t, "Project A", answers.Name)
	assert.Equal(t, "Announcement", answers.Announcement)
	assert.True(t, answers.ShowAnnouncement)
}

func TestAskProjectWithPrompter_Update(t *testing.T) {
	p := NewMockPrompter().
		WithInputResponses("Project B", "").
		WithConfirmResponses(false, true)

	answers, err := AskProjectWithPrompter(p, true)
	require.NoError(t, err)
	assert.False(t, answers.ShowAnnouncement)
	assert.True(t, answers.IsCompleted)
}

func TestAskSuiteWithPrompter_Create(t *testing.T) {
	p := NewMockPrompter().
		WithInputResponses("Suite A", "Desc")

	answers, err := AskSuiteWithPrompter(p, false)
	require.NoError(t, err)
	assert.Equal(t, "Suite A", answers.Name)
	assert.Equal(t, "Desc", answers.Description)
}

func TestAskSuiteWithPrompter_Update(t *testing.T) {
	p := NewMockPrompter().
		WithInputResponses("Suite B", "").
		WithConfirmResponses(true)

	answers, err := AskSuiteWithPrompter(p, true)
	require.NoError(t, err)
	assert.True(t, answers.IsCompleted)
}

func TestAskCaseWithPrompter(t *testing.T) {
	p := NewMockPrompter().
		WithInputResponses("Case title", "JIRA-1").
		WithSelectResponses(SelectResponse{Index: 2}, SelectResponse{Index: 1})

	answers, err := AskCaseWithPrompter(p, false)
	require.NoError(t, err)
	assert.Equal(t, "Case title", answers.Title)
	assert.Equal(t, int64(3), answers.TypeID)
	assert.Equal(t, int64(2), answers.PriorityID)
	assert.Equal(t, "JIRA-1", answers.Refs)
}

func TestAskConfirmWithPrompter(t *testing.T) {
	p := NewMockPrompter().WithConfirmResponses(true)

	confirmed, err := AskConfirmWithPrompter(p, "Proceed?")
	require.NoError(t, err)
	assert.True(t, confirmed)
}

func TestAskProjectWithPrompter_RequiresName(t *testing.T) {
	p := NewMockPrompter().WithInputResponses("")

	answers, err := AskProjectWithPrompter(p, false)
	assert.Error(t, err)
	assert.Nil(t, answers)
	assert.Contains(t, err.Error(), "project name is required")
}
