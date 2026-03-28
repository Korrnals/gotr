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

// setupAddTest настраивает тестовое окружение для add команды
func setupAddTest(t *testing.T, mock *client.MockClient) *cobra.Command {
	// Создаем новую команду для теста
	cmd := &cobra.Command{
		Use:   addCmd.Use,
		Short: addCmd.Short,
		Long:  addCmd.Long,
		RunE:  runAdd,
	}

	// Добавляем флаги
	cmd.Flags().StringP("name", "n", "", "Название ресурса")
	cmd.Flags().String("description", "", "Описание/announcement")
	cmd.Flags().String("announcement", "", "Announcement (для проекта)")
	cmd.Flags().Bool("show-announcement", false, "Показывать announcement")
	cmd.Flags().Int64("suite-id", 0, "ID сьюта")
	cmd.Flags().Int64("section-id", 0, "ID секции")
	cmd.Flags().Int64("milestone-id", 0, "ID milestone")
	cmd.Flags().Int64("template-id", 0, "ID шаблона (для case)")
	cmd.Flags().Int64("type-id", 0, "ID типа (для case)")
	cmd.Flags().Int64("priority-id", 0, "ID приоритета (для case)")
	cmd.Flags().String("title", "", "Заголовок (для case)")
	cmd.Flags().String("refs", "", "Ссылки (references)")
	cmd.Flags().String("comment", "", "Комментарий (для result)")
	cmd.Flags().Int64("status-id", 0, "ID статуса (для result)")
	cmd.Flags().String("elapsed", "", "Time выполнения (для result)")
	cmd.Flags().String("defects", "", "Дефекты (для result)")
	cmd.Flags().Int64("assignedto-id", 0, "ID назначенного пользователя")
	cmd.Flags().String("case-ids", "", "ID кейсов через запятую (для run)")
	cmd.Flags().Bool("include-all", true, "Включить все кейсы (для run)")
	cmd.Flags().String("json-file", "", "Путь к JSON-файлу с данными")
	output.AddFlag(cmd)

	// Создаем контекст с mock clientом
	ctx := context.WithValue(context.Background(), httpClientKey, mock)
	cmd.SetContext(ctx)

	return cmd
}

func TestParseCaseIDs(t *testing.T) {
	assert.Equal(t, []int64{1, 2, 3}, parseCaseIDs("1,2,3"))
	assert.Equal(t, []int64{10}, parseCaseIDs("10"))
	assert.Empty(t, parseCaseIDs(""))
	assert.Empty(t, parseCaseIDs("abc,xyz"))
	result := parseCaseIDs("1,bad,3")
	assert.Equal(t, []int64{1, 3}, result)
}

func TestSplitAndTrim(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, splitAndTrim("a,b,c", ","))
	assert.Equal(t, []string{"a", "b"}, splitAndTrim("a,,b", ","))
	assert.Empty(t, splitAndTrim("", ","))
	assert.Equal(t, []string{"hello"}, splitAndTrim("hello", ","))
}

func TestSplitString(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, splitString("a,b,c", ","))
	assert.Equal(t, []string{"a", "", "b"}, splitString("a,,b", ","))
	assert.Equal(t, []string{""}, splitString("", ","))
}

func TestParseOptionalID(t *testing.T) {
	id, err := parseOptionalID("")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), id)

	id, err = parseOptionalID("123")
	assert.NoError(t, err)
	assert.Equal(t, int64(123), id)

	_, err = parseOptionalID("bad-id")
	assert.Error(t, err)
}

func TestAddProjectInteractive_Cancelled(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupAddTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Project X", "Announcement").
		WithConfirmResponses(true, false)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := addProjectInteractive(mock, cmd)
	assert.NoError(t, err)
}

func TestAddSuiteInteractive_Cancelled(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupAddTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Suite X", "Description").
		WithConfirmResponses(false)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := addSuiteInteractive(mock, cmd, 10)
	assert.NoError(t, err)
}

func TestAddCaseInteractive_Cancelled(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupAddTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Case title", "REF-1").
		WithSelectResponses(interactive.SelectResponse{Index: 0}, interactive.SelectResponse{Index: 0}).
		WithConfirmResponses(false)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := addCaseInteractive(mock, cmd, 20)
	assert.NoError(t, err)
}

func TestAddRunInteractive_Cancelled(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupAddTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Run name", "Run description", "77").
		WithConfirmResponses(true, false)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := addRunInteractive(mock, cmd, 30)
	assert.NoError(t, err)
}

func TestRunAddDryRun_Project(t *testing.T) {
	cmd := setupAddTest(t, &client.MockClient{})
	_ = cmd.Flags().Set("name", "Dry Project")
	dr := output.NewDryRunPrinter("add project")

	err := runAddDryRun(cmd, dr, "project", 0, nil)
	assert.NoError(t, err)
}

func TestAddSection_NameRequired(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupAddTest(t, mock)

	err := addSection(mock, cmd, 10, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--name is required")
}

func TestAddResultForCase_StatusRequired(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupAddTest(t, mock)

	err := addResultForCase(mock, cmd, 10, 20, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--status-id is required")
}

func TestAddSectionInteractive_Success(t *testing.T) {
	mock := &client.MockClient{
		AddSectionFunc: func(ctx context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
			assert.Equal(t, int64(10), projectID)
			assert.Equal(t, "Section A", req.Name)
			assert.Equal(t, int64(2), req.SuiteID)
			assert.Equal(t, int64(3), req.ParentID)
			return &data.Section{ID: 123, Name: req.Name}, nil
		},
	}
	cmd := setupAddTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Section A", "Description A", "2", "3").
		WithConfirmResponses(true)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := addSectionInteractive(mock, cmd, 10)
	assert.NoError(t, err)
}

func TestAddSharedStepInteractive_Success(t *testing.T) {
	mock := &client.MockClient{
		AddSharedStepFunc: func(ctx context.Context, projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
			assert.Equal(t, int64(15), projectID)
			assert.Equal(t, "Shared A", req.Title)
			return &data.SharedStep{ID: 77, Title: req.Title}, nil
		},
	}
	cmd := setupAddTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Shared A").
		WithConfirmResponses(true)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := addSharedStepInteractive(mock, cmd, 15)
	assert.NoError(t, err)
}

func TestAddSectionInteractive_ErrorBranches(t *testing.T) {
	t.Run("name required", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		p := interactive.NewMockPrompter().WithInputResponses("")
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := addSectionInteractive(&client.MockClient{}, cmd, 10)
		assert.ErrorContains(t, err, "section name is required")
	})

	t.Run("invalid suite id", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		p := interactive.NewMockPrompter().
			WithInputResponses("Section A", "Desc", "not-a-number")
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := addSectionInteractive(&client.MockClient{}, cmd, 10)
		assert.ErrorContains(t, err, "invalid suite id")
	})

	t.Run("invalid parent section id", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		p := interactive.NewMockPrompter().
			WithInputResponses("Section A", "Desc", "2", "bad-parent")
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := addSectionInteractive(&client.MockClient{}, cmd, 10)
		assert.ErrorContains(t, err, "invalid parent section id")
	})

	t.Run("cancelled", func(t *testing.T) {
		called := false
		mock := &client.MockClient{
			AddSectionFunc: func(ctx context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
				called = true
				return &data.Section{ID: 1, Name: req.Name}, nil
			},
		}
		cmd := setupAddTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Section A", "Desc", "2", "3").
			WithConfirmResponses(false)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := addSectionInteractive(mock, cmd, 10)
		assert.NoError(t, err)
		assert.False(t, called)
	})

	t.Run("client error", func(t *testing.T) {
		mock := &client.MockClient{
			AddSectionFunc: func(ctx context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
				return nil, errors.New("section boom")
			},
		}
		cmd := setupAddTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Section A", "Desc", "2", "3").
			WithConfirmResponses(true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := addSectionInteractive(mock, cmd, 10)
		assert.ErrorContains(t, err, "failed to create section")
	})
}

