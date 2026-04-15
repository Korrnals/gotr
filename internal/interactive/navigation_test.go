package interactive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsGoBack(t *testing.T) {
	assert.True(t, IsGoBack(ErrGoBack))
	assert.False(t, IsGoBack(ErrExit))
	assert.False(t, IsGoBack(nil))
}

func TestIsExit(t *testing.T) {
	assert.True(t, IsExit(ErrExit))
	assert.False(t, IsExit(ErrGoBack))
	assert.False(t, IsExit(nil))
}
