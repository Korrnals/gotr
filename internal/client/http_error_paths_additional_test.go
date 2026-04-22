package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestWave6C_UsersTargets(t *testing.T) {
	t.Run("GetUsersByProject decode error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/index.php?/api/v2/get_users/7" {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		})
		defer s.Close()

		_, err := c.GetUsersByProject(context.Background(), 7)
		if err == nil {
			t.Fatal("expected decode error")
		}
	})

	t.Run("GetUser decode error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/index.php?/api/v2/get_user/9" {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		})
		defer s.Close()

		_, err := c.GetUser(context.Background(), 9)
		if err == nil {
			t.Fatal("expected decode error")
		}
	})

	t.Run("GetTemplates transport error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		_, err := c.GetTemplates(context.Background(), 30)
		if err == nil {
			t.Fatal("expected transport error")
		}
	})

	t.Run("GetTemplates success edge project id 0", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/index.php?/api/v2/get_templates/0" {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.GetTemplatesResponse{{ID: 1, Name: "T"}})
		})
		defer s.Close()

		resp, err := c.GetTemplates(context.Background(), 0)
		if err != nil {
			t.Fatalf("GetTemplates error: %v", err)
		}
		if len(resp) != 1 {
			t.Fatalf("expected 1 template, got %d", len(resp))
		}
	})
}

func TestWave6C_ProjectsTargets(t *testing.T) {
	t.Run("AddProject decode error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/index.php?/api/v2/add_project" {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		})
		defer s.Close()

		_, err := c.AddProject(context.Background(), &data.AddProjectRequest{Name: "P"})
		if err == nil {
			t.Fatal("expected decode error")
		}
	})

	t.Run("AddProject transport error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		_, err := c.AddProject(context.Background(), &data.AddProjectRequest{Name: "P"})
		if err == nil {
			t.Fatal("expected transport error")
		}
	})

	t.Run("UpdateProject decode error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/index.php?/api/v2/update_project/5" {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		})
		defer s.Close()

		_, err := c.UpdateProject(context.Background(), 5, &data.UpdateProjectRequest{Name: "U"})
		if err == nil {
			t.Fatal("expected decode error")
		}
	})

	t.Run("UpdateProject transport error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		_, err := c.UpdateProject(context.Background(), 5, &data.UpdateProjectRequest{Name: "U"})
		if err == nil {
			t.Fatal("expected transport error")
		}
	})
}

func TestWave6C_ReportsTargets(t *testing.T) {
	t.Run("RunReport decode error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/index.php?/api/v2/run_report/1" {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		})
		defer s.Close()

		_, err := c.RunReport(context.Background(), 1)
		if err == nil {
			t.Fatal("expected decode error")
		}
	})

	t.Run("RunReport transport error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		_, err := c.RunReport(context.Background(), 1)
		if err == nil {
			t.Fatal("expected transport error")
		}
	})

	t.Run("RunCrossProjectReport transport error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		_, err := c.RunCrossProjectReport(context.Background(), 2)
		if err == nil {
			t.Fatal("expected transport error")
		}
	})

	t.Run("GetCrossProjectReports decode error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/index.php?/api/v2/get_cross_project_reports" {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		})
		defer s.Close()

		_, err := c.GetCrossProjectReports(context.Background())
		if err == nil {
			t.Fatal("expected decode error")
		}
	})

	t.Run("GetCrossProjectReports transport error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		_, err := c.GetCrossProjectReports(context.Background())
		if err == nil {
			t.Fatal("expected transport error")
		}
	})
}

