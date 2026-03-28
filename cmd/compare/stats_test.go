package compare

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatCompleteness(t *testing.T) {
	ok := formatCompleteness(100, 100, 2, 2, 0)
	assert.Contains(t, ok, "100/100")

	unknownExpectedNoErrors := formatCompleteness(10, 0, 2, 2, 0)
	assert.Contains(t, unknownExpectedNoErrors, "10")
	assert.Contains(t, unknownExpectedNoErrors, "2/2 suites")

	unknownExpectedWithErrors := formatCompleteness(10, 0, 0, 0, 3)
	assert.Contains(t, unknownExpectedWithErrors, "errors: 3 pages")

	moreThanExpected := formatCompleteness(120, 100, 2, 1, 1)
	assert.Contains(t, moreThanExpected, "possible duplicates")

	partial := formatCompleteness(80, 100, 2, 1, 1)
	assert.Contains(t, partial, "80/100")
}

func TestFormatIntegrityCheck(t *testing.T) {
	assert.Equal(t, "", formatIntegrityCheck(10, 0, 10, 0))

	ok := formatIntegrityCheck(10, 2, 10, 0)
	assert.Contains(t, ok, "10 (2 suites")

	withEmpty := formatIntegrityCheck(10, 2, 10, 1)
	assert.Contains(t, withEmpty, "empty: 1")

	delta := formatIntegrityCheck(12, 2, 10, 0)
	assert.Contains(t, delta, "delta +2")
}
