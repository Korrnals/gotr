package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
)

func closedMockClient(t *testing.T) *HTTPClient {
	t.Helper()
	c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {})
	s.Close()
	return c
}

func TestWave7_CasesMutationTargets(t *testing.T) {
	t.Run("success and decode/non-OK branches", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.String() {
			case "/index.php?/api/v2/add_case/11":
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Case{ID: 101, Title: "A"})
			case "/index.php?/api/v2/add_case/12":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{"))
			case "/index.php?/api/v2/update_case/21":
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Case{ID: 21, Title: "U"})
			case "/index.php?/api/v2/update_case/22":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{"))
			case "/index.php?/api/v2/update_cases/31":
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.GetCasesResponse{{ID: 1}})
			case "/index.php?/api/v2/delete_case/41":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{}"))
			case "/index.php?/api/v2/delete_case/42":
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("delete err"))
			case "/index.php?/api/v2/delete_cases/51":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{}"))
			case "/index.php?/api/v2/copy_cases_to_section/61":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{}"))
			case "/index.php?/api/v2/move_cases_to_section/71":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{}"))
			case "/index.php?/api/v2/add_case_field":
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.AddCaseFieldResponse{ID: 301, Name: "custom_f"})
			case "/index.php?/api/v2/add_case_field_decode":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{"))
			default:
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
		})
		defer s.Close()

		ctx := context.Background()

		if _, err := c.AddCase(ctx, 11, &data.AddCaseRequest{Title: "A"}); err != nil {
			t.Fatalf("AddCase success: %v", err)
		}
		if _, err := c.AddCase(ctx, 12, &data.AddCaseRequest{Title: "B"}); err == nil {
			t.Fatal("expected AddCase decode error")
		}

		if _, err := c.UpdateCase(ctx, 21, &data.UpdateCaseRequest{}); err != nil {
			t.Fatalf("UpdateCase success: %v", err)
		}
		if _, err := c.UpdateCase(ctx, 22, &data.UpdateCaseRequest{}); err == nil {
			t.Fatal("expected UpdateCase decode error")
		}

		if _, err := c.UpdateCases(ctx, 31, &data.UpdateCasesRequest{CaseIDs: []int64{1}}); err != nil {
			t.Fatalf("UpdateCases success: %v", err)
		}

		if err := c.DeleteCase(ctx, 41); err != nil {
			t.Fatalf("DeleteCase success: %v", err)
		}
		if err := c.DeleteCase(ctx, 42); err == nil {
			t.Fatal("expected DeleteCase non-OK error")
		}

		if err := c.DeleteCases(ctx, 51, &data.DeleteCasesRequest{CaseIDs: []int64{1}}); err != nil {
			t.Fatalf("DeleteCases success: %v", err)
		}
		if err := c.CopyCasesToSection(ctx, 61, &data.CopyCasesRequest{CaseIDs: []int64{1}}); err != nil {
			t.Fatalf("CopyCasesToSection success: %v", err)
		}
		if err := c.MoveCasesToSection(ctx, 71, &data.MoveCasesRequest{CaseIDs: []int64{1}}); err != nil {
			t.Fatalf("MoveCasesToSection success: %v", err)
		}

		if _, err := c.AddCaseField(ctx, &data.AddCaseFieldRequest{Name: "custom_f", Type: "string", Label: "F"}); err != nil {
			t.Fatalf("AddCaseField success: %v", err)
		}
	})

	t.Run("AddCaseField decode error via dedicated endpoint", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/index.php?/api/v2/add_case_field" {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		})
		defer s.Close()

		if _, err := c.AddCaseField(context.Background(), &data.AddCaseFieldRequest{Name: "custom_d", Type: "string", Label: "D"}); err == nil {
			t.Fatal("expected AddCaseField decode error")
		}
	})

	t.Run("request errors", func(t *testing.T) {
		c := closedMockClient(t)
		ctx := context.Background()

		if _, err := c.AddCase(ctx, 1, &data.AddCaseRequest{Title: "x"}); err == nil {
			t.Fatal("expected AddCase request error")
		}
		if _, err := c.UpdateCase(ctx, 1, &data.UpdateCaseRequest{}); err == nil {
			t.Fatal("expected UpdateCase request error")
		}
		if _, err := c.UpdateCases(ctx, 1, &data.UpdateCasesRequest{CaseIDs: []int64{1}}); err == nil {
			t.Fatal("expected UpdateCases request error")
		}
		if err := c.DeleteCase(ctx, 1); err == nil {
			t.Fatal("expected DeleteCase request error")
		}
		if err := c.DeleteCases(ctx, 1, &data.DeleteCasesRequest{CaseIDs: []int64{1}}); err == nil {
			t.Fatal("expected DeleteCases request error")
		}
		if err := c.CopyCasesToSection(ctx, 1, &data.CopyCasesRequest{CaseIDs: []int64{1}}); err == nil {
			t.Fatal("expected CopyCasesToSection request error")
		}
		if err := c.MoveCasesToSection(ctx, 1, &data.MoveCasesRequest{CaseIDs: []int64{1}}); err == nil {
			t.Fatal("expected MoveCasesToSection request error")
		}
		if _, err := c.AddCaseField(ctx, &data.AddCaseFieldRequest{Name: "f", Type: "string", Label: "F"}); err == nil {
			t.Fatal("expected AddCaseField request error")
		}
	})
}

