package cmd

import (
	"context"
	"errors"
	"os"
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

func TestUpdateRun_FieldMapping_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateRunFunc: func(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(1000), runID)
			require.NotNil(t, req.Name)
			require.NotNil(t, req.Description)
			require.NotNil(t, req.MilestoneID)
			require.NotNil(t, req.AssignedTo)
			require.NotNil(t, req.IncludeAll)
			assert.Equal(t, "Run Name", *req.Name)
			assert.Equal(t, "Run Desc", *req.Description)
			assert.Equal(t, int64(11), *req.MilestoneID)
			assert.Equal(t, int64(22), *req.AssignedTo)
			assert.False(t, *req.IncludeAll)
			assert.Equal(t, []int64{1, 3}, req.CaseIDs)
			return &data.Run{ID: runID, Name: *req.Name}, nil
		},
	}

	cmd := setupUpdateTest(t, mock)
	require.NoError(t, cmd.Flags().Set("name", "Run Name"))
	require.NoError(t, cmd.Flags().Set("description", "Run Desc"))
	require.NoError(t, cmd.Flags().Set("milestone-id", "11"))
	require.NoError(t, cmd.Flags().Set("assignedto-id", "22"))
	require.NoError(t, cmd.Flags().Set("include-all", "false"))
	require.NoError(t, cmd.Flags().Set("case-ids", "1,bad,3"))

	err := updateRun(mock, cmd, 1000, nil)
	assert.NoError(t, err)
}

func TestUpdateRun_JSONSuccess(t *testing.T) {
	mock := &client.MockClient{
		UpdateRunFunc: func(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(1000), runID)
			require.NotNil(t, req.Name)
			assert.Equal(t, "JSON Run", *req.Name)
			return &data.Run{ID: runID, Name: *req.Name}, nil
		},
	}
	cmd := setupUpdateTest(t, mock)

	err := updateRun(mock, cmd, 1000, []byte(`{"name":"JSON Run"}`))
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

func TestRunUpdate_OrchestrationBranches(t *testing.T) {
	t.Run("dry-run branch", func(t *testing.T) {
		cmd := setupUpdateTest(t, &client.MockClient{})
		cmd.Flags().Bool("dry-run", false, "")
		require.NoError(t, cmd.Flags().Set("dry-run", "true"))
		require.NoError(t, cmd.Flags().Set("name", "Dry Project"))
		err := runUpdate(cmd, []string{"project", "1"})
		assert.NoError(t, err)
	})

	t.Run("interactive branch", func(t *testing.T) {
		mock := &client.MockClient{
			UpdateProjectFunc: func(ctx context.Context, projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error) {
				return &data.GetProjectResponse{ID: projectID, Name: req.Name}, nil
			},
		}
		cmd := setupUpdateTest(t, mock)
		cmd.Flags().BoolP("interactive", "i", false, "")
		p := interactive.NewMockPrompter().
			WithInputResponses("Project", "Announcement").
			WithConfirmResponses(true, true, true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))
		require.NoError(t, cmd.Flags().Set("interactive", "true"))

		err := runUpdate(cmd, []string{"project", "1"})
		assert.NoError(t, err)
	})

	t.Run("json-file success branch", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "update-json-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		_, _ = tmpFile.WriteString(`{"name":"FromFile"}`)
		_ = tmpFile.Close()

		mock := &client.MockClient{
			UpdateProjectFunc: func(ctx context.Context, projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error) {
				return &data.GetProjectResponse{ID: projectID, Name: req.Name}, nil
			},
		}
		cmd := setupUpdateTest(t, mock)
		require.NoError(t, cmd.Flags().Set("json-file", tmpFile.Name()))

		err = runUpdate(cmd, []string{"project", "1"})
		assert.NoError(t, err)
	})
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

