// internal/client/cases_test.go
// Тесты для Cases API POST-методов
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestAddCase(t *testing.T) {
	tests := []struct {
		name         string
		sectionID    int64
		request      *data.AddCaseRequest
		mockStatus   int
		mockResponse interface{}
		wantErr      bool
	}{
		{
			name:      "successful case creation",
			sectionID: 100,
			request: &data.AddCaseRequest{
				Title:      "Login test",
				TypeID:     1,
				PriorityID: 2,
				TemplateID: 1,
				Refs:       "REQ-123",
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Case{
				ID:         999,
				SectionID:  100,
				Title:      "Login test",
				TypeID:     1,
				PriorityID: 2,
			},
			wantErr: false,
		},
		{
			name:      "case with custom fields",
			sectionID: 100,
			request: &data.AddCaseRequest{
				Title:          "Test with steps",
				TypeID:         1,
				CustomSteps:    "Step 1\nStep 2",
				CustomExpected: "Expected result",
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Case{
				ID:             999,
				Title:          "Test with steps",
				CustomSteps:    "Step 1\nStep 2",
				CustomExpected: "Expected result",
			},
			wantErr: false,
		},
		{
			name:         "section not found",
			sectionID:    99999,
			request:      &data.AddCaseRequest{Title: "Test"},
			mockStatus:   http.StatusNotFound,
			mockResponse: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				expectedPath := fmt.Sprintf("/index.php?/api/v2/add_case/%d", tt.sectionID)
				assert.Equal(t, expectedPath, r.URL.String())

				var req data.AddCaseRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, tt.request.Title, req.Title)

				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			ctx := context.Background()
			c, err := client.AddCase(ctx, tt.sectionID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, c)
				assert.Equal(t, tt.request.Title, c.Title)
			}
		})
	}
}

func TestUpdateCase(t *testing.T) {
	tests := []struct {
		name         string
		caseID       int64
		request      *data.UpdateCaseRequest
		mockStatus   int
		mockResponse interface{}
		wantErr      bool
	}{
		{
			name:   "successful update",
			caseID: 999,
			request: &data.UpdateCaseRequest{
				Title:      ptr("Updated title"),
				TypeID:     ptr(int64(2)),
				PriorityID: ptr(int64(3)),
				SectionID:  ptr(int64(200)),
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Case{
				ID:         999,
				Title:      "Updated title",
				TypeID:     2,
				PriorityID: 3,
				SectionID:  200,
			},
			wantErr: false,
		},
		{
			name:   "update with template",
			caseID: 999,
			request: &data.UpdateCaseRequest{
				TemplateID: ptr(int64(3)),
			},
			mockStatus: http.StatusOK,
			mockResponse: data.Case{
				ID:         999,
				TemplateID: 3,
			},
			wantErr: false,
		},
		{
			name:         "case not found",
			caseID:       99999,
			request:      &data.UpdateCaseRequest{Title: ptr("Test")},
			mockStatus:   http.StatusNotFound,
			mockResponse: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				expectedPath := fmt.Sprintf("/index.php?/api/v2/update_case/%d", tt.caseID)
				assert.Equal(t, expectedPath, r.URL.String())

				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			ctx := context.Background()
			c, err := client.UpdateCase(ctx, tt.caseID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, c)
			}
		})
	}
}

