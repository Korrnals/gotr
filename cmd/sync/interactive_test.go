package sync

import (
	"fmt"
	"os"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// mockStdin подменяет os.Stdin для тестирования интерактивного ввода
// Возвращает функцию для восстановления оригинального os.Stdin
func mockStdin(t *testing.T, input string) func() {
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	
	_, err = w.WriteString(input)
	assert.NoError(t, err)
	err = w.Close()
	assert.NoError(t, err)
	
	oldStdin := os.Stdin
	os.Stdin = r
	
	return func() {
		os.Stdin = oldStdin
	}
}

// ==================== Тесты для selectProjectInteractively ====================

func TestSelectProjectInteractively_Success(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func() (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1"},
				{ID: 2, Name: "Project 2"},
			}, nil
		},
	}

	// Симулируем ввод "1"
	restore := mockStdin(t, "1\n")
	defer restore()

	id, err := selectProjectInteractively(mock, "Test prompt:")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)
}

func TestSelectProjectInteractively_SecondProject(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func() (projects data.GetProjectsResponse, err error) {
			return data.GetProjectsResponse{
				{ID: 10, Name: "Alpha"},
				{ID: 20, Name: "Beta"},
				{ID: 30, Name: "Gamma"},
			}, nil
		},
	}

	// Симулируем ввод "2" - выбираем второй проект
	restore := mockStdin(t, "2\n")
	defer restore()

	id, err := selectProjectInteractively(mock, "Select project:")
	assert.NoError(t, err)
	assert.Equal(t, int64(20), id)
}

func TestSelectProjectInteractively_GetProjectsError(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func() (data.GetProjectsResponse, error) {
			return nil, fmt.Errorf("failed to fetch projects")
		},
	}

	id, err := selectProjectInteractively(mock, "Test prompt:")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось получить список проектов")
	assert.Equal(t, int64(0), id)
}

func TestSelectProjectInteractively_NoProjects(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func() (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{}, nil
		},
	}

	id, err := selectProjectInteractively(mock, "Test prompt:")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не найдено проектов")
	assert.Equal(t, int64(0), id)
}

func TestSelectProjectInteractively_InvalidInput(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func() (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1"},
				{ID: 2, Name: "Project 2"},
			}, nil
		},
	}

	// Симулируем ввод "invalid"
	restore := mockStdin(t, "invalid\n")
	defer restore()

	id, err := selectProjectInteractively(mock, "Test prompt:")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "неверный выбор")
	assert.Equal(t, int64(0), id)
}

func TestSelectProjectInteractively_OutOfRange(t *testing.T) {
	mock := &client.MockClient{
		GetProjectsFunc: func() (data.GetProjectsResponse, error) {
			return data.GetProjectsResponse{
				{ID: 1, Name: "Project 1"},
				{ID: 2, Name: "Project 2"},
			}, nil
		},
	}

	// Симулируем ввод "5" - вне диапазона
	restore := mockStdin(t, "5\n")
	defer restore()

	id, err := selectProjectInteractively(mock, "Test prompt:")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "неверный выбор")
	assert.Equal(t, int64(0), id)
}

// ==================== Тесты для selectSuiteInteractively ====================

func TestSelectSuiteInteractively_Success(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetSuitesResponse{
				{ID: 10, Name: "Suite 1"},
				{ID: 20, Name: "Suite 2"},
			}, nil
		},
	}

	// Симулируем ввод "2"
	restore := mockStdin(t, "2\n")
	defer restore()

	id, err := selectSuiteInteractively(mock, 1, "Test prompt:")
	assert.NoError(t, err)
	assert.Equal(t, int64(20), id)
}

func TestSelectSuiteInteractively_AutoSelectSingleSuite(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			assert.Equal(t, int64(1), projectID)
			return data.GetSuitesResponse{
				{ID: 100, Name: "Single Suite"},
			}, nil
		},
	}

	// При одном сьюте выбор автоматический, stdin не нужен
	id, err := selectSuiteInteractively(mock, 1, "Test prompt:")
	assert.NoError(t, err)
	assert.Equal(t, int64(100), id)
}

func TestSelectSuiteInteractively_GetSuitesError(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return nil, fmt.Errorf("failed to fetch suites")
		},
	}

	id, err := selectSuiteInteractively(mock, 1, "Test prompt:")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось получить список сьютов")
	assert.Equal(t, int64(0), id)
}

func TestSelectSuiteInteractively_NoSuites(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{}, nil
		},
	}

	id, err := selectSuiteInteractively(mock, 1, "Test prompt:")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не найдено сьютов")
	assert.Equal(t, int64(0), id)
}

func TestSelectSuiteInteractively_InvalidInput(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{
				{ID: 10, Name: "Suite 1"},
				{ID: 20, Name: "Suite 2"},
			}, nil
		},
	}

	// Симулируем ввод "invalid"
	restore := mockStdin(t, "invalid\n")
	defer restore()

	id, err := selectSuiteInteractively(mock, 1, "Test prompt:")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "неверный выбор")
	assert.Equal(t, int64(0), id)
}

func TestSelectSuiteInteractively_OutOfRange(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{
				{ID: 10, Name: "Suite 1"},
				{ID: 20, Name: "Suite 2"},
			}, nil
		},
	}

	// Симулируем ввод "10" - вне диапазона
	restore := mockStdin(t, "10\n")
	defer restore()

	id, err := selectSuiteInteractively(mock, 1, "Test prompt:")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "неверный выбор")
	assert.Equal(t, int64(0), id)
}

func TestSelectSuiteInteractively_WithDescription(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{
				{ID: 10, Name: "Suite 1", Description: "This is a long description that might be truncated"},
				{ID: 20, Name: "Suite 2", Description: "Short"},
			}, nil
		},
	}

	// Симулируем ввод "1"
	restore := mockStdin(t, "1\n")
	defer restore()

	id, err := selectSuiteInteractively(mock, 1, "Test prompt:")
	assert.NoError(t, err)
	assert.Equal(t, int64(10), id)
}