func TestAddSharedStepInteractive_ErrorBranches(t *testing.T) {
	t.Run("title required", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		p := interactive.NewMockPrompter().WithInputResponses("")
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := addSharedStepInteractive(&client.MockClient{}, cmd, 15)
		assert.ErrorContains(t, err, "shared step title is required")
	})

	t.Run("cancelled", func(t *testing.T) {
		called := false
		mock := &client.MockClient{
			AddSharedStepFunc: func(ctx context.Context, projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
				called = true
				return &data.SharedStep{ID: 1, Title: req.Title}, nil
			},
		}
		cmd := setupAddTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Shared A").
			WithConfirmResponses(false)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := addSharedStepInteractive(mock, cmd, 15)
		assert.NoError(t, err)
		assert.False(t, called)
	})

	t.Run("client error", func(t *testing.T) {
		mock := &client.MockClient{
			AddSharedStepFunc: func(ctx context.Context, projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
				return nil, errors.New("shared step boom")
			},
		}
		cmd := setupAddTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Shared A").
			WithConfirmResponses(true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := addSharedStepInteractive(mock, cmd, 15)
		assert.ErrorContains(t, err, "failed to create shared step")
	})
}

func TestAddSuiteInteractive_Success(t *testing.T) {
	mock := &client.MockClient{
		AddSuiteFunc: func(ctx context.Context, projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
			assert.Equal(t, int64(10), projectID)
			assert.Equal(t, "Suite Ok", req.Name)
			assert.Equal(t, "Suite Desc", req.Description)
			return &data.Suite{ID: 900, Name: req.Name}, nil
		},
	}
	cmd := setupAddTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Suite Ok", "Suite Desc").
		WithConfirmResponses(true)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := addSuiteInteractive(mock, cmd, 10)
	assert.NoError(t, err)
}

func TestAddCaseInteractive_Success(t *testing.T) {
	mock := &client.MockClient{
		AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(20), sectionID)
			assert.Equal(t, "Case Ok", req.Title)
			assert.Equal(t, int64(2), req.TypeID)
			assert.Equal(t, int64(3), req.PriorityID)
			assert.Equal(t, "REF-2", req.Refs)
			return &data.Case{ID: 901, Title: req.Title}, nil
		},
	}
	cmd := setupAddTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Case Ok", "REF-2").
		WithSelectResponses(interactive.SelectResponse{Index: 1}, interactive.SelectResponse{Index: 2}).
		WithConfirmResponses(true)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := addCaseInteractive(mock, cmd, 20)
	assert.NoError(t, err)
}

func TestAddRunInteractive_Success(t *testing.T) {
	mock := &client.MockClient{
		AddRunFunc: func(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, "Run Ok", req.Name)
			assert.Equal(t, "Run Desc", req.Description)
			assert.Equal(t, int64(44), req.SuiteID)
			assert.False(t, req.IncludeAll)
			return &data.Run{ID: 902, Name: req.Name}, nil
		},
	}
	cmd := setupAddTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Run Ok", "Run Desc", "44").
		WithConfirmResponses(false, true)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := addRunInteractive(mock, cmd, 30)
	assert.NoError(t, err)
}

func TestAddInteractive_ClientErrorBranches(t *testing.T) {
	t.Run("add suite interactive client error", func(t *testing.T) {
		mock := &client.MockClient{
			AddSuiteFunc: func(ctx context.Context, projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
				return nil, errors.New("suite boom")
			},
		}
		cmd := setupAddTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Suite", "Desc").
			WithConfirmResponses(true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := addSuiteInteractive(mock, cmd, 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create suite")
	})

	t.Run("add case interactive client error", func(t *testing.T) {
		mock := &client.MockClient{
			AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
				return nil, errors.New("case boom")
			},
		}
		cmd := setupAddTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Case", "REF").
			WithSelectResponses(interactive.SelectResponse{Index: 0}, interactive.SelectResponse{Index: 0}).
			WithConfirmResponses(true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := addCaseInteractive(mock, cmd, 20)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create case")
	})

	t.Run("add run interactive client error", func(t *testing.T) {
		mock := &client.MockClient{
			AddRunFunc: func(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error) {
				return nil, errors.New("run boom")
			},
		}
		cmd := setupAddTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Run", "Desc", "11").
			WithConfirmResponses(true, true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := addRunInteractive(mock, cmd, 30)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create run")
	})
}

func TestRunAddInteractive_Unsupported(t *testing.T) {
	err := runAddInteractive(&client.MockClient{}, &cobra.Command{}, "unknown", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interactive mode not supported")
}

func TestShouldAutoRunAddInteractive_Branches(t *testing.T) {
	t.Run("no prompter in context", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		assert.False(t, shouldAutoRunAddInteractive(cmd, "project", 0, false))
	})

	t.Run("json file disables auto interactive", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		assert.False(t, shouldAutoRunAddInteractive(cmd, "project", 0, true))
	})

	t.Run("project no changed flags", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		assert.True(t, shouldAutoRunAddInteractive(cmd, "project", 0, false))
	})

	t.Run("project with changed flag", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		require.NoError(t, cmd.Flags().Set("name", "Changed"))
		assert.False(t, shouldAutoRunAddInteractive(cmd, "project", 0, false))
	})

	t.Run("parent dependent endpoint without parent", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		assert.False(t, shouldAutoRunAddInteractive(cmd, "suite", 0, false))
	})

	t.Run("parent dependent endpoint with parent", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		assert.True(t, shouldAutoRunAddInteractive(cmd, "suite", 10, false))
	})

	t.Run("section with and without changed flags", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		assert.True(t, shouldAutoRunAddInteractive(cmd, "section", 10, false))
		require.NoError(t, cmd.Flags().Set("suite-id", "1"))
		assert.False(t, shouldAutoRunAddInteractive(cmd, "section", 10, false))
	})

	t.Run("case with and without changed flags", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		assert.True(t, shouldAutoRunAddInteractive(cmd, "case", 10, false))
		require.NoError(t, cmd.Flags().Set("title", "Case"))
		assert.False(t, shouldAutoRunAddInteractive(cmd, "case", 10, false))
	})

	t.Run("run with and without changed flags", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		assert.True(t, shouldAutoRunAddInteractive(cmd, "run", 10, false))
		require.NoError(t, cmd.Flags().Set("case-ids", "1,2"))
		assert.False(t, shouldAutoRunAddInteractive(cmd, "run", 10, false))
	})

	t.Run("shared-step with and without changed flags", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		assert.True(t, shouldAutoRunAddInteractive(cmd, "shared-step", 10, false))
		require.NoError(t, cmd.Flags().Set("title", "Step"))
		assert.False(t, shouldAutoRunAddInteractive(cmd, "shared-step", 10, false))
	})

	t.Run("unknown endpoint", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewMockPrompter()))
		assert.False(t, shouldAutoRunAddInteractive(cmd, "unknown", 10, false))
	})
}

