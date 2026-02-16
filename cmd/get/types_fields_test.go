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

// ==================== Тесты для get case-types ====================

func TestCaseTypesCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetCaseTypesFunc: func() (data.GetCaseTypesResponse, error) {
			return data.GetCaseTypesResponse{
				{ID: 1, Name: "Acceptance", IsDefault: false},
				{ID: 2, Name: "Accessibility", IsDefault: false},
				{ID: 3, Name: "Automated", IsDefault: false},
				{ID: 4, Name: "Compatibility", IsDefault: false},
				{ID: 5, Name: "Destructive", IsDefault: false},
				{ID: 6, Name: "Functional", IsDefault: true},
				{ID: 7, Name: "Other", IsDefault: false},
				{ID: 8, Name: "Performance", IsDefault: false},
				{ID: 9, Name: "Regression", IsDefault: false},
				{ID: 10, Name: "Security", IsDefault: false},
				{ID: 11, Name: "Smoke & Sanity", IsDefault: false},
				{ID: 12, Name: "Usability", IsDefault: false},
			}, nil
		},
	}

	cmd := newCaseTypesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCaseTypesCmd_EmptyList(t *testing.T) {
	mock := &client.MockClient{
		GetCaseTypesFunc: func() (data.GetCaseTypesResponse, error) {
			return data.GetCaseTypesResponse{}, nil
		},
	}

	cmd := newCaseTypesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCaseTypesCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetCaseTypesFunc: func() (data.GetCaseTypesResponse, error) {
			return nil, fmt.Errorf("failed to get case types")
		},
	}

	cmd := newCaseTypesCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get case types")
}

// ==================== Тесты для get case-fields ====================

func TestCaseFieldsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetCaseFieldsFunc: func() (data.GetCaseFieldsResponse, error) {
			return data.GetCaseFieldsResponse{
				{
					ID:          1,
					Name:        "custom_preconds",
					Label:       "Preconditions",
					Description: "The preconditions for this test case",
					SystemName:  "custom_preconds",
					TypeID:      3,
					DisplayOrder: 1,
				},
				{
					ID:          2,
					Name:        "custom_steps",
					Label:       "Steps",
					Description: "The steps for this test case",
					SystemName:  "custom_steps",
					TypeID:      3,
					DisplayOrder: 2,
				},
			}, nil
		},
	}

	cmd := newCaseFieldsCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCaseFieldsCmd_EmptyList(t *testing.T) {
	mock := &client.MockClient{
		GetCaseFieldsFunc: func() (data.GetCaseFieldsResponse, error) {
			return data.GetCaseFieldsResponse{}, nil
		},
	}

	cmd := newCaseFieldsCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCaseFieldsCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetCaseFieldsFunc: func() (data.GetCaseFieldsResponse, error) {
			return nil, fmt.Errorf("failed to get case fields")
		},
	}

	cmd := newCaseFieldsCmd(testhelper.GetClientForTests)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get case fields")
}

func TestCaseTypesCmd_NilClient(t *testing.T) {
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newCaseTypesCmd(nilClientFunc)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}

func TestCaseFieldsCmd_NilClient(t *testing.T) {
	nilClientFunc := func(cmd *cobra.Command) client.ClientInterface {
		return nil
	}

	cmd := newCaseFieldsCmd(nilClientFunc)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP клиент не инициализирован")
}
