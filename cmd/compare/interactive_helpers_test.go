package compare

import (
	"bytes"
	"context"
	"strings"
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

// ==================== Tests for renderTableLines ====================

func TestRenderTableLines_Empty(t *testing.T) {
	result := CompareResult{
		Resource:   "labels",
		Project1ID: 1,
		Project2ID: 2,
	}

	lines := renderTableLines(result, "P1", "P2")
	require.NotEmpty(t, lines)

	joined := joinLines(lines)
	assert.Contains(t, joined, "labels")
	assert.Contains(t, joined, "(none)")
}

func TestRenderTableLines_WithData(t *testing.T) {
	result := CompareResult{
		Resource:     "suites",
		Project1ID:   10,
		Project2ID:   20,
		OnlyInFirst:  []ItemInfo{{ID: 1, Name: "Suite A"}},
		OnlyInSecond: []ItemInfo{{ID: 2, Name: "Suite B"}},
		Common: []CommonItemInfo{
			{Name: "Suite C", ID1: 3, ID2: 4, IDsMatch: true},
			{Name: "Suite D", ID1: 5, ID2: 6, IDsMatch: false},
		},
	}

	lines := renderTableLines(result, "Project Alpha", "Project Beta")
	joined := joinLines(lines)

	assert.Contains(t, joined, "Suite A")
	assert.Contains(t, joined, "Suite B")
	assert.Contains(t, joined, "Suite C")
	assert.Contains(t, joined, "Suite D")
	assert.Contains(t, joined, "✓ Match")
	assert.Contains(t, joined, "⚠ Differ")
	assert.Contains(t, joined, "ID mapping")
}

func TestRenderTableLines_NoIDMismatch(t *testing.T) {
	result := CompareResult{
		Resource:   "groups",
		Project1ID: 1,
		Project2ID: 2,
		Common: []CommonItemInfo{
			{Name: "Grp", ID1: 10, ID2: 10, IDsMatch: true},
		},
	}

	lines := renderTableLines(result, "P1", "P2")
	joined := joinLines(lines)
	assert.Contains(t, joined, "(all IDs match)")
}

// ==================== Tests for comparePostAction ====================

func TestComparePostAction_NonInteractive(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter())

	result := CompareResult{Resource: "cases"}
	action := comparePostAction(ctx, nil, result, "P1", "P2")
	assert.Equal(t, actionExit, action)
}

func TestComparePostAction_Exit(t *testing.T) {
	mock := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Value: interactive.OptExit})
	ctx := interactive.WithPrompter(context.Background(), mock)

	result := CompareResult{Resource: "cases"}
	action := comparePostAction(ctx, nil, result, "P1", "P2")
	assert.Equal(t, actionExit, action)
}

func TestComparePostAction_Sync(t *testing.T) {
	mock := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 2, Value: "→ Sync: migrate differences"})
	ctx := interactive.WithPrompter(context.Background(), mock)

	result := CompareResult{
		Resource:    "cases",
		OnlyInFirst: []ItemInfo{{ID: 1, Name: "A"}},
	}
	action := comparePostAction(ctx, nil, result, "P1", "P2")
	assert.Equal(t, actionSync, action)
}

func TestComparePostAction_SyncDisabledNoDiffs(t *testing.T) {
	// When no differences, sync is disabled → disabled re-prompt loop → exit.
	mock := interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 2, Value: "→ Sync: migrate differences"},
			interactive.SelectResponse{Index: 0, Value: interactive.OptExit},
		)
	ctx := interactive.WithPrompter(context.Background(), mock)

	result := CompareResult{Resource: "cases"}
	action := comparePostAction(ctx, nil, result, "P1", "P2")
	assert.Equal(t, actionExit, action)
}

// ==================== Tests for collectDrillDownResources ====================

func TestCollectDrillDownResources_AllNil(t *testing.T) {
	result := &allResult{}
	entries := collectDrillDownResources(result)
	assert.Empty(t, entries)
}

func TestCollectDrillDownResources_SomeComplete(t *testing.T) {
	result := &allResult{
		Cases:  &CompareResult{Resource: "cases", Status: CompareStatusComplete},
		Suites: &CompareResult{Resource: "suites", Status: CompareStatusInterrupted},
		Labels: &CompareResult{Resource: "labels", Status: CompareStatusComplete},
	}
	entries := collectDrillDownResources(result)
	assert.Len(t, entries, 2) // cases + labels
	assert.Equal(t, "Cases", entries[0].name)
	assert.Equal(t, "Labels", entries[1].name)
}

// ==================== Tests for compareAllPostAction ====================

func TestCompareAllPostAction_Exit(t *testing.T) {
	mock := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0, Value: interactive.OptExit})
	ctx := interactive.WithPrompter(context.Background(), mock)

	result := &allResult{}
	action := compareAllPostAction(ctx, nil, result, "P1", "P2", 1, 2)
	assert.Equal(t, actionExit, action)
}

func TestCompareAllPostAction_Save(t *testing.T) {
	mock := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 2, Value: "💾 Save results to file"})
	ctx := interactive.WithPrompter(context.Background(), mock)

	result := &allResult{}
	action := compareAllPostAction(ctx, nil, result, "P1", "P2", 1, 2)
	assert.Equal(t, actionSave, action)
}

// ==================== Tests for line rendering helpers ====================

func TestHBorder(t *testing.T) {
	line := hBorder("┌", "┬", "┐", []int{5, 10})
	assert.True(t, strings.HasPrefix(line, "┌"))
	assert.True(t, strings.HasSuffix(line, "┐"))
	assert.Contains(t, line, "┬")
}

func TestRowLine(t *testing.T) {
	line := rowLine([]string{"hello", "world"}, []int{10, 10})
	assert.Contains(t, line, "hello")
	assert.Contains(t, line, "world")
	assert.True(t, strings.HasPrefix(line, "│"))
	assert.True(t, strings.HasSuffix(line, "│"))
}

func TestHeaderLine(t *testing.T) {
	line := headerLine("Title", 30)
	assert.Contains(t, line, "Title")
	assert.True(t, strings.HasPrefix(line, "│"))
}

func TestSeparatorLine(t *testing.T) {
	line := separatorLine([]int{5, 10})
	assert.True(t, strings.HasPrefix(line, "├"))
	assert.True(t, strings.HasSuffix(line, "┤"))
	assert.Contains(t, line, "┼")
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}