func TestRunAddInteractive_SupportedEndpoints(t *testing.T) {
	t.Run("project", func(t *testing.T) {
		called := false
		mock := &client.MockClient{
			AddProjectFunc: func(ctx context.Context, req *data.AddProjectRequest) (*data.GetProjectResponse, error) {
				called = true
				return &data.GetProjectResponse{ID: 1, Name: req.Name}, nil
			},
		}
		cmd := setupAddTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Project", "Announcement").
			WithConfirmResponses(true, true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := runAddInteractive(mock, cmd, "project", 0)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("suite", func(t *testing.T) {
		called := false
		mock := &client.MockClient{
			AddSuiteFunc: func(ctx context.Context, projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
				called = true
				return &data.Suite{ID: 2, Name: req.Name}, nil
			},
		}
		cmd := setupAddTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Suite", "Desc").
			WithConfirmResponses(true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := runAddInteractive(mock, cmd, "suite", 10)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("section", func(t *testing.T) {
		called := false
		mock := &client.MockClient{
			AddSectionFunc: func(ctx context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
				called = true
				return &data.Section{ID: 3, Name: req.Name}, nil
			},
		}
		cmd := setupAddTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Section", "Desc", "1", "2").
			WithConfirmResponses(true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := runAddInteractive(mock, cmd, "section", 11)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("case", func(t *testing.T) {
		called := false
		mock := &client.MockClient{
			AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
				called = true
				return &data.Case{ID: 4, Title: req.Title}, nil
			},
		}
		cmd := setupAddTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Case", "REF").
			WithSelectResponses(interactive.SelectResponse{Index: 0}, interactive.SelectResponse{Index: 0}).
			WithConfirmResponses(true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := runAddInteractive(mock, cmd, "case", 12)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("run", func(t *testing.T) {
		called := false
		mock := &client.MockClient{
			AddRunFunc: func(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error) {
				called = true
				return &data.Run{ID: 5, Name: req.Name}, nil
			},
		}
		cmd := setupAddTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Run", "Desc", "44").
			WithConfirmResponses(false, true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := runAddInteractive(mock, cmd, "run", 13)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("shared-step", func(t *testing.T) {
		called := false
		mock := &client.MockClient{
			AddSharedStepFunc: func(ctx context.Context, projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
				called = true
				return &data.SharedStep{ID: 6, Title: req.Title}, nil
			},
		}
		cmd := setupAddTest(t, mock)
		p := interactive.NewMockPrompter().
			WithInputResponses("Shared").
			WithConfirmResponses(true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := runAddInteractive(mock, cmd, "shared-step", 14)
		assert.NoError(t, err)
		assert.True(t, called)
	})
}

func TestRunAddDryRun_SwitchEndpoints(t *testing.T) {
	cmd := setupAddTest(t, &client.MockClient{})
	_ = cmd.Flags().Set("name", "N")
	_ = cmd.Flags().Set("title", "T")
	_ = cmd.Flags().Set("description", "D")
	_ = cmd.Flags().Set("announcement", "A")
	_ = cmd.Flags().Set("show-announcement", "true")
	_ = cmd.Flags().Set("suite-id", "11")
	_ = cmd.Flags().Set("section-id", "22")
	_ = cmd.Flags().Set("milestone-id", "33")
	_ = cmd.Flags().Set("template-id", "44")
	_ = cmd.Flags().Set("type-id", "55")
	_ = cmd.Flags().Set("priority-id", "66")
	_ = cmd.Flags().Set("refs", "REF")
	_ = cmd.Flags().Set("comment", "C")
	_ = cmd.Flags().Set("status-id", "1")
	_ = cmd.Flags().Set("elapsed", "1m")
	_ = cmd.Flags().Set("defects", "BUG-1")
	_ = cmd.Flags().Set("assignedto-id", "99")
	_ = cmd.Flags().Set("case-ids", "1,2")
	_ = cmd.Flags().Set("include-all", "true")

	dr := output.NewDryRunPrinter("add")

	assert.NoError(t, runAddDryRun(cmd, dr, "project", 0, nil))
	assert.NoError(t, runAddDryRun(cmd, dr, "suite", 1, nil))
	assert.NoError(t, runAddDryRun(cmd, dr, "section", 1, nil))
	assert.NoError(t, runAddDryRun(cmd, dr, "case", 2, nil))
	assert.NoError(t, runAddDryRun(cmd, dr, "run", 1, nil))
	assert.NoError(t, runAddDryRun(cmd, dr, "result", 3, nil))
	assert.NoError(t, runAddDryRun(cmd, dr, "shared-step", 1, nil))

	err := runAddDryRun(cmd, dr, "attachment", 1, nil)
	assert.Error(t, err)
	err = runAddDryRun(cmd, dr, "bad-endpoint", 1, nil)
	assert.Error(t, err)
}

func TestRunAddInteractive_ParentIDRequired(t *testing.T) {
	cmd := setupAddTest(t, &client.MockClient{})

	err := runAddInteractive(&client.MockClient{}, cmd, "suite", 0)
	assert.Error(t, err)

	err = runAddInteractive(&client.MockClient{}, cmd, "section", 0)
	assert.Error(t, err)

	err = runAddInteractive(&client.MockClient{}, cmd, "case", 0)
	assert.Error(t, err)

	err = runAddInteractive(&client.MockClient{}, cmd, "run", 0)
	assert.Error(t, err)

	err = runAddInteractive(&client.MockClient{}, cmd, "shared-step", 0)
	assert.Error(t, err)
}

func TestAddSection_Success(t *testing.T) {
	mock := &client.MockClient{
		AddSectionFunc: func(ctx context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
			assert.Equal(t, int64(5), projectID)
			assert.Equal(t, "Section X", req.Name)
			return &data.Section{ID: 44, Name: req.Name}, nil
		},
	}
	cmd := setupAddTest(t, mock)
	_ = cmd.Flags().Set("name", "Section X")

	err := addSection(mock, cmd, 5, nil)
	assert.NoError(t, err)
}

func TestAddSection_JSONAndClientErrorBranches(t *testing.T) {
	t.Run("json success", func(t *testing.T) {
		mock := &client.MockClient{
			AddSectionFunc: func(ctx context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
				assert.Equal(t, int64(5), projectID)
				assert.Equal(t, "Json Section", req.Name)
				return &data.Section{ID: 45, Name: req.Name}, nil
			},
		}
		cmd := setupAddTest(t, mock)
		err := addSection(mock, cmd, 5, []byte(`{"name":"Json Section"}`))
		assert.NoError(t, err)
	})

	t.Run("client error", func(t *testing.T) {
		mock := &client.MockClient{
			AddSectionFunc: func(ctx context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
				return nil, errors.New("section boom")
			},
		}
		cmd := setupAddTest(t, mock)
		require.NoError(t, cmd.Flags().Set("name", "Section X"))
		err := addSection(mock, cmd, 5, nil)
		assert.ErrorContains(t, err, "failed to create section")
	})
}

func TestAddResultForCase_Success(t *testing.T) {
	mock := &client.MockClient{
		AddResultForCaseFunc: func(ctx context.Context, runID int64, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
			assert.Equal(t, int64(10), runID)
			assert.Equal(t, int64(20), caseID)
			assert.Equal(t, int64(1), req.StatusID)
			return &data.Result{ID: 999, StatusID: req.StatusID}, nil
		},
	}
	cmd := setupAddTest(t, mock)
	_ = cmd.Flags().Set("status-id", "1")

	err := addResultForCase(mock, cmd, 10, 20, nil)
	assert.NoError(t, err)
}

func TestAddResultForCase_JSONAndClientErrorBranches(t *testing.T) {
	t.Run("json success", func(t *testing.T) {
		mock := &client.MockClient{
			AddResultForCaseFunc: func(ctx context.Context, runID int64, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
				assert.Equal(t, int64(10), runID)
				assert.Equal(t, int64(20), caseID)
				assert.Equal(t, int64(2), req.StatusID)
				return &data.Result{ID: 1000, StatusID: req.StatusID}, nil
			},
		}
		cmd := setupAddTest(t, mock)
		err := addResultForCase(mock, cmd, 10, 20, []byte(`{"status_id":2}`))
		assert.NoError(t, err)
	})

	t.Run("json parse error", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		err := addResultForCase(&client.MockClient{}, cmd, 10, 20, []byte("{"))
		assert.ErrorContains(t, err, "JSON parse error")
	})

	t.Run("client error", func(t *testing.T) {
		mock := &client.MockClient{
			AddResultForCaseFunc: func(ctx context.Context, runID int64, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
				return nil, errors.New("result for case boom")
			},
		}
		cmd := setupAddTest(t, mock)
		require.NoError(t, cmd.Flags().Set("status-id", "1"))
		err := addResultForCase(mock, cmd, 10, 20, nil)
		assert.ErrorContains(t, err, "failed to add result")
	})
}

// TestAdd_Project_Success проверяет создание проекта
func TestAdd_Project_Success(t *testing.T) {
	mock := &client.MockClient{
		AddProjectFunc: func(ctx context.Context, req *data.AddProjectRequest) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: 999, Name: req.Name}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"project", "--name", "Test Project", "--announcement", "Test Announcement"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_Suite_Success проверяет создание сьюта
func TestAdd_Suite_Success(t *testing.T) {
	mock := &client.MockClient{
		AddSuiteFunc: func(ctx context.Context, projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
			assert.Equal(t, int64(1), projectID)
			return &data.Suite{ID: 100, Name: req.Name}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"suite", "1", "--name", "Test Suite", "--description", "Suite desc"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_Case_Success проверяет создание кейса
func TestAdd_Case_Success(t *testing.T) {
	mock := &client.MockClient{
		AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(100), sectionID)
			assert.Equal(t, "Test Case", req.Title)
			return &data.Case{ID: 999, Title: req.Title}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"case", "100", "--title", "Test Case", "--template-id", "1", "--priority-id", "2"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_Run_Success проверяет создание рана
func TestAdd_Run_Success(t *testing.T) {
	mock := &client.MockClient{
		AddRunFunc: func(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error) {
			assert.Equal(t, int64(1), projectID)
			assert.Equal(t, "Test Run", req.Name)
			return &data.Run{ID: 999, Name: req.Name}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"run", "1", "--name", "Test Run", "--suite-id", "100"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_Result_Success проверяет добавление результата
func TestAdd_Result_Success(t *testing.T) {
	mock := &client.MockClient{
		AddResultFunc: func(ctx context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error) {
			assert.Equal(t, int64(12345), testID)
			assert.Equal(t, int64(1), req.StatusID)
			return &data.Result{ID: 999, StatusID: req.StatusID}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"result", "12345", "--status-id", "1", "--comment", "Test passed"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAddHandlers_JSONParseErrors(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupAddTest(t, mock)
	badJSON := []byte("{")

	assert.ErrorContains(t, addProject(mock, cmd, badJSON), "JSON parse error")
	assert.ErrorContains(t, addSuite(mock, cmd, 1, badJSON), "JSON parse error")
	assert.ErrorContains(t, addCase(mock, cmd, 1, badJSON), "JSON parse error")
	assert.ErrorContains(t, addRun(mock, cmd, 1, badJSON), "JSON parse error")
	assert.ErrorContains(t, addResult(mock, cmd, 1, badJSON), "JSON parse error")
	assert.ErrorContains(t, addSharedStep(mock, cmd, 1, badJSON), "JSON parse error")
}

func TestAddHandlers_ClientErrors(t *testing.T) {
	mock := &client.MockClient{
		AddProjectFunc: func(ctx context.Context, req *data.AddProjectRequest) (*data.GetProjectResponse, error) {
			return nil, errors.New("project boom")
		},
		AddSuiteFunc: func(ctx context.Context, projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
			return nil, errors.New("suite boom")
		},
		AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			return nil, errors.New("case boom")
		},
		AddRunFunc: func(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error) {
			return nil, errors.New("run boom")
		},
		AddResultFunc: func(ctx context.Context, testID int64, req *data.AddResultRequest) (*data.Result, error) {
			return nil, errors.New("result boom")
		},
		AddSharedStepFunc: func(ctx context.Context, projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
			return nil, errors.New("shared step boom")
		},
	}

	t.Run("add project client error", func(t *testing.T) {
		cmd := setupAddTest(t, mock)
		require.NoError(t, cmd.Flags().Set("name", "Project X"))
		err := addProject(mock, cmd, nil)
		assert.ErrorContains(t, err, "failed to create project")
	})

	t.Run("add suite client error", func(t *testing.T) {
		cmd := setupAddTest(t, mock)
		require.NoError(t, cmd.Flags().Set("name", "Suite X"))
		err := addSuite(mock, cmd, 10, nil)
		assert.ErrorContains(t, err, "failed to create suite")
	})

	t.Run("add case client error", func(t *testing.T) {
		cmd := setupAddTest(t, mock)
		require.NoError(t, cmd.Flags().Set("title", "Case X"))
		err := addCase(mock, cmd, 10, nil)
		assert.ErrorContains(t, err, "failed to create case")
	})

	t.Run("add run client error", func(t *testing.T) {
		cmd := setupAddTest(t, mock)
		require.NoError(t, cmd.Flags().Set("name", "Run X"))
		err := addRun(mock, cmd, 10, nil)
		assert.ErrorContains(t, err, "failed to create run")
	})

	t.Run("add result client error", func(t *testing.T) {
		cmd := setupAddTest(t, mock)
		require.NoError(t, cmd.Flags().Set("status-id", "1"))
		err := addResult(mock, cmd, 10, nil)
		assert.ErrorContains(t, err, "failed to add result")
	})

	t.Run("add shared step client error", func(t *testing.T) {
		cmd := setupAddTest(t, mock)
		require.NoError(t, cmd.Flags().Set("title", "Step X"))
		err := addSharedStep(mock, cmd, 10, nil)
		assert.ErrorContains(t, err, "failed to create shared step")
	})
}

func TestAddSharedStep_RequiresTitle(t *testing.T) {
	cmd := setupAddTest(t, &client.MockClient{})
	err := addSharedStep(&client.MockClient{}, cmd, 1, nil)
	assert.ErrorContains(t, err, "--title is required")
}

func TestRunAddDryRun_MissingParentIDBranches(t *testing.T) {
	cmd := setupAddTest(t, &client.MockClient{})
	dr := output.NewDryRunPrinter("add")

	assert.ErrorContains(t, runAddDryRun(cmd, dr, "suite", 0, nil), "project_id required")
	assert.ErrorContains(t, runAddDryRun(cmd, dr, "section", 0, nil), "project_id required")
	assert.ErrorContains(t, runAddDryRun(cmd, dr, "case", 0, nil), "section_id required")
	assert.ErrorContains(t, runAddDryRun(cmd, dr, "run", 0, nil), "project_id required")
	assert.ErrorContains(t, runAddDryRun(cmd, dr, "result", 0, nil), "test_id required")
	assert.ErrorContains(t, runAddDryRun(cmd, dr, "shared-step", 0, nil), "project_id required")
	assert.ErrorContains(t, runAddDryRun(cmd, dr, "attachment", 1, nil), "specific attachment subcommand")
}

func TestRunAddDryRun_JSONBranches(t *testing.T) {
	cmd := setupAddTest(t, &client.MockClient{})
	dr := output.NewDryRunPrinter("add-json")

	assert.NoError(t, runAddDryRun(cmd, dr, "project", 0, []byte(`{"name":"P"}`)))
	assert.NoError(t, runAddDryRun(cmd, dr, "suite", 1, []byte(`{"name":"S"}`)))
	assert.NoError(t, runAddDryRun(cmd, dr, "section", 1, []byte(`{"name":"Sec"}`)))
	assert.NoError(t, runAddDryRun(cmd, dr, "case", 2, []byte(`{"title":"Case"}`)))
	assert.NoError(t, runAddDryRun(cmd, dr, "run", 1, []byte(`{"name":"Run"}`)))
	assert.NoError(t, runAddDryRun(cmd, dr, "result", 3, []byte(`{"status_id":1}`)))
	assert.NoError(t, runAddDryRun(cmd, dr, "shared-step", 1, []byte(`{"title":"Step"}`)))
}

func TestRunAdd_OrchestrationErrors(t *testing.T) {
	cmd := setupAddTest(t, &client.MockClient{})

	t.Run("endpoint required", func(t *testing.T) {
		err := runAdd(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "endpoint required")
	})

	t.Run("invalid id", func(t *testing.T) {
		err := runAdd(cmd, []string{"suite", "bad-id"})
		assert.Error(t, err)
	})

	t.Run("json file read error", func(t *testing.T) {
		require.NoError(t, cmd.Flags().Set("json-file", "/tmp/not-existing-gotr-add-test.json"))
		err := runAdd(cmd, []string{"project"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JSON file read error")
		require.NoError(t, cmd.Flags().Set("json-file", ""))
	})
}

func TestRunAdd_OrchestrationBranches(t *testing.T) {
	t.Run("dry-run branch", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		cmd.Flags().Bool("dry-run", false, "")
		require.NoError(t, cmd.Flags().Set("dry-run", "true"))
		require.NoError(t, cmd.Flags().Set("name", "Dry Project"))

		err := runAdd(cmd, []string{"project"})
		assert.NoError(t, err)
	})

	t.Run("interactive flag branch", func(t *testing.T) {
		mock := &client.MockClient{
			AddProjectFunc: func(ctx context.Context, req *data.AddProjectRequest) (*data.GetProjectResponse, error) {
				return &data.GetProjectResponse{ID: 900, Name: req.Name}, nil
			},
		}
		cmd := setupAddTest(t, mock)
		cmd.Flags().BoolP("interactive", "i", false, "")
		require.NoError(t, cmd.Flags().Set("interactive", "true"))
		p := interactive.NewMockPrompter().
			WithInputResponses("Project I", "Announcement").
			WithConfirmResponses(true, true)
		cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

		err := runAdd(cmd, []string{"project"})
		assert.NoError(t, err)
	})

	t.Run("json-file success branch", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "add-json-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		_, _ = tmpFile.WriteString(`{"name":"FromFile"}`)
		_ = tmpFile.Close()

		mock := &client.MockClient{
			AddProjectFunc: func(ctx context.Context, req *data.AddProjectRequest) (*data.GetProjectResponse, error) {
				return &data.GetProjectResponse{ID: 901, Name: req.Name}, nil
			},
		}
		cmd := setupAddTest(t, mock)
		require.NoError(t, cmd.Flags().Set("json-file", tmpFile.Name()))

		err = runAdd(cmd, []string{"project"})
		assert.NoError(t, err)
	})

	t.Run("result-for-case success", func(t *testing.T) {
		mock := &client.MockClient{
			AddResultForCaseFunc: func(ctx context.Context, runID int64, caseID int64, req *data.AddResultRequest) (*data.Result, error) {
				assert.Equal(t, int64(10), runID)
				assert.Equal(t, int64(20), caseID)
				assert.Equal(t, int64(1), req.StatusID)
				return &data.Result{ID: 902, StatusID: req.StatusID}, nil
			},
		}
		cmd := setupAddTest(t, mock)
		require.NoError(t, cmd.Flags().Set("status-id", "1"))

		err := runAdd(cmd, []string{"result-for-case", "10", "20"})
		assert.NoError(t, err)
	})

	t.Run("result-for-case invalid case id", func(t *testing.T) {
		cmd := setupAddTest(t, &client.MockClient{})
		err := runAdd(cmd, []string{"result-for-case", "10", "bad-case"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "case_id")
	})
}

func TestRunAdd_AutoInteractive_Project(t *testing.T) {
	mock := &client.MockClient{
		AddProjectFunc: func(ctx context.Context, req *data.AddProjectRequest) (*data.GetProjectResponse, error) {
			return &data.GetProjectResponse{ID: 500, Name: req.Name}, nil
		},
	}
	cmd := setupAddTest(t, mock)
	p := interactive.NewMockPrompter().
		WithInputResponses("Auto Project", "Announcement").
		WithConfirmResponses(true, true)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))

	err := runAdd(cmd, []string{"project"})
	assert.NoError(t, err)
}

func TestResolveAddParentID_CaseMultiSelection(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter().
		WithSelectResponses(
			interactive.SelectResponse{Index: 0}, // project
			interactive.SelectResponse{Index: 1}, // suite
			interactive.SelectResponse{Index: 1}, // section
		),
	)

	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 10, Name: "P1"},
				{ID: 20, Name: "P2"},
			}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{
				{ID: 100, Name: "S1"},
				{ID: 200, Name: "S2"},
			}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{
				{ID: 1000, Name: "Sec1"},
				{ID: 2000, Name: "Sec2"},
			}, nil
		},
	}

	got, err := resolveAddParentID(ctx, interactive.PrompterFromContext(ctx), mock, "case", 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2000), got)
}

func TestResolveAddParentID_CaseSuiteFetchError(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}),
	)

	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 10, Name: "P1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return nil, errors.New("suite boom")
		},
	}

	_, err := resolveAddParentID(ctx, interactive.PrompterFromContext(ctx), mock, "case", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get suites for project")
}

func TestResolveAddParentID_ShortCircuitBranches(t *testing.T) {
	mock := &client.MockClient{}

	t.Run("current id already set", func(t *testing.T) {
		got, err := resolveAddParentID(context.Background(), nil, mock, "suite", 99)
		assert.NoError(t, err)
		assert.Equal(t, int64(99), got)
	})

	t.Run("no prompter in context", func(t *testing.T) {
		got, err := resolveAddParentID(context.Background(), nil, mock, "suite", 0)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), got)
	})

	t.Run("default endpoint", func(t *testing.T) {
		ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter())
		got, err := resolveAddParentID(ctx, interactive.PrompterFromContext(ctx), mock, "project", 0)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), got)
	})
}

func TestResolveAddParentID_ProjectSelectionEndpoints(t *testing.T) {
	for _, endpoint := range []string{"suite", "section", "run", "shared-step"} {
		t.Run(endpoint, func(t *testing.T) {
			ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter().
				WithSelectResponses(interactive.SelectResponse{Index: 0}),
			)

			mock := &client.MockClient{
				GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
					return data.GetProjectsResponse{{ID: 77, Name: "P77"}}, nil
				},
			}

			got, err := resolveAddParentID(ctx, interactive.PrompterFromContext(ctx), mock, endpoint, 0)
			assert.NoError(t, err)
			assert.Equal(t, int64(77), got)
		})
	}
}

func TestResolveAddParentID_CaseSingleSuiteSingleSection(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}),
	)

	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 10, Name: "P1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 100, Name: "S1"}}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{{ID: 2000, Name: "Sec1"}}, nil
		},
	}

	got, err := resolveAddParentID(ctx, interactive.PrompterFromContext(ctx), mock, "case", 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2000), got)
}

