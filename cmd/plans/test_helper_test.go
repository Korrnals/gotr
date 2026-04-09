package plans

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSetupTestCmd(t *testing.T) {
	mock := &client.MockClient{}
	cmd := setupTestCmd(t, mock)
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.Context())
}

func TestGetClientForTests(t *testing.T) {
	t.Run("nil context value", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.SetContext(context.Background())
		got := getClientForTests(cmd)
		assert.Nil(t, got)
	})

	t.Run("mock client value", func(t *testing.T) {
		mock := &client.MockClient{}
		cmd := &cobra.Command{}
		cmd.SetContext(context.WithValue(context.Background(), testHTTPClientKey, mock))
		got := getClientForTests(cmd)
		assert.Equal(t, mock, got)
	})

	t.Run("unsupported type", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.SetContext(context.WithValue(context.Background(), testHTTPClientKey, 123))
		got := getClientForTests(cmd)
		assert.Nil(t, got)
	})
}
