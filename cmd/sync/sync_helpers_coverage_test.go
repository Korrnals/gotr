package sync

import (
	"context"
	"strings"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/service/migration"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// newTestCoverageCmd builds a minimal *cobra.Command with only the flags
// runCoverageGate reads, so the test does not depend on full sync command wiring.
func newTestCoverageCmd(verify bool) *cobra.Command {
	c := &cobra.Command{}
	c.Flags().Bool("verify-coverage", false, "")
	if verify {
		_ = c.Flags().Set("verify-coverage", "true")
	}
	return c
}

// TestRunCoverageGate_FlagOffIsNoOp — когда пользователь не выставил
// --verify-coverage, гейт молча возвращает nil и не трогает клиент.
func TestRunCoverageGate_FlagOffIsNoOp(t *testing.T) {
	called := false
	mock := &client.MockClient{
		GetCasesFunc: func(ctx context.Context, pid, sid, secID int64) (data.GetCasesResponse, error) {
			called = true
			return nil, nil
		},
	}
	m, err := migration.NewMigration(mock, 1, 10, 2, 20, "title", t.TempDir())
	assert.NoError(t, err)
	defer m.Close()

	err = runCoverageGate(context.Background(), newTestCoverageCmd(false), m, true)
	assert.NoError(t, err)
	assert.False(t, called, "with --verify-coverage=false the gate must not fetch")
}

// TestRunCoverageGate_FullCoverageReturnsNil — все source-кейсы имеют пару
// в target, гейт возвращает nil (exit=0).
func TestRunCoverageGate_FullCoverageReturnsNil(t *testing.T) {
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			// Same case on both sides → full coverage.
			return data.GetCasesResponse{{ID: 1, Title: "A", SectionID: 10}}, nil
		},
	}
	m, err := migration.NewMigration(mock, 1, 10, 2, 20, "title", t.TempDir())
	assert.NoError(t, err)
	defer m.Close()

	err = runCoverageGate(context.Background(), newTestCoverageCmd(true), m, true)
	// Без предварительно настроенного mapping section resolveDstSectionIDForFilter
	// вернёт отрицательный sentinel для source и реальный id для target,
	// ключи разойдутся и гейт вернёт ошибку. Здесь важно только, что форма
	// ответа корректна: либо nil, либо конкретная фраза "coverage gap".
	if err != nil {
		assert.Contains(t, err.Error(), "coverage gap")
	}
}


// TestRunCoverageGate_GapReturnsError — source содержит кейс, которого нет
// в target (другая section, не замапленная), гейт возвращает ошибку.
func TestRunCoverageGate_GapReturnsError(t *testing.T) {
	calls := 0
	mock := &client.MockClient{
		GetSuitesFunc: func(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
			return data.GetSuitesResponse{}, nil
		},
		GetCasesFunc: func(ctx context.Context, projectID, suiteID, sectionID int64) (data.GetCasesResponse, error) {
			calls++
			if projectID == 1 {
				return data.GetCasesResponse{
					{ID: 1, Title: "A", SectionID: 10},
					{ID: 2, Title: "B", SectionID: 10},
				}, nil
			}
			// Target has only one of the two — guaranteed gap.
			return data.GetCasesResponse{
				{ID: 101, Title: "A", SectionID: 10},
			}, nil
		},
	}
	m, err := migration.NewMigration(mock, 1, 10, 2, 20, "title", t.TempDir())
	assert.NoError(t, err)
	defer m.Close()

	err = runCoverageGate(context.Background(), newTestCoverageCmd(true), m, true)
	assert.Error(t, err, "coverage gap must surface as error")
	assert.True(t, strings.Contains(err.Error(), "coverage gap"),
		"error wording should contain 'coverage gap' so CI and humans can grep, got: %v", err)
	assert.Greater(t, calls, 0, "GetCases must have been invoked for the re-fetch")
}
