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

func TestAskProjectWithPrompter_ErrorBranches(t *testing.T) {
	t.Run("name input error", func(t *testing.T) {
		p := NewMockPrompter()
		answers, err := AskProjectWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "project name input failed")
	})

	t.Run("announcement input error", func(t *testing.T) {
		p := NewMockPrompter().
			WithInputResponses("Project")
		answers, err := AskProjectWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "announcement input failed")
	})

	t.Run("show announcement confirm error", func(t *testing.T) {
		p := NewMockPrompter().
			WithInputResponses("Project", "Ann")
		answers, err := AskProjectWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "show announcement confirm failed")
	})

	t.Run("update completed confirm error", func(t *testing.T) {
		p := NewMockPrompter().
			WithInputResponses("Project", "Ann").
			WithConfirmResponses(true)
		answers, err := AskProjectWithPrompter(p, true)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "is completed confirm failed")
	})
}

func TestAskSuiteWithPrompter_ErrorBranches(t *testing.T) {
	t.Run("suite name input error", func(t *testing.T) {
		p := NewMockPrompter()
		answers, err := AskSuiteWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "suite name input failed")
	})

	t.Run("suite name required", func(t *testing.T) {
		p := NewMockPrompter().WithInputResponses("")
		answers, err := AskSuiteWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "suite name is required")
	})

	t.Run("description input error", func(t *testing.T) {
		p := NewMockPrompter().
			WithInputResponses("Suite")
		answers, err := AskSuiteWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "suite description input failed")
	})

	t.Run("update completion confirm error", func(t *testing.T) {
		p := NewMockPrompter().
			WithInputResponses("Suite", "Desc")
		answers, err := AskSuiteWithPrompter(p, true)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "suite completion confirm failed")
	})
}

func TestAskCaseWithPrompter_ErrorBranches(t *testing.T) {
	t.Run("title input error", func(t *testing.T) {
		p := NewMockPrompter()
		answers, err := AskCaseWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "case title input failed")
	})

	t.Run("title required", func(t *testing.T) {
		p := NewMockPrompter().WithInputResponses("")
		answers, err := AskCaseWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "case title is required")
	})

	t.Run("type selection error", func(t *testing.T) {
		p := NewMockPrompter().
			WithInputResponses("Case")
		answers, err := AskCaseWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "case type selection failed")
	})

	t.Run("priority selection error", func(t *testing.T) {
		p := NewMockPrompter().
			WithInputResponses("Case").
			WithSelectResponses(SelectResponse{Index: 0})
		answers, err := AskCaseWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "priority selection failed")
	})

	t.Run("references input error", func(t *testing.T) {
		p := NewMockPrompter().
			WithInputResponses("Case").
			WithSelectResponses(SelectResponse{Index: 0}, SelectResponse{Index: 0})
		answers, err := AskCaseWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "references input failed")
	})

	t.Run("type selection index out of range", func(t *testing.T) {
		p := NewMockPrompter().
			WithInputResponses("Case").
			WithSelectResponses(SelectResponse{Index: 99})
		answers, err := AskCaseWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "case type selection failed")
	})
}

func TestAskRunWithPrompter_ErrorBranches(t *testing.T) {
	t.Run("run name input error", func(t *testing.T) {
		p := NewMockPrompter()
		answers, err := AskRunWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "run name input failed")
	})

	t.Run("description input error", func(t *testing.T) {
		p := NewMockPrompter().
			WithInputResponses("Run")
		answers, err := AskRunWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "run description input failed")
	})

	t.Run("suite id input error", func(t *testing.T) {
		p := NewMockPrompter().
			WithInputResponses("Run", "Desc")
		answers, err := AskRunWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "suite id input failed")
	})

	t.Run("suite id required", func(t *testing.T) {
		p := NewMockPrompter().
			WithInputResponses("Run", "Desc", "")
		answers, err := AskRunWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "suite id is required")
	})

	t.Run("suite id invalid", func(t *testing.T) {
		p := NewMockPrompter().
			WithInputResponses("Run", "Desc", "abc")
		answers, err := AskRunWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "invalid suite id")
	})

	t.Run("include all confirm error", func(t *testing.T) {
		p := NewMockPrompter().
			WithInputResponses("Run", "Desc", "11")
		answers, err := AskRunWithPrompter(p, false)
		assert.Error(t, err)
		assert.Nil(t, answers)
		assert.Contains(t, err.Error(), "include all confirm failed")
	})
}

func TestAskConfirmWithPrompter_Error(t *testing.T) {
	p := NewMockPrompter()
	confirmed, err := AskConfirmWithPrompter(p, "Proceed?")
	assert.Error(t, err)
	assert.False(t, confirmed)
	assert.Contains(t, err.Error(), "confirmation failed")
}
