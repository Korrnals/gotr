package interactive

import (
	"context"
	"errors"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/AlecAivazis/survey/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerminalPrompter_Input(t *testing.T) {
	original := surveyAskOne
	defer func() { surveyAskOne = original }()

	surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		out, ok := response.(*string)
		require.True(t, ok)
		*out = "typed"
		return nil
	}

	tp := &TerminalPrompter{}
	got, err := tp.Input("msg", "default")
	require.NoError(t, err)
	assert.Equal(t, "typed", got)
}

func TestTerminalPrompter_Confirm(t *testing.T) {
	original := surveyAskOne
	defer func() { surveyAskOne = original }()

	surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		out, ok := response.(*bool)
		require.True(t, ok)
		*out = true
		return nil
	}

	tp := &TerminalPrompter{}
	got, err := tp.Confirm("confirm?", false)
	require.NoError(t, err)
	assert.True(t, got)
}

func TestTerminalPrompter_MultilineInput(t *testing.T) {
	original := surveyAskOne
	defer func() { surveyAskOne = original }()

	surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		out, ok := response.(*string)
		require.True(t, ok)
		*out = "line1\nline2"
		return nil
	}

	tp := &TerminalPrompter{}
	got, err := tp.MultilineInput("body", "")
	require.NoError(t, err)
	assert.Equal(t, "line1\nline2", got)
}

func TestTerminalPrompter_InputError(t *testing.T) {
	original := surveyAskOne
	defer func() { surveyAskOne = original }()

	surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		return errors.New("ask failed")
	}

	tp := &TerminalPrompter{}
	_, err := tp.Input("msg", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get input")
}

func TestTerminalPrompter_Select_EmptyOptions(t *testing.T) {
	tp := &TerminalPrompter{}
	_, _, err := tp.Select("pick", []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "select options list is empty")
}

