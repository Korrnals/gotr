package result

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/Korrnals/gotr/cmd/internal/testhelper"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/stretchr/testify/assert"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create os.Pipe: %v", err)
	}

	os.Stdout = w
	defer func() {
		os.Stdout = originalStdout
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close writer pipe: %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read captured stdout: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("failed to close reader pipe: %v", err)
	}

	return buf.String()
}

func TestFieldsCmd_Success(t *testing.T) {
	mock := &client.MockClient{
		GetResultFieldsFunc: func(ctx context.Context) (data.GetResultFieldsResponse, error) {
			return []data.ResultField{
				{ID: 1, Name: "Status", SystemName: "status_id", IsActive: true},
				{ID: 2, Name: "Comment", SystemName: "comment", IsActive: true},
				{ID: 3, Name: "Version", SystemName: "version", IsActive: true},
			}, nil
		},
	}

	cmd := newFieldsCmd(testhelper.GetClientForTests)
	output.AddFlag(cmd)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestFieldsCmd_EmptyList(t *testing.T) {
	mock := &client.MockClient{
		GetResultFieldsFunc: func(ctx context.Context) (data.GetResultFieldsResponse, error) {
			return []data.ResultField{}, nil
		},
	}

	cmd := newFieldsCmd(testhelper.GetClientForTests)
	output.AddFlag(cmd)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestFieldsCmd_APIError(t *testing.T) {
	mock := &client.MockClient{
		GetResultFieldsFunc: func(ctx context.Context) (data.GetResultFieldsResponse, error) {
			return nil, fmt.Errorf("failed to get result fields")
		},
	}

	cmd := newFieldsCmd(testhelper.GetClientForTests)
	output.AddFlag(cmd)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}

func TestFieldsCmd_DryRun(t *testing.T) {
	mock := &client.MockClient{}

	cmd := newFieldsCmd(testhelper.GetClientForTests)
	output.AddFlag(cmd)
	cmd.SetContext(testhelper.SetupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"--dry-run"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestPrintJSON_Success(t *testing.T) {
	input := struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	}{
		Name: "ok",
		ID:   42,
	}

	var err error
	output := captureStdout(t, func() {
		err = printJSON(input)
	})

	assert.NoError(t, err)
	assert.Equal(t, "{\n  \"name\": \"ok\",\n  \"id\": 42\n}\n", output)
}

func TestPrintJSON_MarshalError(t *testing.T) {
	input := struct {
		Ch chan int `json:"ch"`
	}{
		Ch: make(chan int),
	}

	err := printJSON(input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JSON serialization error")
	assert.Contains(t, err.Error(), "unsupported type")
}
