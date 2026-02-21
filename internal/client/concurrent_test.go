package client

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
)

func TestGetCasesParallel(t *testing.T) {
	tests := []struct {
		name      string
		suiteIDs  []int64
		workers   int
		wantErr   bool
		setupMock func(*MockClient)
	}{
		{
			name:     "successful parallel fetch",
			suiteIDs: []int64{1, 2, 3},
			workers:  2,
			wantErr:  false,
			setupMock: func(m *MockClient) {
				m.GetCasesFunc = func(projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
					return data.GetCasesResponse{
						{ID: suiteID * 100, Title: "Case 1", SuiteID: suiteID},
						{ID: suiteID*100 + 1, Title: "Case 2", SuiteID: suiteID},
					}, nil
				}
			},
		},
		{
			name:     "empty suite list",
			suiteIDs: []int64{},
			workers:  5,
			wantErr:  false,
			setupMock: func(m *MockClient) {
				// Should not be called
			},
		},
		{
			name:     "default workers",
			suiteIDs: []int64{1, 2},
			workers:  0, // Should use default
			wantErr:  false,
			setupMock: func(m *MockClient) {
				m.GetCasesFunc = func(projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
					return data.GetCasesResponse{
						{ID: suiteID, Title: "Case", SuiteID: suiteID},
					}, nil
				}
			},
		},
		{
			name:     "partial failure",
			suiteIDs: []int64{1, 2, 3},
			workers:  2,
			wantErr:  true,
			setupMock: func(m *MockClient) {
				m.GetCasesFunc = func(projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
					if suiteID == 2 {
						return nil, fmt.Errorf("suite %d not found", suiteID)
					}
					return data.GetCasesResponse{{ID: suiteID, Title: "Case"}}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockClient{}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			results, err := mock.GetCasesParallel(30, tt.suiteIDs, tt.workers, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCasesParallel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check results count matches suite count
				if len(results) != len(tt.suiteIDs) {
					t.Errorf("GetCasesParallel() returned %d results, want %d", len(results), len(tt.suiteIDs))
				}
			}
		})
	}
}

func TestGetSuitesParallel(t *testing.T) {
	tests := []struct {
		name       string
		projectIDs []int64
		workers    int
		wantErr    bool
		setupMock  func(*MockClient)
	}{
		{
			name:       "successful parallel fetch",
			projectIDs: []int64{30, 31},
			workers:    2,
			wantErr:    false,
			setupMock: func(m *MockClient) {
				m.GetSuitesFunc = func(projectID int64) (data.GetSuitesResponse, error) {
					return data.GetSuitesResponse{
						{ID: projectID * 10, Name: "Suite 1", ProjectID: projectID},
						{ID: projectID*10 + 1, Name: "Suite 2", ProjectID: projectID},
					}, nil
				}
			},
		},
		{
			name:       "empty project list",
			projectIDs: []int64{},
			workers:    5,
			wantErr:    false,
			setupMock:  func(m *MockClient) {},
		},
		{
			name:       "single project",
			projectIDs: []int64{30},
			workers:    5,
			wantErr:    false,
			setupMock: func(m *MockClient) {
				m.GetSuitesFunc = func(projectID int64) (data.GetSuitesResponse, error) {
					return data.GetSuitesResponse{{ID: 1, Name: "Only Suite"}}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockClient{}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			results, err := mock.GetSuitesParallel(tt.projectIDs, tt.workers, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSuitesParallel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(results) != len(tt.projectIDs) {
					t.Errorf("GetSuitesParallel() returned %d results, want %d", len(results), len(tt.projectIDs))
				}
			}
		})
	}
}

func TestGetCasesForSuitesParallel(t *testing.T) {
	tests := []struct {
		name      string
		suiteIDs  []int64
		workers   int
		wantErr   bool
		setupMock func(*MockClient)
	}{
		{
			name:     "flatten results",
			suiteIDs: []int64{1, 2},
			workers:  2,
			wantErr:  false,
			setupMock: func(m *MockClient) {
				m.GetCasesFunc = func(projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
					return data.GetCasesResponse{
						{ID: suiteID*100 + 1, Title: "Case 1", SuiteID: suiteID},
						{ID: suiteID*100 + 2, Title: "Case 2", SuiteID: suiteID},
					}, nil
				}
			},
		},
		{
			name:     "empty suite list",
			suiteIDs: []int64{},
			workers:  5,
			wantErr:  false,
			setupMock: func(m *MockClient) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockClient{}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			cases, err := mock.GetCasesForSuitesParallel(30, tt.suiteIDs, tt.workers, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCasesForSuitesParallel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.name == "flatten results" {
				// Should have 4 cases total (2 from each suite)
				if len(cases) != 4 {
					t.Errorf("GetCasesForSuitesParallel() returned %d cases, want 4", len(cases))
				}
			}
		})
	}
}
