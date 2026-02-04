// internal/client/cases_test.go
// Тесты для Cases API POST-методов
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

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
				Title:       "Login test",
				TypeID:      1,
				PriorityID:  2,
				TemplateID:  1,
				Refs:        "REQ-123",
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
				Title:           "Test with steps",
				TypeID:          1,
				CustomSteps:     "Step 1\nStep 2",
				CustomExpected:  "Expected result",
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

			c, err := client.AddCase(tt.sectionID, tt.request)

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
				Title:       ptr("Updated title"),
				TypeID:      ptr(int64(2)),
				PriorityID:  ptr(int64(3)),
				SectionID:   ptr(int64(200)),
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

			c, err := client.UpdateCase(tt.caseID, tt.request)

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

			err := client.DeleteCase(tt.caseID)

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

			err := client.CopyCasesToSection(tt.sectionID, tt.request)

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

			err := client.MoveCasesToSection(tt.sectionID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}