func TestResolveAddParentID_CaseSectionsFetchError(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}),
	)

	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 10, Name: "P1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 100, Name: "S1"}}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			return nil, errors.New("sections boom")
		},
	}

	_, err := resolveAddParentID(ctx, interactive.PrompterFromContext(ctx), mock, "case", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get sections for project")
}

func TestResolveAddParentID_CaseSelectSuiteError(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}),
	)

	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 10, Name: "P1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 100, Name: "S1"}, {ID: 200, Name: "S2"}}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{{ID: 1, Name: "Sec"}}, nil
		},
	}

	_, err := resolveAddParentID(ctx, interactive.PrompterFromContext(ctx), mock, "case", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "select suite")
}

func TestResolveAddParentID_CaseSelectSectionError(t *testing.T) {
	ctx := interactive.WithPrompter(context.Background(), interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 0}, interactive.SelectResponse{Index: 0}),
	)

	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 10, Name: "P1"}}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{{ID: 100, Name: "S1"}, {ID: 200, Name: "S2"}}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			return data.GetSectionsResponse{{ID: 1, Name: "Sec1"}, {ID: 2, Name: "Sec2"}}, nil
		},
	}

	_, err := resolveAddParentID(ctx, interactive.PrompterFromContext(ctx), mock, "case", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "select section")
}

