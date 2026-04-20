package interactive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActionMenu_SelectOption(t *testing.T) {
	// Options: [Exit, Save] → raw index 1 = Save (skip auto-adjust: ActionMenu uses OptExit as label)
	mock := NewMockPrompter().WithSelectResponses(SelectResponse{Index: 1, Raw: true})

	key, err := ActionMenu(context.Background(), mock, "What next?", []ActionOption{
		{Label: OptExit, Key: "exit"},
		{Label: "💾 Save", Key: "save"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "save", key)
}

func TestActionMenu_Empty(t *testing.T) {
	mock := NewMockPrompter()
	_, err := ActionMenu(context.Background(), mock, "What next?", nil)
	assert.ErrorIs(t, err, ErrExit)
}

func TestActionMenu_Interrupt(t *testing.T) {
	mock := &interruptPrompter{}

	_, err := ActionMenu(context.Background(), mock, "What next?", []ActionOption{
		{Label: "Test", Key: "test"},
	})
	assert.ErrorIs(t, err, ErrExit)
}
