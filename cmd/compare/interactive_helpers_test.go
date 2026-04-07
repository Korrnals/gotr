package compare

import (
	"bytes"
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Interactive pid selection tests ====================

func TestParseCommonFlags_InteractivePid1(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 10, Name: "Project Alpha"},
				{ID: 20, Name: "Project Beta"},
			}, nil
		},
	}

	// Mock prompter: select index 0 for pid1
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}). // pid1
		WithConfirmResponses(false)                                // do not save

	cmd := &cobra.Command{}
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid2=20"})
	_ = cmd.Execute()

	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	pid1, pid2, format, _, err := parseCommonFlags(cmd, mock)
	require.NoError(t, err)
	assert.Equal(t, int64(10), pid1)
	assert.Equal(t, int64(20), pid2)
	assert.Equal(t, "table", format)
}

func TestParseCommonFlags_InteractivePid2(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 10, Name: "Project Alpha"},
				{ID: 20, Name: "Project Beta"},
			}, nil
		},
	}

	// Mock prompter: select index 1 for pid2
	mp := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 1}). // pid2
		WithConfirmResponses(false)                                // do not save

	cmd := &cobra.Command{}
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=10"})
	_ = cmd.Execute()

	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	pid1, pid2, _, _, err := parseCommonFlags(cmd, mock)
	require.NoError(t, err)
	assert.Equal(t, int64(10), pid1)
	assert.Equal(t, int64(20), pid2)
}

func TestParseCommonFlags_InteractiveBothPids(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 10, Name: "Project Alpha"},
				{ID: 20, Name: "Project Beta"},
				{ID: 30, Name: "Project Gamma"},
			}, nil
		},
	}

	// Mock prompter: select index 0 for pid1, index 2 for pid2
	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 0}, // pid1
			interactive.SelectResponse{Index: 2}, // pid2
		).
		WithConfirmResponses(false) // do not save

	cmd := &cobra.Command{}
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{})
	_ = cmd.Execute()

	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	pid1, pid2, _, _, err := parseCommonFlags(cmd, mock)
	require.NoError(t, err)
	assert.Equal(t, int64(10), pid1)
	assert.Equal(t, int64(30), pid2)
}

func TestParseCommonFlags_FlagsProvidedSkipsInteractive(t *testing.T) {
	mock := &client.MockClient{} // no GetProjectsFunc — must not be called

	cmd := &cobra.Command{}
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=5", "--pid2=7"})
	_ = cmd.Execute()

	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter())
	cmd.SetContext(ctx)

	pid1, pid2, format, _, err := parseCommonFlags(cmd, mock)
	require.NoError(t, err)
	assert.Equal(t, int64(5), pid1)
	assert.Equal(t, int64(7), pid2)
	assert.Equal(t, "table", format)
}

func TestParseCommonFlags_NonInteractiveFailsWithoutPids(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "P1"},
			}, nil
		},
	}

	cmd := &cobra.Command{}
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{})
	_ = cmd.Execute()

	// NonInteractivePrompter rejects all prompts
	ctx := interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter())
	cmd.SetContext(ctx)

	_, _, _, _, err := parseCommonFlags(cmd, mock)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pid1 not specified")
	assert.Contains(t, err.Error(), "non-interactive")
}

func TestParseCommonFlags_NonInteractivePid2Missing(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 10, Name: "P1"},
			}, nil
		},
	}

	cmd := &cobra.Command{}
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{"--pid1=10"})
	_ = cmd.Execute()

	ctx := interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter())
	cmd.SetContext(ctx)

	_, _, _, _, err := parseCommonFlags(cmd, mock)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pid2 not specified")
	assert.Contains(t, err.Error(), "non-interactive")
}

// ==================== End-to-end command interactive tests ====================

func TestSuitesCmd_InteractivePids(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "P1"},
				{ID: 2, Name: "P2"},
			}, nil
		},
		GetProjectFunc: func(ctx context.Context, projectID int64) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: projectID, Name: "Test"}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return []data.Suite{}, nil
		},
	}
	SetGetClientForTests(func(cmd *cobra.Command) client.ClientInterface {
		return mock
	})

	mp := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 0}, // pid1
			interactive.SelectResponse{Index: 1}, // pid2
		).
		WithConfirmResponses(false) // do not save

	cmd := newSuitesCmd()
	addPersistentFlagsForTests(cmd)
	cmd.SetArgs([]string{}) // no --pid1, --pid2

	ctx := interactive.WithPrompter(context.Background(), mp)
	cmd.SetContext(ctx)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
}
