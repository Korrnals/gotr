package interactive

import (
	"context"
	"errors"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type spyPrompter struct {
	lastMessage string
	lastOptions []string
	idx         int
	err         error
}

func (s *spyPrompter) Input(message, defaultVal string) (string, error) {
	return "", nil
}

func (s *spyPrompter) Confirm(message string, def bool) (bool, error) {
	return false, nil
}

func (s *spyPrompter) Select(message string, options []string) (int, string, error) {
	s.lastMessage = message
	s.lastOptions = append([]string(nil), options...)
	if s.err != nil {
		return 0, "", s.err
	}
	if s.idx < 0 || s.idx >= len(options) {
		return 0, "", errors.New("bad index")
	}
	return s.idx, options[s.idx], nil
}

func (s *spyPrompter) MultilineInput(message, defaultVal string) (string, error) {
	return "", nil
}

func TestSelectProject_ErrorBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("get projects error", func(t *testing.T) {
		cli := &client.MockClient{
			GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
				return nil, errors.New("boom")
			},
		}

		_, err := SelectProject(ctx, &spyPrompter{}, cli, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get projects list")
	})

	t.Run("no projects", func(t *testing.T) {
		cli := &client.MockClient{
			GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
				return data.GetProjectsResponse{}, nil
			},
		}

		_, err := SelectProject(ctx, &spyPrompter{}, cli, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no projects found")
	})

	t.Run("select error", func(t *testing.T) {
		cli := &client.MockClient{
			GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
				return data.GetProjectsResponse{{ID: 10, Name: "P"}}, nil
			},
		}
		p := &spyPrompter{err: errors.New("select failed")}

		_, err := SelectProject(ctx, p, cli, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "select failed")
	})
}

func TestSelectRun_DefaultPromptAndCompletedStatus(t *testing.T) {
	// Browse options: [✕ Exit, (active) Active Run, (completed) Closed Run]
	// Index 2 selects "Closed Run"
	p := &spyPrompter{idx: 2}
	runs := data.GetRunsResponse{
		{ID: 11, Name: "Active Run", IsCompleted: false},
		{ID: 22, Name: "Closed Run", IsCompleted: true},
	}

	id, err := SelectRun(context.Background(), p, runs, "")
	require.NoError(t, err)
	assert.Equal(t, int64(22), id)
	assert.Contains(t, p.lastMessage, "Select run:")
	require.Len(t, p.lastOptions, 3) // Exit + 2 runs
	assert.Contains(t, p.lastOptions[1], "active")
	assert.Contains(t, p.lastOptions[2], "completed")
}

func TestSelectSuiteForProject_Branches(t *testing.T) {
	ctx := context.Background()

	t.Run("get suites error", func(t *testing.T) {
		cli := &client.MockClient{
			GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
				return nil, errors.New("boom")
			},
		}

		_, err := SelectSuiteForProject(ctx, &spyPrompter{}, cli, 99, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get suites for project 99")
	})

	t.Run("select suite error propagation", func(t *testing.T) {
		cli := &client.MockClient{
			GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
				return data.GetSuitesResponse{}, nil
			},
		}

		_, err := SelectSuiteForProject(ctx, &spyPrompter{}, cli, 99, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no suites found")
	})
}

