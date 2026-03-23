package interactive

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
