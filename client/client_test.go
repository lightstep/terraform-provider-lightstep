package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew_public(t *testing.T) {
	t.Parallel()

	c := NewClient("api-key", "org-name", "public")
	assert.Equal(t, "https://api.lightstep.com/public/v0.2/org-name", c.baseURL)
}

func TestNew_other(t *testing.T) {
	t.Parallel()
	c := NewClient("api-key", "org-name", "other")
	assert.Equal(t, "https://api-other.lightstep.com/public/v0.2/org-name", c.baseURL)
}