func TestSelectSuite_Run_Section_ErrorBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("SelectSuite no suites", func(t *testing.T) {
		_, err := SelectSuite(ctx, &spyPrompter{}, data.GetSuitesResponse{}, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no suites found")
	})

	t.Run("SelectSuite select error", func(t *testing.T) {
		p := &spyPrompter{err: errors.New("select fail")}
		_, err := SelectSuite(ctx, p, data.GetSuitesResponse{{ID: 1, Name: "S"}}, "Custom suite prompt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "select fail")
	})

	t.Run("SelectRun no runs", func(t *testing.T) {
		_, err := SelectRun(ctx, &spyPrompter{}, data.GetRunsResponse{}, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no runs found")
	})

	t.Run("SelectRun select error", func(t *testing.T) {
		p := &spyPrompter{err: errors.New("select fail")}
		_, err := SelectRun(ctx, p, data.GetRunsResponse{{ID: 2, Name: "R"}}, "Custom run prompt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "select fail")
	})

	t.Run("SelectSection no sections", func(t *testing.T) {
		_, err := SelectSection(ctx, &spyPrompter{}, data.GetSectionsResponse{}, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no sections found")
	})

	t.Run("SelectSection select error", func(t *testing.T) {
		p := &spyPrompter{err: errors.New("select fail")}
		_, err := SelectSection(ctx, p, data.GetSectionsResponse{{ID: 3, Name: "Sec"}}, "Custom section prompt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "select fail")
	})
}

func TestSelectSuiteAndSection_DefaultPromptSuccess(t *testing.T) {
	ctx := context.Background()

	t.Run("SelectSuite default prompt", func(t *testing.T) {
		p := &spyPrompter{idx: 1} // +1 for Exit option
		id, err := SelectSuite(ctx, p, data.GetSuitesResponse{{ID: 101, Name: "Suite"}}, "")
		require.NoError(t, err)
		assert.Equal(t, int64(101), id)
		assert.Contains(t, p.lastMessage, "Select suite:")
	})

	t.Run("SelectSection default prompt", func(t *testing.T) {
		p := &spyPrompter{idx: 1} // +1 for Exit option
		id, err := SelectSection(ctx, p, data.GetSectionsResponse{{ID: 202, Name: "Section"}}, "")
		require.NoError(t, err)
		assert.Equal(t, int64(202), id)
		assert.Contains(t, p.lastMessage, "Select section:")
	})
}

func TestSelectCase(t *testing.T) {
	ctx := context.Background()

	t.Run("success selects correct case", func(t *testing.T) {
		cases := data.GetCasesResponse{
			{ID: 100, Title: "Case A"},
			{ID: 200, Title: "Case B"},
		}
		p := NewMockPrompter().WithSelectResponses(SelectResponse{Index: 1})
		id, err := SelectCase(ctx, p, cases, "")
		require.NoError(t, err)
		assert.Equal(t, int64(200), id)
	})

	t.Run("empty list returns error", func(t *testing.T) {
		p := NewMockPrompter()
		id, err := SelectCase(ctx, p, data.GetCasesResponse{}, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no cases found")
		assert.Zero(t, id)
	})

	t.Run("select error propagates", func(t *testing.T) {
		cases := data.GetCasesResponse{{ID: 1, Title: "C"}}
		p := NewNonInteractivePrompter()
		id, err := SelectCase(ctx, p, cases, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to select case")
		assert.Zero(t, id)
	})

	t.Run("custom prompt is used", func(t *testing.T) {
		cases := data.GetCasesResponse{{ID: 1, Title: "C"}}
		p := &spyPrompter{idx: 1} // +1 for Exit
		id, err := SelectCase(ctx, p, cases, "Pick a case:")
		require.NoError(t, err)
		assert.Equal(t, int64(1), id)
		assert.Contains(t, p.lastMessage, "Pick a case:")
	})

	t.Run("exit returns ErrExit", func(t *testing.T) {
		cases := data.GetCasesResponse{{ID: 1, Title: "C"}}
		p := NewMockPrompter().WithSelectResponses(SelectResponse{Index: 0, Raw: true}) // Exit
		_, err := SelectCase(ctx, p, cases, "")
		assert.True(t, IsExit(err))
	})

	t.Run("aligned labels have header", func(t *testing.T) {
		cases := data.GetCasesResponse{
			{ID: 100, Title: "Login Test"},
			{ID: 200, Title: "Logout Test"},
		}
		p := &spyPrompter{idx: 1} // +1 for Exit
		_, err := SelectCase(ctx, p, cases, "")
		require.NoError(t, err)
		assert.Contains(t, p.lastMessage, "ID")
		assert.Contains(t, p.lastMessage, "Title")
	})
}

func TestSelectSharedStep(t *testing.T) {
	ctx := context.Background()

	t.Run("success selects correct step", func(t *testing.T) {
		steps := data.GetSharedStepsResponse{
			{ID: 555, Title: "Step A"},
			{ID: 666, Title: "Step B"},
		}
		p := NewMockPrompter().WithSelectResponses(SelectResponse{Index: 0})
		id, err := SelectSharedStep(ctx, p, steps, "")
		require.NoError(t, err)
		assert.Equal(t, int64(555), id)
	})

	t.Run("empty list returns error", func(t *testing.T) {
		p := NewMockPrompter()
		id, err := SelectSharedStep(ctx, p, data.GetSharedStepsResponse{}, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no shared steps found")
		assert.Zero(t, id)
	})

	t.Run("select error propagates", func(t *testing.T) {
		steps := data.GetSharedStepsResponse{{ID: 1, Title: "S"}}
		p := NewNonInteractivePrompter()
		id, err := SelectSharedStep(ctx, p, steps, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to select shared step")
		assert.Zero(t, id)
	})

	t.Run("default prompt", func(t *testing.T) {
		steps := data.GetSharedStepsResponse{{ID: 1, Title: "S"}}
		p := &spyPrompter{idx: 1} // +1 for Exit
		_, err := SelectSharedStep(ctx, p, steps, "")
		require.NoError(t, err)
		assert.Contains(t, p.lastMessage, "Select shared step:")
	})
}