func TestWave6C_TestsAndSuitesTargets(t *testing.T) {
	t.Run("GetTest decode error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/index.php?/api/v2/get_test/13" {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		})
		defer s.Close()

		_, err := c.GetTest(context.Background(), 13)
		if err == nil {
			t.Fatal("expected decode error")
		}
	})

	t.Run("GetTest transport error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		_, err := c.GetTest(context.Background(), 13)
		if err == nil {
			t.Fatal("expected transport error")
		}
	})

	t.Run("AddSuite decode error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/index.php?/api/v2/add_suite/3" {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		})
		defer s.Close()

		_, err := c.AddSuite(context.Background(), 3, &data.AddSuiteRequest{Name: "S"})
		if err == nil {
			t.Fatal("expected decode error")
		}
	})

	t.Run("UpdateSuite decode error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/index.php?/api/v2/update_suite/4" {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		})
		defer s.Close()

		_, err := c.UpdateSuite(context.Background(), 4, &data.UpdateSuiteRequest{Name: "S"})
		if err == nil {
			t.Fatal("expected decode error")
		}
	})

	t.Run("AddSuite transport error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		_, err := c.AddSuite(context.Background(), 3, &data.AddSuiteRequest{Name: "S"})
		if err == nil {
			t.Fatal("expected transport error")
		}
	})

	t.Run("UpdateSuite transport error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		_, err := c.UpdateSuite(context.Background(), 4, &data.UpdateSuiteRequest{Name: "S"})
		if err == nil {
			t.Fatal("expected transport error")
		}
	})
}

func TestWave6C_CasesExtendedMilestonesAndPaginatorTargets(t *testing.T) {
	t.Run("UpdateCases transport error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		_, err := c.UpdateCases(context.Background(), 1, &data.UpdateCasesRequest{CaseIDs: []int64{1}})
		if err == nil {
			t.Fatal("expected transport error")
		}
	})

	t.Run("DeleteCases transport error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		err := c.DeleteCases(context.Background(), 1, &data.DeleteCasesRequest{CaseIDs: []int64{1}})
		if err == nil {
			t.Fatal("expected transport error")
		}
	})

	t.Run("CopyCasesToSection transport error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		err := c.CopyCasesToSection(context.Background(), 1, &data.CopyCasesRequest{CaseIDs: []int64{1}})
		if err == nil {
			t.Fatal("expected transport error")
		}
	})

	t.Run("MoveCasesToSection transport error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		err := c.MoveCasesToSection(context.Background(), 1, &data.MoveCasesRequest{CaseIDs: []int64{1}})
		if err == nil {
			t.Fatal("expected transport error")
		}
	})

	t.Run("GetGroups transport error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		_, err := c.GetGroups(context.Background(), 1)
		if err == nil {
			t.Fatal("expected transport error")
		}
	})

	t.Run("GetMilestone transport error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
		s.Close()

		_, err := c.GetMilestone(context.Background(), 1)
		if err == nil {
			t.Fatal("expected transport error")
		}
	})

	t.Run("isJSONArrayBody edge inputs", func(t *testing.T) {
		cases := []struct {
			name string
			body string
			want bool
		}{
			{name: "empty", body: "", want: false},
			{name: "whitespace only", body: " \n\t\r", want: false},
			{name: "array with leading spaces", body: "  [1,2]", want: true},
			{name: "object", body: "{\"k\":1}", want: false},
			{name: "non-json leading char", body: "x[]", want: false},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got := isJSONArrayBody([]byte(tc.body))
				if got != tc.want {
					t.Fatalf("isJSONArrayBody(%q) = %v, want %v", tc.body, got, tc.want)
				}
			})
		}
	})
}

func TestWave6D_CasesAndExtendedResidualBranches(t *testing.T) {
	t.Run("DeleteCases non-OK", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/index.php?/api/v2/delete_cases/55" {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("delete failed"))
		})
		defer s.Close()

		err := c.DeleteCases(context.Background(), 55, &data.DeleteCasesRequest{CaseIDs: []int64{1, 2}})
		if err == nil {
			t.Fatal("expected non-OK error")
		}
	})

	t.Run("CopyCasesToSection non-OK", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/index.php?/api/v2/copy_cases_to_section/66" {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte("copy conflict"))
		})
		defer s.Close()

		err := c.CopyCasesToSection(context.Background(), 66, &data.CopyCasesRequest{CaseIDs: []int64{3}})
		if err == nil {
			t.Fatal("expected non-OK error")
		}
	})

	t.Run("MoveCasesToSection non-OK", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/index.php?/api/v2/move_cases_to_section/77" {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("move failed"))
		})
		defer s.Close()

		err := c.MoveCasesToSection(context.Background(), 77, &data.MoveCasesRequest{CaseIDs: []int64{5}})
		if err == nil {
			t.Fatal("expected non-OK error")
		}
	})

	t.Run("GetGroups non-OK", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/index.php?/api/v2/get_groups/88" {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("denied"))
		})
		defer s.Close()

		_, err := c.GetGroups(context.Background(), 88)
		if err == nil {
			t.Fatal("expected non-OK error")
		}
	})

	t.Run("GetGroups decode error", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/index.php?/api/v2/get_groups/99" {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		})
		defer s.Close()

		_, err := c.GetGroups(context.Background(), 99)
		if err == nil {
			t.Fatal("expected decode error")
		}
	})
}

