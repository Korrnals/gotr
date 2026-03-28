package client

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClientOptions_AppliedInNewClient(t *testing.T) {
	c, err := NewClient("https://example.com/path", "u", "k", false, WithSkipTlsVerify(true), WithTimeout(5*time.Second))
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, 5*time.Second, c.client.Timeout)

	auth, ok := c.client.Transport.(authTransport)
	if !assert.True(t, ok) {
		return
	}

	tr, ok := auth.base.(*http.Transport)
	if !assert.True(t, ok) {
		return
	}
	assert.True(t, tr.TLSClientConfig.InsecureSkipVerify)
}

func TestClientOptions_DefaultsAndInvalidURL(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		c, err := NewClient("https://example.com", "u", "k", false)
		assert.NoError(t, err)
		assert.NotNil(t, c)
		assert.Equal(t, 30*time.Second, c.client.Timeout)
	})

	t.Run("invalid base url", func(t *testing.T) {
		c, err := NewClient(" ", "u", "k", false)
		assert.Error(t, err)
		assert.Nil(t, c)
		assert.Contains(t, err.Error(), "invalid or empty base URL")
	})
}
