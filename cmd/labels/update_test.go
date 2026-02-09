package labels

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseLabels_Single tests parsing single label
func TestParseLabels_Single(t *testing.T) {
	labels := parseLabels("smoke")
	assert.Equal(t, []string{"smoke"}, labels)
}

// TestParseLabels_Multiple tests parsing multiple labels
func TestParseLabels_Multiple(t *testing.T) {
	labels := parseLabels("smoke,critical,regression")
	assert.Equal(t, []string{"smoke", "critical", "regression"}, labels)
}

// TestParseLabels_WithSpaces tests parsing labels with spaces
func TestParseLabels_WithSpaces(t *testing.T) {
	labels := parseLabels("smoke, critical, regression")
	assert.Equal(t, []string{"smoke", "critical", "regression"}, labels)
}

// TestParseLabels_Empty tests parsing empty string
func TestParseLabels_Empty(t *testing.T) {
	labels := parseLabels("")
	assert.Empty(t, labels)
}

// TestParseIntList_Single tests parsing single ID
func TestParseIntList_Single(t *testing.T) {
	ids := parseIntList("123")
	assert.Equal(t, []int64{123}, ids)
}

// TestParseIntList_Multiple tests parsing multiple IDs
func TestParseIntList_Multiple(t *testing.T) {
	ids := parseIntList("1,2,3,4,5")
	assert.Equal(t, []int64{1, 2, 3, 4, 5}, ids)
}

// TestParseIntList_WithSpaces tests parsing IDs with spaces
func TestParseIntList_WithSpaces(t *testing.T) {
	ids := parseIntList("1, 2, 3")
	assert.Equal(t, []int64{1, 2, 3}, ids)
}

// TestParseIntList_Invalid tests parsing with invalid values
func TestParseIntList_Invalid(t *testing.T) {
	ids := parseIntList("1,invalid,3")
	assert.Equal(t, []int64{1, 3}, ids)
}

// TestParseIntList_ZeroAndNegative tests filtering zero and negative
func TestParseIntList_ZeroAndNegative(t *testing.T) {
	ids := parseIntList("1,0,-1,2")
	assert.Equal(t, []int64{1, 2}, ids)
}

// TestParseIntList_Empty tests parsing empty string
func TestParseIntList_Empty(t *testing.T) {
	ids := parseIntList("")
	assert.Empty(t, ids)
}

// TestNewUpdateTestCmd_Creation tests command creation
func TestNewUpdateTestCmd_Creation(t *testing.T) {
	cmd := newUpdateTestCmd(nil)
	assert.Equal(t, "test <test_id>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	// Check required flag
	labelsFlag := cmd.Flags().Lookup("labels")
	assert.NotNil(t, labelsFlag)
}

// TestNewUpdateTestsCmd_Creation tests command creation
func TestNewUpdateTestsCmd_Creation(t *testing.T) {
	cmd := newUpdateTestsCmd(nil)
	assert.Equal(t, "tests", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	// Check required flags
	assert.NotNil(t, cmd.Flags().Lookup("run-id"))
	assert.NotNil(t, cmd.Flags().Lookup("test-ids"))
	assert.NotNil(t, cmd.Flags().Lookup("labels"))
}