func TestWave7_ConfigsAndUsersAndPlansMutationTargets(t *testing.T) {
	t.Run("configs and AddUser/AddPlanEntry success+decode+non-OK", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.String() {
			case "/index.php?/api/v2/add_config_group/1":
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.ConfigGroup{ID: 10, Name: "G"})
			case "/index.php?/api/v2/add_config_group/2":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{"))
			case "/index.php?/api/v2/add_config/10":
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Config{ID: 11, Name: "C"})
			case "/index.php?/api/v2/add_config/11":
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("bad config"))
			case "/index.php?/api/v2/update_config_group/12":
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.ConfigGroup{ID: 12, Name: "UG"})
			case "/index.php?/api/v2/update_config_group/13":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{"))
			case "/index.php?/api/v2/update_config/14":
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Config{ID: 14, Name: "UC"})
			case "/index.php?/api/v2/update_config/15":
				w.WriteHeader(http.StatusBadGateway)
				_, _ = w.Write([]byte("bad update"))
			case "/index.php?/api/v2/add_user":
				var req data.AddUserRequest
				_ = json.NewDecoder(r.Body).Decode(&req)
				if req.Name == "decode" {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("{"))
					return
				}
				if req.Name == "nonok" {
					w.WriteHeader(http.StatusConflict)
					_, _ = w.Write([]byte("dup"))
					return
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.User{ID: 1, Name: req.Name, Email: req.Email})
			case "/index.php?/api/v2/add_plan_entry/100":
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(data.Plan{ID: 100})
			case "/index.php?/api/v2/add_plan_entry/101":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{"))
			case "/index.php?/api/v2/add_plan_entry/102":
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("entry error"))
			default:
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
		})
		defer s.Close()

		ctx := context.Background()

		if _, err := c.AddConfigGroup(ctx, 1, &data.AddConfigGroupRequest{Name: "G"}); err != nil {
			t.Fatalf("AddConfigGroup success: %v", err)
		}
		if _, err := c.AddConfigGroup(ctx, 2, &data.AddConfigGroupRequest{Name: "G2"}); err == nil {
			t.Fatal("expected AddConfigGroup decode error")
		}

		if _, err := c.AddConfig(ctx, 10, &data.AddConfigRequest{Name: "C"}); err != nil {
			t.Fatalf("AddConfig success: %v", err)
		}
		if _, err := c.AddConfig(ctx, 11, &data.AddConfigRequest{Name: "C2"}); err == nil {
			t.Fatal("expected AddConfig non-OK error")
		}

		if _, err := c.UpdateConfigGroup(ctx, 12, &data.UpdateConfigGroupRequest{Name: "UG"}); err != nil {
			t.Fatalf("UpdateConfigGroup success: %v", err)
		}
		if _, err := c.UpdateConfigGroup(ctx, 13, &data.UpdateConfigGroupRequest{Name: "UG2"}); err == nil {
			t.Fatal("expected UpdateConfigGroup decode error")
		}

		if _, err := c.UpdateConfig(ctx, 14, &data.UpdateConfigRequest{Name: "UC"}); err != nil {
			t.Fatalf("UpdateConfig success: %v", err)
		}
		if _, err := c.UpdateConfig(ctx, 15, &data.UpdateConfigRequest{Name: "UC2"}); err == nil {
			t.Fatal("expected UpdateConfig non-OK error")
		}

		if _, err := c.AddUser(ctx, data.AddUserRequest{Name: "ok", Email: "ok@example.com"}); err != nil {
			t.Fatalf("AddUser success: %v", err)
		}
		if _, err := c.AddUser(ctx, data.AddUserRequest{Name: "decode", Email: "d@example.com"}); err == nil {
			t.Fatal("expected AddUser decode error")
		}
		if _, err := c.AddUser(ctx, data.AddUserRequest{Name: "nonok", Email: "n@example.com"}); err == nil {
			t.Fatal("expected AddUser non-OK error")
		}

		if _, err := c.AddPlanEntry(ctx, 100, &data.AddPlanEntryRequest{Name: "E", SuiteID: 1}); err != nil {
			t.Fatalf("AddPlanEntry success: %v", err)
		}
		if _, err := c.AddPlanEntry(ctx, 101, &data.AddPlanEntryRequest{Name: "E2", SuiteID: 1}); err == nil {
			t.Fatal("expected AddPlanEntry decode error")
		}
		if _, err := c.AddPlanEntry(ctx, 102, &data.AddPlanEntryRequest{Name: "E3", SuiteID: 1}); err == nil {
			t.Fatal("expected AddPlanEntry non-OK error")
		}
	})

	t.Run("configs and AddUser/AddPlanEntry request errors", func(t *testing.T) {
		c := closedMockClient(t)
		ctx := context.Background()

		if _, err := c.AddConfigGroup(ctx, 1, &data.AddConfigGroupRequest{Name: "G"}); err == nil {
			t.Fatal("expected AddConfigGroup request error")
		}
		if _, err := c.AddConfig(ctx, 1, &data.AddConfigRequest{Name: "C"}); err == nil {
			t.Fatal("expected AddConfig request error")
		}
		if _, err := c.UpdateConfigGroup(ctx, 1, &data.UpdateConfigGroupRequest{Name: "UG"}); err == nil {
			t.Fatal("expected UpdateConfigGroup request error")
		}
		if _, err := c.UpdateConfig(ctx, 1, &data.UpdateConfigRequest{Name: "UC"}); err == nil {
			t.Fatal("expected UpdateConfig request error")
		}
		if _, err := c.AddUser(ctx, data.AddUserRequest{Name: "A", Email: "a@b.c"}); err == nil {
			t.Fatal("expected AddUser request error")
		}
		if _, err := c.AddPlanEntry(ctx, 1, &data.AddPlanEntryRequest{Name: "E", SuiteID: 1}); err == nil {
			t.Fatal("expected AddPlanEntry request error")
		}
	})
}

