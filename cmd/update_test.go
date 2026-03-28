package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupUpdateTest настраивает тестовое окружение для update команды
func setupUpdateTest(t *testing.T, mock *client.MockClient) *cobra.Command {
	// Создаем новую команду для теста
	cmd := &cobra.Command{
		Use:   updateCmd.Use,
		Short: updateCmd.Short,
		Long:  updateCmd.Long,
		RunE:  runUpdate,
	}

	// Добавляем флаги
	cmd.Flags().StringP("name", "n", "", "Название ресурса")
	cmd.Flags().String("description", "", "Описание")
	cmd.Flags().String("announcement", "", "Announcement (для проекта)")
	cmd.Flags().Bool("show-announcement", false, "Показывать announcement")
	cmd.Flags().Bool("is-completed", false, "Отметить как завершённый")
	cmd.Flags().String("title", "", "Заголовок (для case)")
	cmd.Flags().Int64("type-id", 0, "ID типа (для case)")
	cmd.Flags().Int64("priority-id", 0, "ID приоритета (для case)")
	cmd.Flags().String("refs", "", "Ссылки (references)")
	cmd.Flags().Int64("suite-id", 0, "ID сьюта")
	cmd.Flags().Int64("milestone-id", 0, "ID milestone")
	cmd.Flags().Int64("assignedto-id", 0, "ID назначенного пользователя")
	cmd.Flags().String("case-ids", "", "ID кейсов через запятую (для run)")
	cmd.Flags().Bool("include-all", false, "Включить все кейсы (для run)")
	cmd.Flags().String("json-file", "", "Путь к JSON-файлу с данными")
	output.AddFlag(cmd)

	// Создаем контекст с mock clientом
	ctx := context.WithValue(context.Background(), httpClientKey, mock)
	cmd.SetContext(ctx)

	return cmd
}

func TestParseLabels(t *testing.T) {
	assert.Equal(t, []string{"bug", "regression"}, parseLabels("bug,regression"))
	assert.Equal(t, []string{"single"}, parseLabels("single"))
	assert.Empty(t, parseLabels(""))
	assert.Equal(t, []string{"a", "b"}, parseLabels("a,,b"))
}

// TestUpdate_Project_Success проверяет обновление проекта
func TestUpdate_Project_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateProjectFunc: func(ctx context.Context, projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, "Updated Project", req.Name)
			return &data.GetProjectResponse{ID: projectID, Name: req.Name}, nil
		},
	}

	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{"project", "1", "--name", "Updated Project"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestUpdate_Suite_Success проверяет обновление сьюта
func TestUpdate_Suite_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateSuiteFunc: func(ctx context.Context, suiteID int64, req *data.UpdateSuiteRequest) (*data.Suite, error) {
			assert.Equal(t, int64(100), suiteID)
			assert.Equal(t, "Updated Suite", req.Name)
			return &data.Suite{ID: suiteID, Name: req.Name}, nil
		},
	}

	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{"suite", "100", "--name", "Updated Suite"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestUpdate_Section_Success проверяет обновление секции
func TestUpdate_Section_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateSectionFunc: func(ctx context.Context, sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error) {
			assert.Equal(t, int64(200), sectionID)
			assert.Equal(t, "Updated Section", req.Name)
			return &data.Section{ID: sectionID, Name: req.Name}, nil
		},
	}

	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{"section", "200", "--name", "Updated Section"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestUpdate_Case_Success проверяет обновление кейса
func TestUpdate_Case_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateCaseFunc: func(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(12345), caseID)
			// Title передается как указатель
			assert.NotNil(t, req.Title)
			assert.Equal(t, "Updated Case", *req.Title)
			return &data.Case{ID: caseID, Title: *req.Title}, nil
		},
	}

	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{"case", "12345", "--title", "Updated Case", "--priority-id", "1"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestUpdate_Run_Success проверяет обновление рана
func TestUpdate_Run_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateRunFunc: func(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(1000), runID)
			// Name передается как указатель
			assert.NotNil(t, req.Name)
			assert.Equal(t, "Updated Run", *req.Name)
			return &data.Run{ID: runID, Name: *req.Name}, nil
		},
	}

	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{"run", "1000", "--name", "Updated Run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestUpdate_SharedStep_Success проверяет обновление shared step
func TestUpdate_SharedStep_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateSharedStepFunc: func(ctx context.Context, stepID int64, req *data.UpdateSharedStepRequest) (*data.SharedStep, error) {
			assert.Equal(t, int64(50), stepID)
			assert.Equal(t, "Updated Step", req.Title)
			return &data.SharedStep{ID: stepID, Title: req.Title}, nil
		},
	}

	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{"shared-step", "50", "--title", "Updated Step"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestUpdate_NoArgs проверяет ошибку при отсутствии аргументов
func TestUpdate_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint and id")
}

