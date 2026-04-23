package interactive

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupBy_Basic(t *testing.T) {
	items := []string{"apple", "avocado", "banana", "blueberry", "cherry"}
	groups := GroupBy(items, func(s string) string {
		return string(s[0]) // group by first letter
	})

	require.Len(t, groups, 3)

	// Sorted: a, b, c
	assert.Equal(t, "a", groups[0].Label)
	assert.Equal(t, []string{"apple", "avocado"}, groups[0].Items)

	assert.Equal(t, "b", groups[1].Label)
	assert.Equal(t, []string{"banana", "blueberry"}, groups[1].Items)

	assert.Equal(t, "c", groups[2].Label)
	assert.Equal(t, []string{"cherry"}, groups[2].Items)
}

func TestGroupBy_Empty(t *testing.T) {
	groups := GroupBy([]int{}, func(i int) string { return "x" })
	assert.Empty(t, groups)
}

func TestGroupBy_SingleGroup(t *testing.T) {
	groups := GroupBy([]int{1, 2, 3}, func(i int) string { return "all" })
	require.Len(t, groups, 1)
	assert.Equal(t, "all", groups[0].Label)
	assert.Equal(t, []int{1, 2, 3}, groups[0].Items)
}
