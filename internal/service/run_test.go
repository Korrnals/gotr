// internal/service/run_test.go
package service

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// mockRunClient — мок clientа для тестирования RunService
type mockRunClient struct {
	getRun    func(int64) (*data.Run, error)
	getRuns   func(int64) (data.GetRunsResponse, error)
	addRun    func(int64, *data.AddRunRequest) (*data.Run, error)
	updateRun func(int64, *data.UpdateRunRequest) (*data.Run, error)
	closeRun  func(int64) (*data.Run, error)
	deleteRun func(int64) error
}

func (m *mockRunClient) GetRun(id int64) (*data.Run, error) {
	return m.getRun(id)
}

func (m *mockRunClient) GetRuns(projectID int64) (data.GetRunsResponse, error) {
	return m.getRuns(projectID)
}

func (m *mockRunClient) AddRun(projectID int64, req *data.AddRunRequest) (*data.Run, error) {
	return m.addRun(projectID, req)
}

func (m *mockRunClient) UpdateRun(runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
	return m.updateRun(runID, req)
}

func (m *mockRunClient) CloseRun(runID int64) (*data.Run, error) {
	return m.closeRun(runID)
}

func (m *mockRunClient) DeleteRun(runID int64) error {
	return m.deleteRun(runID)
}

func TestRunService_validateID(t *testing.T) {
	svc := &RunService{}

	tests := []struct {
		name      string
		id        int64
		fieldName string
		wantErr   bool
	}{
		{"valid positive ID", 123, "run_id", false},
		{"zero ID", 0, "run_id", true},
		{"negative ID", -1, "project_id", true},
		{"large valid ID", 999999, "run_id", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateID(tt.id, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.fieldName)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunService_validateCreateRequest(t *testing.T) {
	svc := &RunService{}

	tests := []struct {
		name    string
		req     *data.AddRunRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "nil",
		},
		{
			name: "empty name",
			req: &data.AddRunRequest{
				Name:    "",
				SuiteID: 100,
			},
			wantErr: true,
			errMsg:  "name",
		},
		{
			name: "zero suite_id",
			req: &data.AddRunRequest{
				Name:    "Test Run",
				SuiteID: 0,
			},
			wantErr: true,
			errMsg:  "suite_id",
		},
		{
			name: "negative suite_id",
			req: &data.AddRunRequest{
				Name:    "Test Run",
				SuiteID: -5,
			},
			wantErr: true,
			errMsg:  "suite_id",
		},
		{
			name: "valid request",
			req: &data.AddRunRequest{
				Name:    "Test Run",
				SuiteID: 100,
			},
			wantErr: false,
		},
		{
			name: "valid request with all fields",
			req: &data.AddRunRequest{
				Name:        "Full Test Run",
				Description: "Description",
				SuiteID:     100,
				MilestoneID: 50,
				AssignedTo:  10,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateCreateRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunService_ParseID(t *testing.T) {
	svc := &RunService{}

	tests := []struct {
		name     string
		args     []string
		index    int
		wantID   int64
		wantErr  bool
		errMatch string
	}{
		{
			name:   "valid ID",
			args:   []string{"12345"},
			index:  0,
			wantID: 12345,
		},
		{
			name:     "index out of range",
			args:     []string{"123"},
			index:    5,
			wantErr:  true,
			errMatch: "missing ID argument",
		},
		{
			name:     "invalid string",
			args:     []string{"abc"},
			index:    0,
			wantErr:  true,
			errMatch: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			id, err := svc.ParseID(ctx, tt.args, tt.index)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMatch)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, id)
			}
		})
	}
}

func TestRunService_Methods(t *testing.T) {
	ctx := context.Background()
	mock := &client.MockClient{}
	svc := NewRunServiceFromInterface(mock)

	t.Run("Get invalid id", func(t *testing.T) {
		_, err := svc.Get(ctx, 0)
		assert.Error(t, err)
	})

	t.Run("Get success", func(t *testing.T) {
		mock.GetRunFunc = func(ctx context.Context, runID int64) (*data.Run, error) {
			return &data.Run{ID: runID, Name: "run"}, nil
		}
		run, err := svc.Get(ctx, 11)
		assert.NoError(t, err)
		assert.Equal(t, int64(11), run.ID)
	})

	t.Run("GetByProject success", func(t *testing.T) {
		mock.GetRunsFunc = func(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
			return data.GetRunsResponse{{ID: 1, ProjectID: projectID}}, nil
		}
		runs, err := svc.GetByProject(ctx, 22)
		assert.NoError(t, err)
		assert.Len(t, runs, 1)
	})

	t.Run("GetByProject invalid id", func(t *testing.T) {
		_, err := svc.GetByProject(ctx, 0)
		assert.Error(t, err)
	})

	t.Run("Create validation error", func(t *testing.T) {
		_, err := svc.Create(ctx, 22, &data.AddRunRequest{Name: "", SuiteID: 0})
		assert.Error(t, err)
	})

	t.Run("Create invalid project id", func(t *testing.T) {
		_, err := svc.Create(ctx, 0, &data.AddRunRequest{Name: "n", SuiteID: 1})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "project_id")
	})

	t.Run("Create client error", func(t *testing.T) {
		mock.AddRunFunc = func(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error) {
			return nil, errors.New("add failed")
		}
		_, err := svc.Create(ctx, 22, &data.AddRunRequest{Name: "n", SuiteID: 1})
		assert.Error(t, err)
	})

	t.Run("Create success", func(t *testing.T) {
		mock.AddRunFunc = func(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error) {
			return &data.Run{ID: 33, ProjectID: projectID, Name: req.Name}, nil
		}
		run, err := svc.Create(ctx, 22, &data.AddRunRequest{Name: "smoke", SuiteID: 1})
		assert.NoError(t, err)
		assert.Equal(t, int64(33), run.ID)
	})

	t.Run("Update invalid id", func(t *testing.T) {
		_, err := svc.Update(ctx, 0, &data.UpdateRunRequest{})
		assert.Error(t, err)
	})

	t.Run("Update success", func(t *testing.T) {
		mock.UpdateRunFunc = func(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			return &data.Run{ID: runID, Name: "updated"}, nil
		}
		run, err := svc.Update(ctx, 44, &data.UpdateRunRequest{})
		assert.NoError(t, err)
		assert.Equal(t, int64(44), run.ID)
	})

	t.Run("Update client error", func(t *testing.T) {
		mock.UpdateRunFunc = func(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
			return nil, errors.New("update failed")
		}
		_, err := svc.Update(ctx, 44, &data.UpdateRunRequest{})
		assert.Error(t, err)
	})

	t.Run("Close success", func(t *testing.T) {
		mock.CloseRunFunc = func(ctx context.Context, runID int64) (*data.Run, error) {
			return &data.Run{ID: runID, CompletedOn: 123}, nil
		}
		run, err := svc.Close(ctx, 55)
		assert.NoError(t, err)
		assert.Equal(t, int64(55), run.ID)
	})

	t.Run("Close invalid id", func(t *testing.T) {
		_, err := svc.Close(ctx, 0)
		assert.Error(t, err)
	})

	t.Run("Close client error", func(t *testing.T) {
		mock.CloseRunFunc = func(ctx context.Context, runID int64) (*data.Run, error) {
			return nil, errors.New("close failed")
		}
		_, err := svc.Close(ctx, 55)
		assert.Error(t, err)
	})

	t.Run("Delete success", func(t *testing.T) {
		mock.DeleteRunFunc = func(ctx context.Context, runID int64) error {
			return nil
		}
		err := svc.Delete(ctx, 66)
		assert.NoError(t, err)
	})

	t.Run("Delete error", func(t *testing.T) {
		mock.DeleteRunFunc = func(ctx context.Context, runID int64) error {
			return errors.New("delete failed")
		}
		err := svc.Delete(ctx, 66)
		assert.Error(t, err)
	})

	t.Run("Delete invalid id", func(t *testing.T) {
		err := svc.Delete(ctx, 0)
		assert.Error(t, err)
	})
}

func TestRunService_Create_NilRequest_DoesNotPanic(t *testing.T) {
	mock := &client.MockClient{}
	svc := NewRunServiceFromInterface(mock)

	assert.NotPanics(t, func() {
		run, err := svc.Create(context.Background(), 100, nil)
		assert.Nil(t, run)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request validation")
	})
}

func TestRunService_ConstructorsAndWrappers(t *testing.T) {
	httpClient := &client.HTTPClient{}
	runSvc := NewRunService(httpClient)
	assert.NotNil(t, runSvc)

	t.Run("Output writes JSON", func(t *testing.T) {
		cmd := &cobra.Command{Use: "run-test"}
		cmd.Flags().Bool("quiet", false, "")
		cmd.Flags().String("output", "", "")

		out := captureStdout(t, func() {
			err := runSvc.Output(context.Background(), cmd, map[string]any{"ok": true})
			assert.NoError(t, err)
		})

		assert.Contains(t, out, "\"ok\": true")
	})

	t.Run("PrintSuccess prints when not quiet", func(t *testing.T) {
		cmd := &cobra.Command{Use: "run-test"}
		cmd.Flags().Bool("quiet", false, "")

		out := captureStdout(t, func() {
			runSvc.PrintSuccess(context.Background(), cmd, "run %d created", 7)
		})

		assert.Contains(t, out, "run 7 created")
	})

	t.Run("PrintSuccess silent in quiet mode", func(t *testing.T) {
		cmd := &cobra.Command{Use: "run-test"}
		cmd.Flags().Bool("quiet", true, "")

		out := captureStdout(t, func() {
			runSvc.PrintSuccess(context.Background(), cmd, "hidden")
		})

		assert.Equal(t, "", out)
	})
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	assert.NoError(t, err)

	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()

	fn()
	assert.NoError(t, w.Close())

	b, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.NoError(t, r.Close())

	return string(b)
}
