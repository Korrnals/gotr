package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
)

type noopProgressMonitor struct{}

func (noopProgressMonitor) Increment()      {}
func (noopProgressMonitor) IncrementBy(int) {}

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
				m.GetCasesFunc = func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
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
				m.GetCasesFunc = func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
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
				m.GetCasesFunc = func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
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

			ctx := context.Background()
			results, err := mock.GetCasesParallel(ctx, 30, tt.suiteIDs, tt.workers, nil)

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
				m.GetSuitesFunc = func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
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
				m.GetSuitesFunc = func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
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

			ctx := context.Background()
			results, err := mock.GetSuitesParallel(ctx, tt.projectIDs, tt.workers, nil)

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

func TestHTTPClient_GetSuitesParallel_EmptyProjectIDs(t *testing.T) {
	httpClient := &HTTPClient{}
	result, err := httpClient.GetSuitesParallel(context.Background(), nil, 2, nil)
	if err != nil {
		t.Fatalf("GetSuitesParallel() error = %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty map, got %d entries", len(result))
	}
}

func TestHTTPClient_GetSuitesParallel_SuccessAndPartialFailure(t *testing.T) {
	t.Run("success with monitor and default workers", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET method, got %s", r.Method)
			}
			if r.URL.Path != "/index.php" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			if !containsAny(r.URL.String(), []string{"get_suites/30", "get_suites/31"}) {
				t.Fatalf("unexpected url: %s", r.URL.String())
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.GetSuitesResponse{{ID: 10, Name: "Suite"}})
		})
		defer server.Close()

		result, err := client.GetSuitesParallel(context.Background(), []int64{30, 31}, 0, noopProgressMonitor{})
		if err != nil {
			t.Fatalf("GetSuitesParallel() unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 project results, got %d", len(result))
		}
	})

	t.Run("partial failure", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if containsAny(r.URL.String(), []string{"get_suites/30"}) {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.GetSuitesResponse{{ID: 101, Name: "Suite P1"}})
				return
			}
			if containsAny(r.URL.String(), []string{"get_suites/31"}) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`boom`))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		})
		defer server.Close()

		result, err := client.GetSuitesParallel(context.Background(), []int64{30, 31}, 2, nil)
		if err == nil {
			t.Fatalf("expected partial failure error")
		}
		if len(result) != 1 {
			t.Fatalf("expected one successful project result, got %d", len(result))
		}
	})
}

func containsAny(s string, parts []string) bool {
	for _, part := range parts {
		if part != "" && strings.Contains(s, part) {
			return true
		}
	}
	return false
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
				m.GetCasesFunc = func(ctx context.Context, projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
					return data.GetCasesResponse{
						{ID: suiteID*100 + 1, Title: "Case 1", SuiteID: suiteID},
						{ID: suiteID*100 + 2, Title: "Case 2", SuiteID: suiteID},
					}, nil
				}
			},
		},
		{
			name:      "empty suite list",
			suiteIDs:  []int64{},
			workers:   5,
			wantErr:   false,
			setupMock: func(m *MockClient) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockClient{}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			ctx := context.Background()
			cases, err := mock.GetCasesForSuitesParallel(ctx, 30, tt.suiteIDs, tt.workers, nil)

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

func TestGetCasesForSuitesParallel_ErrorBranches(t *testing.T) {
	t.Run("returns nil when no results and error", func(t *testing.T) {
		mock := &MockClient{}
		mock.GetCasesParallelFunc = func(ctx context.Context, projectID int64, suiteIDs []int64, workers int) (map[int64]data.GetCasesResponse, error) {
			return map[int64]data.GetCasesResponse{}, fmt.Errorf("upstream failed")
		}

		cases, err := mock.GetCasesForSuitesParallel(context.Background(), 30, []int64{1, 2}, 2, nil)
		if err == nil {
			t.Fatal("expected error")
		}
		if cases != nil {
			t.Fatalf("expected nil cases, got %v", cases)
		}
	})

	t.Run("mock keeps nil on error even with partial map", func(t *testing.T) {
		mock := &MockClient{}
		mock.GetCasesParallelFunc = func(ctx context.Context, projectID int64, suiteIDs []int64, workers int) (map[int64]data.GetCasesResponse, error) {
			return map[int64]data.GetCasesResponse{
				1: {
					{ID: 11, Title: "A", SuiteID: 1},
				},
			}, fmt.Errorf("partial failure")
		}

		cases, err := mock.GetCasesForSuitesParallel(context.Background(), 30, []int64{1, 2}, 2, nil)
		if err == nil {
			t.Fatal("expected error")
		}
		if cases != nil {
			t.Fatalf("expected nil cases for mock path, got %+v", cases)
		}
	})
}
