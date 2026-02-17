package get

import (
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для get sections section ====================

func TestSectionGetCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetSectionFunc: func(sectionID int64) (*data.Section, error) {
			assert.Equal(t, int64(100), sectionID)
			return &data.Section{
				ID:      100,
				Name:    "Test Section",
				SuiteID: 20069,
				Depth:   0,
			}, nil
		},
	}

	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"100"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSectionGetCmd_InvalidSectionID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "section_id должен быть положительным числом")
}

func TestSectionGetCmd_ZeroSectionID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "section_id должен быть положительным числом")
}

func TestSectionGetCmd_NegativeSectionID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"-5"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSectionGetCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSectionGetCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetSectionFunc: func(sectionID int64) (*data.Section, error) {
			return nil, fmt.Errorf("section not found")
		},
	}

	cmd := newSectionGetCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== Тесты для get sections list ====================

func TestSectionsListCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetSectionsFunc: func(projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(0), suiteID)
			return data.GetSectionsResponse{
				{ID: 1, Name: "Section 1", SuiteID: 20069},
				{ID: 2, Name: "Section 2", SuiteID: 20069},
			}, nil
		},
	}

	cmd := newSectionsListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSectionsListCmd_WithSuiteID(t *testing.T) {
	mock := &client.MockClient{
		GetSectionsFunc: func(projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			assert.Equal(t, int64(30), projectID)
			assert.Equal(t, int64(20069), suiteID)
			return data.GetSectionsResponse{
				{ID: 1, Name: "Section 1", SuiteID: 20069},
			}, nil
		},
	}

	cmd := newSectionsListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"30", "--suite-id", "20069"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSectionsListCmd_InvalidProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSectionsListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project_id должен быть положительным числом")
}

func TestSectionsListCmd_ZeroProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSectionsListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project_id должен быть положительным числом")
}

func TestSectionsListCmd_NegativeProjectID(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSectionsListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"-10"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSectionsListCmd_NoArgs(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newSectionsListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSectionsListCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetSectionsFunc: func(projectID int64, suiteID int64) (data.GetSectionsResponse, error) {
			return nil, fmt.Errorf("project not found")
		},
	}

	cmd := newSectionsListCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99999"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