func TestWave6D_UsersResidualBranches(t *testing.T) {
	c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.String() {
		case "/index.php?/api/v2/get_users":
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("bad users"))
		case "/index.php?/api/v2/get_users/10":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("bad project users"))
		case "/index.php?/api/v2/get_user/20":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("no user"))
		case "/index.php?/api/v2/get_user_by_email&email=a%40b.c", "/index.php?/api/v2/get_user_by_email?email=a%40b.c":
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("forbidden"))
		case "/index.php?/api/v2/add_user":
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte("dup"))
		case "/index.php?/api/v2/update_user/20":
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("invalid update"))
		case "/index.php?/api/v2/get_priorities":
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("prio down"))
		case "/index.php?/api/v2/get_statuses":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("status down"))
		case "/index.php?/api/v2/get_templates/30":
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("templates down"))
		default:
			t.Fatalf("unexpected URL: %s", r.URL.String())
		}
	})
	defer s.Close()

	if _, err := c.GetUsers(context.Background()); err == nil {
		t.Fatal("expected GetUsers non-OK error")
	}
	if _, err := c.GetUsersByProject(context.Background(), 10); err == nil {
		t.Fatal("expected GetUsersByProject non-OK error")
	}
	if _, err := c.GetUser(context.Background(), 20); err == nil {
		t.Fatal("expected GetUser non-OK error")
	}
	if _, err := c.GetUserByEmail(context.Background(), "a@b.c"); err == nil {
		t.Fatal("expected GetUserByEmail non-OK error")
	}
	if _, err := c.AddUser(context.Background(), data.AddUserRequest{Name: "A", Email: "a@b.c"}); err == nil {
		t.Fatal("expected AddUser non-OK error")
	}
	if _, err := c.UpdateUser(context.Background(), 20, data.UpdateUserRequest{Name: "U"}); err == nil {
		t.Fatal("expected UpdateUser non-OK error")
	}
	if _, err := c.GetPriorities(context.Background()); err == nil {
		t.Fatal("expected GetPriorities non-OK error")
	}
	if _, err := c.GetStatuses(context.Background()); err == nil {
		t.Fatal("expected GetStatuses non-OK error")
	}
	if _, err := c.GetTemplates(context.Background(), 30); err == nil {
		t.Fatal("expected GetTemplates non-OK error")
	}
}

