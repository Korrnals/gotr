package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestHTTPClient_GetCasesParallel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.String(), "get_cases/30")

			suiteID, _ := strconv.ParseInt(r.URL.Query().Get("suite_id"), 10, 64)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.GetCasesResponse{{ID: suiteID*100 + 1, SuiteID: suiteID}})
		})
		defer server.Close()

		results, err := client.GetCasesParallel(context.Background(), 30, []int64{1, 2}, 2, nil)
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, int64(101), results[1][0].ID)
		assert.Equal(t, int64(201), results[2][0].ID)
	})

	t.Run("partial failure", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			suiteID := r.URL.Query().Get("suite_id")
			if suiteID == "2" {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("boom"))
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data.GetCasesResponse{{ID: 101, SuiteID: 1}})
		})
		defer server.Close()

		results, err := client.GetCasesParallel(context.Background(), 30, []int64{1, 2}, 2, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parallel execution failed")
		assert.Len(t, results, 1)
	})
}

func TestHTTPClient_GetSuitesParallel(t *testing.T) {
	client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.True(t, strings.Contains(r.URL.String(), "get_suites/"))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data.GetSuitesResponse{{ID: 1, Name: "S"}})
	})
	defer server.Close()

	results, err := client.GetSuitesParallel(context.Background(), []int64{30, 31}, 2, nil)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestHTTPClient_GetCasesForSuitesParallel(t *testing.T) {
	client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.String(), "get_cases/30")
		suiteID, _ := strconv.ParseInt(r.URL.Query().Get("suite_id"), 10, 64)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data.GetCasesResponse{{ID: suiteID*100 + 1, SuiteID: suiteID}})
	})
	defer server.Close()

	all, err := client.GetCasesForSuitesParallel(context.Background(), 30, []int64{1, 2, 3}, 2, nil)
	assert.NoError(t, err)
	assert.Len(t, all, 3)
}

func TestHTTPClient_GetCasesForSuitesParallel_ErrorBranches(t *testing.T) {
	t.Run("returns nil when all suites failed", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		})
		defer server.Close()

		all, err := client.GetCasesForSuitesParallel(context.Background(), 30, []int64{1}, 1, nil)
		assert.Error(t, err)
		assert.Nil(t, all)
	})

	t.Run("returns partial flattened cases with error", func(t *testing.T) {
		client, server := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			suiteID := r.URL.Query().Get("suite_id")
			if suiteID == "2" {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("suite failed"))
				return
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(data.GetCasesResponse{{ID: 101, SuiteID: 1}})
		})
		defer server.Close()

		all, err := client.GetCasesForSuitesParallel(context.Background(), 30, []int64{1, 2}, 2, nil)
		assert.Error(t, err)
		assert.Len(t, all, 1)
		assert.Equal(t, int64(101), all[0].ID)
	})
}
