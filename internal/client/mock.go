// internal/client/mock.go
// Mock-клиент для тестирования sync операций
package client

import "github.com/Korrnals/gotr/internal/models/data"

// MockSyncClient реализует SyncClientInterface для тестирования
type MockSyncClient struct {
	GetProjectsFunc      func() (data.GetProjectsResponse, error)
	GetSuitesFunc        func(projectID int64) (data.GetSuitesResponse, error)
	GetSectionsFunc      func(projectID, suiteID int64) (data.GetSectionsResponse, error)
	GetCasesFunc         func(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error)
	GetSharedStepsFunc   func(projectID int64) (data.GetSharedStepsResponse, error)
	AddCaseFunc          func(suiteID int64, req *data.AddCaseRequest) (*data.GetCaseResponse, error)
	AddSharedStepFunc    func(projectID int64, req *data.AddSharedStepRequest) (*data.GetSharedStepResponse, error)
	AddSectionFunc       func(projectID int64, req *data.AddSectionRequest) (*data.GetSectionResponse, error)
	AddSuiteFunc         func(projectID int64, req *data.AddSuiteRequest) (*data.GetSuiteResponse, error)
}

var _ SyncClientInterface = (*MockSyncClient)(nil)

func (m *MockSyncClient) GetProjects() (data.GetProjectsResponse, error) {
	if m.GetProjectsFunc != nil {
		return m.GetProjectsFunc()
	}
	return data.GetProjectsResponse{}, nil
}

func (m *MockSyncClient) GetSuites(projectID int64) (data.GetSuitesResponse, error) {
	if m.GetSuitesFunc != nil {
		return m.GetSuitesFunc(projectID)
	}
	return data.GetSuitesResponse{}, nil
}

func (m *MockSyncClient) GetSections(projectID, suiteID int64) (data.GetSectionsResponse, error) {
	if m.GetSectionsFunc != nil {
		return m.GetSectionsFunc(projectID, suiteID)
	}
	return data.GetSectionsResponse{}, nil
}

func (m *MockSyncClient) GetCases(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
	if m.GetCasesFunc != nil {
		return m.GetCasesFunc(projectID, suiteID, sectionID)
	}
	return data.GetCasesResponse{}, nil
}

func (m *MockSyncClient) GetSharedSteps(projectID int64) (data.GetSharedStepsResponse, error) {
	if m.GetSharedStepsFunc != nil {
		return m.GetSharedStepsFunc(projectID)
	}
	return data.GetSharedStepsResponse{}, nil
}

func (m *MockSyncClient) AddCase(suiteID int64, req *data.AddCaseRequest) (*data.GetCaseResponse, error) {
	if m.AddCaseFunc != nil {
		return m.AddCaseFunc(suiteID, req)
	}
	return nil, nil
}

func (m *MockSyncClient) AddSharedStep(projectID int64, req *data.AddSharedStepRequest) (*data.GetSharedStepResponse, error) {
	if m.AddSharedStepFunc != nil {
		return m.AddSharedStepFunc(projectID, req)
	}
	return nil, nil
}

func (m *MockSyncClient) AddSection(projectID int64, req *data.AddSectionRequest) (*data.GetSectionResponse, error) {
	if m.AddSectionFunc != nil {
		return m.AddSectionFunc(projectID, req)
	}
	return nil, nil
}

func (m *MockSyncClient) AddSuite(projectID int64, req *data.AddSuiteRequest) (*data.GetSuiteResponse, error) {
	if m.AddSuiteFunc != nil {
		return m.AddSuiteFunc(projectID, req)
	}
	return nil, nil
}
