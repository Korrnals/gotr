package testrailapi

import "testing"

func TestNew_InitializesResources(t *testing.T) {
	api := New()

	if api == nil {
		t.Fatal("New returned nil")
	}

	if api.Attachments == nil || api.Cases == nil || api.CaseFields == nil || api.CaseTypes == nil {
		t.Fatal("expected core resources to be initialized")
	}

	if api.Configurations == nil || api.Milestones == nil || api.Plans == nil || api.Priorities == nil {
		t.Fatal("expected management resources to be initialized")
	}

	if api.Projects == nil || api.Results == nil || api.ResultFields == nil || api.Runs == nil {
		t.Fatal("expected execution resources to be initialized")
	}

	if api.Sections == nil || api.Statuses == nil || api.Suites == nil || api.Tests == nil || api.Users == nil {
		t.Fatal("expected test tree resources to be initialized")
	}

	if api.Reports == nil || api.Roles == nil || api.Templates == nil || api.Groups == nil {
		t.Fatal("expected administration resources to be initialized")
	}

	if api.SharedSteps == nil || api.Variables == nil || api.Labels == nil || api.Datasets == nil || api.BDDs == nil {
		t.Fatal("expected extension resources to be initialized")
	}
}

func TestPaths_AggregatesAllResources(t *testing.T) {
	api := New()

	all := api.Paths()
	if len(all) == 0 {
		t.Fatal("expected non-empty paths list")
	}

	expected := 0
	expected += len(api.Attachments.Paths())
	expected += len(api.Cases.Paths())
	expected += len(api.CaseFields.Paths())
	expected += len(api.CaseTypes.Paths())
	expected += len(api.Configurations.Paths())
	expected += len(api.Milestones.Paths())
	expected += len(api.Plans.Paths())
	expected += len(api.Priorities.Paths())
	expected += len(api.Projects.Paths())
	expected += len(api.Results.Paths())
	expected += len(api.ResultFields.Paths())
	expected += len(api.Runs.Paths())
	expected += len(api.Sections.Paths())
	expected += len(api.Statuses.Paths())
	expected += len(api.Suites.Paths())
	expected += len(api.Tests.Paths())
	expected += len(api.Users.Paths())
	expected += len(api.Reports.Paths())
	expected += len(api.Roles.Paths())
	expected += len(api.Templates.Paths())
	expected += len(api.SharedSteps.Paths())
	expected += len(api.Variables.Paths())
	expected += len(api.Labels.Paths())
	expected += len(api.Datasets.Paths())
	expected += len(api.BDDs.Paths())

	if len(all) != expected {
		t.Fatalf("unexpected total paths count: got=%d want=%d", len(all), expected)
	}

	hasEndpoint := func(method, uri string) bool {
		for _, p := range all {
			if p.Method == method && p.URI == uri {
				return true
			}
		}
		return false
	}

	if !hasEndpoint("GET", "index.php?/api/v2/get_cases/{project_id}") {
		t.Fatal("expected cases list endpoint in aggregated paths")
	}

	if !hasEndpoint("POST", "index.php?/api/v2/add_attachment_to_case/{case_id}") {
		t.Fatal("expected attachment endpoint in aggregated paths")
	}

	if !hasEndpoint("GET", "index.php?/api/v2/get_bdd/{case_id}") {
		t.Fatal("expected bdd endpoint in aggregated paths")
	}

	// Phase 13.5: verify newly added endpoints
	if !hasEndpoint("GET", "index.php?/api/v2/get_attachment/{attachment_id}") {
		t.Fatal("expected get_attachment endpoint in aggregated paths")
	}
	if !hasEndpoint("GET", "index.php?/api/v2/get_attachments_for_case/{case_id}") {
		t.Fatal("expected get_attachments_for_case endpoint in aggregated paths")
	}
	if !hasEndpoint("GET", "index.php?/api/v2/get_attachments_for_project/{project_id}") {
		t.Fatal("expected get_attachments_for_project endpoint in aggregated paths")
	}
	if !hasEndpoint("POST", "index.php?/api/v2/delete_attachment/{attachment_id}") {
		t.Fatal("expected delete_attachment endpoint in aggregated paths")
	}
	if !hasEndpoint("GET", "index.php?/api/v2/get_users/{project_id}") {
		t.Fatal("expected get_users/{project_id} endpoint in aggregated paths")
	}
	if !hasEndpoint("POST", "index.php?/api/v2/add_user") {
		t.Fatal("expected add_user endpoint in aggregated paths")
	}
	if !hasEndpoint("POST", "index.php?/api/v2/update_user/{user_id}") {
		t.Fatal("expected update_user endpoint in aggregated paths")
	}
	if !hasEndpoint("GET", "index.php?/api/v2/get_cross_project_reports") {
		t.Fatal("expected get_cross_project_reports endpoint in aggregated paths")
	}
	if !hasEndpoint("GET", "index.php?/api/v2/get_label/{label_id}") {
		t.Fatal("expected get_label endpoint in aggregated paths")
	}
	if !hasEndpoint("GET", "index.php?/api/v2/get_labels/{project_id}") {
		t.Fatal("expected get_labels endpoint in aggregated paths")
	}

	groups := api.Groups.Paths()
	if len(groups) == 0 {
		t.Fatal("expected groups paths to be non-empty")
	}

	if groups[0].URI == "" || groups[0].Method == "" {
		t.Fatal("expected valid first groups path")
	}

	for i, p := range all {
		if p.Method == "" {
			t.Fatalf("path[%d]: empty method", i)
		}
		if p.URI == "" {
			t.Fatalf("path[%d]: empty URI", i)
		}
		if p.Description == "" {
			t.Fatalf("path[%d]: empty description", i)
		}
	}
}
