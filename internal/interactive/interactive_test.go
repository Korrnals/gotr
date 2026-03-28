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
		assert.Contains(t, err.Error(), "failed to select project")
		assert.Equal(t, "Select project:", p.lastMessage)
	})
}

func TestSelectRun_DefaultPromptAndCompletedStatus(t *testing.T) {
	p := &spyPrompter{idx: 1}
	runs := data.GetRunsResponse{
		{ID: 11, Name: "Active Run", IsCompleted: false},
		{ID: 22, Name: "Closed Run", IsCompleted: true},
	}

	id, err := SelectRun(context.Background(), p, runs, "")
	require.NoError(t, err)
	assert.Equal(t, int64(22), id)
	assert.Equal(t, "Select run:", p.lastMessage)
	require.Len(t, p.lastOptions, 2)
	assert.Contains(t, p.lastOptions[0], "(active)")
	assert.Contains(t, p.lastOptions[1], "(completed)")
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
