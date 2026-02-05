package sync

import (
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/service/migration"
	"github.com/spf13/cobra"
)

// init устанавливает getClient для тестов
func init() {
	SetGetClientForTests(func(cmd *cobra.Command) *client.HTTPClient {
		// В тестах используется mockClient через newMigration
		return nil
	})
}

// migrationMock — мок для migration.ClientInterface
type migrationMock struct {
	addSharedStep func(projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error)
	addSuite      func(projectID int64, req *data.AddSuiteRequest) (*data.Suite, error)
	addSection    func(projectID int64, req *data.AddSectionRequest) (*data.Section, error)
	addCase       func(suiteID int64, req *data.AddCaseRequest) (*data.Case, error)

	getSharedSteps func(projectID int64) (data.GetSharedStepsResponse, error)
	getSuites      func(projectID int64) (data.GetSuitesResponse, error)
	getSections    func(projectID, suiteID int64) (data.GetSectionsResponse, error)
	getCases       func(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error)
}

// Проверка, что migrationMock реализует migration.ClientInterface
var _ migration.ClientInterface = (*migrationMock)(nil)

func (m *migrationMock) AddSharedStep(projectID int64, req *data.AddSharedStepRequest) (*data.SharedStep, error) {
	if m.addSharedStep != nil {
		return m.addSharedStep(projectID, req)
	}
	return nil, nil
}

func (m *migrationMock) AddSuite(projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
	if m.addSuite != nil {
		return m.addSuite(projectID, req)
	}
	return nil, nil
}

func (m *migrationMock) AddSection(projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
	if m.addSection != nil {
		return m.addSection(projectID, req)
	}
	return nil, nil
}

func (m *migrationMock) AddCase(suiteID int64, req *data.AddCaseRequest) (*data.Case, error) {
	if m.addCase != nil {
		return m.addCase(suiteID, req)
	}
	return nil, nil
}

func (m *migrationMock) GetSharedSteps(projectID int64) (data.GetSharedStepsResponse, error) {
	if m.getSharedSteps != nil {
		return m.getSharedSteps(projectID)
	}
	return data.GetSharedStepsResponse{}, nil
}

func (m *migrationMock) GetSuites(projectID int64) (data.GetSuitesResponse, error) {
	if m.getSuites != nil {
		return m.getSuites(projectID)
	}
	return data.GetSuitesResponse{}, nil
}

func (m *migrationMock) GetCases(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
	if m.getCases != nil {
		return m.getCases(projectID, suiteID, sectionID)
	}
	return data.GetCasesResponse{}, nil
}

func (m *migrationMock) GetSections(projectID, suiteID int64) (data.GetSectionsResponse, error) {
	if m.getSections != nil {
		return m.getSections(projectID, suiteID)
	}
	return data.GetSectionsResponse{}, nil
}

// newMigrationFactoryFromMock создаёт factory для migration с mock клиентом
func newMigrationFactoryFromMock(t *testing.T, mock migration.ClientInterface) func(migration.ClientInterface, int64, int64, int64, int64, string, string) (*migration.Migration, error) {
	return func(client migration.ClientInterface, srcProject, srcSuite, dstProject, dstSuite int64, compareField, logDir string) (*migration.Migration, error) {
		tmp := t.TempDir()
		return migration.NewMigration(mock, srcProject, srcSuite, dstProject, dstSuite, compareField, tmp)
	}
}