func TestDeleteCase(t *testing.T) {
	tests := []struct {
		name       string
		caseID     int64
		mockStatus int
		wantErr    bool
	}{
		{
			name:       "successful delete",
			caseID:     999,
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "case not found",
			caseID:     99999,
			mockStatus: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:       "case has results",
			caseID:     999,
			mockStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				expectedPath := fmt.Sprintf("/index.php?/api/v2/delete_case/%d", tt.caseID)
				assert.Equal(t, expectedPath, r.URL.String())

				w.WriteHeader(tt.mockStatus)
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			ctx := context.Background()
			err := client.DeleteCase(ctx, tt.caseID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCopyCasesToSection(t *testing.T) {
	tests := []struct {
		name       string
		sectionID  int64
		request    *data.CopyCasesRequest
		mockStatus int
		wantErr    bool
	}{
		{
			name:      "successful copy",
			sectionID: 200,
			request: &data.CopyCasesRequest{
				CaseIDs: []int64{1, 2, 3},
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:      "copy single case",
			sectionID: 200,
			request: &data.CopyCasesRequest{
				CaseIDs: []int64{999},
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "section not found",
			sectionID:  99999,
			request:    &data.CopyCasesRequest{CaseIDs: []int64{1}},
			mockStatus: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				expectedPath := fmt.Sprintf("/index.php?/api/v2/copy_cases_to_section/%d", tt.sectionID)
				assert.Equal(t, expectedPath, r.URL.String())

				var req data.CopyCasesRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, tt.request.CaseIDs, req.CaseIDs)

				w.WriteHeader(tt.mockStatus)
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			ctx := context.Background()
			err := client.CopyCasesToSection(ctx, tt.sectionID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMoveCasesToSection(t *testing.T) {
	tests := []struct {
		name       string
		sectionID  int64
		request    *data.MoveCasesRequest
		mockStatus int
		wantErr    bool
	}{
		{
			name:      "successful move",
			sectionID: 200,
			request: &data.MoveCasesRequest{
				CaseIDs: []int64{1, 2, 3},
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:      "move with suite change",
			sectionID: 200,
			request: &data.MoveCasesRequest{
				CaseIDs: []int64{999},
				SuiteID: 100,
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "invalid case id",
			sectionID:  200,
			request:    &data.MoveCasesRequest{CaseIDs: []int64{0}},
			mockStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/index.php?/api/v2/move_cases_to_section/200", r.URL.String())

				var req data.MoveCasesRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, tt.request.CaseIDs, req.CaseIDs)

				w.WriteHeader(tt.mockStatus)
			}

			client, server := mockClient(t, handler)
			defer server.Close()

			ctx := context.Background()
			err := client.MoveCasesToSection(ctx, tt.sectionID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type progressSpy struct {
	count int
}

func (p *progressSpy) Increment() {
	p.count++
}

func (p *progressSpy) IncrementBy(n int) {
	p.count += n
}

func TestDecodeCasesResponseWithSize(t *testing.T) {
	tests := []struct {
		name      string
		payload   string
		wantLen   int
		wantSize  int64
		wantErr   bool
		errSubstr string
	}{
		{
			name:     "empty payload",
			payload:  "",
			wantLen:  0,
			wantSize: -1,
			wantErr:  false,
		},
		{
			name:     "paginated payload with size",
			payload:  `{"offset":0,"limit":250,"size":2,"cases":[{"id":1,"title":"A"},{"id":2,"title":"B"}]}`,
			wantLen:  2,
			wantSize: 2,
			wantErr:  false,
		},
		{
			name:     "paginated payload without size",
			payload:  `{"offset":0,"limit":250,"size":0,"cases":[{"id":1,"title":"A"}]}`,
			wantLen:  1,
			wantSize: -1,
			wantErr:  false,
		},
		{
			name:     "flat payload",
			payload:  `[{"id":1,"title":"A"}]`,
			wantLen:  1,
			wantSize: -1,
			wantErr:  false,
		},
		{
			name:      "unexpected format",
			payload:   "x",
			wantErr:   true,
			errSubstr: "unexpected response format",
		},
		{
			name:      "bad paginated json",
			payload:   `{"cases":`,
			wantErr:   true,
			errSubstr: "decode paginated response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cases, size, err := decodeCasesResponseWithSize([]byte(tt.payload))

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
				return
			}

			assert.NoError(t, err)
			assert.Len(t, cases, tt.wantLen)
			assert.Equal(t, tt.wantSize, size)
		})
	}
}

func TestDecodeCasesResponse(t *testing.T) {
	cases, err := decodeCasesResponse([]byte(`[{"id":10,"title":"T"}]`))
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	assert.Equal(t, int64(10), cases[0].ID)
}

func TestGetCasesWithProgressAndWrapper(t *testing.T) {
	callCount := 0
	firstPage := make(data.GetCasesResponse, 250)
	for i := range firstPage {
		firstPage[i] = data.Case{ID: int64(i + 1), Title: "case"}
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		callCount++
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/index.php", r.URL.Path)

		query := r.URL.Query()
		if callCount == 1 {
			assert.Equal(t, "0", query.Get("offset"))
			assert.Equal(t, "250", query.Get("limit"))
			assert.Equal(t, "7", query.Get("suite_id"))
			assert.Equal(t, "8", query.Get("section_id"))
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(firstPage)
			return
		}

		assert.Equal(t, "250", query.Get("offset"))
		assert.Equal(t, "250", query.Get("limit"))
		assert.Equal(t, "7", query.Get("suite_id"))
		assert.Equal(t, "8", query.Get("section_id"))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data.GetCasesResponse{{ID: 251, Title: "tail"}})
	}

	client, server := mockClient(t, handler)
	defer server.Close()

	monitor := &progressSpy{}
	all, err := client.GetCasesWithProgress(context.Background(), 77, 7, 8, monitor)
	assert.NoError(t, err)
	assert.Len(t, all, 251)
	assert.Equal(t, 2, monitor.count)
	assert.Equal(t, 2, callCount)

	wrapperCall := 0
	wrapperClient, wrapperServer := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		wrapperCall++
		assert.Equal(t, "0", r.URL.Query().Get("offset"))
		assert.Equal(t, "250", r.URL.Query().Get("limit"))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data.GetCasesResponse{{ID: 1, Title: "one"}})
	})
	defer wrapperServer.Close()

	wrapped, err := wrapperClient.GetCases(context.Background(), 77, 0, 0)
	assert.NoError(t, err)
	assert.Len(t, wrapped, 1)
	assert.Equal(t, 1, wrapperCall)
}

func TestGetCasesPage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php", r.URL.Path)
			assert.Equal(t, "4", r.URL.Query().Get("suite_id"))
			assert.Equal(t, "10", r.URL.Query().Get("offset"))
			assert.Equal(t, "2", r.URL.Query().Get("limit"))
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.GetCasesResponse{{ID: 100}, {ID: 101}})
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		page, err := client.GetCasesPage(context.Background(), 42, 4, 10, 2)
		assert.NoError(t, err)
		assert.Len(t, page, 2)
	})

	t.Run("decode error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		_, err := client.GetCasesPage(context.Background(), 42, 4, 0, 2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error cases page")
	})
}

func TestGetCaseAndHistoryForCase(t *testing.T) {
	t.Run("get case success", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/index.php?/api/v2/get_case/11", r.URL.String())
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.Case{ID: 11, Title: "single"})
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		kase, err := client.GetCase(context.Background(), 11)
		assert.NoError(t, err)
		assert.NotNil(t, kase)
		assert.Equal(t, int64(11), kase.ID)
	})

	t.Run("get case api error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("missing"))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		_, err := client.GetCase(context.Background(), 11)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API returned")
	})

	t.Run("get case decode error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		_, err := client.GetCase(context.Background(), 11)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error case")
	})

	t.Run("get history success", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/index.php?/api/v2/get_history_for_case/22", r.URL.String())
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"offset":0,"limit":250,"size":1,"history":[{"id":1,"type_id":2,"created_on":100,"user_id":10,"changes":[]}]} `))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		h, err := client.GetHistoryForCase(context.Background(), 22)
		assert.NoError(t, err)
		assert.NotNil(t, h)
		assert.Len(t, h.History, 1)
	})

	t.Run("get history api error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		_, err := client.GetHistoryForCase(context.Background(), 22)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API returned")
	})

	t.Run("get history decode error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		_, err := client.GetHistoryForCase(context.Background(), 22)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error history case")
	})
}

func TestBulkCaseAndMetaEndpoints(t *testing.T) {
	t.Run("update cases success", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/index.php?/api/v2/update_cases/90", r.URL.String())
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.GetCasesResponse{{ID: 1}, {ID: 2}})
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		resp, err := client.UpdateCases(context.Background(), 90, &data.UpdateCasesRequest{CaseIDs: []int64{1, 2}, PriorityID: 3})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, *resp, 2)
	})

	t.Run("delete cases api error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("invalid"))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		err := client.DeleteCases(context.Background(), 90, &data.DeleteCasesRequest{CaseIDs: []int64{1}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request error DeleteCases")
	})

	t.Run("delete cases success", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/index.php?/api/v2/delete_cases/90", r.URL.String())
			w.WriteHeader(http.StatusOK)
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		err := client.DeleteCases(context.Background(), 90, &data.DeleteCasesRequest{CaseIDs: []int64{1, 2}})
		assert.NoError(t, err)
	})

	t.Run("get case types success", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/index.php?/api/v2/get_case_types", r.URL.String())
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"id":1,"name":"Functional","is_default":true}]`))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		types, err := client.GetCaseTypes(context.Background())
		assert.NoError(t, err)
		assert.Len(t, types, 1)
		assert.Equal(t, "Functional", types[0].Name)
	})

	t.Run("get case types decode error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		_, err := client.GetCaseTypes(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error case types")
	})

	t.Run("get case types api error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("bad"))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		_, err := client.GetCaseTypes(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API returned")
	})

	t.Run("get case fields success", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/index.php?/api/v2/get_case_fields", r.URL.String())
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"configs":[],"description":"d","display_order":1,"id":10,"label":"L","name":"custom_x","system_name":"sys","type_id":1}]`))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		fields, err := client.GetCaseFields(context.Background())
		assert.NoError(t, err)
		assert.Len(t, fields, 1)
		assert.Equal(t, "custom_x", fields[0].Name)
	})

	t.Run("get case fields decode error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		_, err := client.GetCaseFields(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error case fields")
	})

	t.Run("get case fields api error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("missing"))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		_, err := client.GetCaseFields(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API returned")
	})

	t.Run("add case field success", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/index.php?/api/v2/add_case_field", r.URL.String())
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":5,"name":"custom_a","system_name":"system","entity_id":1,"label":"A","description":"desc","type_id":1,"location_id":1,"display_order":1,"configs":"[]","is_multi":0,"is_active":1,"status_id":1,"is_system":0,"include_all":1,"template_ids":[]}`))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		resp, err := client.AddCaseField(context.Background(), &data.AddCaseFieldRequest{Type: "string", Name: "custom_a", Label: "A"})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int64(5), resp.ID)
	})

	t.Run("add case field decode error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		_, err := client.AddCaseField(context.Background(), &data.AddCaseFieldRequest{Type: "string", Name: "custom_a", Label: "A"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error created field case")
	})

	t.Run("add case field api error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("invalid"))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		_, err := client.AddCaseField(context.Background(), &data.AddCaseFieldRequest{Type: "string", Name: "custom_a", Label: "A"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API returned")
	})
}