func TestWave7_DeleteTargetsAcrossModules(t *testing.T) {
	t.Run("success and non-OK", func(t *testing.T) {
		c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.String() {
			case "/index.php?/api/v2/delete_attachment/1",
				"/index.php?/api/v2/delete_config_group/1",
				"/index.php?/api/v2/delete_config/1",
				"/index.php?/api/v2/delete_group/1",
				"/index.php?/api/v2/delete_dataset/1",
				"/index.php?/api/v2/delete_variable/1",
				"/index.php?/api/v2/delete_milestone/1",
				"/index.php?/api/v2/delete_plan/1",
				"/index.php?/api/v2/delete_plan_entry/1/entry-1",
				"/index.php?/api/v2/delete_project/1",
				"/index.php?/api/v2/delete_run/1",
				"/index.php?/api/v2/delete_section/1",
				"/index.php?/api/v2/delete_suite/1":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{}"))
			case "/index.php?/api/v2/delete_attachment/2",
				"/index.php?/api/v2/delete_config_group/2",
				"/index.php?/api/v2/delete_config/2",
				"/index.php?/api/v2/delete_group/2",
				"/index.php?/api/v2/delete_dataset/2",
				"/index.php?/api/v2/delete_variable/2",
				"/index.php?/api/v2/delete_milestone/2",
				"/index.php?/api/v2/delete_plan/2",
				"/index.php?/api/v2/delete_plan_entry/2/entry-2",
				"/index.php?/api/v2/delete_project/2",
				"/index.php?/api/v2/delete_run/2",
				"/index.php?/api/v2/delete_section/2",
				"/index.php?/api/v2/delete_suite/2":
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("delete failed"))
			default:
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
		})
		defer s.Close()

		ctx := context.Background()

		if err := c.DeleteAttachment(ctx, 1); err != nil {
			t.Fatalf("DeleteAttachment success: %v", err)
		}
		if err := c.DeleteAttachment(ctx, 2); err == nil {
			t.Fatal("expected DeleteAttachment non-OK error")
		}

		if err := c.DeleteConfigGroup(ctx, 1); err != nil {
			t.Fatalf("DeleteConfigGroup success: %v", err)
		}
		if err := c.DeleteConfigGroup(ctx, 2); err == nil {
			t.Fatal("expected DeleteConfigGroup non-OK error")
		}

		if err := c.DeleteConfig(ctx, 1); err != nil {
			t.Fatalf("DeleteConfig success: %v", err)
		}
		if err := c.DeleteConfig(ctx, 2); err == nil {
			t.Fatal("expected DeleteConfig non-OK error")
		}

		if err := c.DeleteGroup(ctx, 1); err != nil {
			t.Fatalf("DeleteGroup success: %v", err)
		}
		if err := c.DeleteGroup(ctx, 2); err == nil {
			t.Fatal("expected DeleteGroup non-OK error")
		}

		if err := c.DeleteDataset(ctx, 1); err != nil {
			t.Fatalf("DeleteDataset success: %v", err)
		}
		if err := c.DeleteDataset(ctx, 2); err == nil {
			t.Fatal("expected DeleteDataset non-OK error")
		}

		if err := c.DeleteVariable(ctx, 1); err != nil {
			t.Fatalf("DeleteVariable success: %v", err)
		}
		if err := c.DeleteVariable(ctx, 2); err == nil {
			t.Fatal("expected DeleteVariable non-OK error")
		}

		if err := c.DeleteMilestone(ctx, 1); err != nil {
			t.Fatalf("DeleteMilestone success: %v", err)
		}
		if err := c.DeleteMilestone(ctx, 2); err == nil {
			t.Fatal("expected DeleteMilestone non-OK error")
		}

		if err := c.DeletePlan(ctx, 1); err != nil {
			t.Fatalf("DeletePlan success: %v", err)
		}
		if err := c.DeletePlan(ctx, 2); err == nil {
			t.Fatal("expected DeletePlan non-OK error")
		}

		if err := c.DeletePlanEntry(ctx, 1, "entry-1"); err != nil {
			t.Fatalf("DeletePlanEntry success: %v", err)
		}
		if err := c.DeletePlanEntry(ctx, 2, "entry-2"); err == nil {
			t.Fatal("expected DeletePlanEntry non-OK error")
		}

		if err := c.DeleteProject(ctx, 1); err != nil {
			t.Fatalf("DeleteProject success: %v", err)
		}
		if err := c.DeleteProject(ctx, 2); err == nil {
			t.Fatal("expected DeleteProject non-OK error")
		}

		if err := c.DeleteRun(ctx, 1); err != nil {
			t.Fatalf("DeleteRun success: %v", err)
		}
		if err := c.DeleteRun(ctx, 2); err == nil {
			t.Fatal("expected DeleteRun non-OK error")
		}

		if err := c.DeleteSection(ctx, 1); err != nil {
			t.Fatalf("DeleteSection success: %v", err)
		}
		if err := c.DeleteSection(ctx, 2); err == nil {
			t.Fatal("expected DeleteSection non-OK error")
		}

		if err := c.DeleteSuite(ctx, 1); err != nil {
			t.Fatalf("DeleteSuite success: %v", err)
		}
		if err := c.DeleteSuite(ctx, 2); err == nil {
			t.Fatal("expected DeleteSuite non-OK error")
		}
	})

	t.Run("request errors", func(t *testing.T) {
		c := closedMockClient(t)
		ctx := context.Background()

		if err := c.DeleteAttachment(ctx, 1); err == nil {
			t.Fatal("expected DeleteAttachment request error")
		}
		if err := c.DeleteConfigGroup(ctx, 1); err == nil {
			t.Fatal("expected DeleteConfigGroup request error")
		}
		if err := c.DeleteConfig(ctx, 1); err == nil {
			t.Fatal("expected DeleteConfig request error")
		}
		if err := c.DeleteGroup(ctx, 1); err == nil {
			t.Fatal("expected DeleteGroup request error")
		}
		if err := c.DeleteDataset(ctx, 1); err == nil {
			t.Fatal("expected DeleteDataset request error")
		}
		if err := c.DeleteVariable(ctx, 1); err == nil {
			t.Fatal("expected DeleteVariable request error")
		}
		if err := c.DeleteMilestone(ctx, 1); err == nil {
			t.Fatal("expected DeleteMilestone request error")
		}
		if err := c.DeletePlan(ctx, 1); err == nil {
			t.Fatal("expected DeletePlan request error")
		}
		if err := c.DeletePlanEntry(ctx, 1, "entry"); err == nil {
			t.Fatal("expected DeletePlanEntry request error")
		}
		if err := c.DeleteProject(ctx, 1); err == nil {
			t.Fatal("expected DeleteProject request error")
		}
		if err := c.DeleteRun(ctx, 1); err == nil {
			t.Fatal("expected DeleteRun request error")
		}
		if err := c.DeleteSection(ctx, 1); err == nil {
			t.Fatal("expected DeleteSection request error")
		}
		if err := c.DeleteSuite(ctx, 1); err == nil {
			t.Fatal("expected DeleteSuite request error")
		}
	})
}