// TestAdd_SharedStep_Success проверяет создание shared step
func TestAdd_SharedStep_Success(t *testing.T) {
	mock := &client.MockClient{
		AddSharedStepFunc: func(ctx context.Context, projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
			assert.Equal(t, int64(1), projectID)
			return &data.SharedStep{ID: 999, Title: req.Title}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"shared-step", "1", "--title", "Test Shared Step"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_NoEndpoint проверяет ошибку при отсутствии endpoint
func TestAdd_NoEndpoint(t *testing.T) {
	mock := &client.MockClient{}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint")
}

func TestAdd_Section_NonInteractive_AutoWizard_NoMutatingCall(t *testing.T) {
	called := false
	mock := &client.MockClient{
		AddSectionFunc: func(ctx context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
			called = true
			return &data.Section{ID: 1, Name: "section"}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{"section", "1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.False(t, called)
}

func TestAdd_SharedStep_NonInteractive_AutoWizard_NoMutatingCall(t *testing.T) {
	called := false
	mock := &client.MockClient{
		AddSharedStepFunc: func(ctx context.Context, projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
			called = true
			return &data.SharedStep{ID: 1, Title: "step"}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{"shared-step", "1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.False(t, called)
}

func TestAdd_Suite_AutoSelectProject_WhenIDMissing(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 77, Name: "Project 77"}}, nil
		},
		AddSuiteFunc: func(ctx context.Context, projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
			assert.Equal(t, int64(77), projectID)
			assert.Equal(t, "Auto Suite", req.Name)
			return &data.Suite{ID: 100, Name: req.Name}, nil
		},
	}

	p := interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})
	cmd := setupAddTest(t, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))
	cmd.SetArgs([]string{"suite", "--name", "Auto Suite"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAdd_Case_AutoSelectSectionChain_WhenIDMissing(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 10, Name: "Project A"},
				{ID: 20, Name: "Project B"},
			}, nil
		},
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(20), projectID)
			return data.GetSuitesResponse{
				{ID: 501, Name: "Suite 1"},
				{ID: 502, Name: "Suite 2"},
			}, nil
		},
		GetSectionsFunc: func(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
			assert.Equal(t, int64(20), projectID)
			assert.Equal(t, int64(502), suiteID)
			return data.GetSectionsResponse{
				{ID: 9001, Name: "Section A"},
				{ID: 9002, Name: "Section B"},
			}, nil
		},
		AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(9002), sectionID)
			assert.Equal(t, "Auto Case", req.Title)
			return &data.Case{ID: 111, Title: req.Title}, nil
		},
	}

	p := interactive.NewMockPrompter().
		WithSelectResponses(interactive.SelectResponse{Index: 1}).
		WithSelectResponses(interactive.SelectResponse{Index: 1}).
		WithSelectResponses(interactive.SelectResponse{Index: 1})
	cmd := setupAddTest(t, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), p))
	cmd.SetArgs([]string{"case", "--title", "Auto Case"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestAdd_Suite_NonInteractive_MissingID_NoMutatingCall(t *testing.T) {
	called := false
	mock := &client.MockClient{
		GetProjectsFunc: func(ctx context.Context) (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{{ID: 1, Name: "Project 1"}}, nil
		},
		AddSuiteFunc: func(ctx context.Context, projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
			called = true
			return &data.Suite{ID: 1, Name: req.Name}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetContext(interactive.WithPrompter(cmd.Context(), interactive.NewNonInteractivePrompter()))
	cmd.SetArgs([]string{"suite", "--name", "No ID"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
	assert.False(t, called)
}

// TestAdd_UnsupportedEndpoint проверяет ошибку при неподдерживаемом endpoint
func TestAdd_UnsupportedEndpoint(t *testing.T) {
	mock := &client.MockClient{}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"unsupported", "1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

// ==================== Attachment Tests ====================

// TestAdd_AttachmentCase_Success проверяет добавление вложения к кейсу
func TestAdd_AttachmentCase_Success(t *testing.T) {
	// Создаем временный файл
	tmpFile, err := os.CreateTemp("", "test-attachment-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("test content")
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToCaseFunc: func(ctx context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error) {
			assert.Equal(t, int64(12345), caseID)
			assert.Equal(t, tmpFile.Name(), filePath)
			return &data.AttachmentResponse{AttachmentID: 999, URL: "https://example.com/attachment/999"}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "case", "12345", tmpFile.Name()})

	err = cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_AttachmentPlan_Success проверяет добавление вложения к плану
func TestAdd_AttachmentPlan_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-plan-*.pdf")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("plan content")
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToPlanFunc: func(ctx context.Context, planID int64, filePath string) (*data.AttachmentResponse, error) {
			assert.Equal(t, int64(100), planID)
			assert.Equal(t, tmpFile.Name(), filePath)
			return &data.AttachmentResponse{AttachmentID: 888, URL: "https://example.com/attachment/888"}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "plan", "100", tmpFile.Name()})

	err = cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_AttachmentPlanEntry_Success проверяет добавление вложения к plan entry
func TestAdd_AttachmentPlanEntry_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-entry-*.doc")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("entry content")
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToPlanEntryFunc: func(ctx context.Context, planID int64, entryID string, filePath string) (*data.AttachmentResponse, error) {
			assert.Equal(t, int64(200), planID)
			assert.Equal(t, "entry-abc123", entryID)
			assert.Equal(t, tmpFile.Name(), filePath)
			return &data.AttachmentResponse{AttachmentID: 777, URL: "https://example.com/attachment/777"}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "plan-entry", "200", "entry-abc123", tmpFile.Name()})

	err = cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_AttachmentResult_Success проверяет добавление вложения к результату
func TestAdd_AttachmentResult_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-result-*.log")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("result log content")
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToResultFunc: func(ctx context.Context, resultID int64, filePath string) (*data.AttachmentResponse, error) {
			assert.Equal(t, int64(98765), resultID)
			assert.Equal(t, tmpFile.Name(), filePath)
			return &data.AttachmentResponse{AttachmentID: 666, URL: "https://example.com/attachment/666"}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "result", "98765", tmpFile.Name()})

	err = cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_AttachmentRun_Success проверяет добавление вложения к рану
func TestAdd_AttachmentRun_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-run-*.png")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("png binary content")
	tmpFile.Close()

	mock := &client.MockClient{
		AddAttachmentToRunFunc: func(ctx context.Context, runID int64, filePath string) (*data.AttachmentResponse, error) {
			assert.Equal(t, int64(555), runID)
			assert.Equal(t, tmpFile.Name(), filePath)
			return &data.AttachmentResponse{AttachmentID: 555, URL: "https://example.com/attachment/555"}, nil
		},
	}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "run", "555", tmpFile.Name()})

	err = cmd.Execute()
	assert.NoError(t, err)
}

// TestAdd_Attachment_MissingArgs проверяет ошибку при недостаточных аргументах
func TestAdd_Attachment_MissingArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "attachment type")
}

