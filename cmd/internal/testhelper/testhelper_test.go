package testhelper

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSetupTestCmd(t *testing.T) {
	mock := &client.MockClient{}
	cmd := SetupTestCmd(t, mock)

	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.Context())
	assert.Equal(t, mock, GetClientForTests(cmd))
}

func TestSetupTestCmdWithBuffer(t *testing.T) {
	mock := &client.MockClient{}
	cmd, out := SetupTestCmdWithBuffer(t, mock)

	assert.NotNil(t, cmd)
	assert.NotNil(t, out)
	assert.Same(t, cmd, out)

	ctx := cmd.Context()
	assert.NotNil(t, ctx)
	stored := GetClientForTests(cmd)
	assert.Equal(t, mock, stored)
}

func TestGetClientForTests_NilCmd(t *testing.T) {
	assert.Nil(t, GetClientForTests(nil))
}

func TestGetClientForTests_EmptyContext(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	assert.Nil(t, GetClientForTests(cmd))
}

type clientHolder struct {
	client.ClientInterface
}

func TestGetClientForTests_InterfaceValue(t *testing.T) {
	cmd := &cobra.Command{}
	wrapped := clientHolder{ClientInterface: &client.MockClient{}}
	cmd.SetContext(context.WithValue(context.Background(), HTTPClientKey, wrapped))

	got := GetClientForTests(cmd)
	assert.NotNil(t, got)
	assert.IsType(t, wrapped, got)
}