func TestWave6D_ExtendedNonOKBranches(t *testing.T) {
	c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.String() {
		case "/index.php?/api/v2/get_group/1",
			"/index.php?/api/v2/add_group/1",
			"/index.php?/api/v2/update_group/2",
			"/index.php?/api/v2/delete_group/2",
			"/index.php?/api/v2/get_roles",
			"/index.php?/api/v2/get_role/7",
			"/index.php?/api/v2/get_result_fields",
			"/index.php?/api/v2/get_datasets/1",
			"/index.php?/api/v2/get_dataset/2",
			"/index.php?/api/v2/add_dataset/1",
			"/index.php?/api/v2/update_dataset/2",
			"/index.php?/api/v2/delete_dataset/2",
			"/index.php?/api/v2/get_variables/3",
			"/index.php?/api/v2/add_variable/3",
			"/index.php?/api/v2/update_variable/4",
			"/index.php?/api/v2/delete_variable/4",
			"/index.php?/api/v2/get_bdd/5",
			"/index.php?/api/v2/add_bdd/5",
			"/index.php?/api/v2/update_test_labels/6",
			"/index.php?/api/v2/update_tests_labels/7",
			"/index.php?/api/v2/get_labels/1",
			"/index.php?/api/v2/get_label/2",
			"/index.php?/api/v2/update_label/2":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("extended error"))
		default:
			t.Fatalf("unexpected URL: %s", r.URL.String())
		}
	})
	defer s.Close()

	ctx := context.Background()

	if _, err := c.GetGroup(ctx, 1); err == nil {
		t.Fatal("expected GetGroup non-OK error")
	}
	if _, err := c.AddGroup(ctx, 1, "g", []int64{1}); err == nil {
		t.Fatal("expected AddGroup non-OK error")
	}
	if _, err := c.UpdateGroup(ctx, 2, "g2", []int64{2}); err == nil {
		t.Fatal("expected UpdateGroup non-OK error")
	}
	if err := c.DeleteGroup(ctx, 2); err == nil {
		t.Fatal("expected DeleteGroup non-OK error")
	}
	if _, err := c.GetRoles(ctx); err == nil {
		t.Fatal("expected GetRoles non-OK error")
	}
	if _, err := c.GetRole(ctx, 7); err == nil {
		t.Fatal("expected GetRole non-OK error")
	}
	if _, err := c.GetResultFields(ctx); err == nil {
		t.Fatal("expected GetResultFields non-OK error")
	}
	if _, err := c.GetDatasets(ctx, 1); err == nil {
		t.Fatal("expected GetDatasets non-OK error")
	}
	if _, err := c.GetDataset(ctx, 2); err == nil {
		t.Fatal("expected GetDataset non-OK error")
	}
	if _, err := c.AddDataset(ctx, 1, "ds"); err == nil {
		t.Fatal("expected AddDataset non-OK error")
	}
	if _, err := c.UpdateDataset(ctx, 2, "ds2"); err == nil {
		t.Fatal("expected UpdateDataset non-OK error")
	}
	if err := c.DeleteDataset(ctx, 2); err == nil {
		t.Fatal("expected DeleteDataset non-OK error")
	}
	if _, err := c.GetVariables(ctx, 3); err == nil {
		t.Fatal("expected GetVariables non-OK error")
	}
	if _, err := c.AddVariable(ctx, 3, "var"); err == nil {
		t.Fatal("expected AddVariable non-OK error")
	}
	if _, err := c.UpdateVariable(ctx, 4, "var2"); err == nil {
		t.Fatal("expected UpdateVariable non-OK error")
	}
	if err := c.DeleteVariable(ctx, 4); err == nil {
		t.Fatal("expected DeleteVariable non-OK error")
	}
	if _, err := c.GetBDD(ctx, 5); err == nil {
		t.Fatal("expected GetBDD non-OK error")
	}
	if _, err := c.AddBDD(ctx, 5, "Given step"); err == nil {
		t.Fatal("expected AddBDD non-OK error")
	}
	if err := c.UpdateTestLabels(ctx, 6, []string{"l1"}); err == nil {
		t.Fatal("expected UpdateTestLabels non-OK error")
	}
	if err := c.UpdateTestsLabels(ctx, 7, []int64{8}, []string{"l2"}); err == nil {
		t.Fatal("expected UpdateTestsLabels non-OK error")
	}
	if _, err := c.GetLabels(ctx, 1); err == nil {
		t.Fatal("expected GetLabels non-OK error")
	}
	if _, err := c.GetLabel(ctx, 2); err == nil {
		t.Fatal("expected GetLabel non-OK error")
	}
	if _, err := c.UpdateLabel(ctx, 2, data.UpdateLabelRequest{ProjectID: 1, Title: "lbl"}); err == nil {
		t.Fatal("expected UpdateLabel non-OK error")
	}
}