func TestTerminalPrompter_Select_AskError(t *testing.T) {
	original := surveyAskOne
	defer func() { surveyAskOne = original }()

	surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		return errors.New("ask failed")
	}

	tp := &TerminalPrompter{}
	_, _, err := tp.Select("pick", []string{"a", "b"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select option")
}

func TestTerminalPrompter_Select_SelectedNotInList(t *testing.T) {
	original := surveyAskOne
	defer func() { surveyAskOne = original }()

	surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		out, ok := response.(*string)
		require.True(t, ok)
		*out = "c"
		return nil
	}

	tp := &TerminalPrompter{}
	_, _, err := tp.Select("pick", []string{"a", "b"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "selected option is not in list")
}

func TestTerminalPrompter_ConfirmError(t *testing.T) {
	original := surveyAskOne
	defer func() { surveyAskOne = original }()

	surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		return errors.New("ask failed")
	}

	tp := &TerminalPrompter{}
	_, err := tp.Confirm("confirm?", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get confirmation")
}

func TestTerminalPrompter_MultilineInputError(t *testing.T) {
	original := surveyAskOne
	defer func() { surveyAskOne = original }()

	surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		return errors.New("ask failed")
	}

	tp := &TerminalPrompter{}
	_, err := tp.MultilineInput("body", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get multiline input")
}

func TestPrompterFromContext_Default(t *testing.T) {
	p := PrompterFromContext(context.Background())
	assert.NotNil(t, p)
}

func TestPrompterFromContext_WithPrompter(t *testing.T) {
	mock := NewMockPrompter()
	ctx := WithPrompter(context.Background(), mock)
	p := PrompterFromContext(ctx)
	assert.Equal(t, mock, p)
}

func TestPrompterFromContext_NilContext(t *testing.T) {
	p := PrompterFromContext(nil)
	assert.NotNil(t, p)
}

func TestWithPrompter_NilContext(t *testing.T) {
	ctx := WithPrompter(nil, NewMockPrompter())
	assert.NotNil(t, ctx)
}

func TestHasPrompterInContext(t *testing.T) {
	assert.False(t, HasPrompterInContext(nil))
	assert.False(t, HasPrompterInContext(context.Background()))

	ctx := WithPrompter(context.Background(), NewMockPrompter())
	assert.True(t, HasPrompterInContext(ctx))
}

func TestIsNonInteractive(t *testing.T) {
	assert.False(t, IsNonInteractive(nil))
	assert.False(t, IsNonInteractive(context.Background()))

	ctxInteractive := WithPrompter(context.Background(), NewMockPrompter())
	assert.False(t, IsNonInteractive(ctxInteractive))

	ctxNonInteractive := WithPrompter(context.Background(), NewNonInteractivePrompter())
	assert.True(t, IsNonInteractive(ctxNonInteractive))
}

func TestNonInteractivePrompter_AllMethods(t *testing.T) {
	p := NewNonInteractivePrompter()

	_, err := p.Input("name", "")
	assert.ErrorIs(t, err, ErrNonInteractive)

	_, err = p.MultilineInput("body", "")
	assert.ErrorIs(t, err, ErrNonInteractive)

	_, err = p.Confirm("confirm", true)
	assert.ErrorIs(t, err, ErrNonInteractive)

	_, _, err = p.Select("select", []string{"a", "b"})
	assert.ErrorIs(t, err, ErrNonInteractive)
}

func TestMockPrompter_Queues(t *testing.T) {
	m := NewMockPrompter().
		WithInputResponses("hello", "world").
		WithConfirmResponses(true).
		WithSelectResponses(SelectResponse{Index: 1})

	input, err := m.Input("msg", "")
	require.NoError(t, err)
	assert.Equal(t, "hello", input)

	multi, err := m.MultilineInput("msg", "")
	require.NoError(t, err)
	assert.Equal(t, "world", multi)

	ok, err := m.Confirm("confirm", false)
	require.NoError(t, err)
	assert.True(t, ok)

	idx, value, err := m.Select("select", []string{"x", "y"})
	require.NoError(t, err)
	assert.Equal(t, 1, idx)
	assert.Equal(t, "y", value)
}

func TestMockPrompter_ExhaustedQueue(t *testing.T) {
	m := NewMockPrompter()
	_, err := m.Input("msg", "")
	assert.Error(t, err)
}

func TestSelectProject_WithMockPrompter(t *testing.T) {
	ctx := context.Background()
	mockPrompt := NewMockPrompter().WithSelectResponses(SelectResponse{Index: 1})
	mockClient := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 11, Name: "Project A"},
				{ID: 22, Name: "Project B"},
			}, nil
		},
	}

	projectID, err := SelectProject(ctx, mockPrompt, mockClient, "Select project")
	require.NoError(t, err)
	assert.Equal(t, int64(22), projectID)
}

func TestSelectSuite_WithMockPrompter(t *testing.T) {
	ctx := context.Background()
	mockPrompt := NewMockPrompter().WithSelectResponses(SelectResponse{Index: 0})
	suites := data.GetSuitesResponse{{ID: 101, Name: "Suite 1"}}

	suiteID, err := SelectSuite(ctx, mockPrompt, suites, "Select suite")
	require.NoError(t, err)
	assert.Equal(t, int64(101), suiteID)
}

func TestSelectRun_WithMockPrompter(t *testing.T) {
	ctx := context.Background()
	mockPrompt := NewMockPrompter().WithSelectResponses(SelectResponse{Index: 0})
	runs := data.GetRunsResponse{{ID: 301, Name: "Run 1"}}

	runID, err := SelectRun(ctx, mockPrompt, runs, "Select run")
	require.NoError(t, err)
	assert.Equal(t, int64(301), runID)
}

func TestSelectSection_WithMockPrompter(t *testing.T) {
	ctx := context.Background()
	mockPrompt := NewMockPrompter().WithSelectResponses(SelectResponse{Index: 1})
	sections := data.GetSectionsResponse{
		{ID: 401, Name: "Section A"},
		{ID: 402, Name: "Section B"},
	}

	sectionID, err := SelectSection(ctx, mockPrompt, sections, "Select section")
	require.NoError(t, err)
	assert.Equal(t, int64(402), sectionID)
}
