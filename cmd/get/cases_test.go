package get

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для get cases ====================

func TestCasesCmd_WithSuiteID(t *testing.T) {
	mock := &client.MockClient{
		GetCasesFunc: func(projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(20069), suiteID)
			assert.Equal(t, int64(0), sectionID)
			return data.GetCasesResponse{
				{ID: 1, Title: "Test Case 1"},
				{ID: 2, Title: "Test Case 2"},
			}, nil
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--suite-id", "20069"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCasesCmd_WithProjectIDFlag(t *testing.T) {
	mock := &client.MockClient{
		GetCasesFunc: func(projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(20069), suiteID)
			return data.GetCasesResponse{
				{ID: 1, Title: "Test Case 1"},
			}, nil
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--project-id", "30", "--suite-id", "20069"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCasesCmd_WithSectionID(t *testing.T) {
	mock := &client.MockClient{
		GetCasesFunc: func(projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(20069), suiteID)
			assert.Equal(t, int64(100), sectionID)
			return data.GetCasesResponse{
				{ID: 1, Title: "Test Case in Section"},
			}, nil
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--suite-id", "20069", "--section-id", "100"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCasesCmd_AutoSelectSingleSuite(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetSuitesResponse{
				{ID: 20069, Name: "Master Suite"},
			}, nil
		},
		GetCasesFunc: func(projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(20069), suiteID)
			return data.GetCasesResponse{
				{ID: 1, Title: "Test Case 1"},
			}, nil
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30"}) // Без --suite-id

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCasesCmd_AllSuites(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(30), projectID)
			return data.GetSuitesResponse{
				{ID: 20069, Name: "Suite 1"},
				{ID: 20070, Name: "Suite 2"},
			}, nil
		},
		GetCasesFunc: func(projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			if suiteID == 20069 {
				return data.GetCasesResponse{{ID: 1, Title: "Case 1"}}, nil
			}
			if suiteID == 20070 {
				return data.GetCasesResponse{{ID: 2, Title: "Case 2"}}, nil
			}
			return nil, fmt.Errorf("unexpected suite ID: %d", suiteID)
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--all-suites"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCasesCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный ID проекта")
}

func TestCasesCmd_InvalidProjectIDFlag(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--project-id", "invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCasesCmd_NoSuites(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{}, nil // Нет сьютов
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не найдено сьютов")
}

func TestCasesCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetCasesFunc: func(projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--suite-id", "20069"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCasesCmd_GetSuitesError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return nil, fmt.Errorf("failed to get suites")
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось получить список сьютов")
}

// ==================== Тесты для get case ====================

func TestCaseCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetCaseFunc: func(caseID int64) (*data.Case, error) {
			assert.Equal(t, int64(12345), caseID)
			return &data.Case{
				ID:    12345,
				Title: "Test Case",
			}, nil
		},
	}

	cmd := newCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCaseCmd_InvalidCaseID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный ID кейса")
}

func TestCaseCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCaseCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetCaseFunc: func(caseID int64) (*data.Case, error) {
			return nil, fmt.Errorf("case not found")
		},
	}

	cmd := newCaseCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== Тесты для fetchCasesFromAllSuites ====================

func TestFetchCasesFromAllSuites_WithError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{
				{ID: 20069, Name: "Suite 1"},
				{ID: 20070, Name: "Suite 2"},
			}, nil
		},
		GetCasesFunc: func(projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			if suiteID == 20070 {
				return nil, fmt.Errorf("suite not found")
			}
			return data.GetCasesResponse{{ID: 1, Title: "Case 1"}}, nil
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--all-suites"})

	err := cmd.Execute()
	// Ошибка в одном сьюте не прерывает выполнение, просто выводится сообщение
	assert.NoError(t, err)
}

func TestCasesCmd_AllSuites_WithSectionID(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{
				{ID: 20069, Name: "Suite 1"},
			}, nil
		},
		GetCasesFunc: func(projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
			assert.Equal(t, int64(100), sectionID)
			return data.GetCasesResponse{{ID: 1, Title: "Case in Section"}}, nil
		},
	}

	cmd := newCasesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--all-suites", "--section-id", "100"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCasesCmd_NilClient(t *testing.T) {
	// Тестируем случай когда клиент не инициализирован
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newCasesCmd(nilClientFunc)
	cmd.SetArgs([]string{"30", "--suite-id", "20069"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}

func TestCaseCmd_NilClient(t *testing.T) {
	// Тестируем случай когда клиент не инициализирован
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newCaseCmd(nilClientFunc)
	cmd.SetArgs([]string{"12345"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}