func TestCasesEqualByField(t *testing.T) {
	c1 := data.Case{ID: 1, Title: "A", PriorityID: 2, CustomPreconds: "p", SuiteID: 10, CreatedBy: 5, SectionID: 20}
	c2 := data.Case{ID: 1, Title: "A", PriorityID: 2, CustomPreconds: "p", SuiteID: 10, CreatedBy: 5, SectionID: 20}

	assert.True(t, casesEqualByField(c1, c2, "title"))
	assert.True(t, casesEqualByField(c1, c2, "priority_id"))
	assert.True(t, casesEqualByField(c1, c2, "custom_preconds"))
	assert.True(t, casesEqualByField(c1, c2, "id"))
	assert.True(t, casesEqualByField(c1, c2, "suite_id"))
	assert.True(t, casesEqualByField(c1, c2, "created_by"))
	assert.True(t, casesEqualByField(c1, c2, "section_id"))

	c2.Title = "B"
	assert.False(t, casesEqualByField(c1, c2, "title"))
	assert.False(t, casesEqualByField(c1, c2, "unknown"))
}

func TestDiffCasesData(t *testing.T) {
	t.Run("success diff by title", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			assert.Equal(t, "0", q.Get("offset"))
			assert.Equal(t, "250", q.Get("limit"))

			switch {
			case strings.Contains(r.URL.String(), "get_cases/1"):
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(data.GetCasesResponse{
					{ID: 1, Title: "A"},
					{ID: 2, Title: "OnlyFirst"},
				})
			case strings.Contains(r.URL.String(), "get_cases/2"):
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(data.GetCasesResponse{
					{ID: 1, Title: "B"},
					{ID: 3, Title: "OnlySecond"},
				})
			default:
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		diff, err := client.DiffCasesData(context.Background(), 1, 2, "title")
		assert.NoError(t, err)
		if err != nil {
			return
		}
		assert.NotNil(t, diff)
		assert.Len(t, diff.OnlyInFirst, 1)
		assert.Len(t, diff.OnlyInSecond, 1)
		assert.Len(t, diff.DiffByField, 1)
		assert.Equal(t, int64(1), diff.DiffByField[0].CaseID)
	})

	t.Run("upstream error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.String(), "get_cases/2") {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("boom"))
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.GetCasesResponse{{ID: 1, Title: "A"}})
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		_, err := client.DiffCasesData(context.Background(), 1, 2, "title")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get cases from project")
	})
}

