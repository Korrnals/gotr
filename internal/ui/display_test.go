package ui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFmtDuration(t *testing.T) {
	assert.Equal(t, "500ms", fmtDuration(500*time.Millisecond))
	assert.Equal(t, "2s", fmtDuration(2*time.Second))
	assert.Equal(t, "1m05s", fmtDuration(65*time.Second))
}

func TestFmtCount(t *testing.T) {
	assert.Equal(t, "999", fmtCount(999))
	assert.Equal(t, "1.0K", fmtCount(1000))
	assert.Equal(t, "1.5K", fmtCount(1500))
	assert.Equal(t, "1.0M", fmtCount(1_000_000))
}
