package cases

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseIDList_Single tests parsing single ID
func TestParseIDList_Single(t *testing.T) {
	ids := parseIDList([]string{"123"})
	assert.Equal(t, []int64{123}, ids)
}

// TestParseIDList_MultipleArgs tests parsing multiple args
func TestParseIDList_MultipleArgs(t *testing.T) {
	ids := parseIDList([]string{"1", "2", "3"})
	assert.Equal(t, []int64{1, 2, 3}, ids)
}

// TestParseIDList_CommaSeparated tests parsing comma-separated IDs
func TestParseIDList_CommaSeparated(t *testing.T) {
	ids := parseIDList([]string{"1,2,3"})
	assert.Equal(t, []int64{1, 2, 3}, ids)
}

// TestParseIDList_Mixed tests parsing mixed format
func TestParseIDList_Mixed(t *testing.T) {
	ids := parseIDList([]string{"1,2", "3", "4,5,6"})
	assert.Equal(t, []int64{1, 2, 3, 4, 5, 6}, ids)
}

// TestParseIDList_WithSpaces tests parsing with spaces
func TestParseIDList_WithSpaces(t *testing.T) {
	ids := parseIDList([]string{"1, 2, 3"})
	assert.Equal(t, []int64{1, 2, 3}, ids)
}

// TestParseIDList_Invalid tests filtering invalid values
func TestParseIDList_Invalid(t *testing.T) {
	ids := parseIDList([]string{"1", "invalid", "3"})
	assert.Equal(t, []int64{1, 3}, ids)
}

// TestParseIDList_ZeroAndNegative tests filtering zero and negative
func TestParseIDList_ZeroAndNegative(t *testing.T) {
	ids := parseIDList([]string{"1", "0", "-1", "2"})
	assert.Equal(t, []int64{1, 2}, ids)
}

// TestParseIDList_Empty tests parsing empty
func TestParseIDList_Empty(t *testing.T) {
	ids := parseIDList([]string{})
	assert.Empty(t, ids)
}

// TestParseIDList_EmptyString tests parsing empty string
func TestParseIDList_EmptyString(t *testing.T) {
	ids := parseIDList([]string{""})
	assert.Empty(t, ids)
}

// TestNewUpdateCmd_Creation tests command creation
func TestNewUpdateCmd_Creation(t *testing.T) {
	cmd := newUpdateCmd(nil)
	assert.Equal(t, "update <case_ids...>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("suite-id"))
	assert.NotNil(t, cmd.Flags().Lookup("priority-id"))
	assert.NotNil(t, cmd.Flags().Lookup("estimate"))
}

// TestNewDeleteCmd_Creation tests command creation
func TestNewDeleteCmd_Creation(t *testing.T) {
	cmd := newDeleteCmd(nil)
	assert.Equal(t, "delete <case_ids...>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	assert.NotNil(t, cmd.Flags().Lookup("suite-id"))
}

// TestNewCopyCmd_Creation tests command creation
func TestNewCopyCmd_Creation(t *testing.T) {
	cmd := newCopyCmd(nil)
	assert.Equal(t, "copy <case_ids...>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	assert.NotNil(t, cmd.Flags().Lookup("section-id"))
}

// TestNewMoveCmd_Creation tests command creation
func TestNewMoveCmd_Creation(t *testing.T) {
	cmd := newMoveCmd(nil)
	assert.Equal(t, "move <case_ids...>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	assert.NotNil(t, cmd.Flags().Lookup("section-id"))
}
