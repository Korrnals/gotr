// internal/service/run_test.go
package service

import (
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// mockRunClient — мок клиента для тестирования RunService
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
			errMatch: "отсутствует аргумент",
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
			id, err := svc.ParseID(tt.args, tt.index)
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
