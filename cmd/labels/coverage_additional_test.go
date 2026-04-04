package labels

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestLabelsCmds_InteractiveResolveErrors(t *testing.T) {
	getProjectsErrMock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return nil, assert.AnError
		},
	}

	t.Run("get", func(t *testing.T) {
		cmd := newGetCmd(getClientForTests)
		cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, getProjectsErrMock).Context(), interactive.NewMockPrompter()))
		cmd.SetArgs([]string{})
		err := cmd.Execute()
		assert.Error(t, err)
	})

	t.Run("list", func(t *testing.T) {
		cmd := newListCmd(getClientForTests)
		cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, getProjectsErrMock).Context(), interactive.NewMockPrompter()))
		cmd.SetArgs([]string{})
		err := cmd.Execute()
		assert.Error(t, err)
	})

	t.Run("update-test", func(t *testing.T) {
		cmd := newUpdateTestCmd(getClientForTests)
		cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, getProjectsErrMock).Context(), interactive.NewMockPrompter()))
		cmd.SetArgs([]string{"--labels", "smoke"})
		err := cmd.Execute()
		assert.Error(t, err)
	})

	t.Run("update-label", func(t *testing.T) {
		cmd := newUpdateLabelCmd(getClientForTests)
		cmd.SetContext(interactive.WithPrompter(setupTestCmd(t, getProjectsErrMock).Context(), interactive.NewMockPrompter()))
		cmd.SetArgs([]string{"--project", "1", "--title", "Label"})
		err := cmd.Execute()
		assert.Error(t, err)
	})
}