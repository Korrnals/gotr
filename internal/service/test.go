// internal/service/test.go
package service

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/utils"
	"github.com/spf13/cobra"
)

// TestService предоставляет бизнес-логику для работы с тестами
type TestService struct {
	client client.ClientInterface
}

// NewTestService создаёт новый сервис для работы с тестами
func NewTestService(client client.ClientInterface) *TestService {
	return &TestService{client: client}
}

// Get получает тест по ID
func (s *TestService) Get(testID int64) (*data.Test, error) {
	if testID <= 0 {
		return nil, fmt.Errorf("ID теста должен быть положительным числом")
	}

	test, err := s.client.GetTest(testID)
	if err != nil {
		return nil, err
	}

	return test, nil
}

// GetForRun получает список тестов для рана
func (s *TestService) GetForRun(runID int64, filters map[string]string) ([]data.Test, error) {
	if runID <= 0 {
		return nil, fmt.Errorf("ID рана должен быть положительным числом")
	}

	tests, err := s.client.GetTests(runID, filters)
	if err != nil {
		return nil, err
	}

	return tests, nil
}

// Update обновляет тест
func (s *TestService) Update(testID int64, req *data.UpdateTestRequest) (*data.Test, error) {
	if testID <= 0 {
		return nil, fmt.Errorf("ID теста должен быть положительным числом")
	}

	if req == nil {
		return nil, fmt.Errorf("запрос не может быть пустым")
	}

	// Валидация
	if req.StatusID < 0 {
		return nil, fmt.Errorf("status_id не может быть отрицательным")
	}

	test, err := s.client.UpdateTest(testID, req)
	if err != nil {
		return nil, err
	}

	return test, nil
}

// ParseID парсит ID из аргументов командной строки
func (s *TestService) ParseID(args []string, index int) (int64, error) {
	if len(args) <= index {
		return 0, fmt.Errorf("необходимо указать ID")
	}

	id, err := strconv.ParseInt(args[index], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("некорректный ID: %s", args[index])
	}

	if id <= 0 {
		return 0, fmt.Errorf("ID должен быть положительным числом")
	}

	return id, nil
}

// PrintSuccess выводит сообщение об успехе
func (s *TestService) PrintSuccess(cmd *cobra.Command, format string, args ...interface{}) {
	utils.PrintSuccess(cmd, format, args...)
}

// Output выводит результат в JSON
func (s *TestService) Output(cmd *cobra.Command, data interface{}) error {
	return utils.OutputResult(cmd, data)
}