func TestGetCasesParallelCtx(t *testing.T) {
	t.Run("empty suite ids", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf("unexpected request: %s", r.URL.String())
		})
		defer server.Close()

		resp, result, err := client.GetCasesParallelCtx(context.Background(), 55, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, resp)
		assert.Empty(t, result.Cases)
	})

	t.Run("single suite success", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/index.php", r.URL.Path)
			assert.Contains(t, r.URL.String(), "get_cases/55")
			assert.Equal(t, "7", r.URL.Query().Get("suite_id"))
			assert.Equal(t, "0", r.URL.Query().Get("offset"))
			assert.Equal(t, "2", r.URL.Query().Get("limit"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"offset":0,"limit":2,"size":2,"cases":[{"id":1,"title":"A"},{"id":2,"title":"B"}]}`))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		cfg := &concurrency.ControllerConfig{
			MaxConcurrentSuites:      1,
			MaxConcurrentPages:       1,
			RequestsPerMinute:        0,
			Timeout:                  0,
			PageSize:                 2,
			MaxRetries:               0,
			MaxConsecutiveErrorWaves: 1,
		}

		resp, result, err := client.GetCasesParallelCtx(context.Background(), 55, []int64{7}, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, resp, 2)
		assert.Equal(t, 1, result.Stats.CompletedSuites)
	})
}

func TestCasesFetcherFetchPageCtx(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/index.php", r.URL.Path)
			assert.Contains(t, r.URL.String(), "get_cases/9")
			assert.Equal(t, "3", r.URL.Query().Get("suite_id"))
			assert.Equal(t, "4", r.URL.Query().Get("offset"))
			assert.Equal(t, "5", r.URL.Query().Get("limit"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"offset":4,"limit":5,"size":11,"cases":[{"id":10},{"id":11}]}`))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		f := &casesFetcher{client: client}
		cases, total, err := f.FetchPageCtx(context.Background(), concurrency.PageRequest{
			SuiteTask: concurrency.SuiteTask{ProjectID: 9, SuiteID: 3},
			Offset:    4,
			Limit:     5,
		})

		assert.NoError(t, err)
		assert.Len(t, cases, 2)
		assert.Equal(t, int64(11), total)
	})

	t.Run("decode error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("x"))
		}

		client, server := mockClient(t, handler)
		defer server.Close()

		f := &casesFetcher{client: client}
		_, _, err := f.FetchPageCtx(context.Background(), concurrency.PageRequest{
			SuiteTask: concurrency.SuiteTask{ProjectID: 9, SuiteID: 3},
			Offset:    0,
			Limit:     1,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected response format")
	})
}
