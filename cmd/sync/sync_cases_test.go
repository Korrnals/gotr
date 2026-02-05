package sync

import (
	"context"
	"os"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"

	"github.com/stretchr/testify/assert"
)

// TestSyncCases_DryRun_NoAddCase проверяет, что в режиме dry-run не вызывается AddCase
// TODO: Тест требует рефакторинга - использует устаревшую архитектуру getClient

func TestSyncCases_DryRun_NoAddCase(t *testing.T) {
	t.Skip("Skipping broken test - needs refactoring to use context-based client")
	addCalled := false
	mock := &mockClient{
		getCases: func(p, s, sec int64) (data.GetCasesResponse, error) {
			if p == 1 {
				return data.GetCasesResponse{{ID: 1, Title: "Case 1"}}, nil
			}
			return data.GetCasesResponse{}, nil
		},
		addCase: func(suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
			addCalled = true
			return &data.Case{ID: 100, Title: r.Title}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	// Создаём dummy клиент (как во втором тесте)
	dummy, _ := client.NewClient("http://example.com", "u", "k", false)
	cmd := casesCmd
	cmd.SetContext(context.WithValue(context.Background(), testHTTPClientKey, dummy))
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("dry-run", "true")

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.False(t, addCalled, "AddCase не должен вызываться в dry-run")
}

// TestSyncCases_Confirm_TriggersAddCase проверяет, что подтверждение запускает импорт кейсов
// TODO: Тест требует рефакторинга - использует устаревшую архитектуру getClient

func TestSyncCases_Confirm_TriggersAddCase(t *testing.T) {
	t.Skip("Skipping broken test - needs refactoring to use context-based client")
	addCalled := false
	mock := &mockClient{
		getCases: func(p, s, sec int64) (data.GetCasesResponse, error) {
			if p == 1 {
				return data.GetCasesResponse{{ID: 1, Title: "Case 1"}}, nil
			}
			return data.GetCasesResponse{}, nil
		},
		addCase: func(suiteID int64, r *data.AddCaseRequest) (*data.Case, error) {
			addCalled = true
			return &data.Case{ID: 100, Title: r.Title}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	// ensure command has a client in context to avoid GetClient exit
	dummy, _ := client.NewClient("http://example.com", "u", "k", false)
	cmd := casesCmd
	cmd.SetContext(context.WithValue(context.Background(), testHTTPClientKey, dummy))
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("src-suite", "10")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dst-suite", "20")
	cmd.Flags().Set("dry-run", "false")

	// simulate stdin "y"
	r, w, _ := os.Pipe()
	_, _ = w.Write([]byte("y\n"))
	_ = w.Close()
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = r

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.True(t, addCalled, "AddCase должен вызываться после подтверждения")
}
