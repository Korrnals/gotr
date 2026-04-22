package client

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// Regression tests for Stage 0.5: ensure delete-* endpoints default to soft-delete
// (soft=1) and *Permanent variants emit soft=0. See rollback 403 Forbidden bug.

func TestDeleteCase_SoftByDefault(t *testing.T) {
	var seen string
	handler := func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		w.WriteHeader(http.StatusOK)
	}
	c, srv := mockClient(t, handler)
	defer srv.Close()

	err := c.DeleteCase(context.Background(), 42)
	assert.NoError(t, err)
	assert.Equal(t, "/index.php?/api/v2/delete_case/42&soft=1", seen)
}

func TestDeleteCasePermanent_SoftZero(t *testing.T) {
	var seen string
	handler := func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		w.WriteHeader(http.StatusOK)
	}
	c, srv := mockClient(t, handler)
	defer srv.Close()

	err := c.DeleteCasePermanent(context.Background(), 42)
	assert.NoError(t, err)
	assert.Equal(t, "/index.php?/api/v2/delete_case/42&soft=0", seen)
}

func TestDeleteCases_SoftByDefault(t *testing.T) {
	var seen string
	handler := func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		w.WriteHeader(http.StatusOK)
	}
	c, srv := mockClient(t, handler)
	defer srv.Close()

	err := c.DeleteCases(context.Background(), 7, &data.DeleteCasesRequest{CaseIDs: []int64{1, 2, 3}})
	assert.NoError(t, err)
	assert.Equal(t, "/index.php?/api/v2/delete_cases/7&soft=1", seen)
}

func TestDeleteCasesPermanent_SoftZero(t *testing.T) {
	var seen string
	handler := func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		w.WriteHeader(http.StatusOK)
	}
	c, srv := mockClient(t, handler)
	defer srv.Close()

	err := c.DeleteCasesPermanent(context.Background(), 7, &data.DeleteCasesRequest{CaseIDs: []int64{1, 2, 3}})
	assert.NoError(t, err)
	assert.Equal(t, "/index.php?/api/v2/delete_cases/7&soft=0", seen)
}

func TestDeleteSection_SoftByDefault(t *testing.T) {
	var seen string
	handler := func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		w.WriteHeader(http.StatusOK)
	}
	c, srv := mockClient(t, handler)
	defer srv.Close()

	err := c.DeleteSection(context.Background(), 11)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("/index.php?/api/v2/delete_section/%d&soft=1", 11), seen)
}

func TestDeleteSectionPermanent_SoftZero(t *testing.T) {
	var seen string
	handler := func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		w.WriteHeader(http.StatusOK)
	}
	c, srv := mockClient(t, handler)
	defer srv.Close()

	err := c.DeleteSectionPermanent(context.Background(), 11)
	assert.NoError(t, err)
	assert.Equal(t, "/index.php?/api/v2/delete_section/11&soft=0", seen)
}

func TestDeleteSuite_SoftByDefault(t *testing.T) {
	var seen string
	handler := func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		w.WriteHeader(http.StatusOK)
	}
	c, srv := mockClient(t, handler)
	defer srv.Close()

	err := c.DeleteSuite(context.Background(), 22)
	assert.NoError(t, err)
	assert.Equal(t, "/index.php?/api/v2/delete_suite/22&soft=1", seen)
}

func TestDeleteSuitePermanent_SoftZero(t *testing.T) {
	var seen string
	handler := func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		w.WriteHeader(http.StatusOK)
	}
	c, srv := mockClient(t, handler)
	defer srv.Close()

	err := c.DeleteSuitePermanent(context.Background(), 22)
	assert.NoError(t, err)
	assert.Equal(t, "/index.php?/api/v2/delete_suite/22&soft=0", seen)
}
