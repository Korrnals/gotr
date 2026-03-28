package testhelper

import (
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/stretchr/testify/assert"
)

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