func TestWave8_HotspotMethods_NonOKBranches(t *testing.T) {
	c, s := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.String() {
		case "/index.php?/api/v2/add_case/901",
			"/index.php?/api/v2/update_case/902",
			"/index.php?/api/v2/update_cases/903",
			"/index.php?/api/v2/add_case_field",
			"/index.php?/api/v2/add_config_group/904",
			"/index.php?/api/v2/add_config/905",
			"/index.php?/api/v2/update_config_group/906",
			"/index.php?/api/v2/update_config/907",
			"/index.php?/api/v2/add_plan_entry/908",
			"/index.php?/api/v2/add_user":
			w.WriteHeader(http.StatusTeapot)
			_, _ = w.Write([]byte("nope"))
		default:
			t.Fatalf("unexpected URL: %s", r.URL.String())
		}
	})
	defer s.Close()

	ctx := context.Background()

	if _, err := c.AddCase(ctx, 901, &data.AddCaseRequest{Title: "A"}); err == nil {
		t.Fatal("expected AddCase non-OK error")
	}
	if _, err := c.UpdateCase(ctx, 902, &data.UpdateCaseRequest{}); err == nil {
		t.Fatal("expected UpdateCase non-OK error")
	}
	if _, err := c.UpdateCases(ctx, 903, &data.UpdateCasesRequest{CaseIDs: []int64{1}}); err == nil {
		t.Fatal("expected UpdateCases non-OK error")
	}
	if _, err := c.AddCaseField(ctx, &data.AddCaseFieldRequest{Name: "custom_nok", Type: "string", Label: "NOK"}); err == nil {
		t.Fatal("expected AddCaseField non-OK error")
	}

	if _, err := c.AddConfigGroup(ctx, 904, &data.AddConfigGroupRequest{Name: "G"}); err == nil {
		t.Fatal("expected AddConfigGroup non-OK error")
	}
	if _, err := c.AddConfig(ctx, 905, &data.AddConfigRequest{Name: "C"}); err == nil {
		t.Fatal("expected AddConfig non-OK error")
	}
	if _, err := c.UpdateConfigGroup(ctx, 906, &data.UpdateConfigGroupRequest{Name: "UG"}); err == nil {
		t.Fatal("expected UpdateConfigGroup non-OK error")
	}
	if _, err := c.UpdateConfig(ctx, 907, &data.UpdateConfigRequest{Name: "UC"}); err == nil {
		t.Fatal("expected UpdateConfig non-OK error")
	}

	if _, err := c.AddPlanEntry(ctx, 908, &data.AddPlanEntryRequest{Name: "PE", SuiteID: 1}); err == nil {
		t.Fatal("expected AddPlanEntry non-OK error")
	}
	if _, err := c.AddUser(ctx, data.AddUserRequest{Name: "U", Email: "u@example.com"}); err == nil {
		t.Fatal("expected AddUser non-OK error")
	}
}