func TestUpdate_Section_NonInteractive_AutoWizard_NoMutatingCall(t *testing.T) {
	called := false
	mock := &client.MockClient{
		UpdateSectionFunc: func(ctx context.Context, sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error) {
			called = true
			return &data.Section{ID: sectionID, Name: "updated"}, nil
		},
	}

	cmd := setupUpdateTest(t, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{"section", "200"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.False(t, called)
}

func TestUpdate_Run_NonInteractive_AutoWizard_NoMutatingCall(t *testing.T) {
	called := false
	mock := &client.MockClient{
		UpdateRunFunc: func(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			called = true
			name := ""
			if req.Name != nil {
				name = *req.Name
			}
			return &data.Run{ID: runID, Name: name}, nil
		},
	}

	cmd := setupUpdateTest(t, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{"run", "1000"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.False(t, called)
}

func TestUpdate_SharedStep_NonInteractive_AutoWizard_NoMutatingCall(t *testing.T) {
	called := false
	mock := &client.MockClient{
		UpdateSharedStepFunc: func(ctx context.Context, stepID int64, req *data.UpdateSharedStepRequest) (*data.SharedStep, error) {
			called = true
			return &data.SharedStep{ID: stepID, Title: req.Title}, nil
		},
	}

	cmd := setupUpdateTest(t, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{"shared-step", "50"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.False(t, called)
}

// TestUpdate_InvalidID проверяет ошибку при неверном ID
func TestUpdate_InvalidID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{"project", "invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID")
}

// TestUpdate_UnsupportedEndpoint проверяет ошибку при неподдерживаемом endpoint
func TestUpdate_UnsupportedEndpoint(t *testing.T) {
	mock := &client.MockClient{}

	cmd := setupUpdateTest(t, mock)
	cmd.SetArgs([]string{"unsupported", "1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestUpdateProjectInteractive_Cancelled(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupUpdateTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Project", "Announcement").
		WithConfirmResponses(true, false, false)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := updateProjectInteractive(mock, cmd, 1)
	assert.NoError(t, err)
}

func TestUpdateSuiteInteractive_Cancelled(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupUpdateTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Suite", "Desc").
		WithConfirmResponses(true, false)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := updateSuiteInteractive(mock, cmd, 2)
	assert.NoError(t, err)
}

func TestUpdateCaseInteractive_Cancelled(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupUpdateTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Case title", "REF-1").
		WithSelectResponses(interactive.SelectResponse{Index: 0}, interactive.SelectResponse{Index: 0}).
		WithConfirmResponses(false)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := updateCaseInteractive(mock, cmd, 3)
	assert.NoError(t, err)
}

func TestRunUpdateDryRun_Project(t *testing.T) {
	cmd := setupUpdateTest(t, &client.MockClient{})
	_ = cmd.Flags().Set("name", "Updated Project")
	dr := output.NewDryRunPrinter("update project")

	err := runUpdateDryRun(cmd, dr, "project", 1, nil)
	assert.NoError(t, err)
}

func TestUpdateLabels_DryRun(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupUpdateTest(t, mock)
	cmd.Flags().String("labels", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	_ = cmd.Flags().Set("labels", "bug,regression")
	_ = cmd.Flags().Set("dry-run", "true")

	err := updateLabels(mock, cmd, 99)
	assert.NoError(t, err)
}

func TestUpdateSectionInteractive_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateSectionFunc: func(ctx context.Context, sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error) {
			assert.Equal(t, int64(5), sectionID)
			assert.Equal(t, "Section X", req.Name)
			assert.Equal(t, "Description X", req.Description)
			return &data.Section{ID: sectionID, Name: req.Name}, nil
		},
	}
	cmd := setupUpdateTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Section X", "Description X").
		WithConfirmResponses(true)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := updateSectionInteractive(mock, cmd, 5)
	assert.NoError(t, err)
}

func TestUpdateRunInteractive_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateRunFunc: func(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(8), runID)
			assert.NotNil(t, req.Name)
			assert.Equal(t, "Run X", *req.Name)
			return &data.Run{ID: runID, Name: *req.Name}, nil
		},
	}
	cmd := setupUpdateTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Run X", "Run Desc").
		WithConfirmResponses(true, true)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := updateRunInteractive(mock, cmd, 8)
	assert.NoError(t, err)
}

func TestRunUpdateInteractive_Unsupported(t *testing.T) {
	err := runUpdateInteractive(&client.MockClient{}, &cobra.Command{}, "unknown", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interactive mode not supported")
}

func TestRunUpdateDryRun_SwitchEndpoints(t *testing.T) {
	cmd := setupUpdateTest(t, &client.MockClient{})
	cmd.Flags().String("labels", "", "")
	_ = cmd.Flags().Set("name", "N")
	_ = cmd.Flags().Set("title", "T")
	_ = cmd.Flags().Set("description", "D")
	_ = cmd.Flags().Set("announcement", "A")
	_ = cmd.Flags().Set("show-announcement", "true")
	_ = cmd.Flags().Set("is-completed", "true")
	_ = cmd.Flags().Set("milestone-id", "11")
	_ = cmd.Flags().Set("assignedto-id", "22")
	_ = cmd.Flags().Set("include-all", "true")
	_ = cmd.Flags().Set("type-id", "2")
	_ = cmd.Flags().Set("priority-id", "3")
	_ = cmd.Flags().Set("refs", "REF")
	_ = cmd.Flags().Set("case-ids", "1,2")
	_ = cmd.Flags().Set("labels", "bug,regression")

	dr := output.NewDryRunPrinter("update")

	assert.NoError(t, runUpdateDryRun(cmd, dr, "project", 1, nil))
	assert.NoError(t, runUpdateDryRun(cmd, dr, "suite", 1, nil))
	assert.NoError(t, runUpdateDryRun(cmd, dr, "section", 1, nil))
	assert.NoError(t, runUpdateDryRun(cmd, dr, "case", 1, nil))
	assert.NoError(t, runUpdateDryRun(cmd, dr, "run", 1, nil))
	assert.NoError(t, runUpdateDryRun(cmd, dr, "shared-step", 1, nil))
	assert.NoError(t, runUpdateDryRun(cmd, dr, "labels", 1, nil))

	err := runUpdateDryRun(cmd, dr, "bad-endpoint", 1, nil)
	assert.Error(t, err)
}

func TestUpdateSharedStepInteractive_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateSharedStepFunc: func(ctx context.Context, sharedStepID int64, req *data.UpdateSharedStepRequest) (*data.SharedStep, error) {
			assert.Equal(t, int64(12), sharedStepID)
			assert.Equal(t, "New shared step", req.Title)
			return &data.SharedStep{ID: sharedStepID, Title: req.Title}, nil
		},
	}
	cmd := setupUpdateTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("New shared step").
		WithConfirmResponses(true)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := updateSharedStepInteractive(mock, cmd, 12)
	assert.NoError(t, err)
}

func TestUpdateSharedStepInteractive_RequiresTitle(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupUpdateTest(t, mock)
	p := interactive.NewMockPrompter().WithInputResponses("")
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := updateSharedStepInteractive(mock, cmd, 12)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shared step title is required")
}

func TestUpdateHandlers_JSONParseErrors(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupUpdateTest(t, mock)
	badJSON := []byte("{")

	assert.ErrorContains(t, updateProject(mock, cmd, 1, badJSON), "JSON parse error")
	assert.ErrorContains(t, updateSuite(mock, cmd, 1, badJSON), "JSON parse error")
	assert.ErrorContains(t, updateSection(mock, cmd, 1, badJSON), "JSON parse error")
	assert.ErrorContains(t, updateCase(mock, cmd, 1, badJSON), "JSON parse error")
	assert.ErrorContains(t, updateRun(mock, cmd, 1, badJSON), "JSON parse error")
	assert.ErrorContains(t, updateSharedStep(mock, cmd, 1, badJSON), "JSON parse error")
}

func TestUpdateHandlers_ClientErrors(t *testing.T) {
	mock := &client.MockClient{
		UpdateProjectFunc: func(ctx context.Context, projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error) {
			return nil, errors.New("project boom")
		},
		UpdateSuiteFunc: func(ctx context.Context, suiteID int64, req *data.UpdateSuiteRequest) (*data.Suite, error) {
			return nil, errors.New("suite boom")
		},
		UpdateSectionFunc: func(ctx context.Context, sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error) {
			return nil, errors.New("section boom")
		},
		UpdateCaseFunc: func(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			return nil, errors.New("case boom")
		},
		UpdateRunFunc: func(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			return nil, errors.New("run boom")
		},
		UpdateSharedStepFunc: func(ctx context.Context, stepID int64, req *data.UpdateSharedStepRequest) (*data.SharedStep, error) {
			return nil, errors.New("step boom")
		},
	}

	t.Run("update project client error", func(t *testing.T) {
		cmd := setupUpdateTest(t, mock)
		err := updateProject(mock, cmd, 1, nil)
		assert.ErrorContains(t, err, "failed to update project")
	})

	t.Run("update suite client error", func(t *testing.T) {
		cmd := setupUpdateTest(t, mock)
		err := updateSuite(mock, cmd, 1, nil)
		assert.ErrorContains(t, err, "failed to update suite")
	})

	t.Run("update section client error", func(t *testing.T) {
		cmd := setupUpdateTest(t, mock)
		err := updateSection(mock, cmd, 1, nil)
		assert.ErrorContains(t, err, "failed to update section")
	})

	t.Run("update case client error", func(t *testing.T) {
		cmd := setupUpdateTest(t, mock)
		err := updateCase(mock, cmd, 1, nil)
		assert.ErrorContains(t, err, "failed to update case")
	})

	t.Run("update run client error", func(t *testing.T) {
		cmd := setupUpdateTest(t, mock)
		err := updateRun(mock, cmd, 1, nil)
		assert.ErrorContains(t, err, "failed to update run")
	})

	t.Run("update shared step client error", func(t *testing.T) {
		cmd := setupUpdateTest(t, mock)
		err := updateSharedStep(mock, cmd, 1, nil)
		assert.ErrorContains(t, err, "failed to update shared step")
	})
}

func TestUpdateLabels_ErrorBranches(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupUpdateTest(t, mock)
	cmd.Flags().String("labels", "", "")
	cmd.Flags().Bool("dry-run", false, "")

	err := updateLabels(mock, cmd, 99)
	assert.ErrorContains(t, err, "--labels is required")

	require.NoError(t, cmd.Flags().Set("labels", ",,"))
	err = updateLabels(mock, cmd, 99)
	assert.ErrorContains(t, err, "labels not specified")
}
