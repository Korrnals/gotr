package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestTestServiceGet(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		svc := NewTestService(&client.MockClient{})
		if _, err := svc.Get(context.Background(), 0); err == nil {
			t.Fatalf("expected validation error for non-positive test ID")
		}
	})

	t.Run("client error", func(t *testing.T) {
		mock := &client.MockClient{}
		mock.GetTestFunc = func(ctx context.Context, testID int64) (*data.Test, error) {
			return nil, errors.New("boom")
		}
		svc := NewTestService(mock)
		if _, err := svc.Get(context.Background(), 10); err == nil {
			t.Fatalf("expected client error")
		}
	})

	t.Run("success", func(t *testing.T) {
		mock := &client.MockClient{}
		mock.GetTestFunc = func(ctx context.Context, testID int64) (*data.Test, error) {
			return &data.Test{ID: testID, Title: "Smoke"}, nil
		}
		svc := NewTestService(mock)

		res, err := svc.Get(context.Background(), 10)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if res.ID != 10 || res.Title != "Smoke" {
			t.Fatalf("unexpected test result: %+v", res)
		}
	})
}

func TestTestServiceGetForRun(t *testing.T) {
	t.Run("invalid run id", func(t *testing.T) {
		svc := NewTestService(&client.MockClient{})
		if _, err := svc.GetForRun(context.Background(), -1, nil); err == nil {
			t.Fatalf("expected validation error for non-positive run ID")
		}
	})

	t.Run("success", func(t *testing.T) {
		mock := &client.MockClient{}
		mock.GetTestsFunc = func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return []data.Test{{ID: 1, RunID: runID, Title: "t1"}}, nil
		}

		svc := NewTestService(mock)
		res, err := svc.GetForRun(context.Background(), 20, map[string]string{"status_id": "1"})
		if err != nil {
			t.Fatalf("GetForRun() error = %v", err)
		}
		if len(res) != 1 || res[0].RunID != 20 {
			t.Fatalf("unexpected tests result: %+v", res)
		}
	})

	t.Run("client error", func(t *testing.T) {
		mock := &client.MockClient{}
		mock.GetTestsFunc = func(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
			return nil, errors.New("get tests failed")
		}

		svc := NewTestService(mock)
		res, err := svc.GetForRun(context.Background(), 20, nil)
		assert.Nil(t, res)
		assert.Error(t, err)
	})
}

func TestTestServiceUpdate(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		svc := NewTestService(&client.MockClient{})
		if _, err := svc.Update(context.Background(), 0, &data.UpdateTestRequest{StatusID: 1}); err == nil {
			t.Fatalf("expected validation error for non-positive test ID")
		}
	})

	t.Run("nil request", func(t *testing.T) {
		svc := NewTestService(&client.MockClient{})
		if _, err := svc.Update(context.Background(), 1, nil); err == nil {
			t.Fatalf("expected validation error for nil request")
		}
	})

	t.Run("negative status", func(t *testing.T) {
		svc := NewTestService(&client.MockClient{})
		if _, err := svc.Update(context.Background(), 1, &data.UpdateTestRequest{StatusID: -1}); err == nil {
			t.Fatalf("expected validation error for negative status")
		}
	})

	t.Run("success", func(t *testing.T) {
		mock := &client.MockClient{}
		mock.UpdateTestFunc = func(ctx context.Context, testID int64, req *data.UpdateTestRequest) (*data.Test, error) {
			return &data.Test{ID: testID, StatusID: req.StatusID}, nil
		}

		svc := NewTestService(mock)
		res, err := svc.Update(context.Background(), 100, &data.UpdateTestRequest{StatusID: 5})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if res.ID != 100 || res.StatusID != 5 {
			t.Fatalf("unexpected updated test: %+v", res)
		}
	})

	t.Run("client error", func(t *testing.T) {
		mock := &client.MockClient{}
		mock.UpdateTestFunc = func(ctx context.Context, testID int64, req *data.UpdateTestRequest) (*data.Test, error) {
			return nil, errors.New("update failed")
		}

		svc := NewTestService(mock)
		res, err := svc.Update(context.Background(), 100, &data.UpdateTestRequest{StatusID: 5})
		assert.Nil(t, res)
		assert.Error(t, err)
	})
}

func TestTestServiceParseID(t *testing.T) {
	svc := NewTestService(&client.MockClient{})

	t.Run("missing arg", func(t *testing.T) {
		if _, err := svc.ParseID(context.Background(), []string{}, 0); err == nil {
			t.Fatalf("expected error for missing ID")
		}
	})

	t.Run("invalid number", func(t *testing.T) {
		if _, err := svc.ParseID(context.Background(), []string{"abc"}, 0); err == nil {
			t.Fatalf("expected parse error")
		}
	})

	t.Run("non-positive number", func(t *testing.T) {
		if _, err := svc.ParseID(context.Background(), []string{"0"}, 0); err == nil {
			t.Fatalf("expected validation error for non-positive ID")
		}
	})

	t.Run("success", func(t *testing.T) {
		id, err := svc.ParseID(context.Background(), []string{"123"}, 0)
		if err != nil {
			t.Fatalf("ParseID() error = %v", err)
		}
		if id != 123 {
			t.Fatalf("ParseID() = %d, want 123", id)
		}
	})

	t.Run("index out of range", func(t *testing.T) {
		_, err := svc.ParseID(context.Background(), []string{"100"}, 2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID is required")
	})
}

func TestTestService_OutputAndPrintSuccess(t *testing.T) {
	svc := NewTestService(&client.MockClient{})

	t.Run("Output writes JSON", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test-service"}
		cmd.Flags().Bool("quiet", false, "")
		cmd.Flags().String("output", "", "")

		out := captureStdout(t, func() {
			err := svc.Output(context.Background(), cmd, map[string]any{"kind": "test"})
			assert.NoError(t, err)
		})

		assert.Contains(t, out, "\"kind\": \"test\"")
	})

	t.Run("PrintSuccess prints when not quiet", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test-service"}
		cmd.Flags().Bool("quiet", false, "")

		out := captureStdout(t, func() {
			svc.PrintSuccess(context.Background(), cmd, "test %d updated", 9)
		})

		assert.Contains(t, out, "test 9 updated")
	})

	t.Run("PrintSuccess silent in quiet mode", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test-service"}
		cmd.Flags().Bool("quiet", true, "")

		out := captureStdout(t, func() {
			svc.PrintSuccess(context.Background(), cmd, "hidden")
		})

		assert.Equal(t, "", out)
	})
}
