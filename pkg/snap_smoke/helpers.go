//go:build smoke

package snap_smoke

import (
	"context"
	"fmt"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
)

// testCase creates a test case in the given section and registers cleanup.
// Returns the created case (with ID assigned by TestRail).
func testCase(t *testing.T, cli *client.HTTPClient, sectionID int64, title string) *data.Case {
	t.Helper()
	ctx := context.Background()

	c, err := cli.AddCase(ctx, sectionID, &data.AddCaseRequest{
		Title:      fmt.Sprintf("[smoke] %s", title),
		PriorityID: 3,
		TypeID:     1,
	})
	if err != nil {
		t.Fatalf("testCase: add_case failed: %v", err)
	}
	t.Logf("Created test case ID=%d title=%q in section=%d", c.ID, c.Title, sectionID)

	t.Cleanup(func() {
		if err := cli.DeleteCase(context.Background(), c.ID); err != nil {
			t.Logf("cleanup: delete_case %d failed (may already be deleted): %v", c.ID, err)
		} else {
			t.Logf("cleanup: deleted case %d", c.ID)
		}
	})

	return c
}

// testSection finds or creates a section for smoke tests.
// Uses a well-known name so repeated runs reuse the same section.
func testSection(t *testing.T, cli *client.HTTPClient, projectID, suiteID int64) int64 {
	t.Helper()
	ctx := context.Background()

	const sectionName = "[smoke] snap-rollback-tests"

	sections, err := cli.GetSections(ctx, projectID, suiteID)
	if err != nil {
		t.Fatalf("testSection: get_sections failed: %v", err)
	}

	for _, s := range sections {
		if s.Name == sectionName {
			t.Logf("Reusing existing section ID=%d name=%q", s.ID, s.Name)
			return s.ID
		}
	}

	s, err := cli.AddSection(ctx, projectID, &data.AddSectionRequest{
		Name:    sectionName,
		SuiteID: suiteID,
	})
	if err != nil {
		t.Fatalf("testSection: add_section failed: %v", err)
	}
	t.Logf("Created section ID=%d name=%q", s.ID, s.Name)
	return s.ID
}