func TestUpdateProjectInteractive_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateProjectFunc: func(ctx context.Context, projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error) {
			assert.Equal(t, int64(11), projectID)
			assert.Equal(t, "Project Updated", req.Name)
			assert.Equal(t, "Announcement Updated", req.Announcement)
			assert.True(t, req.ShowAnnouncement)
			assert.True(t, req.IsCompleted)
			return &data.GetProjectResponse{ID: projectID, Name: req.Name}, nil
		},
	}
	cmd := setupUpdateTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Project Updated", "Announcement Updated").
		WithConfirmResponses(true, true, true)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := updateProjectInteractive(mock, cmd, 11)
	assert.NoError(t, err)
}

func TestUpdateSuiteInteractive_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateSuiteFunc: func(ctx context.Context, suiteID int64, req *data.UpdateSuiteRequest) (*data.Suite, error) {
			assert.Equal(t, int64(22), suiteID)
			assert.Equal(t, "Suite Updated", req.Name)
			assert.Equal(t, "Suite Description", req.Description)
			assert.True(t, req.IsCompleted)
			return &data.Suite{ID: suiteID, Name: req.Name}, nil
		},
	}
	cmd := setupUpdateTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Suite Updated", "Suite Description").
		WithConfirmResponses(true, true)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := updateSuiteInteractive(mock, cmd, 22)
	assert.NoError(t, err)
}

func TestUpdateCaseInteractive_Success(t *testing.T) {
	mock := &client.MockClient{
		UpdateCaseFunc: func(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(33), caseID)
			require.NotNil(t, req.Title)
			require.NotNil(t, req.TypeID)
			require.NotNil(t, req.PriorityID)
			require.NotNil(t, req.Refs)
			assert.Equal(t, "Case Updated", *req.Title)
			assert.Equal(t, int64(2), *req.TypeID)
			assert.Equal(t, int64(3), *req.PriorityID)
			assert.Equal(t, "REF-2", *req.Refs)
			return &data.Case{ID: caseID, Title: *req.Title}, nil
		},
	}
	cmd := setupUpdateTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Case Updated", "REF-2").
		WithSelectResponses(interactive.SelectResponse{Index: 1}, interactive.SelectResponse{Index: 2}).
		WithConfirmResponses(true)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := updateCaseInteractive(mock, cmd, 33)
	assert.NoError(t, err)
}

func TestUpdateInteractive_InputErrorBranches(t *testing.T) {
	nonInteractive := interactive.NewNonInteractivePrompter()

	t.Run("project input error", func(t *testing.T) {
		cmd := setupUpdateTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), nonInteractive))
		err := updateProjectInteractive(&client.MockClient{}, cmd, 1)
		assert.ErrorContains(t, err, "input error")
	})

	t.Run("suite input error", func(t *testing.T) {
		cmd := setupUpdateTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), nonInteractive))
		err := updateSuiteInteractive(&client.MockClient{}, cmd, 2)
		assert.ErrorContains(t, err, "input error")
	})

	t.Run("case input error", func(t *testing.T) {
		cmd := setupUpdateTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), nonInteractive))
		err := updateCaseInteractive(&client.MockClient{}, cmd, 3)
		assert.ErrorContains(t, err, "input error")
	})
}

