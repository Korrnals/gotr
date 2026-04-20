package crud

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCmdArgs(t *testing.T) {
	root := &cobra.Command{Use: "gotr"}
	child := &cobra.Command{Use: "cases"}
	sub := &cobra.Command{Use: "update"}
	root.AddCommand(child)
	child.AddCommand(sub)

	args := cmdArgs(sub)
	assert.Equal(t, []string{"gotr", "cases", "update"}, args)
}

func TestCmdArgs_Single(t *testing.T) {
	cmd := &cobra.Command{Use: "gotr"}
	args := cmdArgs(cmd)
	assert.Equal(t, []string{"gotr"}, args)
}

func TestExtractID_Struct(t *testing.T) {
	type resp struct {
		ID int64
	}
	assert.Equal(t, int64(42), extractID(resp{ID: 42}))
}

func TestExtractID_Pointer(t *testing.T) {
	type resp struct {
		ID int64
	}
	assert.Equal(t, int64(99), extractID(&resp{ID: 99}))
}

func TestExtractID_NilPointer(t *testing.T) {
	type resp struct {
		ID int64
	}
	var p *resp
	assert.Equal(t, int64(0), extractID(p))
}

func TestExtractID_NonStruct(t *testing.T) {
	assert.Equal(t, int64(0), extractID("hello"))
	assert.Equal(t, int64(0), extractID(42))
}

func TestExtractID_NoIDField(t *testing.T) {
	type resp struct {
		Name string
	}
	assert.Equal(t, int64(0), extractID(resp{Name: "test"}))
}

func TestExtractID_WrongIDType(t *testing.T) {
	type resp struct {
		ID string
	}
	assert.Equal(t, int64(0), extractID(resp{ID: "test"}))
}
