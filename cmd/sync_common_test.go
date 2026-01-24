package cmd

import (
	"gotr/internal/migration"
	"gotr/internal/models/data"
	"testing"
)

// mockClient — универсальный мок для migration.ClientInterface, используемый в тестах cmd
type mockClient struct {
	addSharedStep func(projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error)
	addSuite      func(projectID int64, req *data.AddSuiteRequest) (*data.Suite, error)
	addSection    func(projectID int64, req *data.AddSectionRequest) (*data.Section, error)
	addCase       func(suiteID int64, req *data.AddCaseRequest) (*data.Case, error)

	getSharedSteps func(projectID int64) (data.GetSharedStepsResponse, error)
	getSuites      func(projectID int64) (data.GetSuitesResponse, error)
	getSections    func(projectID, suiteID int64) (data.GetSectionsResponse, error)
	getCases       func(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error)
}

func (m *mockClient) AddSharedStep(projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
	if m.addSharedStep != nil {
		return m.addSharedStep(projectID, req)
	}
	return nil, nil
}
func (m *mockClient) AddSuite(projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
	if m.addSuite != nil {
		return m.addSuite(projectID, req)
	}
	return nil, nil
}
func (m *mockClient) AddSection(projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
	if m.addSection != nil {
		return m.addSection(projectID, req)
	}
	return nil, nil
}
func (m *mockClient) AddCase(suiteID int64, req *data.AddCaseRequest) (*data.Case, error) {
	if m.addCase != nil {
		return m.addCase(suiteID, req)
	}
	return nil, nil
}

func (m *mockClient) GetSharedSteps(projectID int64) (data.GetSharedStepsResponse, error) {
	if m.getSharedSteps != nil {
		return m.getSharedSteps(projectID)
	}
	return data.GetSharedStepsResponse{}, nil
}
func (m *mockClient) GetSuites(projectID int64) (data.GetSuitesResponse, error) {
	if m.getSuites != nil {
		return m.getSuites(projectID)
	}
	return data.GetSuitesResponse{}, nil
}
func (m *mockClient) GetCases(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
	if m.getCases != nil {
		return m.getCases(projectID, suiteID, sectionID)
	}
	return data.GetCasesResponse{}, nil
}
func (m *mockClient) GetSections(projectID, suiteID int64) (data.GetSectionsResponse, error) {
	if m.getSections != nil {
		return m.getSections(projectID, suiteID)
	}
	return data.GetSectionsResponse{}, nil
}

// helper: returns a newMigration factory which constructs migration with mock and t.TempDir()
func newMigrationFactoryFromMock(t *testing.T, mock migration.ClientInterface) func(migration.ClientInterface, int64, int64, int64, int64, string, string) (*migration.Migration, error) {
	return func(client migration.ClientInterface, srcProject, srcSuite, dstProject, dstSuite int64, compareField, logDir string) (*migration.Migration, error) {
		tmp := t.TempDir()
		return migration.NewMigration(mock, srcProject, srcSuite, dstProject, dstSuite, compareField, tmp)
	}
}
