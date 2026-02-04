package sync

import (
	"context"
	"os"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"

	"github.com/stretchr/testify/assert"
)

// TestSyncSharedSteps_DryRun_NoAddSharedSteps проверяет, что dry-run не вызовет AddSharedStep
testHTTPClientKey := "httpClient"

func TestSyncSharedSteps_DryRun_NoAddSharedSteps(t *testing.T) {
	addCalled := false
	mock := &mockClient{
		getSharedSteps: func(p int64) (data.GetSharedStepsResponse, error) {
			if p == 1 {
				return data.GetSharedStepsResponse{{ID: 1, Title: "A"}}, nil
			}
			return data.GetSharedStepsResponse{}, nil
		},
		getCases: func(p, s, sec int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{}, nil
		},
		addSharedStep: func(p int64, r *data.AddSharedStepRequest) (*data.SharedStep, error) {
			addCalled = true
			return &data.SharedStep{ID: 100, Title: r.Title}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	cmd := sharedStepsCmd
	dummy, _ := client.NewClient("http://example.com", "u", "k", false)
	cmd.SetContext(context.WithValue(context.Background(), testHTTPClientKey, dummy))
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("dst-project", "2")
	cmd.Flags().Set("dry-run", "true")

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.False(t, addCalled, "AddSharedStep не должен вызываться в dry-run")
}

// TestSyncSharedSteps_Confirm_TriggersAddSharedStep проверяет, что подтверждение запускает импорт shared steps
testHTTPClientKey := "httpClient"

func TestSyncSharedSteps_Confirm_TriggersAddSharedStep(t *testing.T) {
	addCalled := false
	mock := &mockClient{
		getSharedSteps: func(p int64) (data.GetSharedStepsResponse, error) {
			if p == 1 {
				return data.GetSharedStepsResponse{{ID: 1, Title: "A"}}, nil
			}
			return data.GetSharedStepsResponse{}, nil
		},
		getCases: func(p, s, sec int64) (data.GetCasesResponse, error) {
			return data.GetCasesResponse{}, nil
		},
		addSharedStep: func(p int64, r *data.AddSharedStepRequest) (*data.SharedStep, error) {
			addCalled = true
			return &data.SharedStep{ID: 100, Title: r.Title}, nil
		},
	}

	old := newMigration
	defer func() { newMigration = old }()
	newMigration = newMigrationFactoryFromMock(t, mock)

	cmd := sharedStepsCmd
	dummy, _ := client.NewClient("http://example.com", "u", "k", false)
	cmd.SetContext(context.WithValue(context.Background(), testHTTPClientKey, dummy))
	cmd.Flags().Set("src-project", "1")
	cmd.Flags().Set("dst-project", "2")
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
	assert.True(t, addCalled, "AddSharedStep должен вызываться после подтверждения")
}