func TestWave6D_ExtendedUsersAndCasesTransportBranches(t *testing.T) {
	c, err := NewClient("http://127.0.0.1:1", "test", "test", false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	// Extended transport branches
	if _, err := c.GetGroups(ctx, 1); err == nil {
		t.Fatal("expected GetGroups transport error")
	}
	if _, err := c.GetGroup(ctx, 1); err == nil {
		t.Fatal("expected GetGroup transport error")
	}
	if _, err := c.AddGroup(ctx, 1, "g", []int64{1}); err == nil {
		t.Fatal("expected AddGroup transport error")
	}
	if _, err := c.UpdateGroup(ctx, 1, "g", []int64{1}); err == nil {
		t.Fatal("expected UpdateGroup transport error")
	}
	if err := c.DeleteGroup(ctx, 1); err == nil {
		t.Fatal("expected DeleteGroup transport error")
	}
	if _, err := c.GetRoles(ctx); err == nil {
		t.Fatal("expected GetRoles transport error")
	}
	if _, err := c.GetRole(ctx, 1); err == nil {
		t.Fatal("expected GetRole transport error")
	}
	if _, err := c.GetResultFields(ctx); err == nil {
		t.Fatal("expected GetResultFields transport error")
	}
	if _, err := c.GetDatasets(ctx, 1); err == nil {
		t.Fatal("expected GetDatasets transport error")
	}
	if _, err := c.GetDataset(ctx, 1); err == nil {
		t.Fatal("expected GetDataset transport error")
	}
	if _, err := c.AddDataset(ctx, 1, "d"); err == nil {
		t.Fatal("expected AddDataset transport error")
	}
	if _, err := c.UpdateDataset(ctx, 1, "d"); err == nil {
		t.Fatal("expected UpdateDataset transport error")
	}
	if err := c.DeleteDataset(ctx, 1); err == nil {
		t.Fatal("expected DeleteDataset transport error")
	}
	if _, err := c.GetVariables(ctx, 1); err == nil {
		t.Fatal("expected GetVariables transport error")
	}
	if _, err := c.AddVariable(ctx, 1, "v"); err == nil {
		t.Fatal("expected AddVariable transport error")
	}
	if _, err := c.UpdateVariable(ctx, 1, "v"); err == nil {
		t.Fatal("expected UpdateVariable transport error")
	}
	if err := c.DeleteVariable(ctx, 1); err == nil {
		t.Fatal("expected DeleteVariable transport error")
	}
	if _, err := c.GetBDD(ctx, 1); err == nil {
		t.Fatal("expected GetBDD transport error")
	}
	if _, err := c.AddBDD(ctx, 1, "Given"); err == nil {
		t.Fatal("expected AddBDD transport error")
	}
	if err := c.UpdateTestLabels(ctx, 1, []string{"l"}); err == nil {
		t.Fatal("expected UpdateTestLabels transport error")
	}
	if err := c.UpdateTestsLabels(ctx, 1, []int64{1}, []string{"l"}); err == nil {
		t.Fatal("expected UpdateTestsLabels transport error")
	}
	if _, err := c.GetLabels(ctx, 1); err == nil {
		t.Fatal("expected GetLabels transport error")
	}
	if _, err := c.GetLabel(ctx, 1); err == nil {
		t.Fatal("expected GetLabel transport error")
	}
	if _, err := c.UpdateLabel(ctx, 1, data.UpdateLabelRequest{ProjectID: 1, Title: "lbl"}); err == nil {
		t.Fatal("expected UpdateLabel transport error")
	}

	// Users transport branches
	if _, err := c.GetUsers(ctx); err == nil {
		t.Fatal("expected GetUsers transport error")
	}
	if _, err := c.GetUsersByProject(ctx, 1); err == nil {
		t.Fatal("expected GetUsersByProject transport error")
	}
	if _, err := c.GetUser(ctx, 1); err == nil {
		t.Fatal("expected GetUser transport error")
	}
	if _, err := c.GetUserByEmail(ctx, "a@b.c"); err == nil {
		t.Fatal("expected GetUserByEmail transport error")
	}
	if _, err := c.AddUser(ctx, data.AddUserRequest{Name: "A", Email: "a@b.c"}); err == nil {
		t.Fatal("expected AddUser transport error")
	}
	if _, err := c.UpdateUser(ctx, 1, data.UpdateUserRequest{Name: "B"}); err == nil {
		t.Fatal("expected UpdateUser transport error")
	}
	if _, err := c.GetPriorities(ctx); err == nil {
		t.Fatal("expected GetPriorities transport error")
	}
	if _, err := c.GetStatuses(ctx); err == nil {
		t.Fatal("expected GetStatuses transport error")
	}
	if _, err := c.GetTemplates(ctx, 1); err == nil {
		t.Fatal("expected GetTemplates transport error")
	}

	// Cases transport branches for residual ~80% methods
	if _, err := c.GetCase(ctx, 1); err == nil {
		t.Fatal("expected GetCase transport error")
	}
	if _, err := c.GetHistoryForCase(ctx, 1); err == nil {
		t.Fatal("expected GetHistoryForCase transport error")
	}
	if _, err := c.AddCase(ctx, 1, &data.AddCaseRequest{}); err == nil {
		t.Fatal("expected AddCase transport error")
	}
	if _, err := c.UpdateCase(ctx, 1, &data.UpdateCaseRequest{}); err == nil {
		t.Fatal("expected UpdateCase transport error")
	}
	if err := c.DeleteCase(ctx, 1); err == nil {
		t.Fatal("expected DeleteCase transport error")
	}
	if _, err := c.GetCaseTypes(ctx); err == nil {
		t.Fatal("expected GetCaseTypes transport error")
	}
	if _, err := c.GetCaseFields(ctx); err == nil {
		t.Fatal("expected GetCaseFields transport error")
	}
	if _, err := c.AddCaseField(ctx, &data.AddCaseFieldRequest{}); err == nil {
		t.Fatal("expected AddCaseField transport error")
	}
}

func TestWave6E_ClientCoreEdgeBranches(t *testing.T) {
	t.Run("auth transport fills missing headers", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
		if err != nil {
			t.Fatalf("new request: %v", err)
		}

		tr := authTransport{
			username: "u",
			apiKey:   "k",
			base: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.Header.Get("Content-Type") != "application/json" {
					t.Fatalf("expected default content-type, got %q", r.Header.Get("Content-Type"))
				}
				if r.Header.Get("User-Agent") == "" {
					t.Fatal("expected default user-agent")
				}
				u, p, ok := r.BasicAuth()
				if !ok || u != "u" || p != "k" {
					t.Fatalf("unexpected basic auth: ok=%v u=%q p=%q", ok, u, p)
				}
				return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader("{}"))}, nil
			}),
		}

		_, err = tr.RoundTrip(req)
		if err != nil {
			t.Fatalf("round trip: %v", err)
		}
	})

	t.Run("auth transport keeps existing content type", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "http://example.com", nil)
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		req.Header.Set("Content-Type", "multipart/form-data; boundary=abc")
		req.Header.Set("User-Agent", "custom-agent")

		tr := authTransport{
			username: "u",
			apiKey:   "k",
			base: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.Header.Get("Content-Type") != "multipart/form-data; boundary=abc" {
					t.Fatalf("content-type must be preserved, got %q", r.Header.Get("Content-Type"))
				}
				if r.Header.Get("User-Agent") != "custom-agent" {
					t.Fatalf("user-agent must be preserved, got %q", r.Header.Get("User-Agent"))
				}
				return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader("{}"))}, nil
			}),
		}

		_, err = tr.RoundTrip(req)
		if err != nil {
			t.Fatalf("round trip: %v", err)
		}
	})

	t.Run("new client debug mode branch", func(t *testing.T) {
		c, err := NewClient("https://example.com/some/path", "u", "k", true)
		if err != nil {
			t.Fatalf("new client debug=true: %v", err)
		}
		if c == nil {
			t.Fatal("expected client instance")
		}
	})

	t.Run("do request invalid endpoint", func(t *testing.T) {
		c, err := NewClient("https://example.com", "u", "k", false)
		if err != nil {
			t.Fatalf("new client: %v", err)
		}

		_, err = c.DoRequest(context.Background(), http.MethodGet, "bad\nendpoint", nil, nil)
		if err == nil {
			t.Fatal("expected request build error for invalid endpoint")
		}
	})

	t.Run("format api error read body failure", func(t *testing.T) {
		c, err := NewClient("https://example.com", "u", "k", false)
		if err != nil {
			t.Fatalf("new client: %v", err)
		}

		err = c.formatAPIError(&http.Response{Status: "500 Internal Server Error", Body: requestReadErrorCloser{}})
		if err == nil {
			t.Fatal("expected read body error")
		}
	})
}
