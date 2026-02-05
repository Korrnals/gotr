// internal/client/interfaces.go
// Базовые интерфейсы клиента (минимальный набор для тестирования)
package client

import "github.com/Korrnals/gotr/internal/models/data"

// SyncClientInterface — минимальный интерфейс для sync операций
type SyncClientInterface interface {
	GetProjects() (data.GetProjectsResponse, error)
	GetSuites(projectID int64) (data.GetSuitesResponse, error)
	GetSections(projectID, suiteID int64) (data.GetSectionsResponse, error)
	GetCases(projectID, suiteID, sectionID int64) (data.GetCasesResponse, error)
	GetSharedSteps(projectID int64) (data.GetSharedStepsResponse, error)
	
	AddCase(suiteID int64, req *data.AddCaseRequest) (*data.GetCaseResponse, error)
	AddSharedStep(projectID int64, req *data.AddSharedStepRequest) (*data.GetSharedStepResponse, error)
	AddSection(projectID int64, req *data.AddSectionRequest) (*data.GetSectionResponse, error)
	AddSuite(projectID int64, req *data.AddSuiteRequest) (*data.GetSuiteResponse, error)
}
