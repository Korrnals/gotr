package progress

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		enabled bool
		quiet   bool
	}{
		{
			name:    "default manager",
			opts:    nil,
			enabled: true,
			quiet:   false,
		},
		{
			name:    "quiet mode",
			opts:    []Option{WithQuiet(true)},
			enabled: false,
			quiet:   true,
		},
		{
			name:    "explicit output",
			opts:    []Option{WithOutput(&bytes.Buffer{})},
			enabled: true,
			quiet:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.opts...)
			assert.Equal(t, tt.enabled, m.enabled)
			assert.Equal(t, tt.quiet, m.quiet)
		})
	}
}

func TestManager_NewBar(t *testing.T) {
	t.Run("enabled returns bar", func(t *testing.T) {
		m := NewManager()
		bar := m.NewBar(100, "test")
		assert.NotNil(t, bar)
	})

	t.Run("quiet mode returns nil", func(t *testing.T) {
		m := NewManager(WithQuiet(true))
		bar := m.NewBar(100, "test")
		assert.Nil(t, bar)
	})
}

func TestManager_NewSpinner(t *testing.T) {
	t.Run("enabled returns spinner", func(t *testing.T) {
		m := NewManager()
		spinner := m.NewSpinner("loading")
		assert.NotNil(t, spinner)
	})

	t.Run("quiet mode returns nil", func(t *testing.T) {
		m := NewManager(WithQuiet(true))
		spinner := m.NewSpinner("loading")
		assert.Nil(t, spinner)
	})
}

func TestAdd(t *testing.T) {
	t.Run("nil bar does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Add(nil, 1)
		})
	})
}

func TestFinish(t *testing.T) {
	t.Run("nil bar does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Finish(nil)
		})
	})
}

func TestDescribe(t *testing.T) {
	t.Run("nil bar does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Describe(nil, "new description")
		})
	})
}
