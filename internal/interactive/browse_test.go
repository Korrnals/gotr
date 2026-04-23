package interactive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrowse_SelectItem(t *testing.T) {
	// Items: [Alpha, Bravo]. Mock index 1 = Bravo (auto-adjusted past Exit).
	mock := NewMockPrompter().WithSelectResponses(SelectResponse{Index: 1})

	idx, err := Browse(context.Background(), mock, BrowseConfig{
		Prompt: "Pick one:",
		Items:  []string{"Alpha", "Bravo"},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, idx) // Bravo is index 1 in Items
}

func TestBrowse_Exit(t *testing.T) {
	// Raw index 0 = Exit (skip auto-adjustment)
	mock := NewMockPrompter().WithSelectResponses(SelectResponse{Index: 0, Raw: true})

	_, err := Browse(context.Background(), mock, BrowseConfig{
		Prompt: "Pick:",
		Items:  []string{"A"},
	})
	assert.ErrorIs(t, err, ErrExit)
}

func TestBrowse_Back(t *testing.T) {
	// Raw index 0 = Back (skip auto-adjustment)
	mock := NewMockPrompter().WithSelectResponses(SelectResponse{Index: 0, Raw: true})

	_, err := Browse(context.Background(), mock, BrowseConfig{
		Prompt:    "Pick:",
		Items:     []string{"A"},
		AllowBack: true,
	})
	assert.ErrorIs(t, err, ErrGoBack)
}

func TestBrowse_EmptyItems(t *testing.T) {
	mock := NewMockPrompter()
	_, err := Browse(context.Background(), mock, BrowseConfig{
		Prompt: "Pick:",
		Items:  nil,
	})
	assert.ErrorIs(t, err, ErrGoBack)
}

func TestBrowse_Interrupt(t *testing.T) {
	mock := &interruptPrompter{}

	_, err := Browse(context.Background(), mock, BrowseConfig{
		Prompt: "Pick:",
		Items:  []string{"A"},
	})
	assert.ErrorIs(t, err, ErrExit)
}

// interruptPrompter always returns context.Canceled.
type interruptPrompter struct{}

func (p *interruptPrompter) Input(string, string) (string, error)        { return "", context.Canceled }
func (p *interruptPrompter) Confirm(string, bool) (bool, error)          { return false, context.Canceled }
func (p *interruptPrompter) Select(string, []string) (int, string, error) { return 0, "", context.Canceled }
func (p *interruptPrompter) MultilineInput(string, string) (string, error) { return "", context.Canceled }
