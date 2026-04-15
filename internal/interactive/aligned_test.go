package interactive

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlignedLabels_Basic(t *testing.T) {
	cols := []Column{
		{Header: "#"},
		{Header: "NAME"},
		{Header: "STATUS"},
	}
	rows := [][]string{
		{"[1]", "Alpha", "active"},
		{"[2]", "Bravo Charlie", "closed"},
		{"[3]", "D", "active"},
	}

	header, labels := AlignedLabels(cols, rows)

	// Header uses max widths.
	assert.Contains(t, header, "#")
	assert.Contains(t, header, "NAME")
	assert.Contains(t, header, "STATUS")

	require.Len(t, labels, 3)
	// All labels should have the same length (padded).
	assert.Equal(t, len(labels[0]), len(labels[1]))
	assert.Equal(t, len(labels[1]), len(labels[2]))

	// Second row should contain full name.
	assert.Contains(t, labels[1], "Bravo Charlie")
}

func TestAlignedLabels_EmptyInput(t *testing.T) {
	header, labels := AlignedLabels(nil, nil)
	assert.Empty(t, header)
	assert.Nil(t, labels)
}

func TestAlignedLabels_MinWidth(t *testing.T) {
	cols := []Column{
		{Header: "ID", MinWidth: 10},
		{Header: "VAL"},
	}
	rows := [][]string{
		{"1", "x"},
	}

	header, labels := AlignedLabels(cols, rows)
	// ID column should be at least 10 chars wide.
	assert.True(t, len(header) >= 10)
	assert.Len(t, labels, 1)
}

func TestAlignedLabels_Separators(t *testing.T) {
	cols := []Column{
		{Header: "A"},
		{Header: "B"},
		{Header: "C"},
	}
	rows := [][]string{
		{"1", "2", "3"},
	}

	_, labels := AlignedLabels(cols, rows)
	require.Len(t, labels, 1)
	// Should contain │ separators between columns.
	assert.Contains(t, labels[0], "│")
}
