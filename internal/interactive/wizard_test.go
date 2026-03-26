package interactive

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