func TestUpdateInteractive_ClientErrorBranches(t *testing.T) {
	t.Run("project client error", func(t *testing.T) {
		mock := &client.MockClient{
			UpdateProjectFunc: func(ctx context.Context, projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error) {
				return nil, errors.New("project boom")
			},
		}
		cmd := setupUpdateTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Project", "Announcement").
			WithConfirmResponses(true, true, true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := updateProjectInteractive(mock, cmd, 11)
		assert.ErrorContains(t, err, "failed to update project")
	})

	t.Run("suite client error", func(t *testing.T) {
		mock := &client.MockClient{
			UpdateSuiteFunc: func(ctx context.Context, suiteID int64, req *data.UpdateSuiteRequest) (*data.Suite, error) {
				return nil, errors.New("suite boom")
			},
		}
		cmd := setupUpdateTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Suite", "Description").
			WithConfirmResponses(true, true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := updateSuiteInteractive(mock, cmd, 22)
		assert.ErrorContains(t, err, "failed to update suite")
	})

	t.Run("case client error", func(t *testing.T) {
		mock := &client.MockClient{
			UpdateCaseFunc: func(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
				return nil, errors.New("case boom")
			},
		}
		cmd := setupUpdateTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Case", "REF").
			WithSelectResponses(interactive.SelectResponse{Index: 0}, interactive.SelectResponse{Index: 0}).
			WithConfirmResponses(true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := updateCaseInteractive(mock, cmd, 33)
		assert.ErrorContains(t, err, "failed to update case")
	})
}

func TestUpdateRunInteractive_Cancelled(t *testing.T) {
	called := false
	mock := &client.MockClient{
		UpdateRunFunc: func(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			called = true
			return &data.Run{ID: runID}, nil
		},
	}
	cmd := setupUpdateTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Run", "Desc").
		WithConfirmResponses(true, false)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := updateRunInteractive(mock, cmd, 8)
	assert.NoError(t, err)
	assert.False(t, called)
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

func TestShouldAutoRunUpdateInteractive_Branches(t *testing.T) {
	t.Run("no prompter in context", func(t *testing.T) {
		cmd := setupUpdateTest(t, &client.MockClient{})
		assert.False(t, shouldAutoRunUpdateInteractive(cmd, "project", false))
	})

	t.Run("json file disables interactive", func(t *testing.T) {
		cmd := setupUpdateTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		assert.False(t, shouldAutoRunUpdateInteractive(cmd, "project", true))
	})

	t.Run("project no changed flags", func(t *testing.T) {
		cmd := setupUpdateTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		assert.True(t, shouldAutoRunUpdateInteractive(cmd, "project", false))
	})

	t.Run("project with changed flag", func(t *testing.T) {
		cmd := setupUpdateTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		require.NoError(t, cmd.Flags().Set("name", "Updated"))
		assert.False(t, shouldAutoRunUpdateInteractive(cmd, "project", false))
	})

	t.Run("unsupported endpoint", func(t *testing.T) {
		cmd := setupUpdateTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		assert.False(t, shouldAutoRunUpdateInteractive(cmd, "unsupported", false))
	})

	t.Run("suite with and without changed flags", func(t *testing.T) {
		cmd := setupUpdateTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		assert.True(t, shouldAutoRunUpdateInteractive(cmd, "suite", false))
		require.NoError(t, cmd.Flags().Set("description", "Changed"))
		assert.False(t, shouldAutoRunUpdateInteractive(cmd, "suite", false))
	})

	t.Run("section with and without changed flags", func(t *testing.T) {
		cmd := setupUpdateTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		assert.True(t, shouldAutoRunUpdateInteractive(cmd, "section", false))
		require.NoError(t, cmd.Flags().Set("name", "Changed"))
		assert.False(t, shouldAutoRunUpdateInteractive(cmd, "section", false))
	})

	t.Run("case with and without changed flags", func(t *testing.T) {
		cmd := setupUpdateTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		assert.True(t, shouldAutoRunUpdateInteractive(cmd, "case", false))
		require.NoError(t, cmd.Flags().Set("title", "Changed"))
		assert.False(t, shouldAutoRunUpdateInteractive(cmd, "case", false))
	})

	t.Run("run with and without changed flags", func(t *testing.T) {
		cmd := setupUpdateTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		assert.True(t, shouldAutoRunUpdateInteractive(cmd, "run", false))
		require.NoError(t, cmd.Flags().Set("case-ids", "1,2"))
		assert.False(t, shouldAutoRunUpdateInteractive(cmd, "run", false))
	})

	t.Run("shared-step with and without changed flags", func(t *testing.T) {
		cmd := setupUpdateTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		assert.True(t, shouldAutoRunUpdateInteractive(cmd, "shared-step", false))
		require.NoError(t, cmd.Flags().Set("title", "Changed"))
		assert.False(t, shouldAutoRunUpdateInteractive(cmd, "shared-step", false))
	})
}

func TestRunUpdateInteractive_SupportedEndpoints(t *testing.T) {
	t.Run("project", func(t *testing.T) {
		called := false
		mock := &client.MockClient{
			UpdateProjectFunc: func(ctx context.Context, projectID int64, req *data.UpdateProjectRequest) (*data.GetProjectResponse, error) {
				called = true
				return &data.GetProjectResponse{ID: projectID, Name: req.Name}, nil
			},
		}
		cmd := setupUpdateTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("P", "A").
			WithConfirmResponses(true, true, true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := runUpdateInteractive(mock, cmd, "project", 1)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("suite", func(t *testing.T) {
		called := false
		mock := &client.MockClient{
			UpdateSuiteFunc: func(ctx context.Context, suiteID int64, req *data.UpdateSuiteRequest) (*data.Suite, error) {
				called = true
				return &data.Suite{ID: suiteID, Name: req.Name}, nil
			},
		}
		cmd := setupUpdateTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("S", "D").
			WithConfirmResponses(true, true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := runUpdateInteractive(mock, cmd, "suite", 2)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("section", func(t *testing.T) {
		called := false
		mock := &client.MockClient{
			UpdateSectionFunc: func(ctx context.Context, sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error) {
				called = true
				return &data.Section{ID: sectionID, Name: req.Name}, nil
			},
		}
		cmd := setupUpdateTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Sec", "Desc").
			WithConfirmResponses(true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := runUpdateInteractive(mock, cmd, "section", 3)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("case", func(t *testing.T) {
		called := false
		mock := &client.MockClient{
			UpdateCaseFunc: func(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
				called = true
				title := ""
				if req.Title != nil {
					title = *req.Title
				}
				return &data.Case{ID: caseID, Title: title}, nil
			},
		}
		cmd := setupUpdateTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Case", "REF").
			WithSelectResponses(interactive.SelectResponse{Index: 0}, interactive.SelectResponse{Index: 0}).
			WithConfirmResponses(true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := runUpdateInteractive(mock, cmd, "case", 4)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("run", func(t *testing.T) {
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
		p := interactive.NewMockPrompter().
			WithInputResponses("Run", "Desc").
			WithConfirmResponses(true, true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := runUpdateInteractive(mock, cmd, "run", 5)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("shared-step", func(t *testing.T) {
		called := false
		mock := &client.MockClient{
			UpdateSharedStepFunc: func(ctx context.Context, sharedStepID int64, req *data.UpdateSharedStepRequest) (*data.SharedStep, error) {
				called = true
				return &data.SharedStep{ID: sharedStepID, Title: req.Title}, nil
			},
		}
		cmd := setupUpdateTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Step").
			WithConfirmResponses(true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := runUpdateInteractive(mock, cmd, "shared-step", 6)
		assert.NoError(t, err)
		assert.True(t, called)
	})
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

func TestRunUpdateDryRun_JSONBranches(t *testing.T) {
	cmd := setupUpdateTest(t, &client.MockClient{})
	cmd.Flags().String("labels", "", "")
	require.NoError(t, cmd.Flags().Set("labels", "bug,regression"))
	dr := output.NewDryRunPrinter("update-json")

	assert.NoError(t, runUpdateDryRun(cmd, dr, "project", 1, []byte(`{"name":"P"}`)))
	assert.NoError(t, runUpdateDryRun(cmd, dr, "suite", 1, []byte(`{"name":"S"}`)))
	assert.NoError(t, runUpdateDryRun(cmd, dr, "section", 1, []byte(`{"name":"Sec"}`)))
	assert.NoError(t, runUpdateDryRun(cmd, dr, "case", 1, []byte(`{"title":"Case"}`)))
	assert.NoError(t, runUpdateDryRun(cmd, dr, "run", 1, []byte(`{"name":"Run"}`)))
	assert.NoError(t, runUpdateDryRun(cmd, dr, "shared-step", 1, []byte(`{"title":"Step"}`)))
	assert.NoError(t, runUpdateDryRun(cmd, dr, "labels", 1, []byte(`{"labels":"a"}`)))
}

func TestUpdateSectionInteractive_ErrorBranches(t *testing.T) {
	t.Run("at least one field required", func(t *testing.T) {
		cmd := setupUpdateTest(t, &client.MockClient{})
		p := interactive.NewMockPrompter().WithInputResponses("", "")
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := updateSectionInteractive(&client.MockClient{}, cmd, 9)
		assert.ErrorContains(t, err, "at least one field is required")
	})

	t.Run("cancelled", func(t *testing.T) {
		called := false
		mock := &client.MockClient{
			UpdateSectionFunc: func(ctx context.Context, sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error) {
				called = true
				return &data.Section{ID: sectionID, Name: req.Name}, nil
			},
		}
		cmd := setupUpdateTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Name", "").
			WithConfirmResponses(false)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := updateSectionInteractive(mock, cmd, 9)
		assert.NoError(t, err)
		assert.False(t, called)
	})

	t.Run("client error", func(t *testing.T) {
		mock := &client.MockClient{
			UpdateSectionFunc: func(ctx context.Context, sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error) {
				return nil, errors.New("section boom")
			},
		}
		cmd := setupUpdateTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Name", "").
			WithConfirmResponses(true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := updateSectionInteractive(mock, cmd, 9)
		assert.ErrorContains(t, err, "failed to update section")
	})
}

func TestUpdateRunInteractive_ClientError(t *testing.T) {
	mock := &client.MockClient{
		UpdateRunFunc: func(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			return nil, errors.New("run boom")
		},
	}
	cmd := setupUpdateTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Run X", "Run Desc").
		WithConfirmResponses(true, true)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := updateRunInteractive(mock, cmd, 8)
	assert.ErrorContains(t, err, "failed to update run")
}

func TestUpdateSharedStepInteractive_CancelledAndClientError(t *testing.T) {
	t.Run("cancelled", func(t *testing.T) {
		called := false
		mock := &client.MockClient{
			UpdateSharedStepFunc: func(ctx context.Context, sharedStepID int64, req *data.UpdateSharedStepRequest) (*data.SharedStep, error) {
				called = true
				return &data.SharedStep{ID: sharedStepID, Title: req.Title}, nil
			},
		}
		cmd := setupUpdateTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Step").
			WithConfirmResponses(false)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := updateSharedStepInteractive(mock, cmd, 12)
		assert.NoError(t, err)
		assert.False(t, called)
	})

	t.Run("client error", func(t *testing.T) {
		mock := &client.MockClient{
			UpdateSharedStepFunc: func(ctx context.Context, sharedStepID int64, req *data.UpdateSharedStepRequest) (*data.SharedStep, error) {
				return nil, errors.New("step boom")
			},
		}
		cmd := setupUpdateTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Step").
			WithConfirmResponses(true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := updateSharedStepInteractive(mock, cmd, 12)
		assert.ErrorContains(t, err, "failed to update shared step")
	})
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

func TestUpdateLabels_ClientError(t *testing.T) {
	mock := &client.MockClient{
		UpdateTestLabelsFunc: func(ctx context.Context, testID int64, labels []string) error {
			return errors.New("labels boom")
		},
	}

	cmd := setupUpdateTest(t, mock)
	cmd.Flags().String("labels", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	require.NoError(t, cmd.Flags().Set("labels", "bug,regression"))

	err := updateLabels(mock, cmd, 99)
	assert.ErrorContains(t, err, "failed to update labels")
}

func TestUpdateLabels_Success(t *testing.T) {
	called := false
	mock := &client.MockClient{
		UpdateTestLabelsFunc: func(ctx context.Context, testID int64, labels []string) error {
			called = true
			assert.Equal(t, int64(99), testID)
			assert.Equal(t, []string{"bug", "regression"}, labels)
			return nil
		},
	}

	cmd := setupUpdateTest(t, mock)
	cmd.Flags().String("labels", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	require.NoError(t, cmd.Flags().Set("labels", "bug,regression"))

	err := updateLabels(mock, cmd, 99)
	assert.NoError(t, err)
	assert.True(t, called)
}