// TestAdd_Attachment_InvalidCaseID проверяет ошибку при неверном case_id
func TestAdd_Attachment_InvalidCaseID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "case", "invalid", "/tmp/test.txt"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "case_id")
}

// TestAdd_Attachment_UnsupportedType проверяет ошибку при неподдерживаемом типе
func TestAdd_Attachment_UnsupportedType(t *testing.T) {
	mock := &client.MockClient{}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "invalid-type", "123", "/tmp/test.txt"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported attachment type")
}

// TestAdd_Attachment_MissingFilePath проверяет ошибку при отсутствии пути к файлу
func TestAdd_Attachment_MissingFilePath(t *testing.T) {
	mock := &client.MockClient{}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "case", "12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "usage")
}

// TestAdd_Attachment_MissingPlanEntryArgs проверяет ошибку при недостаточных аргументах для plan-entry
func TestAdd_Attachment_MissingPlanEntryArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := setupAddTest(t, mock)
	cmd.SetArgs([]string{"attachment", "plan-entry", "100", "entry-id"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "usage")
}

func TestAdd_Attachment_MissingPlanFilePath(t *testing.T) {
	cmd := setupAddTest(t, &client.MockClient{})
	cmd.SetArgs([]string{"attachment", "plan", "100"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "usage")
}

func TestAdd_Attachment_MissingResultFilePath(t *testing.T) {
	cmd := setupAddTest(t, &client.MockClient{})
	cmd.SetArgs([]string{"attachment", "result", "123"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "usage")
}

func TestAdd_Attachment_MissingRunFilePath(t *testing.T) {
	cmd := setupAddTest(t, &client.MockClient{})
	cmd.SetArgs([]string{"attachment", "run", "555"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "usage")
}

func TestAdd_Attachment_InvalidPlanID(t *testing.T) {
	cmd := setupAddTest(t, &client.MockClient{})
	cmd.SetArgs([]string{"attachment", "plan", "invalid", "/tmp/test.txt"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan_id")
}

func TestAdd_Attachment_InvalidResultID(t *testing.T) {
	cmd := setupAddTest(t, &client.MockClient{})
	cmd.SetArgs([]string{"attachment", "result", "invalid", "/tmp/test.txt"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "result_id")
}

func TestAdd_Attachment_InvalidRunID(t *testing.T) {
	cmd := setupAddTest(t, &client.MockClient{})
	cmd.SetArgs([]string{"attachment", "run", "invalid", "/tmp/test.txt"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "run_id")
}

func TestAdd_AttachmentPlanEntry_InvalidPlanID(t *testing.T) {
	cmd := setupAddTest(t, &client.MockClient{})
	cmd.SetArgs([]string{"attachment", "plan-entry", "invalid", "entry-1", "/tmp/test.txt"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan_id")
}

func TestAddAttachmentHelpers_DryRun(t *testing.T) {
	cmd := setupAddTest(t, &client.MockClient{})
	cmd.Flags().Bool("dry-run", false, "")
	require.NoError(t, cmd.Flags().Set("dry-run", "true"))

	err := addAttachmentToCase(&client.MockClient{}, cmd, 1, "/tmp/does-not-exist")
	assert.NoError(t, err)

	err = addAttachmentToPlan(&client.MockClient{}, cmd, 2, "/tmp/does-not-exist")
	assert.NoError(t, err)

	err = addAttachmentToPlanEntry(&client.MockClient{}, cmd, 3, "entry-1", "/tmp/does-not-exist")
	assert.NoError(t, err)

	err = addAttachmentToResult(&client.MockClient{}, cmd, 4, "/tmp/does-not-exist")
	assert.NoError(t, err)

	err = addAttachmentToRun(&client.MockClient{}, cmd, 5, "/tmp/does-not-exist")
	assert.NoError(t, err)
}

func TestAddAttachmentHelpers_FileNotFound(t *testing.T) {
	cmd := setupAddTest(t, &client.MockClient{})

	err := addAttachmentToCase(&client.MockClient{}, cmd, 1, "/tmp/does-not-exist")
	assert.ErrorContains(t, err, "file not found")

	err = addAttachmentToPlan(&client.MockClient{}, cmd, 2, "/tmp/does-not-exist")
	assert.ErrorContains(t, err, "file not found")

	err = addAttachmentToPlanEntry(&client.MockClient{}, cmd, 3, "entry-1", "/tmp/does-not-exist")
	assert.ErrorContains(t, err, "file not found")

	err = addAttachmentToResult(&client.MockClient{}, cmd, 4, "/tmp/does-not-exist")
	assert.ErrorContains(t, err, "file not found")

	err = addAttachmentToRun(&client.MockClient{}, cmd, 5, "/tmp/does-not-exist")
	assert.ErrorContains(t, err, "file not found")
}

func TestAddAttachmentHelpers_ClientErrors(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "attachment-helper-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.WriteString("content")
	_ = tmpFile.Close()

	t.Run("case", func(t *testing.T) {
		mock := &client.MockClient{
			AddAttachmentToCaseFunc: func(ctx context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error) {
				return nil, errors.New("case attachment boom")
			},
		}
		cmd := setupAddTest(t, mock)
		err := addAttachmentToCase(mock, cmd, 1, tmpFile.Name())
		assert.ErrorContains(t, err, "failed to add attachment to case")
	})

	t.Run("plan", func(t *testing.T) {
		mock := &client.MockClient{
			AddAttachmentToPlanFunc: func(ctx context.Context, planID int64, filePath string) (*data.AttachmentResponse, error) {
				return nil, errors.New("plan attachment boom")
			},
		}
		cmd := setupAddTest(t, mock)
		err := addAttachmentToPlan(mock, cmd, 2, tmpFile.Name())
		assert.ErrorContains(t, err, "failed to add attachment to plan")
	})

	t.Run("plan-entry", func(t *testing.T) {
		mock := &client.MockClient{
			AddAttachmentToPlanEntryFunc: func(ctx context.Context, planID int64, entryID string, filePath string) (*data.AttachmentResponse, error) {
				return nil, errors.New("plan entry attachment boom")
			},
		}
		cmd := setupAddTest(t, mock)
		err := addAttachmentToPlanEntry(mock, cmd, 3, "entry-1", tmpFile.Name())
		assert.ErrorContains(t, err, "failed to add attachment to plan entry")
	})

	t.Run("result", func(t *testing.T) {
		mock := &client.MockClient{
			AddAttachmentToResultFunc: func(ctx context.Context, resultID int64, filePath string) (*data.AttachmentResponse, error) {
				return nil, errors.New("result attachment boom")
			},
		}
		cmd := setupAddTest(t, mock)
		err := addAttachmentToResult(mock, cmd, 4, tmpFile.Name())
		assert.ErrorContains(t, err, "failed to add attachment to result")
	})

	t.Run("run", func(t *testing.T) {
		mock := &client.MockClient{
			AddAttachmentToRunFunc: func(ctx context.Context, runID int64, filePath string) (*data.AttachmentResponse, error) {
				return nil, errors.New("run attachment boom")
			},
		}
		cmd := setupAddTest(t, mock)
		err := addAttachmentToRun(mock, cmd, 5, tmpFile.Name())
		assert.ErrorContains(t, err, "failed to add attachment to run")
	})
}
